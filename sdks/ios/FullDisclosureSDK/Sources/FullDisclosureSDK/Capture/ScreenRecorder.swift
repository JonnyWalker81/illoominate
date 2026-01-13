import Foundation
import ReplayKit
import AVFoundation

/// Screen recorder using ReplayKit
@MainActor
public final class ScreenRecorder: NSObject {
    private let recorder = RPScreenRecorder.shared()
    private var isCurrentlyRecording = false
    private var outputURL: URL?
    private var assetWriter: AVAssetWriter?
    private var videoInput: AVAssetWriterInput?
    private var startTime: CMTime?
    private var recordingStartDate: Date?
    private var durationTimer: Timer?
    private let maxDuration: TimeInterval

    public var onRecordingStateChanged: ((Bool) -> Void)?

    /// Recording errors
    public enum RecordingError: LocalizedError {
        case notAvailable
        case notRecording
        case alreadyRecording
        case fileCreationFailed
        case writingFailed(String)

        public var errorDescription: String? {
            switch self {
            case .notAvailable:
                return "Screen recording is not available on this device"
            case .notRecording:
                return "No recording in progress"
            case .alreadyRecording:
                return "Recording is already in progress"
            case .fileCreationFailed:
                return "Failed to create recording file"
            case .writingFailed(let message):
                return "Recording failed: \(message)"
            }
        }
    }

    public init(maxDuration: TimeInterval = 60) {
        self.maxDuration = maxDuration
        super.init()
    }

    /// Check if screen recording is available
    public var isAvailable: Bool {
        recorder.isAvailable
    }

    /// Check if currently recording
    public var isRecording: Bool {
        isCurrentlyRecording
    }

    /// Current recording duration in seconds
    public var currentDuration: TimeInterval {
        guard let startDate = recordingStartDate else { return 0 }
        return Date().timeIntervalSince(startDate)
    }

    /// Start recording the screen
    public func startRecording() async throws {
        guard !isCurrentlyRecording else {
            throw RecordingError.alreadyRecording
        }
        guard recorder.isAvailable else {
            throw RecordingError.notAvailable
        }

        Logger.debug("Starting screen recording...")

        // Create output file
        let documentsPath = FileManager.default.temporaryDirectory
        let fileName = "fd_recording_\(Date().timeIntervalSince1970).mp4"
        outputURL = documentsPath.appendingPathComponent(fileName)

        guard let outputURL = outputURL else {
            throw RecordingError.fileCreationFailed
        }

        // Remove existing file if any
        try? FileManager.default.removeItem(at: outputURL)

        // Setup asset writer
        do {
            assetWriter = try AVAssetWriter(outputURL: outputURL, fileType: .mp4)
        } catch {
            throw RecordingError.writingFailed(error.localizedDescription)
        }

        // Video settings
        let screenBounds = await UIScreen.main.bounds
        let screenScale = await UIScreen.main.scale

        let videoSettings: [String: Any] = [
            AVVideoCodecKey: AVVideoCodecType.h264,
            AVVideoWidthKey: Int(screenBounds.width * screenScale),
            AVVideoHeightKey: Int(screenBounds.height * screenScale),
            AVVideoCompressionPropertiesKey: [
                AVVideoAverageBitRateKey: 6_000_000,
                AVVideoProfileLevelKey: AVVideoProfileLevelH264HighAutoLevel
            ]
        ]

        videoInput = AVAssetWriterInput(mediaType: .video, outputSettings: videoSettings)
        videoInput?.expectsMediaDataInRealTime = true

        if let videoInput = videoInput, assetWriter?.canAdd(videoInput) == true {
            assetWriter?.add(videoInput)
        }

        // Start capture
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            recorder.startCapture { [weak self] sampleBuffer, type, error in
                guard let self = self else { return }

                if let error = error {
                    Logger.error("Capture error: \(error.localizedDescription)")
                    return
                }

                Task { @MainActor in
                    self.processSampleBuffer(sampleBuffer, type: type)
                }
            } completionHandler: { error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume()
                }
            }
        }

        isCurrentlyRecording = true
        recordingStartDate = Date()
        onRecordingStateChanged?(true)

        Logger.debug("Screen recording started")

        // Auto-stop after max duration
        startDurationTimer()
    }

    /// Stop recording and return the video URL
    public func stopRecording() async throws -> URL {
        guard isCurrentlyRecording else {
            throw RecordingError.notRecording
        }

        Logger.debug("Stopping screen recording...")

        durationTimer?.invalidate()
        durationTimer = nil

        return try await withCheckedThrowingContinuation { continuation in
            recorder.stopCapture { [weak self] error in
                guard let self = self else { return }

                Task { @MainActor in
                    self.isCurrentlyRecording = false
                    self.recordingStartDate = nil
                    self.onRecordingStateChanged?(false)

                    if let error = error {
                        continuation.resume(throwing: error)
                        return
                    }

                    // Finish writing
                    self.videoInput?.markAsFinished()
                    self.assetWriter?.finishWriting {
                        if let url = self.outputURL {
                            Logger.debug("Screen recording saved to: \(url.path)")
                            continuation.resume(returning: url)
                        } else {
                            continuation.resume(throwing: RecordingError.fileCreationFailed)
                        }
                    }
                }
            }
        }
    }

    // MARK: - Private

    private func processSampleBuffer(_ sampleBuffer: CMSampleBuffer, type: RPSampleBufferType) {
        guard type == .video else { return }

        if assetWriter?.status == .unknown {
            startTime = CMSampleBufferGetPresentationTimeStamp(sampleBuffer)
            assetWriter?.startWriting()
            assetWriter?.startSession(atSourceTime: startTime!)
        }

        guard assetWriter?.status == .writing else { return }

        if let videoInput = videoInput, videoInput.isReadyForMoreMediaData {
            videoInput.append(sampleBuffer)
        }
    }

    private func startDurationTimer() {
        durationTimer = Timer.scheduledTimer(withTimeInterval: maxDuration, repeats: false) { [weak self] _ in
            Task { @MainActor in
                guard let self = self, self.isCurrentlyRecording else { return }
                Logger.debug("Max recording duration reached, stopping...")
                _ = try? await self.stopRecording()
            }
        }
    }
}

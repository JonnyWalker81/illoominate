import Foundation
import os
import OSLog

/// Captures console logs for debugging
public final class ConsoleLogCapture: @unchecked Sendable {
    public static let shared = ConsoleLogCapture()

    private var logBuffer: CircularBuffer<LogEntry>
    private let maxEntries: Int
    private var excludePatterns: [NSRegularExpression]
    private let lock = NSLock()
    private var isCapturing = false

    // For stdout/stderr redirection
    private var originalStdout: Int32 = -1
    private var originalStderr: Int32 = -1
    private var pipe: Pipe?

    private init(maxEntries: Int = 500) {
        self.maxEntries = maxEntries
        self.logBuffer = CircularBuffer(capacity: maxEntries)
        self.excludePatterns = []
    }

    /// Configure exclusion patterns
    public func configure(excludePatterns: [String]) {
        self.excludePatterns = excludePatterns.compactMap { try? NSRegularExpression(pattern: $0) }
    }

    /// Start capturing console logs
    public func startCapturing() {
        guard !isCapturing else { return }

        Logger.debug("Starting console log capture")

        // Capture OSLog entries (iOS 15+)
        Task {
            await captureOSLogs()
        }

        // Redirect stdout/stderr for print() statements
        redirectStandardOutputs()

        isCapturing = true
    }

    /// Stop capturing
    public func stopCapturing() {
        guard isCapturing else { return }

        Logger.debug("Stopping console log capture")
        restoreStandardOutputs()
        isCapturing = false
    }

    /// Get captured logs
    public func getLogs() -> [LogEntry] {
        lock.lock()
        defer { lock.unlock() }
        return Array(logBuffer)
    }

    /// Get logs as formatted strings
    public func getFormattedLogs(maxLines: Int? = nil) -> [String] {
        let logs = getLogs()
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]

        var formatted = logs.map { entry in
            "[\(formatter.string(from: entry.timestamp))] [\(entry.level.rawValue.uppercased())] \(entry.message)"
        }

        if let maxLines = maxLines, formatted.count > maxLines {
            formatted = Array(formatted.suffix(maxLines))
        }

        return formatted
    }

    /// Clear captured logs
    public func clear() {
        lock.lock()
        defer { lock.unlock() }
        logBuffer = CircularBuffer(capacity: maxEntries)
    }

    // MARK: - OSLog Capture

    private func captureOSLogs() async {
        do {
            let store = try OSLogStore(scope: .currentProcessIdentifier)
            let position = store.position(timeIntervalSinceLatestBoot: -3600) // Last hour

            let entries = try store.getEntries(at: position)

            for entry in entries {
                if let logEntry = entry as? OSLogEntryLog {
                    addLogEntry(LogEntry(
                        timestamp: logEntry.date,
                        level: LogLevel.from(osLogType: logEntry.level),
                        message: logEntry.composedMessage,
                        subsystem: logEntry.subsystem,
                        category: logEntry.category
                    ))
                }
            }
        } catch {
            Logger.warning("Failed to capture OSLog: \(error.localizedDescription)")
        }
    }

    // MARK: - Stdout/Stderr Redirect

    private func redirectStandardOutputs() {
        pipe = Pipe()

        guard let pipe = pipe else { return }

        // Save original file descriptors
        originalStdout = dup(STDOUT_FILENO)
        originalStderr = dup(STDERR_FILENO)

        // Redirect stdout to pipe
        setvbuf(stdout, nil, _IONBF, 0)
        dup2(pipe.fileHandleForWriting.fileDescriptor, STDOUT_FILENO)
        dup2(pipe.fileHandleForWriting.fileDescriptor, STDERR_FILENO)

        // Read from pipe
        pipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            guard !data.isEmpty, let string = String(data: data, encoding: .utf8) else { return }

            // Also write to original stdout for debugging
            if let self = self, self.originalStdout != -1 {
                write(self.originalStdout, (string as NSString).utf8String, string.utf8.count)
            }

            // Add to log buffer
            let lines = string.split(separator: "\n", omittingEmptySubsequences: false)
            for line in lines where !line.isEmpty {
                self?.addLogEntry(LogEntry(
                    timestamp: Date(),
                    level: .info,
                    message: String(line),
                    subsystem: "stdout",
                    category: "print"
                ))
            }
        }
    }

    private func restoreStandardOutputs() {
        if originalStdout != -1 {
            dup2(originalStdout, STDOUT_FILENO)
            close(originalStdout)
            originalStdout = -1
        }

        if originalStderr != -1 {
            dup2(originalStderr, STDERR_FILENO)
            close(originalStderr)
            originalStderr = -1
        }

        pipe?.fileHandleForReading.readabilityHandler = nil
        pipe = nil
    }

    // MARK: - Internal

    private func addLogEntry(_ entry: LogEntry) {
        // Check exclusion patterns
        for pattern in excludePatterns {
            let range = NSRange(entry.message.startIndex..<entry.message.endIndex, in: entry.message)
            if pattern.firstMatch(in: entry.message, range: range) != nil {
                return // Skip this entry
            }
        }

        lock.lock()
        defer { lock.unlock() }
        logBuffer.append(entry)
    }
}

// MARK: - Supporting Types

public struct LogEntry: Codable, Sendable {
    public let id: UUID
    public let timestamp: Date
    public let level: LogLevel
    public let message: String
    public let subsystem: String?
    public let category: String?

    public init(
        timestamp: Date,
        level: LogLevel,
        message: String,
        subsystem: String? = nil,
        category: String? = nil
    ) {
        self.id = UUID()
        self.timestamp = timestamp
        self.level = level
        self.message = message
        self.subsystem = subsystem
        self.category = category
    }
}

public enum LogLevel: String, Codable, Sendable {
    case debug
    case info
    case notice
    case warning
    case error
    case fault

    static func from(osLogType: OSLogEntryLog.Level) -> LogLevel {
        switch osLogType {
        case .debug: return .debug
        case .info: return .info
        case .notice: return .notice
        case .error: return .error
        case .fault: return .fault
        default: return .info
        }
    }
}

// MARK: - Circular Buffer

struct CircularBuffer<T>: Sequence {
    private var buffer: [T?]
    private var head = 0
    private var count = 0
    private let capacity: Int

    init(capacity: Int) {
        self.capacity = capacity
        self.buffer = Array(repeating: nil, count: capacity)
    }

    mutating func append(_ element: T) {
        buffer[head] = element
        head = (head + 1) % capacity
        count = Swift.min(count + 1, capacity)
    }

    func makeIterator() -> AnyIterator<T> {
        var index = (head - count + capacity) % capacity
        var remaining = count

        return AnyIterator {
            guard remaining > 0 else { return nil }
            let element = self.buffer[index]
            index = (index + 1) % self.capacity
            remaining -= 1
            return element
        }
    }
}

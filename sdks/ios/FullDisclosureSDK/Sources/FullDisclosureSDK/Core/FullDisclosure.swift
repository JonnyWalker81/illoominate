import UIKit
import SwiftUI

/// Main entry point for the FullDisclosure SDK
@MainActor
public final class FullDisclosure {

    // MARK: - Singleton

    /// Shared instance of the SDK
    public static let shared = FullDisclosure()

    // MARK: - State

    private var configuration: Configuration?
    private var apiClient: APIClient?
    private var uploadManager: UploadManager?
    private var screenshotCapture: ScreenshotCapture?
    private var screenRecorder: ScreenRecorder?
    private var customMetadata: [String: String] = [:]

    /// Current identified user
    public private(set) var currentUser: IdentifiedUser?

    /// Whether the SDK has been initialized
    public private(set) var isInitialized = false

    private init() {}

    // MARK: - Initialization

    /// Initialize the SDK with your project token
    /// - Parameters:
    ///   - token: Your SDK token from the FullDisclosure dashboard
    ///   - configuration: Optional configuration for customization
    public func initialize(
        token: String,
        configuration: Configuration = .default
    ) {
        guard !isInitialized else {
            Logger.warning("FullDisclosure SDK already initialized")
            return
        }

        self.configuration = configuration
        self.apiClient = APIClient(
            token: token,
            baseURL: configuration.baseURL,
            timeout: configuration.timeout
        )
        self.uploadManager = UploadManager(apiClient: apiClient!)
        self.screenshotCapture = ScreenshotCapture(configuration: configuration)
        self.screenRecorder = ScreenRecorder(maxDuration: configuration.maxRecordingDuration)

        // Enable features based on configuration
        Logger.isEnabled = configuration.debugLogging

        if configuration.enableShakeToReport {
            enableShakeToReport()
        }

        if configuration.enableConsoleLogCapture {
            enableConsoleLogCapture()
        }

        if configuration.enableFloatingButton {
            showFloatingButton()
        }

        isInitialized = true
        Logger.info("FullDisclosure SDK initialized successfully")
    }

    // MARK: - User Identification

    /// Identify the current user
    /// - Parameters:
    ///   - userId: Unique user identifier
    ///   - email: User's email address
    ///   - name: User's display name
    ///   - traits: Additional user properties
    public func identify(
        userId: String,
        email: String? = nil,
        name: String? = nil,
        traits: [String: String]? = nil
    ) async throws {
        guard let apiClient = apiClient else {
            throw NetworkError.notInitialized
        }

        let request = IdentifyRequest(
            userId: userId,
            email: email,
            name: name,
            traits: traits
        )

        try await apiClient.identify(request)

        currentUser = IdentifiedUser(
            userId: userId,
            email: email,
            name: name,
            traits: traits
        )

        Logger.info("User identified: \(userId)")
    }

    /// Clear user identification
    public func clearIdentity() {
        currentUser = nil
        Logger.info("User identity cleared")
    }

    // MARK: - Feedback Dialog

    /// Show the feedback dialog
    /// - Parameters:
    ///   - type: Pre-select a feedback type
    ///   - screenshot: Pre-attach a screenshot
    public func showFeedbackDialog(
        type: FeedbackType? = nil,
        screenshot: UIImage? = nil
    ) {
        showFeedbackDialog(type: type, screenshot: screenshot, completion: nil)
    }

    /// Show the feedback dialog with completion handler
    /// - Parameters:
    ///   - type: Pre-select a feedback type
    ///   - screenshot: Pre-attach a screenshot (or auto-capture if nil and shake triggered)
    ///   - completion: Called when feedback is submitted successfully
    public func showFeedbackDialog(
        type: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        completion: ((FeedbackResult) -> Void)?
    ) {
        guard isInitialized, let configuration = configuration, let apiClient = apiClient else {
            Logger.error("SDK not initialized. Call initialize(token:) first.")
            return
        }

        // Auto-capture screenshot if enabled and none provided
        var screenshotToUse = screenshot
        if screenshotToUse == nil && configuration.enableScreenshotOnShake {
            screenshotToUse = captureScreenshot()
        }

        UIWindow.presentFeedbackDialog(
            type: type,
            screenshot: screenshotToUse,
            configuration: configuration,
            theme: configuration.theme,
            apiClient: apiClient,
            onSubmit: completion
        )
    }

    // MARK: - Programmatic Submission

    /// Submit feedback programmatically
    /// - Parameter feedback: The feedback to submit
    /// - Returns: Result containing the feedback ID
    public func submitFeedback(_ feedback: FeedbackSubmission) async throws -> FeedbackResult {
        guard let apiClient = apiClient, let configuration = configuration else {
            throw NetworkError.notInitialized
        }

        // Collect metadata
        var consoleLogs: [String]? = nil
        if feedback.includeConsoleLogs && configuration.enableConsoleLogCapture {
            consoleLogs = ConsoleLogCapture.shared.getFormattedLogs(maxLines: configuration.maxLogLines)
        }

        var mergedMetadata = customMetadata
        for (key, value) in feedback.customMetadata {
            mergedMetadata[key] = value
        }

        let metadata = MetadataCollector.collect(
            consoleLogs: consoleLogs,
            customMetadata: mergedMetadata.isEmpty ? nil : mergedMetadata,
            userTraits: currentUser?.traits
        )

        let request = SubmitFeedbackRequest(
            title: feedback.title,
            description: feedback.description,
            type: feedback.type,
            submitterEmail: feedback.email ?? currentUser?.email,
            submitterName: feedback.name ?? currentUser?.name,
            submitterIdentifier: currentUser?.userId,
            sourceMetadata: metadata
        )

        let response = try await apiClient.submitFeedback(request)

        // Upload attachments
        if !feedback.attachments.isEmpty, let uploadManager = uploadManager {
            for attachment in feedback.attachments {
                _ = try await uploadManager.upload(
                    feedbackId: response.id,
                    attachment: attachment
                )
            }
        }

        return FeedbackResult(
            feedbackId: response.id,
            createdAt: response.createdAt
        )
    }

    // MARK: - Screenshot & Recording

    /// Capture a screenshot of the current screen
    /// - Returns: The captured screenshot, or nil if capture failed
    public func captureScreenshot() -> UIImage? {
        screenshotCapture?.capture()
    }

    /// Check if screen recording is available
    public var isRecordingAvailable: Bool {
        screenRecorder?.isAvailable ?? false
    }

    /// Check if currently recording
    public var isRecording: Bool {
        screenRecorder?.isRecording ?? false
    }

    /// Start screen recording
    public func startRecording() async throws {
        guard let recorder = screenRecorder else {
            throw NetworkError.notInitialized
        }
        try await recorder.startRecording()
    }

    /// Stop screen recording and return the video URL
    public func stopRecording() async throws -> URL {
        guard let recorder = screenRecorder else {
            throw NetworkError.notInitialized
        }
        return try await recorder.stopRecording()
    }

    // MARK: - Console Logs

    /// Enable console log capture
    public func enableConsoleLogCapture() {
        guard let configuration = configuration else { return }
        ConsoleLogCapture.shared.configure(excludePatterns: configuration.excludeConsolePatterns)
        ConsoleLogCapture.shared.startCapturing()
    }

    /// Disable console log capture
    public func disableConsoleLogCapture() {
        ConsoleLogCapture.shared.stopCapturing()
    }

    /// Get captured console logs
    public func getConsoleLogs() -> [String] {
        guard let configuration = configuration else { return [] }
        return ConsoleLogCapture.shared.getFormattedLogs(maxLines: configuration.maxLogLines)
    }

    /// Clear captured console logs
    public func clearConsoleLogs() {
        ConsoleLogCapture.shared.clear()
    }

    // MARK: - Triggers

    /// Enable shake to report
    public func enableShakeToReport() {
        ShakeDetector.shared.enable { [weak self] in
            self?.showFeedbackDialog()
        }
    }

    /// Disable shake to report
    public func disableShakeToReport() {
        ShakeDetector.shared.disable()
    }

    /// Show floating feedback button
    public func showFloatingButton() {
        guard let configuration = configuration else { return }
        FloatingButton.shared.show(theme: configuration.theme) { [weak self] in
            self?.showFeedbackDialog()
        }
    }

    /// Hide floating feedback button
    public func hideFloatingButton() {
        FloatingButton.shared.hide()
    }

    // MARK: - Configuration

    /// Update the theme
    public func updateTheme(_ theme: Theme) {
        configuration?.theme = theme
    }

    /// Set custom metadata to include with all submissions
    public func setCustomMetadata(_ metadata: [String: String]) {
        self.customMetadata = metadata
    }

    /// Add a custom metadata key-value pair
    public func addCustomMetadata(key: String, value: String) {
        customMetadata[key] = value
    }

    /// Remove a custom metadata key
    public func removeCustomMetadata(key: String) {
        customMetadata.removeValue(forKey: key)
    }
}

import Foundation

/// Configuration options for the FullDisclosure SDK
public struct Configuration: Sendable {

    // MARK: - API Settings

    /// Base URL for the API (default: https://api.fulldisclosure.io)
    public var baseURL: URL

    /// Request timeout in seconds
    public var timeout: TimeInterval

    // MARK: - Feature Toggles

    /// Enable shake-to-report trigger
    public var enableShakeToReport: Bool

    /// Automatically capture screenshot when shake is detected
    public var enableScreenshotOnShake: Bool

    /// Enable console log capture
    public var enableConsoleLogCapture: Bool

    /// Enable screen recording feature
    public var enableScreenRecording: Bool

    /// Show floating feedback button
    public var enableFloatingButton: Bool

    // MARK: - Privacy Controls

    /// View tags to redact in screenshots (blur sensitive content)
    public var redactedViewTags: Set<Int>

    /// Regex patterns to exclude from console logs
    public var excludeConsolePatterns: [String]

    /// Maximum number of log lines to capture
    public var maxLogLines: Int

    // MARK: - UI Customization

    /// Theme configuration
    public var theme: Theme

    /// Which feedback types to show in the dialog
    public var feedbackTypes: [FeedbackType]

    /// Require email for feedback submission
    public var requireEmail: Bool

    /// Show contact fields (email/name) in feedback form
    /// Set to false if user is identified via identify() method
    public var showContactFields: Bool

    // MARK: - Attachment Limits

    /// Maximum attachment file size in bytes (default: 25MB)
    public var maxAttachmentSize: Int64

    /// Maximum number of attachments per submission
    public var maxAttachmentCount: Int

    /// JPEG compression quality for screenshots (0.0 - 1.0)
    public var imageCompressionQuality: CGFloat

    /// Maximum screen recording duration in seconds
    public var maxRecordingDuration: TimeInterval

    // MARK: - Logging

    /// Enable SDK debug logging
    public var debugLogging: Bool

    // MARK: - Initialization

    public init(
        baseURL: URL = URL(string: "https://api.fulldisclosure.io")!,
        timeout: TimeInterval = 30,
        enableShakeToReport: Bool = true,
        enableScreenshotOnShake: Bool = true,
        enableConsoleLogCapture: Bool = true,
        enableScreenRecording: Bool = true,
        enableFloatingButton: Bool = false,
        redactedViewTags: Set<Int> = [],
        excludeConsolePatterns: [String] = [],
        maxLogLines: Int = 500,
        theme: Theme = .default,
        feedbackTypes: [FeedbackType] = [.bug, .feature, .general],
        requireEmail: Bool = false,
        showContactFields: Bool = true,
        maxAttachmentSize: Int64 = 25 * 1024 * 1024,
        maxAttachmentCount: Int = 5,
        imageCompressionQuality: CGFloat = 0.8,
        maxRecordingDuration: TimeInterval = 60,
        debugLogging: Bool = false
    ) {
        self.baseURL = baseURL
        self.timeout = timeout
        self.enableShakeToReport = enableShakeToReport
        self.enableScreenshotOnShake = enableScreenshotOnShake
        self.enableConsoleLogCapture = enableConsoleLogCapture
        self.enableScreenRecording = enableScreenRecording
        self.enableFloatingButton = enableFloatingButton
        self.redactedViewTags = redactedViewTags
        self.excludeConsolePatterns = excludeConsolePatterns
        self.maxLogLines = maxLogLines
        self.theme = theme
        self.feedbackTypes = feedbackTypes
        self.requireEmail = requireEmail
        self.showContactFields = showContactFields
        self.maxAttachmentSize = maxAttachmentSize
        self.maxAttachmentCount = maxAttachmentCount
        self.imageCompressionQuality = imageCompressionQuality
        self.maxRecordingDuration = maxRecordingDuration
        self.debugLogging = debugLogging
    }

    // MARK: - Default Configuration

    public static let `default` = Configuration()

    // MARK: - Builder Methods

    public func with(baseURL: URL) -> Configuration {
        var copy = self
        copy.baseURL = baseURL
        return copy
    }

    public func with(theme: Theme) -> Configuration {
        var copy = self
        copy.theme = theme
        return copy
    }

    public func with(enableShakeToReport: Bool) -> Configuration {
        var copy = self
        copy.enableShakeToReport = enableShakeToReport
        return copy
    }

    public func with(enableScreenRecording: Bool) -> Configuration {
        var copy = self
        copy.enableScreenRecording = enableScreenRecording
        return copy
    }

    public func with(requireEmail: Bool) -> Configuration {
        var copy = self
        copy.requireEmail = requireEmail
        return copy
    }

    public func with(feedbackTypes: [FeedbackType]) -> Configuration {
        var copy = self
        copy.feedbackTypes = feedbackTypes
        return copy
    }

    public func with(redactedViewTags: Set<Int>) -> Configuration {
        var copy = self
        copy.redactedViewTags = redactedViewTags
        return copy
    }

    public func with(showContactFields: Bool) -> Configuration {
        var copy = self
        copy.showContactFields = showContactFields
        return copy
    }

    public func with(debugLogging: Bool) -> Configuration {
        var copy = self
        copy.debugLogging = debugLogging
        Logger.isEnabled = debugLogging
        return copy
    }
}

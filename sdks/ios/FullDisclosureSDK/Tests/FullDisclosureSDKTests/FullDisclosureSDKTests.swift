import XCTest
@testable import FullDisclosureSDK

final class FullDisclosureSDKTests: XCTestCase {

    func testFeedbackTypeDisplayName() {
        XCTAssertEqual(FeedbackType.bug.displayName, "Bug Report")
        XCTAssertEqual(FeedbackType.feature.displayName, "Feature Request")
        XCTAssertEqual(FeedbackType.general.displayName, "General Feedback")
    }

    func testFeedbackTypeIconName() {
        XCTAssertEqual(FeedbackType.bug.iconName, "ladybug")
        XCTAssertEqual(FeedbackType.feature.iconName, "lightbulb")
        XCTAssertEqual(FeedbackType.general.iconName, "bubble.left.and.bubble.right")
    }

    func testAttachmentDataFromText() {
        let attachment = AttachmentData(logText: "Test log content")
        XCTAssertEqual(attachment.filename, "console.log")
        XCTAssertEqual(attachment.contentType, "text/plain")
        XCTAssertEqual(attachment.sizeBytes, Int64("Test log content".utf8.count))
    }

    func testAllowedContentTypes() {
        XCTAssertTrue(AllowedContentType.isAllowed("image/jpeg"))
        XCTAssertTrue(AllowedContentType.isAllowed("image/png"))
        XCTAssertTrue(AllowedContentType.isAllowed("video/mp4"))
        XCTAssertTrue(AllowedContentType.isAllowed("text/plain"))
        XCTAssertFalse(AllowedContentType.isAllowed("application/javascript"))
    }

    func testConfigurationDefaults() {
        let config = Configuration.default

        XCTAssertEqual(config.timeout, 30)
        XCTAssertTrue(config.enableShakeToReport)
        XCTAssertTrue(config.enableScreenshotOnShake)
        XCTAssertTrue(config.enableConsoleLogCapture)
        XCTAssertTrue(config.enableScreenRecording)
        XCTAssertFalse(config.enableFloatingButton)
        XCTAssertEqual(config.maxLogLines, 500)
        XCTAssertFalse(config.requireEmail)
        XCTAssertEqual(config.maxAttachmentSize, 25 * 1024 * 1024)
        XCTAssertEqual(config.maxAttachmentCount, 5)
        XCTAssertEqual(config.imageCompressionQuality, 0.8)
    }

    func testConfigurationBuilder() {
        let config = Configuration.default
            .with(requireEmail: true)
            .with(enableShakeToReport: false)
            .with(feedbackTypes: [.bug, .feature])

        XCTAssertTrue(config.requireEmail)
        XCTAssertFalse(config.enableShakeToReport)
        XCTAssertEqual(config.feedbackTypes, [.bug, .feature])
    }

    func testNetworkErrorDescriptions() {
        XCTAssertNotNil(NetworkError.notInitialized.errorDescription)
        XCTAssertNotNil(NetworkError.unauthorized.errorDescription)
        XCTAssertNotNil(NetworkError.rateLimited.errorDescription)

        let validationError = NetworkError.validationError(["title": "Required"])
        XCTAssertTrue(validationError.errorDescription?.contains("title") == true)
    }

    func testNetworkErrorRetryable() {
        XCTAssertTrue(NetworkError.rateLimited.isRetryable)
        XCTAssertTrue(NetworkError.serverError(500).isRetryable)
        XCTAssertTrue(NetworkError.networkUnavailable.isRetryable)
        XCTAssertTrue(NetworkError.timeout.isRetryable)

        XCTAssertFalse(NetworkError.unauthorized.isRetryable)
        XCTAssertFalse(NetworkError.validationError([:]).isRetryable)
    }

    func testFeedbackSubmissionDefaults() {
        let submission = FeedbackSubmission(
            title: "Test",
            description: "Test description"
        )

        XCTAssertEqual(submission.type, .general)
        XCTAssertNil(submission.email)
        XCTAssertNil(submission.name)
        XCTAssertTrue(submission.attachments.isEmpty)
        XCTAssertFalse(submission.includeConsoleLogs)
    }

    func testLogLevel() {
        XCTAssertEqual(LogLevel.debug.rawValue, "debug")
        XCTAssertEqual(LogLevel.error.rawValue, "error")
    }
}

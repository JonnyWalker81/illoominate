import SwiftUI
import Combine

/// ViewModel for the feedback form
@MainActor
final class FeedbackViewModel: ObservableObject {
    // MARK: - Form State

    @Published var selectedType: FeedbackType
    @Published var title: String = ""
    @Published var description: String = ""
    @Published var email: String = ""
    @Published var name: String = ""
    @Published var attachments: [AttachmentItem] = []
    @Published var includeConsoleLogs: Bool = true

    // MARK: - UI State

    @Published var isSubmitting = false
    @Published var uploadProgress: Double = 0
    @Published var currentUploadIndex: Int = 0
    @Published var showError = false
    @Published var errorMessage: String?
    @Published var submissionResult: FeedbackResult?

    // MARK: - Configuration

    let availableTypes: [FeedbackType]
    let requireEmail: Bool
    let showContactFields: Bool
    let hasConsoleLogs: Bool
    let maxAttachments: Int

    private let apiClient: APIClient?
    private let uploadManager: UploadManager?
    private let configuration: Configuration
    private var customMetadata: [String: String] = [:]

    // MARK: - Initialization

    init(
        initialType: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        configuration: Configuration = .default,
        apiClient: APIClient? = nil
    ) {
        self.configuration = configuration
        self.apiClient = apiClient
        self.uploadManager = apiClient.map { UploadManager(apiClient: $0) }
        self.availableTypes = configuration.feedbackTypes
        self.selectedType = initialType ?? configuration.feedbackTypes.first ?? .general
        self.requireEmail = configuration.requireEmail
        self.showContactFields = configuration.showContactFields
        self.hasConsoleLogs = configuration.enableConsoleLogCapture
        self.maxAttachments = configuration.maxAttachmentCount

        // Add initial screenshot if provided
        if let screenshot = screenshot {
            attachments.append(AttachmentItem(image: screenshot))
        }
    }

    // MARK: - Validation

    var canSubmit: Bool {
        !title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        !description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        (!requireEmail || isValidEmail(email)) &&
        !isSubmitting
    }

    var titleError: String? {
        let trimmed = title.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty && !title.isEmpty {
            return "Title is required"
        }
        if trimmed.count > 200 {
            return "Title must be 200 characters or less"
        }
        return nil
    }

    var emailError: String? {
        if requireEmail && !email.isEmpty && !isValidEmail(email) {
            return "Please enter a valid email"
        }
        return nil
    }

    private func isValidEmail(_ email: String) -> Bool {
        let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}"
        let predicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
        return predicate.evaluate(with: email)
    }

    // MARK: - Actions

    func captureScreenshot() async {
        guard attachments.count < maxAttachments else { return }

        let capture = ScreenshotCapture(configuration: configuration)
        if let image = capture.capture() {
            attachments.append(AttachmentItem(image: image))
        }
    }

    func startRecording() async {
        guard attachments.count < maxAttachments else { return }

        let recorder = ScreenRecorder(maxDuration: configuration.maxRecordingDuration)

        do {
            try await recorder.startRecording()
            // Recording will be added when stopped
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    func removeAttachment(_ item: AttachmentItem) {
        attachments.removeAll { $0.id == item.id }
    }

    func setCustomMetadata(_ metadata: [String: String]) {
        self.customMetadata = metadata
    }

    // MARK: - Submission

    func submit() async {
        guard canSubmit, let apiClient = apiClient else {
            if apiClient == nil {
                errorMessage = "SDK not initialized"
                showError = true
            }
            return
        }

        isSubmitting = true
        uploadProgress = 0
        errorMessage = nil

        do {
            // Collect metadata
            let consoleLogs: [String]? = includeConsoleLogs && hasConsoleLogs
                ? ConsoleLogCapture.shared.getFormattedLogs(maxLines: configuration.maxLogLines)
                : nil

            let metadata = MetadataCollector.collect(
                consoleLogs: consoleLogs,
                customMetadata: customMetadata.isEmpty ? nil : customMetadata
            )

            // Determine email and name - use form values if provided, otherwise fall back to identified user
            let currentUser = FullDisclosure.shared.currentUser
            let submitterEmail = !email.isEmpty ? email : currentUser?.email
            let submitterName = !name.isEmpty ? name : currentUser?.name

            // Create feedback request
            let request = SubmitFeedbackRequest(
                title: title.trimmingCharacters(in: .whitespacesAndNewlines),
                description: description.trimmingCharacters(in: .whitespacesAndNewlines),
                type: selectedType,
                submitterEmail: submitterEmail,
                submitterName: submitterName,
                submitterIdentifier: currentUser?.userId,
                sourceMetadata: metadata
            )

            // Submit feedback
            let response = try await apiClient.submitFeedback(request)

            // Upload attachments
            if !attachments.isEmpty, let uploadManager = uploadManager {
                for (index, item) in attachments.enumerated() {
                    currentUploadIndex = index

                    _ = try await uploadManager.upload(
                        feedbackId: response.id,
                        attachment: item.data
                    ) { [weak self] progress in
                        Task { @MainActor in
                            let overallProgress = (Double(index) + progress) / Double(self?.attachments.count ?? 1)
                            self?.uploadProgress = overallProgress
                        }
                    }
                }
            }

            // Success!
            submissionResult = FeedbackResult(
                feedbackId: response.id,
                createdAt: response.createdAt
            )

            Logger.info("Feedback submitted successfully: \(response.id)")

        } catch {
            Logger.error("Feedback submission failed: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
            showError = true
        }

        isSubmitting = false
    }
}

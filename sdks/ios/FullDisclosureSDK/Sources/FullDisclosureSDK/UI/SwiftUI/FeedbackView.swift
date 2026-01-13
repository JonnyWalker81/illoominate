import SwiftUI

/// Main feedback form view
public struct FeedbackView: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: FeedbackViewModel

    private let theme: Theme
    private let onSubmit: ((FeedbackResult) -> Void)?
    private let onCancel: (() -> Void)?

    public init(
        initialType: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        configuration: Configuration = .default,
        apiClient: APIClient? = nil,
        theme: Theme = .default,
        onSubmit: ((FeedbackResult) -> Void)? = nil,
        onCancel: (() -> Void)? = nil
    ) {
        self._viewModel = StateObject(wrappedValue: FeedbackViewModel(
            initialType: initialType,
            screenshot: screenshot,
            configuration: configuration,
            apiClient: apiClient
        ))
        self.theme = theme
        self.onSubmit = onSubmit
        self.onCancel = onCancel
    }

    public var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Type Selector
                    FeedbackTypeSelector(
                        selectedType: $viewModel.selectedType,
                        availableTypes: viewModel.availableTypes,
                        theme: theme
                    )

                    // Title Field
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Text("Title")
                                .font(.subheadline)
                                .fontWeight(.medium)
                                .foregroundColor(theme.textColor)

                            Text("*")
                                .foregroundColor(theme.errorColor)
                        }

                        TextField("Brief summary of your feedback", text: $viewModel.title)
                            .textFieldStyle(FDTextFieldStyle(theme: theme))

                        if let error = viewModel.titleError {
                            Text(error)
                                .font(.caption)
                                .foregroundColor(theme.errorColor)
                        }
                    }

                    // Description Field
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Text("Description")
                                .font(.subheadline)
                                .fontWeight(.medium)
                                .foregroundColor(theme.textColor)

                            Text("*")
                                .foregroundColor(theme.errorColor)
                        }

                        TextEditor(text: $viewModel.description)
                            .frame(minHeight: 120)
                            .padding(8)
                            .background(theme.surfaceColor)
                            .clipShape(RoundedRectangle(cornerRadius: theme.cornerRadius))
                            .overlay(
                                RoundedRectangle(cornerRadius: theme.cornerRadius)
                                    .stroke(theme.borderColor, lineWidth: 1)
                            )
                    }

                    // Attachments Section
                    AttachmentsSection(
                        attachments: $viewModel.attachments,
                        theme: theme,
                        maxCount: viewModel.maxAttachments,
                        onAddScreenshot: { await viewModel.captureScreenshot() },
                        onAddRecording: { await viewModel.startRecording() },
                        onRemove: viewModel.removeAttachment
                    )

                    // Contact Info Section
                    if viewModel.showContactFields {
                        VStack(alignment: .leading, spacing: 16) {
                            Text("Contact (optional)")
                                .font(.subheadline)
                                .fontWeight(.medium)
                                .foregroundColor(theme.textColor)

                            VStack(spacing: 12) {
                                HStack {
                                    Image(systemName: "envelope")
                                        .foregroundColor(theme.secondaryTextColor)
                                        .frame(width: 24)

                                    TextField("Email", text: $viewModel.email)
                                        .textContentType(.emailAddress)
                                        .keyboardType(.emailAddress)
                                        .autocapitalization(.none)
                                }
                                .padding()
                                .background(theme.surfaceColor)
                                .clipShape(RoundedRectangle(cornerRadius: theme.cornerRadius))
                                .overlay(
                                    RoundedRectangle(cornerRadius: theme.cornerRadius)
                                        .stroke(
                                            viewModel.emailError != nil ? theme.errorColor : theme.borderColor,
                                            lineWidth: 1
                                        )
                                )

                                if let error = viewModel.emailError {
                                    Text(error)
                                        .font(.caption)
                                        .foregroundColor(theme.errorColor)
                                }

                                HStack {
                                    Image(systemName: "person")
                                        .foregroundColor(theme.secondaryTextColor)
                                        .frame(width: 24)

                                    TextField("Name", text: $viewModel.name)
                                        .textContentType(.name)
                                }
                                .padding()
                                .background(theme.surfaceColor)
                                .clipShape(RoundedRectangle(cornerRadius: theme.cornerRadius))
                                .overlay(
                                    RoundedRectangle(cornerRadius: theme.cornerRadius)
                                        .stroke(theme.borderColor, lineWidth: 1)
                                )
                            }
                        }
                    }

                    // Console Logs Toggle
                    if viewModel.hasConsoleLogs {
                        Toggle(isOn: $viewModel.includeConsoleLogs) {
                            VStack(alignment: .leading, spacing: 4) {
                                Text("Include console logs")
                                    .font(.subheadline)
                                    .foregroundColor(theme.textColor)

                                Text("Helps debug issues")
                                    .font(.caption)
                                    .foregroundColor(theme.secondaryTextColor)
                            }
                        }
                        .tint(theme.primaryColor)
                        .padding()
                        .background(theme.surfaceColor)
                        .clipShape(RoundedRectangle(cornerRadius: theme.cornerRadius))
                    }

                    // Powered By
                    if theme.showPoweredBy {
                        HStack {
                            Spacer()
                            Text("Powered by FullDisclosure")
                                .font(.caption2)
                                .foregroundColor(theme.secondaryTextColor)
                            Spacer()
                        }
                        .padding(.top, 8)
                    }
                }
                .padding()
            }
            .navigationTitle("Send Feedback")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        onCancel?()
                        dismiss()
                    }
                    .foregroundColor(theme.secondaryTextColor)
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Submit") {
                        Task { await viewModel.submit() }
                    }
                    .fontWeight(.semibold)
                    .foregroundColor(viewModel.canSubmit ? theme.primaryColor : theme.secondaryTextColor)
                    .disabled(!viewModel.canSubmit)
                }
            }
            .overlay {
                if viewModel.isSubmitting {
                    SubmittingOverlay(
                        progress: viewModel.uploadProgress,
                        hasAttachments: !viewModel.attachments.isEmpty,
                        theme: theme
                    )
                }
            }
            .alert("Error", isPresented: $viewModel.showError) {
                Button("OK") {}
            } message: {
                Text(viewModel.errorMessage ?? "An error occurred")
            }
            .onChange(of: viewModel.submissionResult) { _, result in
                if let result = result {
                    onSubmit?(result)
                    dismiss()
                }
            }
        }
        .interactiveDismissDisabled(viewModel.isSubmitting)
    }
}

// MARK: - Text Field Style

struct FDTextFieldStyle: TextFieldStyle {
    let theme: Theme

    func _body(configuration: TextField<Self._Label>) -> some View {
        configuration
            .padding()
            .background(theme.surfaceColor)
            .clipShape(RoundedRectangle(cornerRadius: theme.cornerRadius))
            .overlay(
                RoundedRectangle(cornerRadius: theme.cornerRadius)
                    .stroke(theme.borderColor, lineWidth: 1)
            )
    }
}

// MARK: - Submitting Overlay

struct SubmittingOverlay: View {
    let progress: Double
    let hasAttachments: Bool
    let theme: Theme

    var body: some View {
        ZStack {
            Color.black.opacity(0.4)
                .ignoresSafeArea()

            VStack(spacing: 16) {
                ProgressView()
                    .scaleEffect(1.5)
                    .tint(theme.primaryColor)

                Text(hasAttachments ? "Uploading..." : "Submitting...")
                    .font(.headline)
                    .foregroundColor(.white)

                if hasAttachments && progress > 0 {
                    ProgressView(value: progress)
                        .progressViewStyle(.linear)
                        .tint(theme.primaryColor)
                        .frame(width: 200)

                    Text("\(Int(progress * 100))%")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.8))
                }
            }
            .padding(32)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(Color(.systemGray6).opacity(0.95))
            )
        }
    }
}

#Preview {
    FeedbackView(
        theme: .default
    )
}

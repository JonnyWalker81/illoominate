import Foundation

/// Manages attachment uploads with retry logic
actor UploadManager {
    private let apiClient: APIClient
    private let maxRetries: Int
    private let retryDelay: TimeInterval

    init(apiClient: APIClient, maxRetries: Int = 3, retryDelay: TimeInterval = 1.0) {
        self.apiClient = apiClient
        self.maxRetries = maxRetries
        self.retryDelay = retryDelay
    }

    /// Upload an attachment with automatic retry
    func upload(
        feedbackId: UUID,
        attachment: AttachmentData,
        progress: (@Sendable (Double) -> Void)? = nil
    ) async throws -> AttachmentResponse {

        Logger.debug("Starting upload: \(attachment.filename) (\(attachment.formattedSize))")

        // Step 1: Initiate upload to get presigned URL
        let initRequest = InitiateUploadRequest(
            feedbackId: feedbackId,
            filename: attachment.filename,
            contentType: attachment.contentType,
            sizeBytes: attachment.sizeBytes
        )

        let uploadInfo = try await apiClient.initiateUpload(initRequest)

        guard let uploadURL = URL(string: uploadInfo.uploadUrl) else {
            throw NetworkError.invalidURL
        }

        Logger.debug("Got presigned URL, uploading to storage...")

        // Step 2: Upload to GCS using presigned URL (with retry)
        var lastError: Error?
        for attempt in 1...maxRetries {
            do {
                try await apiClient.uploadToStorage(
                    url: uploadURL,
                    data: attachment.data,
                    contentType: attachment.contentType,
                    progress: progress
                )
                break // Success, exit retry loop
            } catch {
                lastError = error
                Logger.warning("Upload attempt \(attempt) failed: \(error.localizedDescription)")

                if attempt < maxRetries {
                    try await Task.sleep(nanoseconds: UInt64(retryDelay * Double(NSEC_PER_SEC) * Double(attempt)))
                }
            }
        }

        if let error = lastError {
            // Check if last attempt also failed
            do {
                try await apiClient.uploadToStorage(
                    url: uploadURL,
                    data: attachment.data,
                    contentType: attachment.contentType,
                    progress: progress
                )
            } catch {
                throw error
            }
        }

        // Step 3: Confirm upload completion
        Logger.debug("Confirming upload completion...")
        let response = try await apiClient.completeUpload(attachmentId: uploadInfo.attachmentId)

        Logger.debug("Upload completed: \(response.id)")
        return response
    }

    /// Upload multiple attachments
    func uploadAll(
        feedbackId: UUID,
        attachments: [AttachmentData],
        progress: (@Sendable (Int, Int, Double) -> Void)? = nil
    ) async throws -> [AttachmentResponse] {
        var results: [AttachmentResponse] = []

        for (index, attachment) in attachments.enumerated() {
            let response = try await upload(
                feedbackId: feedbackId,
                attachment: attachment
            ) { itemProgress in
                progress?(index, attachments.count, itemProgress)
            }
            results.append(response)
        }

        return results
    }
}

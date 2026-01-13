import Foundation

/// API client for communicating with the FullDisclosure backend
public actor APIClient {
    private let token: String
    private let baseURL: URL
    private let session: URLSession
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    public init(token: String, baseURL: URL, timeout: TimeInterval = 30) {
        self.token = token
        self.baseURL = baseURL

        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = timeout
        config.timeoutIntervalForResource = 300
        self.session = URLSession(configuration: config)

        self.encoder = JSONEncoder()

        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            // Try ISO8601 with fractional seconds
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = formatter.date(from: dateString) {
                return date
            }

            // Try without fractional seconds
            formatter.formatOptions = [.withInternetDateTime]
            if let date = formatter.date(from: dateString) {
                return date
            }

            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Cannot decode date string \(dateString)"
            )
        }
    }

    // MARK: - Generic Request

    private func request<T: Decodable>(
        _ endpoint: Endpoint,
        body: Encodable? = nil
    ) async throws -> T {
        let url = baseURL.appendingPathComponent(endpoint.path)

        var request = URLRequest(url: url)
        request.httpMethod = endpoint.method.rawValue

        // Required headers (matching backend expectations)
        request.setValue(token, forHTTPHeaderField: "X-SDK-Token")
        request.setValue("ios", forHTTPHeaderField: "X-SDK-Source")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(sdkVersion, forHTTPHeaderField: "X-SDK-Version")

        if let body = body {
            request.httpBody = try encoder.encode(body)
        }

        Logger.debug("Request: \(endpoint.method.rawValue) \(url.absoluteString)")

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.invalidResponse
        }

        Logger.debug("Response: HTTP \(httpResponse.statusCode)")

        switch httpResponse.statusCode {
        case 200...299:
            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                Logger.error("Decoding error: \(error)")
                throw NetworkError.decodingError(error.localizedDescription)
            }
        case 401:
            throw NetworkError.unauthorized
        case 400, 422:
            let errorResponse = try? decoder.decode(ErrorResponse.self, from: data)
            throw NetworkError.validationError(errorResponse?.errors ?? [:])
        case 429:
            throw NetworkError.rateLimited
        default:
            throw NetworkError.serverError(httpResponse.statusCode)
        }
    }

    private var sdkVersion: String {
        "ios-1.0.0"
    }

    // MARK: - SDK Endpoints

    /// POST /api/sdk/feedback
    func submitFeedback(_ submission: SubmitFeedbackRequest) async throws -> FeedbackResponse {
        try await request(.submitFeedback, body: submission)
    }

    /// POST /api/sdk/attachments/init
    func initiateUpload(_ request: InitiateUploadRequest) async throws -> UploadInfo {
        try await self.request(.initiateUpload, body: request)
    }

    /// POST /api/sdk/attachments/complete
    func completeUpload(attachmentId: UUID) async throws -> AttachmentResponse {
        try await request(.completeUpload, body: CompleteUploadRequest(attachmentId: attachmentId))
    }

    /// POST /api/sdk/identify
    func identify(_ user: IdentifyRequest) async throws {
        let _: EmptyResponse = try await request(.identify, body: user)
    }

    // MARK: - Direct Upload to GCS

    /// Upload file data directly to the presigned GCS URL
    func uploadToStorage(
        url: URL,
        data: Data,
        contentType: String,
        progress: (@Sendable (Double) -> Void)? = nil
    ) async throws {
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"
        request.setValue(contentType, forHTTPHeaderField: "Content-Type")

        // Use upload task for progress tracking
        let delegate = UploadProgressDelegate(progress: progress)
        let uploadSession = URLSession(configuration: .default, delegate: delegate, delegateQueue: nil)

        let (_, response) = try await uploadSession.upload(for: request, from: data)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.uploadFailed
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            Logger.error("Upload failed with status: \(httpResponse.statusCode)")
            throw NetworkError.uploadFailed
        }

        Logger.debug("Upload completed successfully")
    }
}

// MARK: - Upload Progress Delegate

private final class UploadProgressDelegate: NSObject, URLSessionTaskDelegate, Sendable {
    private let progress: (@Sendable (Double) -> Void)?

    init(progress: (@Sendable (Double) -> Void)?) {
        self.progress = progress
    }

    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        didSendBodyData bytesSent: Int64,
        totalBytesSent: Int64,
        totalBytesExpectedToSend: Int64
    ) {
        let percentComplete = Double(totalBytesSent) / Double(totalBytesExpectedToSend)
        progress?(percentComplete)
    }
}

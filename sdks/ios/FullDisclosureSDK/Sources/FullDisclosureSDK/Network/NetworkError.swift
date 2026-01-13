import Foundation

/// Errors that can occur during network operations
public enum NetworkError: LocalizedError, Sendable {
    case notInitialized
    case invalidURL
    case invalidResponse
    case unauthorized
    case rateLimited
    case validationError([String: String])
    case serverError(Int)
    case uploadFailed
    case decodingError(String)
    case networkUnavailable
    case timeout
    case unknown(String)

    public var errorDescription: String? {
        switch self {
        case .notInitialized:
            return "SDK not initialized. Call FullDisclosure.shared.initialize(token:) first."
        case .invalidURL:
            return "Invalid URL"
        case .invalidResponse:
            return "Invalid response from server"
        case .unauthorized:
            return "Invalid or expired SDK token"
        case .rateLimited:
            return "Rate limit exceeded. Please try again later."
        case .validationError(let errors):
            return "Validation failed: \(errors.map { "\($0.key): \($0.value)" }.joined(separator: ", "))"
        case .serverError(let code):
            return "Server error (HTTP \(code))"
        case .uploadFailed:
            return "Failed to upload attachment"
        case .decodingError(let message):
            return "Failed to decode response: \(message)"
        case .networkUnavailable:
            return "Network connection unavailable"
        case .timeout:
            return "Request timed out"
        case .unknown(let message):
            return message
        }
    }

    public var isRetryable: Bool {
        switch self {
        case .rateLimited, .serverError, .networkUnavailable, .timeout:
            return true
        default:
            return false
        }
    }
}

/// Error response from the API
struct ErrorResponse: Decodable {
    let code: String?
    let message: String?
    let errors: [String: String]?
}

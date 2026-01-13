import Foundation

/// HTTP methods
enum HTTPMethod: String {
    case GET
    case POST
    case PUT
    case DELETE
    case PATCH
}

/// API endpoints
enum Endpoint {
    case submitFeedback
    case initiateUpload
    case completeUpload
    case identify

    var path: String {
        switch self {
        case .submitFeedback:
            return "/api/sdk/feedback"
        case .initiateUpload:
            return "/api/sdk/attachments/init"
        case .completeUpload:
            return "/api/sdk/attachments/complete"
        case .identify:
            return "/api/sdk/identify"
        }
    }

    var method: HTTPMethod {
        switch self {
        case .submitFeedback, .initiateUpload, .completeUpload, .identify:
            return .POST
        }
    }
}

// MARK: - Request Models

/// Request body for submitting feedback
struct SubmitFeedbackRequest: Encodable {
    let title: String
    let description: String
    let type: FeedbackType
    let submitterEmail: String?
    let submitterName: String?
    let submitterIdentifier: String?
    let sourceMetadata: SourceMetadata

    enum CodingKeys: String, CodingKey {
        case title
        case description
        case type
        case submitterEmail = "submitter_email"
        case submitterName = "submitter_name"
        case submitterIdentifier = "submitter_identifier"
        case sourceMetadata = "source_metadata"
    }
}

/// Source metadata included with feedback
struct SourceMetadata: Encodable {
    let deviceModel: String
    let osVersion: String
    let appVersion: String
    let appBuild: String
    let bundleId: String
    let locale: String
    let timezone: String
    let screenResolution: String
    let batteryLevel: Float?
    let batteryState: String?
    let networkType: String?
    let freeMemory: Int64?
    let freeDiskSpace: Int64?
    let consoleLogs: [String]?
    let customMetadata: [String: String]?
    let userTraits: [String: String]?

    enum CodingKeys: String, CodingKey {
        case deviceModel = "device_model"
        case osVersion = "os_version"
        case appVersion = "app_version"
        case appBuild = "app_build"
        case bundleId = "bundle_id"
        case locale
        case timezone
        case screenResolution = "screen_resolution"
        case batteryLevel = "battery_level"
        case batteryState = "battery_state"
        case networkType = "network_type"
        case freeMemory = "free_memory"
        case freeDiskSpace = "free_disk_space"
        case consoleLogs = "console_logs"
        case customMetadata = "custom_metadata"
        case userTraits = "user_traits"
    }

    init(deviceInfo: DeviceInfo, consoleLogs: [String]? = nil, customMetadata: [String: String]? = nil, userTraits: [String: String]? = nil) {
        self.deviceModel = deviceInfo.deviceModel
        self.osVersion = deviceInfo.osVersion
        self.appVersion = deviceInfo.appVersion
        self.appBuild = deviceInfo.appBuild
        self.bundleId = deviceInfo.bundleId
        self.locale = deviceInfo.locale
        self.timezone = deviceInfo.timezone
        self.screenResolution = deviceInfo.screenResolution
        self.batteryLevel = deviceInfo.batteryLevel
        self.batteryState = deviceInfo.batteryState
        self.networkType = deviceInfo.networkType
        self.freeMemory = deviceInfo.freeMemory
        self.freeDiskSpace = deviceInfo.freeDiskSpace
        self.consoleLogs = consoleLogs
        self.customMetadata = customMetadata
        self.userTraits = userTraits
    }
}

/// Request body for initiating upload
struct InitiateUploadRequest: Encodable {
    let feedbackId: UUID
    let filename: String
    let contentType: String
    let sizeBytes: Int64

    enum CodingKeys: String, CodingKey {
        case feedbackId = "feedback_id"
        case filename
        case contentType = "content_type"
        case sizeBytes = "size_bytes"
    }
}

/// Request body for completing upload
struct CompleteUploadRequest: Encodable {
    let attachmentId: UUID

    enum CodingKeys: String, CodingKey {
        case attachmentId = "attachment_id"
    }
}

/// Request body for user identification
struct IdentifyRequest: Encodable {
    let userId: String
    let email: String?
    let name: String?
    let traits: [String: String]?

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case email
        case name
        case traits
    }
}

// MARK: - Response Models

/// Response from feedback submission
struct FeedbackResponse: Decodable {
    let id: UUID
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case createdAt = "created_at"
    }
}

/// Response from initiating upload
struct UploadInfo: Decodable {
    let attachmentId: UUID
    let uploadUrl: String
    let expiresAt: String

    enum CodingKeys: String, CodingKey {
        case attachmentId = "attachment_id"
        case uploadUrl = "upload_url"
        case expiresAt = "expires_at"
    }
}

/// Response from completing upload
struct AttachmentResponse: Decodable {
    let id: UUID
    let filename: String
    let contentType: String
    let sizeBytes: Int64
    let status: String

    enum CodingKeys: String, CodingKey {
        case id
        case filename
        case contentType = "content_type"
        case sizeBytes = "size_bytes"
        case status
    }
}

/// Empty response for endpoints that don't return data
struct EmptyResponse: Decodable {}

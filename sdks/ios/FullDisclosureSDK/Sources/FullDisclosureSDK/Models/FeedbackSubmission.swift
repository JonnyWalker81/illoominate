import Foundation

/// Data for submitting feedback
public struct FeedbackSubmission: Sendable {
    public var title: String
    public var description: String
    public var type: FeedbackType
    public var email: String?
    public var name: String?
    public var attachments: [AttachmentData]
    public var includeConsoleLogs: Bool
    public var customMetadata: [String: String]

    public init(
        title: String,
        description: String,
        type: FeedbackType = .general,
        email: String? = nil,
        name: String? = nil,
        attachments: [AttachmentData] = [],
        includeConsoleLogs: Bool = false,
        customMetadata: [String: String] = [:]
    ) {
        self.title = title
        self.description = description
        self.type = type
        self.email = email
        self.name = name
        self.attachments = attachments
        self.includeConsoleLogs = includeConsoleLogs
        self.customMetadata = customMetadata
    }
}

/// Result of a successful feedback submission
public struct FeedbackResult: Sendable, Equatable {
    public let feedbackId: UUID
    public let createdAt: Date

    public init(feedbackId: UUID, createdAt: Date) {
        self.feedbackId = feedbackId
        self.createdAt = createdAt
    }
}

/// Identified user information
public struct IdentifiedUser: Sendable {
    public let userId: String
    public let email: String?
    public let name: String?
    public let traits: [String: String]?

    public init(
        userId: String,
        email: String? = nil,
        name: String? = nil,
        traits: [String: String]? = nil
    ) {
        self.userId = userId
        self.email = email
        self.name = name
        self.traits = traits
    }
}

import Foundation

/// The type of feedback being submitted
public enum FeedbackType: String, Codable, CaseIterable, Sendable {
    case bug = "bug"
    case feature = "feature"
    case general = "general"

    /// Human-readable display name
    public var displayName: String {
        switch self {
        case .bug:
            return "Bug Report"
        case .feature:
            return "Feature Request"
        case .general:
            return "General Feedback"
        }
    }

    /// SF Symbol icon name
    public var iconName: String {
        switch self {
        case .bug:
            return "ladybug"
        case .feature:
            return "lightbulb"
        case .general:
            return "bubble.left.and.bubble.right"
        }
    }

    /// Description for the type
    public var typeDescription: String {
        switch self {
        case .bug:
            return "Report a problem or unexpected behavior"
        case .feature:
            return "Suggest a new feature or improvement"
        case .general:
            return "Share your thoughts or ask a question"
        }
    }
}

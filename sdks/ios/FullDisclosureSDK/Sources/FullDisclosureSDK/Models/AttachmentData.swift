import Foundation
import UIKit

/// Represents attachment data to be uploaded with feedback
public struct AttachmentData: Sendable {
    public let data: Data
    public let filename: String
    public let contentType: String

    public init(data: Data, filename: String, contentType: String) {
        self.data = data
        self.filename = filename
        self.contentType = contentType
    }

    /// Create an attachment from a UIImage
    public init?(image: UIImage, filename: String = "screenshot.jpg", compressionQuality: CGFloat = 0.8) {
        guard let data = image.jpegData(compressionQuality: compressionQuality) else {
            return nil
        }
        self.data = data
        self.filename = filename
        self.contentType = "image/jpeg"
    }

    /// Create an attachment from a PNG image
    public init?(pngImage: UIImage, filename: String = "screenshot.png") {
        guard let data = pngImage.pngData() else {
            return nil
        }
        self.data = data
        self.filename = filename
        self.contentType = "image/png"
    }

    /// Create an attachment from a video file URL
    public init?(videoURL: URL) {
        guard FileManager.default.fileExists(atPath: videoURL.path),
              let data = try? Data(contentsOf: videoURL) else {
            return nil
        }
        self.data = data
        self.filename = videoURL.lastPathComponent
        self.contentType = "video/mp4"
    }

    /// Create an attachment from a text log
    public init(logText: String, filename: String = "console.log") {
        self.data = Data(logText.utf8)
        self.filename = filename
        self.contentType = "text/plain"
    }

    /// Size in bytes
    public var sizeBytes: Int64 {
        Int64(data.count)
    }

    /// Human-readable size
    public var formattedSize: String {
        let formatter = ByteCountFormatter()
        formatter.countStyle = .file
        return formatter.string(fromByteCount: sizeBytes)
    }
}

/// Allowed content types for attachments (matching backend validation)
public enum AllowedContentType: String, CaseIterable {
    case jpeg = "image/jpeg"
    case png = "image/png"
    case gif = "image/gif"
    case webp = "image/webp"
    case svg = "image/svg+xml"
    case mp4 = "video/mp4"
    case webm = "video/webm"
    case quicktime = "video/quicktime"
    case pdf = "application/pdf"
    case text = "text/plain"

    public static func isAllowed(_ contentType: String) -> Bool {
        allCases.map(\.rawValue).contains(contentType)
    }
}

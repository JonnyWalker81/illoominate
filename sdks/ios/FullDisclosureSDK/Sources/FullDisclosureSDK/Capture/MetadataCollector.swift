import Foundation
import UIKit

/// Collects device and app metadata for feedback submissions
@MainActor
struct MetadataCollector {

    /// Collect all metadata for a feedback submission
    static func collect(
        consoleLogs: [String]? = nil,
        customMetadata: [String: String]? = nil,
        userTraits: [String: String]? = nil
    ) -> SourceMetadata {
        let deviceInfo = DeviceInfo.collect()

        return SourceMetadata(
            deviceInfo: deviceInfo,
            consoleLogs: consoleLogs,
            customMetadata: customMetadata,
            userTraits: userTraits
        )
    }
}

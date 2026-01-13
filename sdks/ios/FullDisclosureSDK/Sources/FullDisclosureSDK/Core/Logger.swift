import Foundation
import os

/// Internal logger for the SDK
enum Logger {
    private static let subsystem = "com.fulldisclosure.sdk"
    private static let logger = os.Logger(subsystem: subsystem, category: "SDK")

    static var isEnabled = true

    static func debug(_ message: String) {
        guard isEnabled else { return }
        logger.debug("\(message, privacy: .public)")
    }

    static func info(_ message: String) {
        guard isEnabled else { return }
        logger.info("\(message, privacy: .public)")
    }

    static func warning(_ message: String) {
        guard isEnabled else { return }
        logger.warning("\(message, privacy: .public)")
    }

    static func error(_ message: String) {
        guard isEnabled else { return }
        logger.error("\(message, privacy: .public)")
    }
}

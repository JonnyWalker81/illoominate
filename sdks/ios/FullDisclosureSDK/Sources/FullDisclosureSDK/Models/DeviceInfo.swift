import Foundation
import UIKit

/// Device and app information collected for feedback
struct DeviceInfo: Codable, Sendable {
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
    let isJailbroken: Bool

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
        case isJailbroken = "is_jailbroken"
    }

    @MainActor
    static func collect() -> DeviceInfo {
        let device = UIDevice.current
        let screen = UIScreen.main
        let bundle = Bundle.main

        // Enable battery monitoring temporarily
        let previousBatteryMonitoringState = device.isBatteryMonitoringEnabled
        device.isBatteryMonitoringEnabled = true

        let batteryLevel = device.batteryLevel >= 0 ? device.batteryLevel : nil
        let batteryState: String? = {
            switch device.batteryState {
            case .charging: return "charging"
            case .full: return "full"
            case .unplugged: return "unplugged"
            case .unknown: return nil
            @unknown default: return nil
            }
        }()

        // Restore battery monitoring state
        device.isBatteryMonitoringEnabled = previousBatteryMonitoringState

        return DeviceInfo(
            deviceModel: deviceModelIdentifier(),
            osVersion: device.systemVersion,
            appVersion: bundle.infoDictionary?["CFBundleShortVersionString"] as? String ?? "unknown",
            appBuild: bundle.infoDictionary?["CFBundleVersion"] as? String ?? "unknown",
            bundleId: bundle.bundleIdentifier ?? "unknown",
            locale: Locale.current.identifier,
            timezone: TimeZone.current.identifier,
            screenResolution: "\(Int(screen.bounds.width * screen.scale))x\(Int(screen.bounds.height * screen.scale))",
            batteryLevel: batteryLevel,
            batteryState: batteryState,
            networkType: nil, // Would require Reachability or NWPathMonitor
            freeMemory: freeMemoryBytes(),
            freeDiskSpace: freeDiskSpaceBytes(),
            isJailbroken: checkJailbroken()
        )
    }

    private static func deviceModelIdentifier() -> String {
        var systemInfo = utsname()
        uname(&systemInfo)
        let machineMirror = Mirror(reflecting: systemInfo.machine)
        return machineMirror.children.reduce("") { identifier, element in
            guard let value = element.value as? Int8, value != 0 else { return identifier }
            return identifier + String(UnicodeScalar(UInt8(value)))
        }
    }

    private static func freeMemoryBytes() -> Int64? {
        var pageSize: vm_size_t = 0
        host_page_size(mach_host_self(), &pageSize)

        var vmStats = vm_statistics64()
        var count = mach_msg_type_number_t(MemoryLayout<vm_statistics64>.size / MemoryLayout<integer_t>.size)

        let result = withUnsafeMutablePointer(to: &vmStats) {
            $0.withMemoryRebound(to: integer_t.self, capacity: Int(count)) {
                host_statistics64(mach_host_self(), HOST_VM_INFO64, $0, &count)
            }
        }

        guard result == KERN_SUCCESS else { return nil }
        return Int64(vmStats.free_count) * Int64(pageSize)
    }

    private static func freeDiskSpaceBytes() -> Int64? {
        let fileURL = URL(fileURLWithPath: NSHomeDirectory())
        do {
            let values = try fileURL.resourceValues(forKeys: [.volumeAvailableCapacityForImportantUsageKey])
            return values.volumeAvailableCapacityForImportantUsage
        } catch {
            return nil
        }
    }

    private static func checkJailbroken() -> Bool {
        #if targetEnvironment(simulator)
        return false
        #else
        // Check for common jailbreak indicators
        let jailbreakPaths = [
            "/Applications/Cydia.app",
            "/Library/MobileSubstrate/MobileSubstrate.dylib",
            "/bin/bash",
            "/usr/sbin/sshd",
            "/etc/apt",
            "/private/var/lib/apt/"
        ]

        for path in jailbreakPaths {
            if FileManager.default.fileExists(atPath: path) {
                return true
            }
        }

        // Check if we can write outside sandbox
        let testPath = "/private/jailbreak_test.txt"
        do {
            try "test".write(toFile: testPath, atomically: true, encoding: .utf8)
            try FileManager.default.removeItem(atPath: testPath)
            return true
        } catch {
            return false
        }
        #endif
    }
}

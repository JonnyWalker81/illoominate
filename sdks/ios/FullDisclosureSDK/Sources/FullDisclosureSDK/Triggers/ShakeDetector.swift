import UIKit
import SwiftUI
import CoreMotion

/// Notification posted when device is shaken
public extension Notification.Name {
    static let deviceDidShake = Notification.Name("com.fulldisclosure.deviceDidShake")
}

/// Detects shake gestures to trigger feedback using Core Motion
@MainActor
public final class ShakeDetector {
    public static let shared = ShakeDetector()

    private var isEnabled = false
    private var onShake: (() -> Void)?

    private let motionManager = CMMotionManager()
    private let motionQueue = OperationQueue()

    // Shake detection parameters
    private let shakeThreshold: Double = 2.5  // Acceleration threshold in g
    private let shakeInterval: TimeInterval = 0.5  // Minimum time between shakes
    private var lastShakeTime: Date = .distantPast

    private init() {
        motionQueue.name = "com.fulldisclosure.shakeDetector"
        motionQueue.maxConcurrentOperationCount = 1
    }

    /// Enable shake detection
    public func enable(onShake: @escaping () -> Void) {
        self.onShake = onShake
        self.isEnabled = true

        startMotionDetection()

        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleShakeNotification),
            name: .deviceDidShake,
            object: nil
        )

        Logger.debug("Shake detection enabled")
    }

    /// Disable shake detection
    public func disable() {
        isEnabled = false
        onShake = nil

        stopMotionDetection()

        NotificationCenter.default.removeObserver(
            self,
            name: .deviceDidShake,
            object: nil
        )

        Logger.debug("Shake detection disabled")
    }

    private func startMotionDetection() {
        guard motionManager.isAccelerometerAvailable else {
            Logger.warning("Accelerometer not available")
            return
        }

        motionManager.accelerometerUpdateInterval = 0.1

        motionManager.startAccelerometerUpdates(to: motionQueue) { [weak self] data, error in
            guard let self = self, let data = data, error == nil else { return }

            // Calculate total acceleration magnitude (excluding gravity would require device motion)
            let acceleration = data.acceleration
            let magnitude = sqrt(
                acceleration.x * acceleration.x +
                acceleration.y * acceleration.y +
                acceleration.z * acceleration.z
            )

            // Subtract 1g for gravity (approximate)
            let netAcceleration = abs(magnitude - 1.0)

            if netAcceleration > self.shakeThreshold {
                self.handleShakeDetected()
            }
        }

        Logger.debug("Motion detection started")
    }

    private func stopMotionDetection() {
        motionManager.stopAccelerometerUpdates()
        Logger.debug("Motion detection stopped")
    }

    private func handleShakeDetected() {
        let now = Date()
        guard now.timeIntervalSince(lastShakeTime) > shakeInterval else { return }
        lastShakeTime = now

        DispatchQueue.main.async {
            NotificationCenter.default.post(name: .deviceDidShake, object: nil)
        }
    }

    @objc private func handleShakeNotification() {
        guard isEnabled else { return }

        Logger.debug("Shake detected!")

        // Haptic feedback
        let generator = UINotificationFeedbackGenerator()
        generator.notificationOccurred(.success)

        onShake?()
    }
}

// MARK: - SwiftUI View Modifier

/// View modifier for handling shake gestures in SwiftUI
struct ShakeViewModifier: ViewModifier {
    let action: () -> Void

    func body(content: Content) -> some View {
        content
            .onReceive(NotificationCenter.default.publisher(for: .deviceDidShake)) { _ in
                action()
            }
    }
}

public extension View {
    /// Trigger an action when the device is shaken
    func onShake(perform action: @escaping () -> Void) -> some View {
        modifier(ShakeViewModifier(action: action))
    }
}

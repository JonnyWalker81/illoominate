import UIKit

/// Captures screenshots with privacy redaction support
@MainActor
public final class ScreenshotCapture {
    private let configuration: Configuration

    init(configuration: Configuration) {
        self.configuration = configuration
    }

    /// Capture the current screen
    public func capture() -> UIImage? {
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let window = windowScene.windows.first(where: { $0.isKeyWindow }) else {
            Logger.warning("No key window found for screenshot")
            return nil
        }

        // Store redaction state
        var redactionOverlays: [UIView] = []

        // Apply redaction before capture
        applyRedaction(to: window, overlays: &redactionOverlays)

        // Capture the window
        let renderer = UIGraphicsImageRenderer(bounds: window.bounds)
        let image = renderer.image { _ in
            window.drawHierarchy(in: window.bounds, afterScreenUpdates: true)
        }

        // Remove redaction overlays
        for overlay in redactionOverlays {
            overlay.removeFromSuperview()
        }

        // Compress the image
        return compress(image)
    }

    /// Capture with a specific view excluded
    public func capture(excluding excludedView: UIView) -> UIImage? {
        let wasHidden = excludedView.isHidden
        excludedView.isHidden = true

        let image = capture()

        excludedView.isHidden = wasHidden
        return image
    }

    // MARK: - Privacy Redaction

    private func applyRedaction(to window: UIWindow, overlays: inout [UIView]) {
        // Find and redact sensitive views
        let sensitiveViews = findSensitiveViews(in: window)

        for view in sensitiveViews {
            // Create blur overlay
            let blurEffect = UIBlurEffect(style: .regular)
            let blurView = UIVisualEffectView(effect: blurEffect)
            blurView.frame = view.convert(view.bounds, to: window)
            blurView.layer.cornerRadius = view.layer.cornerRadius
            blurView.clipsToBounds = true

            window.addSubview(blurView)
            overlays.append(blurView)
        }
    }

    private func findSensitiveViews(in view: UIView) -> [UIView] {
        var sensitiveViews: [UIView] = []

        // Check if this view is marked for redaction
        if configuration.redactedViewTags.contains(view.tag) {
            sensitiveViews.append(view)
        }

        // Check for secure text fields
        if let textField = view as? UITextField, textField.isSecureTextEntry {
            sensitiveViews.append(textField)
        }

        // Check for text views with sensitive content (basic heuristic)
        if let textField = view as? UITextField,
           textField.textContentType == .password ||
            textField.textContentType == .newPassword ||
            textField.textContentType == .creditCardNumber {
            sensitiveViews.append(textField)
        }

        // Recursively check subviews
        for subview in view.subviews {
            sensitiveViews.append(contentsOf: findSensitiveViews(in: subview))
        }

        return sensitiveViews
    }

    // MARK: - Compression

    private func compress(_ image: UIImage) -> UIImage {
        guard let data = image.jpegData(compressionQuality: configuration.imageCompressionQuality),
              let compressed = UIImage(data: data) else {
            return image
        }
        return compressed
    }
}

// MARK: - Static Access

extension ScreenshotCapture {
    /// Capture screenshot using default configuration
    @MainActor
    public static func captureScreen(
        redactedViewTags: Set<Int> = [],
        compressionQuality: CGFloat = 0.8
    ) -> UIImage? {
        let config = Configuration(
            redactedViewTags: redactedViewTags,
            imageCompressionQuality: compressionQuality
        )
        let capture = ScreenshotCapture(configuration: config)
        return capture.capture()
    }
}

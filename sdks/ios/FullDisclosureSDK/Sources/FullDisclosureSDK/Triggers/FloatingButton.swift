import UIKit
import SwiftUI

/// A draggable floating button for triggering feedback
@MainActor
public final class FloatingButton {
    public static let shared = FloatingButton()

    private var buttonWindow: UIWindow?
    private var onTap: (() -> Void)?
    private let buttonSize: CGFloat = 56

    private init() {}

    /// Show the floating button
    public func show(theme: Theme = .default, onTap: @escaping () -> Void) {
        guard buttonWindow == nil else { return }

        self.onTap = onTap

        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene else {
            Logger.warning("No window scene found for floating button")
            return
        }

        // Create a dedicated window for the button
        let window = PassthroughWindow(windowScene: windowScene)
        window.windowLevel = .alert + 1
        window.backgroundColor = .clear
        window.isHidden = false

        // Create the button view
        let buttonView = FloatingButtonView(
            theme: theme,
            size: buttonSize,
            onTap: { [weak self] in
                self?.onTap?()
            }
        )

        let hostingController = UIHostingController(rootView: buttonView)
        hostingController.view.backgroundColor = .clear
        window.rootViewController = hostingController

        buttonWindow = window
        Logger.debug("Floating button shown")
    }

    /// Hide the floating button
    public func hide() {
        buttonWindow?.isHidden = true
        buttonWindow = nil
        onTap = nil
        Logger.debug("Floating button hidden")
    }

    /// Check if button is visible
    public var isVisible: Bool {
        buttonWindow != nil && buttonWindow?.isHidden == false
    }
}

// MARK: - Passthrough Window

/// A window that only handles touches on its subviews
private class PassthroughWindow: UIWindow {
    override func hitTest(_ point: CGPoint, with event: UIEvent?) -> UIView? {
        guard let hitView = super.hitTest(point, with: event) else {
            return nil
        }

        // Only return the hit view if it's not the root view
        return hitView == rootViewController?.view ? nil : hitView
    }
}

// MARK: - SwiftUI Button View

private struct FloatingButtonView: View {
    let theme: Theme
    let size: CGFloat
    let onTap: () -> Void

    @State private var position: CGPoint = .zero
    @State private var isDragging = false
    @GestureState private var dragOffset: CGSize = .zero

    var body: some View {
        GeometryReader { geometry in
            Button(action: onTap) {
                ZStack {
                    Circle()
                        .fill(theme.primaryColor)
                        .shadow(color: .black.opacity(0.3), radius: 8, x: 0, y: 4)

                    Image(systemName: "bubble.left.and.bubble.right.fill")
                        .font(.system(size: size * 0.4))
                        .foregroundColor(.white)
                }
                .frame(width: size, height: size)
            }
            .scaleEffect(isDragging ? 1.1 : 1.0)
            .position(
                x: position.x + dragOffset.width,
                y: position.y + dragOffset.height
            )
            .gesture(
                DragGesture()
                    .updating($dragOffset) { value, state, _ in
                        state = value.translation
                    }
                    .onChanged { _ in
                        isDragging = true
                    }
                    .onEnded { value in
                        isDragging = false
                        position = CGPoint(
                            x: position.x + value.translation.width,
                            y: position.y + value.translation.height
                        )
                        // Snap to edges
                        snapToEdge(in: geometry.size)
                    }
            )
            .onAppear {
                // Initial position: bottom right
                position = CGPoint(
                    x: geometry.size.width - size - 16,
                    y: geometry.size.height - size - 100
                )
            }
            .animation(.spring(response: 0.3), value: isDragging)
        }
    }

    private func snapToEdge(in size: CGSize) {
        let padding: CGFloat = 16
        let halfButton = self.size / 2

        // Snap to nearest horizontal edge
        let leftDistance = position.x
        let rightDistance = size.width - position.x

        if leftDistance < rightDistance {
            position.x = padding + halfButton
        } else {
            position.x = size.width - padding - halfButton
        }

        // Clamp vertical position
        position.y = max(padding + halfButton + 50, min(size.height - padding - halfButton - 50, position.y))
    }
}

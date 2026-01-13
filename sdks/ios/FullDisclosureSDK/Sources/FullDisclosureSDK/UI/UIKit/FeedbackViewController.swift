import UIKit
import SwiftUI

/// UIKit wrapper for the feedback view
@MainActor
public final class FeedbackViewController: UIViewController {

    private var hostingController: UIHostingController<FeedbackView>?
    private let configuration: Configuration
    private let initialType: FeedbackType?
    private let screenshot: UIImage?
    private let theme: Theme
    private let apiClient: APIClient?

    private var onSubmit: ((FeedbackResult) -> Void)?
    private var onCancel: (() -> Void)?

    public init(
        type: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        configuration: Configuration = .default,
        theme: Theme = .default,
        apiClient: APIClient? = nil
    ) {
        self.initialType = type
        self.screenshot = screenshot
        self.configuration = configuration
        self.theme = theme
        self.apiClient = apiClient
        super.init(nibName: nil, bundle: nil)
    }

    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    public override func viewDidLoad() {
        super.viewDidLoad()

        let feedbackView = FeedbackView(
            initialType: initialType,
            screenshot: screenshot,
            configuration: configuration,
            apiClient: apiClient,
            theme: theme,
            onSubmit: { [weak self] result in
                self?.onSubmit?(result)
                self?.dismiss(animated: true)
            },
            onCancel: { [weak self] in
                self?.onCancel?()
                self?.dismiss(animated: true)
            }
        )

        hostingController = UIHostingController(rootView: feedbackView)

        guard let hostingController = hostingController else { return }

        addChild(hostingController)
        view.addSubview(hostingController.view)
        hostingController.view.translatesAutoresizingMaskIntoConstraints = false

        NSLayoutConstraint.activate([
            hostingController.view.topAnchor.constraint(equalTo: view.topAnchor),
            hostingController.view.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            hostingController.view.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            hostingController.view.trailingAnchor.constraint(equalTo: view.trailingAnchor)
        ])

        hostingController.didMove(toParent: self)
    }

    // MARK: - Presentation

    /// Present the feedback view controller from any view controller
    public static func present(
        from viewController: UIViewController,
        type: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        configuration: Configuration = .default,
        theme: Theme = .default,
        apiClient: APIClient? = nil,
        onSubmit: ((FeedbackResult) -> Void)? = nil,
        onCancel: (() -> Void)? = nil
    ) {
        let feedbackVC = FeedbackViewController(
            type: type,
            screenshot: screenshot,
            configuration: configuration,
            theme: theme,
            apiClient: apiClient
        )
        feedbackVC.onSubmit = onSubmit
        feedbackVC.onCancel = onCancel

        let nav = UINavigationController(rootViewController: feedbackVC)
        nav.modalPresentationStyle = .pageSheet

        if let sheet = nav.sheetPresentationController {
            sheet.detents = [.large()]
            sheet.prefersGrabberVisible = true
        }

        viewController.present(nav, animated: true)
    }
}

// MARK: - UIWindow Extension for presenting feedback

public extension UIWindow {

    /// Present feedback dialog from the key window
    @MainActor
    static func presentFeedbackDialog(
        type: FeedbackType? = nil,
        screenshot: UIImage? = nil,
        configuration: Configuration = .default,
        theme: Theme = .default,
        apiClient: APIClient? = nil,
        onSubmit: ((FeedbackResult) -> Void)? = nil,
        onCancel: (() -> Void)? = nil
    ) {
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let rootViewController = windowScene.windows.first(where: { $0.isKeyWindow })?.rootViewController else {
            Logger.warning("No root view controller found")
            return
        }

        // Find the top-most presented view controller
        var topController = rootViewController
        while let presented = topController.presentedViewController {
            topController = presented
        }

        FeedbackViewController.present(
            from: topController,
            type: type,
            screenshot: screenshot,
            configuration: configuration,
            theme: theme,
            apiClient: apiClient,
            onSubmit: onSubmit,
            onCancel: onCancel
        )
    }
}

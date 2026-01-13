# FullDisclosure iOS SDK

Collect user feedback, bug reports, and feature requests directly from your iOS app.

## Requirements

- iOS 15.0+
- Swift 5.9+
- Xcode 15.0+

## Installation

### Swift Package Manager

Add the following to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/yourusername/FullDisclosureSDK.git", from: "1.0.0")
]
```

Or in Xcode:
1. Go to **File > Add Package Dependencies**
2. Enter the repository URL
3. Select the version and add to your target

## Quick Start

### 1. Initialize the SDK

```swift
import FullDisclosureSDK

@main
struct MyApp: App {
    init() {
        FullDisclosure.shared.initialize(token: "your_sdk_token_here")
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

That's it! Users can now shake their device to submit feedback.

### 2. Manual Trigger (Optional)

```swift
Button("Send Feedback") {
    FullDisclosure.shared.showFeedbackDialog()
}
```

## Features

### Shake to Report
Enabled by default. Users shake their device to open the feedback dialog.

```swift
// Disable if needed
FullDisclosure.shared.disableShakeToReport()

// Re-enable
FullDisclosure.shared.enableShakeToReport()
```

### Screenshot Capture
Automatically captures a screenshot when the feedback dialog opens (when triggered by shake).

```swift
// Manual screenshot
let screenshot = FullDisclosure.shared.captureScreenshot()
```

### Screen Recording
Users can record their screen to demonstrate issues.

```swift
// Programmatic recording
try await FullDisclosure.shared.startRecording()
// ... user performs actions ...
let videoURL = try await FullDisclosure.shared.stopRecording()
```

### Console Log Capture
Automatically captures console logs to help debug issues.

```swift
// Get captured logs
let logs = FullDisclosure.shared.getConsoleLogs()

// Clear logs
FullDisclosure.shared.clearConsoleLogs()
```

### User Identification
Link feedback to authenticated users.

```swift
// After user logs in
try await FullDisclosure.shared.identify(
    userId: user.id,
    email: user.email,
    name: user.displayName
)

// After logout
FullDisclosure.shared.clearIdentity()
```

### Floating Button
Show a persistent feedback button.

```swift
FullDisclosure.shared.showFloatingButton()
FullDisclosure.shared.hideFloatingButton()
```

## Configuration

```swift
let config = Configuration.default
    .with(baseURL: URL(string: "https://your-api.com")!)
    .with(theme: .dark)
    .with(enableScreenRecording: true)
    .with(requireEmail: true)
    .with(feedbackTypes: [.bug, .feature])
    .with(redactedViewTags: [100, 101])  // Blur sensitive views
    .with(debugLogging: true)

FullDisclosure.shared.initialize(
    token: "your_sdk_token_here",
    configuration: config
)
```

## Theming

```swift
let customTheme = Theme(
    primaryColor: .blue,
    backgroundColor: .white,
    cornerRadius: 16,
    showPoweredBy: false
)

let config = Configuration.default.with(theme: customTheme)
```

## Privacy Redaction

Mark sensitive views to be blurred in screenshots:

```swift
// In SwiftUI
Text("Credit Card: 1234-5678-9012-3456")
    .tag(100)  // Add to redactedViewTags

// In UIKit
creditCardLabel.tag = 100
```

Then configure:

```swift
let config = Configuration.default
    .with(redactedViewTags: [100])
```

## UIKit Support

```swift
import FullDisclosureSDK

class ViewController: UIViewController {
    @IBAction func sendFeedback() {
        FeedbackViewController.present(
            from: self,
            type: .bug,
            onSubmit: { result in
                print("Feedback submitted: \(result.feedbackId)")
            }
        )
    }
}
```

## Programmatic Submission

```swift
let feedback = FeedbackSubmission(
    title: "App crashes on checkout",
    description: "When I tap the pay button...",
    type: .bug,
    includeConsoleLogs: true
)

let result = try await FullDisclosure.shared.submitFeedback(feedback)
print("Submitted: \(result.feedbackId)")
```

## SwiftUI View Modifier

```swift
ContentView()
    .onShake {
        // Custom shake handling
    }
```

## Data Collected

The SDK collects the following data to help with debugging:

- Device model and OS version
- App version and build number
- Screen resolution
- Battery level (optional)
- Free memory and disk space (optional)
- Console logs (optional, user-controlled)
- Screenshots and screen recordings (user-initiated)

All data collection is transparent and can be controlled via configuration.

## License

MIT License

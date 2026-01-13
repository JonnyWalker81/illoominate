// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "FullDisclosureSDK",
    platforms: [
        .iOS(.v17)
    ],
    products: [
        .library(
            name: "FullDisclosureSDK",
            targets: ["FullDisclosureSDK"]
        )
    ],
    dependencies: [],
    targets: [
        .target(
            name: "FullDisclosureSDK",
            dependencies: [],
            swiftSettings: [
                .enableExperimentalFeature("StrictConcurrency")
            ]
        ),
        .testTarget(
            name: "FullDisclosureSDKTests",
            dependencies: ["FullDisclosureSDK"]
        )
    ]
)

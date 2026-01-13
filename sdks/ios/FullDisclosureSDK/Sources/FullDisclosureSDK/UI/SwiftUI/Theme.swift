import SwiftUI

/// Theme configuration for customizing the SDK appearance
public struct Theme: Sendable {
    public var primaryColor: Color
    public var backgroundColor: Color
    public var surfaceColor: Color
    public var textColor: Color
    public var secondaryTextColor: Color
    public var errorColor: Color
    public var successColor: Color
    public var borderColor: Color
    public var cornerRadius: CGFloat
    public var showPoweredBy: Bool

    public init(
        primaryColor: Color = Color(red: 99/255, green: 102/255, blue: 241/255),
        backgroundColor: Color = Color(.systemBackground),
        surfaceColor: Color = Color(.secondarySystemBackground),
        textColor: Color = .primary,
        secondaryTextColor: Color = .secondary,
        errorColor: Color = .red,
        successColor: Color = .green,
        borderColor: Color = Color(.systemGray4),
        cornerRadius: CGFloat = 12,
        showPoweredBy: Bool = true
    ) {
        self.primaryColor = primaryColor
        self.backgroundColor = backgroundColor
        self.surfaceColor = surfaceColor
        self.textColor = textColor
        self.secondaryTextColor = secondaryTextColor
        self.errorColor = errorColor
        self.successColor = successColor
        self.borderColor = borderColor
        self.cornerRadius = cornerRadius
        self.showPoweredBy = showPoweredBy
    }

    public static let `default` = Theme()

    public static let dark = Theme(
        primaryColor: Color(red: 129/255, green: 140/255, blue: 248/255),
        backgroundColor: Color(.black),
        surfaceColor: Color(.systemGray6),
        textColor: .white,
        secondaryTextColor: Color(.systemGray),
        borderColor: Color(.systemGray5)
    )
}

// MARK: - Color Hex Extension

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 0, 0, 0)
        }

        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

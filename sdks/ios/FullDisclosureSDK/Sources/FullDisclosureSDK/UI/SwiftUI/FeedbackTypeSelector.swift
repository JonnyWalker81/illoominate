import SwiftUI

/// Selector for feedback type (bug, feature, general)
struct FeedbackTypeSelector: View {
    @Binding var selectedType: FeedbackType
    let availableTypes: [FeedbackType]
    let theme: Theme

    var body: some View {
        HStack(spacing: 12) {
            ForEach(availableTypes, id: \.self) { type in
                TypeButton(
                    type: type,
                    isSelected: selectedType == type,
                    theme: theme
                ) {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        selectedType = type
                    }
                }
            }
        }
    }
}

private struct TypeButton: View {
    let type: FeedbackType
    let isSelected: Bool
    let theme: Theme
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: type.iconName)
                    .font(.system(size: 24))

                Text(type.displayName)
                    .font(.caption)
                    .fontWeight(isSelected ? .semibold : .regular)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                RoundedRectangle(cornerRadius: theme.cornerRadius)
                    .fill(isSelected ? theme.primaryColor.opacity(0.1) : Color.clear)
            )
            .foregroundColor(isSelected ? theme.primaryColor : theme.secondaryTextColor)
            .overlay(
                RoundedRectangle(cornerRadius: theme.cornerRadius)
                    .stroke(
                        isSelected ? theme.primaryColor : theme.borderColor,
                        lineWidth: isSelected ? 2 : 1
                    )
            )
        }
        .buttonStyle(.plain)
        .accessibilityLabel(type.displayName)
        .accessibilityHint(type.typeDescription)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}

#Preview {
    FeedbackTypeSelector(
        selectedType: .constant(.bug),
        availableTypes: [.bug, .feature, .general],
        theme: .default
    )
    .padding()
}

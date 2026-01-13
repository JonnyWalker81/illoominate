import SwiftUI
import AVKit

/// Preview and management of attachments
struct AttachmentsSection: View {
    @Binding var attachments: [AttachmentItem]
    let theme: Theme
    let maxCount: Int
    let onAddScreenshot: () async -> Void
    let onAddRecording: () async -> Void
    let onRemove: (AttachmentItem) -> Void

    @State private var isRecording = false
    @State private var showingActionSheet = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Attachments")
                    .font(.subheadline)
                    .fontWeight(.medium)
                    .foregroundColor(theme.textColor)

                Spacer()

                if attachments.count < maxCount {
                    Menu {
                        Button(action: {
                            Task { await onAddScreenshot() }
                        }) {
                            Label("Add Screenshot", systemImage: "camera")
                        }

                        Button(action: {
                            Task { await onAddRecording() }
                        }) {
                            Label("Record Screen", systemImage: "record.circle")
                        }
                    } label: {
                        Label("Add", systemImage: "plus.circle.fill")
                            .font(.subheadline)
                            .foregroundColor(theme.primaryColor)
                    }
                }
            }

            if attachments.isEmpty {
                HStack {
                    Spacer()
                    VStack(spacing: 8) {
                        Image(systemName: "photo.on.rectangle.angled")
                            .font(.largeTitle)
                            .foregroundColor(theme.secondaryTextColor)
                        Text("No attachments")
                            .font(.caption)
                            .foregroundColor(theme.secondaryTextColor)
                    }
                    .padding(.vertical, 24)
                    Spacer()
                }
                .background(
                    RoundedRectangle(cornerRadius: theme.cornerRadius)
                        .stroke(theme.borderColor, style: StrokeStyle(lineWidth: 1, dash: [5]))
                )
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(attachments) { item in
                            AttachmentThumbnail(
                                item: item,
                                theme: theme,
                                onRemove: { onRemove(item) }
                            )
                        }
                    }
                }
            }
        }
    }
}

/// Individual attachment item
struct AttachmentItem: Identifiable {
    let id = UUID()
    let data: AttachmentData
    let thumbnail: UIImage?
    let isVideo: Bool

    init(data: AttachmentData, thumbnail: UIImage? = nil) {
        self.data = data
        self.thumbnail = thumbnail
        self.isVideo = data.contentType.starts(with: "video/")
    }

    init(image: UIImage, filename: String = "screenshot.jpg") {
        self.data = AttachmentData(image: image, filename: filename)!
        self.thumbnail = image
        self.isVideo = false
    }

    init(videoURL: URL) {
        self.data = AttachmentData(videoURL: videoURL)!
        self.thumbnail = Self.generateVideoThumbnail(from: videoURL)
        self.isVideo = true
    }

    private static func generateVideoThumbnail(from url: URL) -> UIImage? {
        let asset = AVAsset(url: url)
        let generator = AVAssetImageGenerator(asset: asset)
        generator.appliesPreferredTrackTransform = true

        do {
            let cgImage = try generator.copyCGImage(at: .zero, actualTime: nil)
            return UIImage(cgImage: cgImage)
        } catch {
            return nil
        }
    }
}

/// Thumbnail view for an attachment
private struct AttachmentThumbnail: View {
    let item: AttachmentItem
    let theme: Theme
    let onRemove: () -> Void

    var body: some View {
        ZStack(alignment: .topTrailing) {
            Group {
                if let thumbnail = item.thumbnail {
                    Image(uiImage: thumbnail)
                        .resizable()
                        .aspectRatio(contentMode: .fill)
                } else {
                    Color.gray.opacity(0.3)
                        .overlay(
                            Image(systemName: item.isVideo ? "video" : "doc")
                                .font(.title2)
                                .foregroundColor(theme.secondaryTextColor)
                        )
                }
            }
            .frame(width: 80, height: 80)
            .clipShape(RoundedRectangle(cornerRadius: 8))
            .overlay(
                RoundedRectangle(cornerRadius: 8)
                    .stroke(theme.borderColor, lineWidth: 1)
            )

            // Video indicator
            if item.isVideo {
                Image(systemName: "play.circle.fill")
                    .font(.title3)
                    .foregroundColor(.white)
                    .shadow(radius: 2)
                    .padding(4)
            }

            // Remove button
            Button(action: onRemove) {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 20))
                    .foregroundColor(.white)
                    .background(Circle().fill(Color.black.opacity(0.5)))
            }
            .offset(x: 8, y: -8)
        }
        .accessibilityLabel(item.isVideo ? "Video attachment" : "Image attachment")
        .accessibilityHint("Double tap to remove")
    }
}

#Preview {
    AttachmentsSection(
        attachments: .constant([]),
        theme: .default,
        maxCount: 5,
        onAddScreenshot: {},
        onAddRecording: {},
        onRemove: { _ in }
    )
    .padding()
}

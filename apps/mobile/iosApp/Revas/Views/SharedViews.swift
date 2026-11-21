import SwiftUI
import Shared

func severityColor(_ code: String) -> Color {
    switch code {
    case "high":
        return .red
    case "medium":
        return .orange
    default:
        return .green
    }
}

struct OfflineBannerView: View {
    let label: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text("Menampilkan data tersimpan (mode offline)")
                .font(.caption.weight(.semibold))
            if let label {
                Text("Terakhir diperbarui \(label)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }
}

struct ErrorBannerView: View {
    let message: String
    let retry: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(message)
                .font(.footnote)
                .foregroundStyle(.red)
            Button("Coba lagi", action: retry)
                .font(.footnote)
        }
    }
}

struct DetailRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack(alignment: .top) {
            Text(label)
                .font(.footnote)
                .foregroundStyle(.secondary)
                .frame(width: 128, alignment: .leading)
            Text(value)
                .font(.subheadline)
        }
    }
}

struct SeverityBadge: View {
    let label: String
    let color: Color

    var body: some View {
        Text(label)
            .font(.caption.weight(.semibold))
            .padding(.horizontal, 8)
            .padding(.vertical, 2)
            .background(color.opacity(0.15), in: Capsule())
            .foregroundStyle(color)
    }
}

struct PrivacyShieldView: View {
    var body: some View {
        ZStack {
            Color(.systemBackground).ignoresSafeArea()
            VStack(spacing: 8) {
                Image(systemName: "lock.shield")
                    .font(.largeTitle)
                    .foregroundStyle(.tint)
                Text("Revas terkunci")
                    .font(.headline)
            }
        }
    }
}

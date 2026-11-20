import SwiftUI
import Shared

@MainActor
struct AlertDetailScreen: View {
    @State private var model: AlertDetailModel

    init(sdk: SharedSdk, alertId: String) {
        _model = State(initialValue: AlertDetailModel(alertId: alertId, sdk: sdk))
    }

    var body: some View {
        Group {
            if let alert = model.alert {
                content(alert)
            } else if model.loading {
                ProgressView("Memuat…")
            } else {
                ErrorBannerView(message: model.error ?? "Data tidak ditemukan.") {
                    Task { await model.load() }
                }
                .padding()
            }
        }
        .navigationTitle("Detail alert")
        .navigationBarTitleDisplayMode(.inline)
        .task { await model.load() }
    }

    private func content(_ alert: Alert) -> some View {
        List {
            Section {
                HStack(spacing: 8) {
                    SeverityBadge(
                        label: MobileLabels.shared.severity(severity: alert.severity),
                        color: severityColor(MobileLabels.shared.severityCode(severity: alert.severity))
                    )
                    Text(Formatters.shared.idr(amount: alert.claimedAmount))
                        .font(.title3.bold())
                }
                DetailRow(label: "Klaim", value: alert.claimId)
                DetailRow(label: "Status", value: MobileLabels.shared.status(status: alert.status))
                DetailRow(label: "Jenis kerugian", value: MobileLabels.shared.lossType(lossType: alert.lossType))
                DetailRow(label: "Jumlah indikasi", value: "\(alert.flagCount)")
                DetailRow(
                    label: "SLA",
                    value: alert.slaBreached
                        ? "Lewat SLA · \(Formatters.shared.dateTime(instant: alert.slaDueAt))"
                        : Formatters.shared.dateTime(instant: alert.slaDueAt)
                )
                DetailRow(label: "Dibuat", value: Formatters.shared.dateTime(instant: alert.createdAt))
                if alert.suppressed {
                    Text("Alert ini ditahan sementara oleh aturan penahanan.")
                        .font(.footnote)
                        .foregroundStyle(.red)
                }
            }
            Section("Indikasi aturan") {
                ForEach(alert.flags, id: \.flagId) { flag in
                    FlagCardView(flag: flag)
                }
            }
            Section {
                Text("Mode baca-saja. Aksi tulis hanya tersedia di aplikasi web.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .listStyle(.insetGrouped)
    }
}

private struct FlagCardView: View {
    let flag: AlertFlag

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(flag.ruleName)
                    .font(.subheadline.weight(.semibold))
                Spacer()
                SeverityBadge(
                    label: MobileLabels.shared.severity(severity: flag.severity),
                    color: severityColor(MobileLabels.shared.severityCode(severity: flag.severity))
                )
            }
            Text("\(flag.family) · \(flag.reasonCode) · v\(flag.ruleVersion)")
                .font(.caption)
                .foregroundStyle(.secondary)
            Text(flag.explanation)
                .font(.footnote)
            DetailRow(label: "Nilai pemicu", value: flag.triggerValue)
            DetailRow(label: "Versi parameter", value: "\(flag.paramsVersion)")
            DetailRow(
                label: "Referensi bukti",
                value: flag.evidenceRefs.isEmpty ? "-" : flag.evidenceRefs.joined(separator: ", ")
            )
            if flag.suppressed {
                Text("Ditahan sementara")
                    .font(.caption)
                    .foregroundStyle(.red)
            }
        }
        .padding(.vertical, 2)
    }
}

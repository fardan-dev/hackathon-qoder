import SwiftUI
import Shared

@MainActor
struct CaseDetailScreen: View {
    @State private var model: CaseDetailModel

    init(sdk: SharedSdk, caseId: String) {
        _model = State(initialValue: CaseDetailModel(caseId: caseId, sdk: sdk))
    }

    var body: some View {
        Group {
            if let detail = model.detail {
                content(detail)
            } else if model.loading {
                ProgressView("Memuat…")
            } else {
                ErrorBannerView(message: model.error ?? "Data tidak ditemukan.") {
                    Task { await model.load() }
                }
                .padding()
            }
        }
        .navigationTitle("Detail kasus")
        .navigationBarTitleDisplayMode(.inline)
        .task { await model.load() }
    }

    private func content(_ detail: CaseDetail) -> some View {
        List {
            Section("Ringkasan") {
                Text(detail.title)
                    .font(.headline)
                Text(detail.summary)
                    .font(.footnote)
                DetailRow(label: "Status", value: detail.status)
                DetailRow(label: "Prioritas", value: "\(detail.priority)")
                DetailRow(label: "Jumlah klaim", value: "\(detail.claimCount)")
                DetailRow(label: "Dibuat", value: Formatters.shared.dateTime(instant: detail.createdAt))
                DetailRow(label: "Diperbarui", value: Formatters.shared.dateTime(instant: detail.updatedAt))
            }
            Section("Klaim terkait") {
                if detail.claims.isEmpty {
                    Text("Belum ada klaim terkait.")
                        .foregroundStyle(.secondary)
                }
                ForEach(detail.claims, id: \.claimId) { claim in
                    NavigationLink(value: Route.alert(claim.alertId)) {
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                Text("Klaim \(claim.claimId)")
                                    .font(.subheadline)
                                Spacer()
                                SeverityBadge(
                                    label: MobileLabels.shared.severity(severity: claim.severity),
                                    color: severityColor(MobileLabels.shared.severityCode(severity: claim.severity))
                                )
                            }
                            Text(
                                "\(Formatters.shared.idr(amount: claim.claimedAmount)) · "
                                + "\(MobileLabels.shared.lossType(lossType: claim.lossType)) · "
                                + "\(claim.flags.count) indikasi"
                            )
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        }
                    }
                }
            }
            Section("Bukti") {
                if detail.evidence.isEmpty {
                    Text("Belum ada bukti terlampir.")
                        .foregroundStyle(.secondary)
                }
                ForEach(detail.evidence, id: \.evidenceId) { evidence in
                    VStack(alignment: .leading, spacing: 4) {
                        Text(evidence.artifactType)
                            .font(.subheadline.weight(.semibold))
                        Text(evidence.source)
                            .font(.caption)
                        Text("SHA-256 \(evidence.contentSha256.prefix(16))…")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        if let notes = evidence.notes, !notes.isEmpty {
                            Text(notes)
                                .font(.caption)
                        }
                        Text(Formatters.shared.dateTime(instant: evidence.createdAt))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            Section("Riwayat keputusan") {
                if detail.decisions.isEmpty {
                    Text("Belum ada keputusan.")
                        .foregroundStyle(.secondary)
                }
                ForEach(detail.decisions, id: \.decisionId) { decision in
                    VStack(alignment: .leading, spacing: 4) {
                        HStack {
                            Text(decision.outcome)
                                .font(.subheadline.weight(.semibold))
                            Spacer()
                            Text(decision.reasonCode)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                        if !decision.notes.isEmpty {
                            Text(decision.notes)
                                .font(.footnote)
                        }
                        Text(Formatters.shared.dateTime(instant: decision.decidedAt))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            Section {
                Text("Mode baca-saja. Aksi tulis hanya tersedia di aplikasi web.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .listStyle(.insetGrouped)
        .privacySensitive()
    }
}

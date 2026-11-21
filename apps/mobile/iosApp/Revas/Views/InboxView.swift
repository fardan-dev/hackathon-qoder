import SwiftUI
import Shared

@MainActor
struct InboxView: View {
    @Bindable var model: InboxModel
    let userName: String
    let onSignOut: () -> Void

    @State private var severitySelection = "all"
    @State private var slaOnly = false

    var body: some View {
        List {
            if model.fromCache {
                Section {
                    OfflineBannerView(label: model.lastSyncedLabel)
                }
            }
            if let error = model.error, model.alerts.isEmpty {
                Section {
                    ErrorBannerView(message: error) {
                        Task { await model.refresh() }
                    }
                }
            }
            Section {
                Picker("Severity", selection: $severitySelection) {
                    Text("Semua").tag("all")
                    Text("Tinggi").tag("high")
                    Text("Sedang").tag("medium")
                    Text("Rendah").tag("low")
                }
                .pickerStyle(.segmented)
                Toggle("Lewat SLA", isOn: $slaOnly)
            } header: {
                Text("Halo, \(userName)")
            }
            Section {
                if model.loading {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                } else if model.alerts.isEmpty {
                    Text("Tidak ada alert pada filter ini")
                        .foregroundStyle(.secondary)
                } else {
                    ForEach(model.alerts, id: \.alertId) { alert in
                        NavigationLink(value: Route.alert(alert.alertId)) {
                            AlertRowView(alert: alert)
                        }
                    }
                }
            } footer: {
                Text("Mode baca-saja. Aksi tulis hanya tersedia di aplikasi web.")
            }
        }
        .listStyle(.insetGrouped)
        .navigationTitle("Alert")
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button("Muat ulang") {
                        Task { await model.refresh() }
                    }
                    Button("Keluar", role: .destructive, action: onSignOut)
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .refreshable { await model.refresh() }
        .task { await model.load() }
        .onChange(of: severitySelection) { _, newValue in
            Task { await model.apply(severityCode: newValue == "all" ? nil : newValue, slaOnly: slaOnly) }
        }
        .onChange(of: slaOnly) { _, newValue in
            Task { await model.apply(severityCode: severitySelection == "all" ? nil : severitySelection, slaOnly: newValue) }
        }
    }
}

struct AlertRowView: View {
    let alert: Alert

    private var code: String { MobileLabels.shared.severityCode(severity: alert.severity) }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                SeverityBadge(
                    label: MobileLabels.shared.severity(severity: alert.severity),
                    color: severityColor(code)
                )
                Text(Formatters.shared.idr(amount: alert.claimedAmount))
                    .font(.headline)
            }
            Text("Klaim \(alert.claimId)")
                .font(.subheadline)
            Text(meta)
                .font(.caption)
                .foregroundStyle(.secondary)
            Text(slaText)
                .font(.caption)
                .foregroundStyle(alert.slaBreached ? Color.red : Color.secondary)
        }
        .padding(.vertical, 2)
    }

    private var meta: String {
        var parts = [
            MobileLabels.shared.status(status: alert.status),
            MobileLabels.shared.lossType(lossType: alert.lossType),
            "\(alert.flagCount) indikasi",
        ]
        if alert.suppressed {
            parts.append("Ditahan sementara")
        }
        return parts.joined(separator: " · ")
    }

    private var slaText: String {
        let due = Formatters.shared.dateTime(instant: alert.slaDueAt)
        return alert.slaBreached ? "Lewat SLA · tenggat \(due)" : "SLA \(due)"
    }
}

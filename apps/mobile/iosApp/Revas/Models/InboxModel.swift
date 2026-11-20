@preconcurrency import Shared
import Observation

@MainActor
@Observable
final class InboxModel {
    private(set) var alerts: [Alert] = []
    private(set) var loading = true
    private(set) var refreshing = false
    private(set) var fromCache = false
    private(set) var lastSyncedLabel: String?
    private(set) var error: String?
    private(set) var severityCode: String?
    private(set) var slaOnly = false

    private let sdk: SharedSdk

    init(sdk: SharedSdk) {
        self.sdk = sdk
    }

    func load() async {
        await fetch(force: false)
    }

    func refresh() async {
        await fetch(force: true)
    }

    func apply(severityCode: String?, slaOnly: Bool) async {
        self.severityCode = severityCode
        self.slaOnly = slaOnly
        await fetch(force: true)
    }

    private func fetch(force: Bool) async {
        guard !refreshing else { return }
        refreshing = true
        loading = alerts.isEmpty
        defer {
            refreshing = false
            loading = false
        }
        do {
            let result = try await sdk.loadAlerts(
                severityCode: severityCode,
                slaBreachedOnly: slaOnly,
                forceRefresh: force
            )
            guard result.isSuccess, let feed = result.feed else {
                error = result.errorMessage ?? "Gagal memuat alert."
                return
            }
            alerts = feed.alerts
            fromCache = feed.fromCache
            if let syncedAt = feed.lastSyncedAt {
                lastSyncedLabel = Formatters.shared.dateTime(instant: syncedAt)
            } else {
                lastSyncedLabel = nil
            }
            error = nil
        } catch {
            self.error = "Tidak dapat terhubung ke server."
        }
    }
}

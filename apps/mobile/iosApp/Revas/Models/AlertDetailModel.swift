@preconcurrency import Shared
import Observation

@MainActor
@Observable
final class AlertDetailModel {
    let alertId: String
    private(set) var alert: Alert?
    private(set) var loading = true
    private(set) var error: String?

    private let sdk: SharedSdk

    init(alertId: String, sdk: SharedSdk) {
        self.alertId = alertId
        self.sdk = sdk
    }

    func load() async {
        guard alert == nil else { return }
        loading = true
        defer { loading = false }
        do {
            let result = try await sdk.loadAlert(alertId: alertId)
            if result.isSuccess {
                alert = result.alert
                error = nil
            } else {
                error = result.errorMessage ?? "Gagal memuat detail alert."
            }
        } catch {
            self.error = "Tidak dapat terhubung ke server."
        }
    }
}

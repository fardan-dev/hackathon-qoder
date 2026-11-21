@preconcurrency import Shared
import Observation

@MainActor
@Observable
final class CaseDetailModel {
    let caseId: String
    private(set) var detail: CaseDetail?
    private(set) var loading = true
    private(set) var error: String?

    private let sdk: SharedSdk

    init(caseId: String, sdk: SharedSdk) {
        self.caseId = caseId
        self.sdk = sdk
    }

    func load() async {
        guard detail == nil else { return }
        loading = true
        defer { loading = false }
        do {
            let result = try await sdk.loadCase(caseId: caseId)
            if result.isSuccess {
                detail = result.case
                error = nil
            } else {
                error = result.errorMessage ?? "Gagal memuat detail kasus."
            }
        } catch {
            self.error = "Tidak dapat terhubung ke server."
        }
    }
}

@preconcurrency import Shared
import Observation

@MainActor
@Observable
final class SessionModel {
    enum Phase {
        case loading
        case signedOut
        case locked
        case unlocked
    }

    private(set) var phase: Phase = .loading
    private(set) var user: UserSession?
    var email = "reviewer@claimshield.local"
    var password = ""
    private(set) var submitting = false
    private(set) var error: String?

    private let sdk: SharedSdk

    init(sdk: SharedSdk) {
        self.sdk = sdk
    }

    func restore() async {
        guard phase == .loading else { return }
        do {
            let result = try await sdk.restoreSession()
            if result.isSuccess {
                user = result.user
                phase = result.user == nil ? .signedOut : .locked
            } else {
                error = result.errorMessage
                phase = .signedOut
            }
        } catch {
            self.error = "Tidak dapat memulihkan sesi."
            phase = .signedOut
        }
    }

    func submit() async {
        guard !submitting else { return }
        submitting = true
        error = nil
        defer { submitting = false }
        do {
            let result = try await sdk.login(email: email, password: password)
            if result.isSuccess, let session = result.user {
                user = session
                password = ""
                phase = .unlocked
            } else {
                error = result.errorMessage ?? "Gagal masuk. Periksa kembali kredensial."
            }
        } catch {
            self.error = "Tidak dapat terhubung ke server."
        }
    }

    func unlock() {
        guard phase == .locked else { return }
        phase = .unlocked
    }

    func lock() {
        guard phase == .unlocked else { return }
        phase = .locked
    }

    func signOut() async {
        _ = try? await sdk.logout()
        user = nil
        password = ""
        phase = .signedOut
    }
}

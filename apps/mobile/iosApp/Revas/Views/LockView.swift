import SwiftUI

@MainActor
struct LockView: View {
    let session: SessionModel

    @State private var error: String?
    @State private var verifying = false

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "lock.shield")
                .font(.system(size: 48))
                .foregroundStyle(.tint)
            Text("Sesi terkunci")
                .font(.title2.bold())
            Text("Verifikasi Face ID untuk membuka akses.")
                .foregroundStyle(.secondary)

            if let error {
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(.red)
            }

            Button("Verifikasi") {
                Task { await verify() }
            }
            .buttonStyle(.borderedProminent)
            .disabled(verifying)

            Button("Keluar dan masuk dengan kata sandi") {
                Task { await session.signOut() }
            }
            .font(.footnote)
        }
        .padding(24)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .task { await verify() }
    }

    private func verify() async {
        guard !verifying else { return }
        verifying = true
        defer { verifying = false }
        if await BiometricGate.authenticate() {
            error = nil
            session.unlock()
        } else {
            error = "Verifikasi gagal atau dibatalkan."
        }
    }
}

import SwiftUI

@MainActor
struct LoginView: View {
    @Bindable var session: SessionModel

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Revas Mobile")
                .font(.largeTitle.bold())
                .foregroundStyle(.tint)
            Text("Pemantauan indikasi klaim dan kasus")
                .foregroundStyle(.secondary)

            TextField("Email", text: $session.email)
                .textContentType(.username)
                .keyboardType(.emailAddress)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .textFieldStyle(.roundedBorder)
                .accessibilityIdentifier("login_email")

            SecureField("Kata sandi", text: $session.password)
                .textContentType(.password)
                .textFieldStyle(.roundedBorder)
                .accessibilityIdentifier("login_password")
                .onSubmit { Task { await session.submit() } }

            if let error = session.error {
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(.red)
            }

            Button {
                Task { await session.submit() }
            } label: {
                if session.submitting {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                } else {
                    Text("Masuk")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(.borderedProminent)
            .disabled(session.email.isEmpty || session.password.isEmpty || session.submitting)

            Text("Aksi tulis (assign, eskalasi, keputusan) hanya tersedia di aplikasi web.")
                .font(.footnote)
                .foregroundStyle(.secondary)
        }
        .padding(24)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

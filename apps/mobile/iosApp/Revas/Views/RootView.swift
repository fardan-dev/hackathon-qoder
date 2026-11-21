import SwiftUI
import Shared

@MainActor
struct RootView: View {
    let container: AppContainer

    @State private var session: SessionModel
    @State private var deepLink: DeepLink?
    @Environment(\.scenePhase) private var scenePhase

    init(container: AppContainer) {
        self.container = container
        _session = State(initialValue: SessionModel(sdk: container.sdk))
    }

    var body: some View {
        Group {
            switch session.phase {
            case .loading:
                SplashView()
            case .signedOut:
                LoginView(session: session)
            case .locked:
                LockView(session: session)
            case .unlocked:
                AppShell(
                    container: container,
                    userName: session.user?.displayName.isEmpty == false
                        ? session.user?.displayName ?? ""
                        : session.user?.email ?? "",
                    deepLink: deepLink,
                    onSignOut: { Task { await session.signOut() } }
                )
            }
        }
        .task { await session.restore() }
        .onOpenURL { url in
            deepLink = DeepLink(url: url)
        }
        .onChange(of: scenePhase) { _, phase in
            if phase == .background {
                session.lock()
            }
        }
        .overlay {
            if scenePhase != .active {
                PrivacyShieldView()
            }
        }
    }
}

@MainActor
struct SplashView: View {
    var body: some View {
        VStack(spacing: 16) {
            Text("Revas Mobile")
                .font(.largeTitle.bold())
                .foregroundStyle(.tint)
            ProgressView()
        }
    }
}

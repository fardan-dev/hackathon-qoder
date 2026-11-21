import SwiftUI
import Shared

@MainActor
struct AppShell: View {
    private let sdk: SharedSdk
    let userName: String
    let deepLink: DeepLink?
    let onSignOut: () -> Void

    @State private var path: [Route] = []
    @State private var inbox: InboxModel
    @Environment(\.scenePhase) private var scenePhase

    init(container: AppContainer, userName: String, deepLink: DeepLink?, onSignOut: @escaping () -> Void) {
        self.sdk = container.sdk
        self.userName = userName
        self.deepLink = deepLink
        self.onSignOut = onSignOut
        _inbox = State(initialValue: InboxModel(sdk: container.sdk))
    }

    var body: some View {
        NavigationStack(path: $path) {
            InboxView(model: inbox, userName: userName, onSignOut: onSignOut)
                .navigationDestination(for: Route.self) { route in
                    switch route {
                    case .alert(let alertId):
                        AlertDetailScreen(sdk: sdk, alertId: alertId)
                    case .caseDetail(let caseId):
                        CaseDetailScreen(sdk: sdk, caseId: caseId)
                    }
                }
        }
        .onAppear { apply(deepLink) }
        .onChange(of: deepLink) { _, newValue in apply(newValue) }
        .onChange(of: scenePhase) { _, phase in
            if phase == .active {
                Task { await inbox.refresh() }
            }
        }
    }

    private func apply(_ link: DeepLink?) {
        guard let link else { return }
        path = [Route(link)]
    }
}

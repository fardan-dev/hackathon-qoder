import Foundation
import Shared

enum APIConfig {
    static var identityBaseUrl: String { base(port: 8081) }
    static var alertBaseUrl: String { base(port: 8084) }
    static var caseBaseUrl: String { base(port: 8085) }

    private static func base(port: Int) -> String {
        #if targetEnvironment(simulator)
        return "http://localhost:\(port)"
        #else
        let host = ProcessInfo.processInfo.environment["REVAS_API_HOST"] ?? "localhost"
        return "http://\(host):\(port)"
        #endif
    }
}

@MainActor
final class AppContainer {
    let sdk: SharedSdk

    init() {
        let config = MobileApiConfig(
            identityBaseUrl: APIConfig.identityBaseUrl,
            alertBaseUrl: APIConfig.alertBaseUrl,
            caseBaseUrl: APIConfig.caseBaseUrl
        )
        sdk = IosSdkKt.createIosSdk(config: config, tokenStorage: KeychainTokenStorage())
    }
}

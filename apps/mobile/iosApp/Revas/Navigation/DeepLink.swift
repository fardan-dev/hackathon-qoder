import Foundation

enum DeepLink: Equatable {
    case alert(String)
    case caseDetail(String)

    init?(url: URL) {
        guard url.scheme == "revas", let host = url.host else { return nil }
        let identifier = url.pathComponents.dropFirst().first ?? ""
        guard !identifier.isEmpty else { return nil }
        switch host {
        case "alert":
            self = .alert(identifier)
        case "case":
            self = .caseDetail(identifier)
        default:
            return nil
        }
    }
}

enum Route: Hashable {
    case alert(String)
    case caseDetail(String)

    init(_ deepLink: DeepLink) {
        switch deepLink {
        case .alert(let id):
            self = .alert(id)
        case .caseDetail(let id):
            self = .caseDetail(id)
        }
    }
}

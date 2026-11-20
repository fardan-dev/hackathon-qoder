import Foundation
import Testing
@testable import Revas

@Suite("Keychain token storage", .serialized)
struct KeychainTokenStorageTests {
    private let storage = KeychainTokenStorage(
        service: "com.revas.mobile.tests",
        account: "auth-token"
    )

    @Test("round trips and overwrites the token")
    func roundTrip() {
        storage.clearToken()
        #expect(storage.readToken() == nil)

        storage.writeToken(token: "token-abc")
        #expect(storage.readToken() == "token-abc")

        storage.writeToken(token: "token-def")
        #expect(storage.readToken() == "token-def")

        storage.clearToken()
        #expect(storage.readToken() == nil)
    }
}

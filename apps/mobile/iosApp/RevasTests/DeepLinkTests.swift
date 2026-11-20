import Foundation
import Testing
@testable import Revas

@Suite("Deep link routing")
struct DeepLinkTests {
    @Test("parses alert links")
    func parsesAlertLink() throws {
        let url = try #require(URL(string: "revas://alert/alert-123"))
        #expect(DeepLink(url: url) == .alert("alert-123"))
    }

    @Test("parses case links")
    func parsesCaseLink() throws {
        let url = try #require(URL(string: "revas://case/case-9"))
        #expect(DeepLink(url: url) == .caseDetail("case-9"))
    }

    @Test("rejects unknown hosts")
    func rejectsUnknownHost() throws {
        let url = try #require(URL(string: "revas://unknown/abc"))
        #expect(DeepLink(url: url) == nil)
    }

    @Test("rejects links without identifier")
    func rejectsMissingIdentifier() throws {
        let url = try #require(URL(string: "revas://alert"))
        #expect(DeepLink(url: url) == nil)
    }

    @Test("rejects foreign schemes")
    func rejectsForeignScheme() throws {
        let url = try #require(URL(string: "https://alert/abc"))
        #expect(DeepLink(url: url) == nil)
    }

    @Test("maps deep links to routes")
    func mapsToRoute() {
        #expect(Route(.alert("a1")) == .alert("a1"))
        #expect(Route(.caseDetail("c1")) == .caseDetail("c1"))
    }
}

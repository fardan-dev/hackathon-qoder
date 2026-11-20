import XCTest

final class LoginFlowUITests: XCTestCase {
    override func setUp() {
        continueAfterFailure = false
    }

    func testLoginScreenShowsReadOnlyNotice() {
        let app = XCUIApplication()
        app.launch()

        let emailField = app.textFields["login_email"]
        XCTAssertTrue(emailField.waitForExistence(timeout: 15))

        XCTAssertTrue(app.secureTextFields["login_password"].exists)
        XCTAssertTrue(app.buttons["Masuk"].exists)
        XCTAssertTrue(app.staticTexts["Aksi tulis (assign, eskalasi, keputusan) hanya tersedia di aplikasi web."].exists)
    }
}

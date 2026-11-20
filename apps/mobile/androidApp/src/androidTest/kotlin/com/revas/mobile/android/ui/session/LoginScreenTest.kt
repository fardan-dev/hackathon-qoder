package com.revas.mobile.android.ui.session

import androidx.compose.ui.test.assertIsEnabled
import androidx.compose.ui.test.assertIsNotEnabled
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performTextInput
import com.revas.mobile.android.ui.theme.RevasTheme
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test

class LoginScreenTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun emptyPasswordKeepsSubmitDisabled() {
        composeRule.setContent {
            RevasTheme {
                LoginScreen(
                    form = LoginUiState(email = "reviewer@claimshield.local", password = ""),
                    onEmailChange = {},
                    onPasswordChange = {},
                    onSubmit = {},
                )
            }
        }

        composeRule.onNodeWithText("Masuk").assertIsNotEnabled()
    }

    @Test
    fun submitIsInvokedWhenFormIsComplete() {
        var submitted = 0
        composeRule.setContent {
            RevasTheme {
                LoginScreen(
                    form = LoginUiState(email = "reviewer@claimshield.local", password = "rahasia"),
                    onEmailChange = {},
                    onPasswordChange = {},
                    onSubmit = { submitted += 1 },
                )
            }
        }

        composeRule.onNodeWithText("Masuk").assertIsEnabled().performClick()

        assertEquals(1, submitted)
    }

    @Test
    fun errorAndWebOnlyNoticeAreVisible() {
        composeRule.setContent {
            RevasTheme {
                LoginScreen(
                    form = LoginUiState(password = "salah", error = "Email atau kata sandi salah"),
                    onEmailChange = {},
                    onPasswordChange = {},
                    onSubmit = {},
                )
            }
        }

        composeRule.onNodeWithText("Email atau kata sandi salah").assertExists()
        composeRule
            .onNodeWithText("Aksi tulis (assign, eskalasi, keputusan) hanya tersedia di aplikasi web.")
            .assertExists()
    }

    @Test
    fun typingIsForwardedToCallbacks() {
        var email = ""
        composeRule.setContent {
            RevasTheme {
                LoginScreen(
                    form = LoginUiState(email = "", password = ""),
                    onEmailChange = { email = it },
                    onPasswordChange = {},
                    onSubmit = {},
                )
            }
        }

        composeRule.onNodeWithTag("login_email").performTextInput("baru@claimshield.local")

        assertEquals("baru@claimshield.local", email)
    }
}

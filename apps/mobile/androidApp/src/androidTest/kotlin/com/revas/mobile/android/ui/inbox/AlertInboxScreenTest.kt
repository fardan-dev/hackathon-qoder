@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.android.ui.inbox

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import com.revas.mobile.android.ui.theme.RevasTheme
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.AlertStatus
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.Severity
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

class AlertInboxScreenTest {

    @get:Rule
    val composeRule = createComposeRule()

    private val alert = sampleAlert()

    @Test
    fun rendersReadOnlyAlertSummary() {
        composeRule.setContent {
            RevasTheme {
                AlertInboxScreen(
                    state = AlertInboxUiState(loading = false, alerts = listOf(alert)),
                    userName = "Rina",
                    onRefresh = {},
                    onSignOut = {},
                    onSelectAlert = {},
                    onSelectSeverity = {},
                    onToggleSla = {},
                )
            }
        }

        composeRule.onNodeWithText("Klaim CLM-0001").assertIsDisplayed()
        composeRule.onNodeWithText("Rp 85.000.000").assertIsDisplayed()
        composeRule.onNodeWithText("Halo, Rina").assertIsDisplayed()
        composeRule.onNodeWithText("tenggat", substring = true).assertIsDisplayed()
        composeRule.onNodeWithText("Mode baca-saja", substring = true).assertExists()
        composeRule.onNodeWithText("Assign").assertDoesNotExist()
        composeRule.onNodeWithText("Eskalasi").assertDoesNotExist()
    }

    @Test
    fun offlineBannerShowsLastSync() {
        composeRule.setContent {
            RevasTheme {
                AlertInboxScreen(
                    state = AlertInboxUiState(
                        loading = false,
                        alerts = listOf(alert),
                        fromCache = true,
                        lastSyncedAt = Instant.parse("2026-11-20T02:00:00Z"),
                    ),
                    userName = "Rina",
                    onRefresh = {},
                    onSignOut = {},
                    onSelectAlert = {},
                    onSelectSeverity = {},
                    onToggleSla = {},
                )
            }
        }

        composeRule.onNodeWithText("Menampilkan data tersimpan (mode offline)").assertIsDisplayed()
        composeRule.onNodeWithText("Terakhir diperbarui 20 Nov 2026 09:00 WIB").assertIsDisplayed()
    }

    @Test
    fun emptyStateIsShownForEmptyFeed() {
        composeRule.setContent {
            RevasTheme {
                AlertInboxScreen(
                    state = AlertInboxUiState(loading = false, alerts = emptyList()),
                    userName = "Rina",
                    onRefresh = {},
                    onSignOut = {},
                    onSelectAlert = {},
                    onSelectSeverity = {},
                    onToggleSla = {},
                )
            }
        }

        composeRule.onNodeWithText("Tidak ada alert pada filter ini").assertIsDisplayed()
    }

    @Test
    fun selectingAlertAndSeverityIsForwarded() {
        var selectedAlert: String? = null
        var selectedSeverity: Severity? = Severity.Low
        composeRule.setContent {
            RevasTheme {
                AlertInboxScreen(
                    state = AlertInboxUiState(loading = false, alerts = listOf(alert)),
                    userName = "Rina",
                    onRefresh = {},
                    onSignOut = {},
                    onSelectAlert = { selectedAlert = it },
                    onSelectSeverity = { selectedSeverity = it },
                    onToggleSla = {},
                )
            }
        }

        composeRule.onNodeWithText("Klaim CLM-0001").performClick()
        composeRule.onNodeWithTag("filter_High").performClick()

        assertEquals(alert.alertId, selectedAlert)
        assertEquals(Severity.High, selectedSeverity)
    }

    @Test
    fun filterStillAppliesLocallyWhenFeedIsCached() {
        val feed = AlertFeed(alerts = listOf(alert), lastSyncedAt = null, fromCache = true)
        assertEquals(listOf(alert), feed.alerts.filter { AlertFilter(severity = Severity.High).matches(it) })
    }
}

private fun sampleAlert(): Alert = Alert(
    alertId = "alert-1",
    claimId = "CLM-0001",
    lossType = LossType.Motor,
    severity = Severity.High,
    severityRank = 1,
    flagCount = 2,
    claimedAmount = 85_000_000,
    flags = emptyList(),
    status = AlertStatus.OPEN,
    suppressed = false,
    slaDueAt = Instant.parse("2026-11-20T10:00:00Z"),
    slaBreached = true,
    createdAt = Instant.parse("2026-11-19T03:00:00Z"),
)

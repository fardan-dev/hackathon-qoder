@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.android.ui.inbox

import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.AlertStatus
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.model.Severity
import com.revas.mobile.domain.repository.AlertRepository
import com.revas.mobile.domain.usecase.LoadAlertsUseCase
import com.revas.mobile.domain.usecase.Outcome
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

class AlertInboxViewModelTest {

    @BeforeEach
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `loads feed on init without forcing refresh`() {
        val repository = FakeAlertRepository(
            Outcome.Success(AlertFeed(listOf(sampleAlert()), lastSyncedAt = null, fromCache = false)),
        )

        val viewModel = AlertInboxViewModel(LoadAlertsUseCase(repository))

        val state = viewModel.state.value
        assertFalse(state.loading)
        assertEquals(1, state.alerts.size)
        assertNull(state.error)
        assertEquals(listOf(AlertFilter() to false), repository.calls)
    }

    @Test
    fun `severity filter is forwarded with forced refresh`() {
        val repository = FakeAlertRepository(
            Outcome.Success(AlertFeed(listOf(sampleAlert()), lastSyncedAt = null, fromCache = false)),
        )
        val viewModel = AlertInboxViewModel(LoadAlertsUseCase(repository))

        viewModel.selectSeverity(Severity.High)

        assertEquals(Severity.High, viewModel.state.value.filter.severity)
        assertEquals(AlertFilter(severity = Severity.High) to true, repository.calls.last())
    }

    @Test
    fun `sla toggle is forwarded with forced refresh`() {
        val repository = FakeAlertRepository(
            Outcome.Success(AlertFeed(listOf(sampleAlert()), lastSyncedAt = null, fromCache = false)),
        )
        val viewModel = AlertInboxViewModel(LoadAlertsUseCase(repository))

        viewModel.toggleSlaOnly()

        assertTrue(viewModel.state.value.filter.slaBreachedOnly)
        assertEquals(AlertFilter(slaBreachedOnly = true) to true, repository.calls.last())
    }

    @Test
    fun `offline feed marks state as cached`() {
        val syncedAt = Instant.parse("2026-11-20T02:00:00Z")
        val repository = FakeAlertRepository(
            Outcome.Success(AlertFeed(listOf(sampleAlert()), lastSyncedAt = syncedAt, fromCache = true)),
        )

        val viewModel = AlertInboxViewModel(LoadAlertsUseCase(repository))

        val state = viewModel.state.value
        assertTrue(state.fromCache)
        assertEquals(syncedAt, state.lastSyncedAt)
    }

    @Test
    fun `failure surfaces a user facing message`() {
        val repository = FakeAlertRepository(Outcome.Failure(MobileError.Network()))

        val viewModel = AlertInboxViewModel(LoadAlertsUseCase(repository))

        val state = viewModel.state.value
        assertFalse(state.loading)
        assertTrue(state.alerts.isEmpty())
        assertEquals("Tidak dapat terhubung ke server. Periksa koneksi lalu coba lagi.", state.error)
    }
}

private class FakeAlertRepository(private val outcome: Outcome<AlertFeed>) : AlertRepository {
    val calls = mutableListOf<Pair<AlertFilter, Boolean>>()

    override suspend fun loadFeed(filter: AlertFilter, forceRefresh: Boolean): Outcome<AlertFeed> {
        calls += filter to forceRefresh
        return outcome
    }

    override suspend fun loadAlert(alertId: String): Outcome<Alert> = Outcome.Success(sampleAlert(alertId))
}

internal fun sampleAlert(alertId: String = "alert-1"): Alert = Alert(
    alertId = alertId,
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

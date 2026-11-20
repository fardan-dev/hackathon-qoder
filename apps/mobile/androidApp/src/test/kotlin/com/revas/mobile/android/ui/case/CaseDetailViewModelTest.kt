@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.android.ui.case

import androidx.lifecycle.SavedStateHandle
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.repository.CaseRepository
import com.revas.mobile.domain.usecase.LoadCaseDetailUseCase
import com.revas.mobile.domain.usecase.Outcome
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

class CaseDetailViewModelTest {

    @BeforeEach
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `loads case using the id from navigation arguments`() {
        val repository = FakeCaseRepository(Outcome.Success(sampleCase("case-42")))
        val viewModel = CaseDetailViewModel(
            savedStateHandle = SavedStateHandle(mapOf("caseId" to "case-42")),
            loadCase = LoadCaseDetailUseCase(repository),
        )

        val state = viewModel.state.value
        assertFalse(state.loading)
        assertEquals("case-42", state.case?.caseId)
        assertEquals(listOf("case-42"), repository.requestedIds)
    }

    @Test
    fun `blank case id fails validation without hitting the repository`() {
        val repository = FakeCaseRepository(Outcome.Success(sampleCase("case-1")))
        val viewModel = CaseDetailViewModel(
            savedStateHandle = SavedStateHandle(),
            loadCase = LoadCaseDetailUseCase(repository),
        )

        val state = viewModel.state.value
        assertNull(state.case)
        assertEquals("ID kasus tidak valid", state.error)
        assertEquals(emptyList<String>(), repository.requestedIds)
    }

    @Test
    fun `not found error is surfaced`() {
        val repository = FakeCaseRepository(Outcome.Failure(MobileError.NotFound()))
        val viewModel = CaseDetailViewModel(
            savedStateHandle = SavedStateHandle(mapOf("caseId" to "case-9")),
            loadCase = LoadCaseDetailUseCase(repository),
        )

        assertEquals("Data tidak ditemukan.", viewModel.state.value.error)
    }
}

private class FakeCaseRepository(private val outcome: Outcome<CaseDetail>) : CaseRepository {
    val requestedIds = mutableListOf<String>()

    override suspend fun loadCase(caseId: String): Outcome<CaseDetail> {
        requestedIds += caseId
        return outcome
    }
}

internal fun sampleCase(caseId: String): CaseDetail = CaseDetail(
    caseId = caseId,
    title = "Kasus dugaan fraud",
    summary = "Dua klaim motor dengan pola serupa.",
    status = "INVESTIGATING",
    priority = 1,
    createdAt = Instant.parse("2026-11-18T03:00:00Z"),
    updatedAt = Instant.parse("2026-11-20T03:00:00Z"),
    claimCount = 2,
    claims = emptyList(),
    evidence = emptyList(),
    decisions = emptyList(),
)

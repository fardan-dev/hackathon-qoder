package com.revas.mobile.domain

import com.revas.mobile.domain.model.ErrorKind
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.repository.SessionRepository
import com.revas.mobile.domain.usecase.LoginUseCase
import com.revas.mobile.domain.usecase.Outcome
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue

private class FakeSessionRepository(
    private val loginResult: Outcome<UserSession> = Outcome.Success(session),
) : SessionRepository {
    var lastEmail: String? = null
    var lastPassword: String? = null

    override suspend fun login(email: String, password: String): Outcome<UserSession> {
        lastEmail = email
        lastPassword = password
        return loginResult
    }

    override suspend fun restore(): Outcome<UserSession?> = Outcome.Success(null)
    override suspend fun logout(): Outcome<Unit> = Outcome.Success(Unit)
    override fun hasStoredToken(): Boolean = false

    companion object {
        val session = UserSession(
            userId = "u1",
            tenantId = "t1",
            email = "reviewer@claimshield.local",
            displayName = "Reviewer Demo",
            role = "reviewer",
            expiresAt = null,
        )
    }
}

class LoginUseCaseTest {
    @Test
    fun rejects_blank_email() = runTest {
        val repository = FakeSessionRepository()
        val result = LoginUseCase(repository)("   ", "secret")
        val failure = assertIs<Outcome.Failure>(result)
        assertEquals(ErrorKind.Validation, failure.error.kind)
    }

    @Test
    fun rejects_malformed_email() = runTest {
        val repository = FakeSessionRepository()
        val result = LoginUseCase(repository)("reviewer", "secret")
        val failure = assertIs<Outcome.Failure>(result)
        assertEquals(ErrorKind.Validation, failure.error.kind)
    }

    @Test
    fun rejects_empty_password() = runTest {
        val repository = FakeSessionRepository()
        val result = LoginUseCase(repository)("reviewer@claimshield.local", "")
        val failure = assertIs<Outcome.Failure>(result)
        assertEquals(ErrorKind.Validation, failure.error.kind)
    }

    @Test
    fun normalizes_email_before_login() = runTest {
        val repository = FakeSessionRepository()
        val result = LoginUseCase(repository)("  Reviewer@ClaimShield.Local ", "secret")
        assertTrue(result is Outcome.Success)
        assertEquals("reviewer@claimshield.local", repository.lastEmail)
        assertEquals("secret", repository.lastPassword)
    }
}

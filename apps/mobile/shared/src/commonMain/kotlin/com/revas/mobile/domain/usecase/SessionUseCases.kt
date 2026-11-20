package com.revas.mobile.domain.usecase

import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.repository.SessionRepository

class LoginUseCase(private val repository: SessionRepository) {
    suspend operator fun invoke(email: String, password: String): Outcome<UserSession> {
        val normalizedEmail = email.trim().lowercase()
        if (normalizedEmail.isEmpty()) {
            return Outcome.Failure(MobileError.Validation("Email wajib diisi"))
        }
        if (!normalizedEmail.contains('@') || normalizedEmail.startsWith('@') || normalizedEmail.endsWith('@')) {
            return Outcome.Failure(MobileError.Validation("Format email tidak valid"))
        }
        if (password.isEmpty()) {
            return Outcome.Failure(MobileError.Validation("Kata sandi wajib diisi"))
        }
        return repository.login(normalizedEmail, password)
    }
}

class RestoreSessionUseCase(private val repository: SessionRepository) {
    suspend operator fun invoke(): Outcome<UserSession?> = repository.restore()
}

class LogoutUseCase(private val repository: SessionRepository) {
    suspend operator fun invoke(): Outcome<Unit> = repository.logout()
}

class HasStoredSessionUseCase(private val repository: SessionRepository) {
    operator fun invoke(): Boolean = repository.hasStoredToken()
}

package com.revas.mobile.data.session

import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.mapper.toDomain
import com.revas.mobile.data.remote.FetchResult
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.repository.SessionRepository
import com.revas.mobile.domain.usecase.Outcome

internal class DefaultSessionRepository(
    private val api: MobileApi,
    private val tokenStorage: TokenStorage,
    private val cache: MobileCache,
) : SessionRepository {
    override suspend fun login(email: String, password: String): Outcome<UserSession> {
        return when (val result = api.login(email, password)) {
            is FetchResult.Success -> {
                tokenStorage.writeToken(result.value.token)
                cache.purge()
                Outcome.Success(result.value.user.toDomain(result.value.expiresAt))
            }

            FetchResult.NotModified -> Outcome.Failure(MobileError.Unknown("Respons login tidak valid"))
            is FetchResult.Failure -> Outcome.Failure(result.error)
        }
    }

    override suspend fun restore(): Outcome<UserSession?> {
        if (tokenStorage.readToken().isNullOrBlank()) {
            return Outcome.Success(null)
        }
        return when (val result = api.currentUser()) {
            is FetchResult.Success -> Outcome.Success(result.value.toDomain(null))
            FetchResult.NotModified -> Outcome.Failure(MobileError.Unknown("Respons sesi tidak valid"))
            is FetchResult.Failure -> {
                if (result.error.kind == com.revas.mobile.domain.model.ErrorKind.Unauthorized) {
                    clearLocalSession()
                    Outcome.Success(null)
                } else {
                    Outcome.Failure(result.error)
                }
            }
        }
    }

    override suspend fun logout(): Outcome<Unit> {
        runCatching { api.logout() }
        clearLocalSession()
        return Outcome.Success(Unit)
    }

    override fun hasStoredToken(): Boolean = !tokenStorage.readToken().isNullOrBlank()

    private fun clearLocalSession() {
        tokenStorage.clearToken()
        cache.purge()
    }
}

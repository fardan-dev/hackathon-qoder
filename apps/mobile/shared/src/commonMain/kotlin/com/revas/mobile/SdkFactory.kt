package com.revas.mobile

import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.data.remote.MobileApiConfig
import com.revas.mobile.data.repository.DefaultAlertRepository
import com.revas.mobile.data.repository.DefaultCaseRepository
import com.revas.mobile.data.session.DefaultSessionRepository
import com.revas.mobile.data.session.TokenStorage
import com.revas.mobile.domain.usecase.HasStoredSessionUseCase
import com.revas.mobile.domain.usecase.LoadAlertDetailUseCase
import com.revas.mobile.domain.usecase.LoadAlertsUseCase
import com.revas.mobile.domain.usecase.LoadCaseDetailUseCase
import com.revas.mobile.domain.usecase.LoginUseCase
import com.revas.mobile.domain.usecase.LogoutUseCase
import com.revas.mobile.domain.usecase.RestoreSessionUseCase
import com.revas.mobile.platform.Clock
import com.revas.mobile.platform.MobileLogger
import io.ktor.client.HttpClient
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

private val mobileJson = Json {
    ignoreUnknownKeys = true
    explicitNulls = false
    encodeDefaults = true
}

internal fun createSdk(
    config: MobileApiConfig,
    cache: MobileCache,
    tokenStorage: TokenStorage,
    clock: Clock,
    logger: MobileLogger,
): SharedSdk {
    val client = HttpClient {
        expectSuccess = false
        install(ContentNegotiation) {
            json(mobileJson)
        }
        install(HttpTimeout) {
            requestTimeoutMillis = 15_000
            connectTimeoutMillis = 10_000
            socketTimeoutMillis = 15_000
        }
    }
    val api = MobileApi(
        client = client,
        config = config,
        tokenProvider = tokenStorage::readToken,
        json = mobileJson,
    )
    val sessionRepository = DefaultSessionRepository(api, tokenStorage, cache)
    val alertRepository = DefaultAlertRepository(api, cache, clock, mobileJson)
    val caseRepository = DefaultCaseRepository(api, cache, clock, mobileJson)
    return SharedSdk(
        api = api,
        cache = cache,
        clock = clock,
        logger = logger,
        useCases = MobileUseCases(
            login = LoginUseCase(sessionRepository),
            restore = RestoreSessionUseCase(sessionRepository),
            logout = LogoutUseCase(sessionRepository),
            hasStoredSession = HasStoredSessionUseCase(sessionRepository),
            loadAlerts = LoadAlertsUseCase(alertRepository),
            loadAlert = LoadAlertDetailUseCase(alertRepository),
            loadCase = LoadCaseDetailUseCase(caseRepository),
        ),
    )
}

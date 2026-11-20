package com.revas.mobile

import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.data.session.DefaultSessionRepository
import com.revas.mobile.data.repository.DefaultAlertRepository
import com.revas.mobile.data.repository.DefaultCaseRepository
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.model.Severity
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.usecase.HasStoredSessionUseCase
import com.revas.mobile.domain.usecase.LoadAlertDetailUseCase
import com.revas.mobile.domain.usecase.LoadAlertsUseCase
import com.revas.mobile.domain.usecase.LoadCaseDetailUseCase
import com.revas.mobile.domain.usecase.LoginUseCase
import com.revas.mobile.domain.usecase.LogoutUseCase
import com.revas.mobile.domain.usecase.Outcome
import com.revas.mobile.domain.usecase.errorOrNull
import com.revas.mobile.domain.usecase.RestoreSessionUseCase
import com.revas.mobile.platform.Clock
import com.revas.mobile.platform.MobileLogger

class MobileUseCases internal constructor(
    val login: LoginUseCase,
    val restore: RestoreSessionUseCase,
    val logout: LogoutUseCase,
    val hasStoredSession: HasStoredSessionUseCase,
    val loadAlerts: LoadAlertsUseCase,
    val loadAlert: LoadAlertDetailUseCase,
    val loadCase: LoadCaseDetailUseCase,
)

class AuthResult(val user: UserSession?, val error: MobileError?) {
    val isSuccess: Boolean get() = error == null
    val errorMessage: String? get() = error?.message
}

class AlertFeedResult(val feed: AlertFeed?, val error: MobileError?) {
    val isSuccess: Boolean get() = error == null
    val errorMessage: String? get() = error?.message
}

class AlertResult(val alert: Alert?, val error: MobileError?) {
    val isSuccess: Boolean get() = error == null
    val errorMessage: String? get() = error?.message
}

class CaseResult(val case: CaseDetail?, val error: MobileError?) {
    val isSuccess: Boolean get() = error == null
    val errorMessage: String? get() = error?.message
}

class SimpleResult(val error: MobileError?) {
    val isSuccess: Boolean get() = error == null
    val errorMessage: String? get() = error?.message
}

class SharedSdk internal constructor(
    private val api: MobileApi,
    private val cache: MobileCache,
    private val clock: Clock,
    private val logger: MobileLogger,
    val useCases: MobileUseCases,
) {
    fun hasStoredSession(): Boolean = useCases.hasStoredSession()

    suspend fun login(email: String, password: String): AuthResult =
        useCases.login(email, password).toAuthResult()

    suspend fun restoreSession(): AuthResult =
        useCases.restore().toNullableAuthResult()

    suspend fun logout(): SimpleResult {
        val result = useCases.logout()
        logger.debug("logout")
        return SimpleResult(result.errorOrNull())
    }

    suspend fun loadAlerts(
        filter: AlertFilter,
        forceRefresh: Boolean = false,
    ): AlertFeedResult = when (val result = useCases.loadAlerts(filter, forceRefresh)) {
        is Outcome.Success -> AlertFeedResult(result.value, null)
        is Outcome.Failure -> AlertFeedResult(null, result.error)
    }

    suspend fun loadAlerts(
        severityCode: String?,
        slaBreachedOnly: Boolean,
        forceRefresh: Boolean,
    ): AlertFeedResult = loadAlerts(
        filter = AlertFilter(
            severity = severityCode?.let { code ->
                Severity.entries.firstOrNull { it.name.lowercase() == code.lowercase() }
            },
            slaBreachedOnly = slaBreachedOnly,
        ),
        forceRefresh = forceRefresh,
    )

    suspend fun loadAlert(alertId: String): AlertResult =
        when (val result = useCases.loadAlert(alertId)) {
            is Outcome.Success -> AlertResult(result.value, null)
            is Outcome.Failure -> AlertResult(null, result.error)
        }

    suspend fun loadCase(caseId: String): CaseResult =
        when (val result = useCases.loadCase(caseId)) {
            is Outcome.Success -> CaseResult(result.value, null)
            is Outcome.Failure -> CaseResult(null, result.error)
        }

    fun purgeLocalData() {
        cache.purge()
    }

    fun close() {
        api.close()
    }

    private fun Outcome<UserSession?>.toNullableAuthResult(): AuthResult = when (this) {
        is Outcome.Success -> AuthResult(value, null)
        is Outcome.Failure -> AuthResult(null, error)
    }

    private fun Outcome<UserSession>.toAuthResult(): AuthResult = when (this) {
        is Outcome.Success -> AuthResult(value, null)
        is Outcome.Failure -> AuthResult(null, error)
    }
}

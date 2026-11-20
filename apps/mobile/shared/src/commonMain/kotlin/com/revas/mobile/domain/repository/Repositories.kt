package com.revas.mobile.domain.repository

import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.usecase.Outcome

interface AlertRepository {
    suspend fun loadFeed(filter: AlertFilter, forceRefresh: Boolean): Outcome<AlertFeed>
    suspend fun loadAlert(alertId: String): Outcome<Alert>
}

interface CaseRepository {
    suspend fun loadCase(caseId: String): Outcome<CaseDetail>
}

interface SessionRepository {
    suspend fun login(email: String, password: String): Outcome<UserSession>
    suspend fun restore(): Outcome<UserSession?>
    suspend fun logout(): Outcome<Unit>
    fun hasStoredToken(): Boolean
}

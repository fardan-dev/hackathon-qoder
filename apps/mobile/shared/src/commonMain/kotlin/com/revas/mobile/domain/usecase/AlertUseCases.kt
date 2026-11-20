package com.revas.mobile.domain.usecase

import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.repository.AlertRepository

class LoadAlertsUseCase(private val repository: AlertRepository) {
    suspend operator fun invoke(filter: AlertFilter, forceRefresh: Boolean = false): Outcome<AlertFeed> =
        repository.loadFeed(filter, forceRefresh)
}

class LoadAlertDetailUseCase(private val repository: AlertRepository) {
    suspend operator fun invoke(alertId: String): Outcome<Alert> {
        if (alertId.isBlank()) {
            return Outcome.Failure(MobileError.Validation("ID alert tidak valid"))
        }
        return repository.loadAlert(alertId)
    }
}

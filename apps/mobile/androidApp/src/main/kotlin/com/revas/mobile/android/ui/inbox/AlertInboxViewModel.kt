@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.android.ui.inbox

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.revas.mobile.android.ui.common.userMessage
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.Severity
import com.revas.mobile.domain.usecase.LoadAlertsUseCase
import com.revas.mobile.domain.usecase.Outcome
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlin.time.ExperimentalTime
import kotlin.time.Instant
import javax.inject.Inject

data class AlertInboxUiState(
    val loading: Boolean = true,
    val refreshing: Boolean = false,
    val alerts: List<Alert> = emptyList(),
    val filter: AlertFilter = AlertFilter(),
    val fromCache: Boolean = false,
    val lastSyncedAt: Instant? = null,
    val error: String? = null,
) {
    val isEmpty: Boolean get() = !loading && alerts.isEmpty()
}

@HiltViewModel
class AlertInboxViewModel @Inject constructor(
    private val loadAlerts: LoadAlertsUseCase,
) : ViewModel() {

    private val _state = MutableStateFlow(AlertInboxUiState())
    val state: StateFlow<AlertInboxUiState> = _state

    init {
        refresh(force = false)
    }

    fun refresh(force: Boolean = true) {
        if (_state.value.refreshing) return
        _state.update { it.copy(loading = it.alerts.isEmpty(), refreshing = true, error = null) }
        viewModelScope.launch {
            when (val outcome = loadAlerts(_state.value.filter, forceRefresh = force)) {
                is Outcome.Success -> applyFeed(outcome.value)
                is Outcome.Failure -> _state.update {
                    it.copy(loading = false, refreshing = false, error = outcome.error.userMessage())
                }
            }
        }
    }

    fun selectSeverity(severity: Severity?) {
        if (_state.value.filter.severity == severity) return
        _state.update {
            it.copy(filter = it.filter.copy(severity = severity), loading = true, refreshing = true, error = null)
        }
        refresh(force = true)
    }

    fun toggleSlaOnly() {
        _state.update {
            it.copy(filter = it.filter.copy(slaBreachedOnly = !it.filter.slaBreachedOnly), loading = true, refreshing = true, error = null)
        }
        refresh(force = true)
    }

    private fun applyFeed(feed: AlertFeed) {
        _state.update {
            it.copy(
                loading = false,
                refreshing = false,
                alerts = feed.alerts,
                fromCache = feed.fromCache,
                lastSyncedAt = feed.lastSyncedAt,
                error = null,
            )
        }
    }
}

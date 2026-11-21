package com.revas.mobile.android.ui.alert

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.revas.mobile.android.ui.common.userMessage
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.usecase.LoadAlertDetailUseCase
import com.revas.mobile.domain.usecase.Outcome
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AlertDetailUiState(
    val loading: Boolean = true,
    val alert: Alert? = null,
    val error: String? = null,
)

@HiltViewModel
class AlertDetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val loadAlert: LoadAlertDetailUseCase,
) : ViewModel() {

    private val alertId: String = savedStateHandle["alertId"] ?: ""

    private val _state = MutableStateFlow(AlertDetailUiState())
    val state: StateFlow<AlertDetailUiState> = _state

    init {
        reload()
    }

    fun reload() {
        _state.update { it.copy(loading = it.alert == null, error = null) }
        viewModelScope.launch {
            when (val outcome = loadAlert(alertId)) {
                is Outcome.Success -> _state.update {
                    it.copy(loading = false, alert = outcome.value, error = null)
                }
                is Outcome.Failure -> _state.update {
                    it.copy(loading = false, error = outcome.error.userMessage())
                }
            }
        }
    }
}

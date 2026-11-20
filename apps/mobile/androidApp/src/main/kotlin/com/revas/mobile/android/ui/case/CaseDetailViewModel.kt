package com.revas.mobile.android.ui.case

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.revas.mobile.android.ui.common.userMessage
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.usecase.LoadCaseDetailUseCase
import com.revas.mobile.domain.usecase.Outcome
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class CaseDetailUiState(
    val loading: Boolean = true,
    val case: CaseDetail? = null,
    val error: String? = null,
)

@HiltViewModel
class CaseDetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val loadCase: LoadCaseDetailUseCase,
) : ViewModel() {

    private val caseId: String = savedStateHandle["caseId"] ?: ""

    private val _state = MutableStateFlow(CaseDetailUiState())
    val state: StateFlow<CaseDetailUiState> = _state

    init {
        reload()
    }

    fun reload() {
        _state.update { it.copy(loading = it.case == null, error = null) }
        viewModelScope.launch {
            when (val outcome = loadCase(caseId)) {
                is Outcome.Success -> _state.update {
                    it.copy(loading = false, case = outcome.value, error = null)
                }
                is Outcome.Failure -> _state.update {
                    it.copy(loading = false, error = outcome.error.userMessage())
                }
            }
        }
    }
}

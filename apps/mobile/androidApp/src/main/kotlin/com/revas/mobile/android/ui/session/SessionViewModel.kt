package com.revas.mobile.android.ui.session

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.revas.mobile.domain.model.UserSession
import com.revas.mobile.domain.usecase.HasStoredSessionUseCase
import com.revas.mobile.domain.usecase.LoginUseCase
import com.revas.mobile.domain.usecase.LogoutUseCase
import com.revas.mobile.domain.usecase.Outcome
import com.revas.mobile.domain.usecase.RestoreSessionUseCase
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val email: String = "reviewer@claimshield.local",
    val password: String = "",
    val submitting: Boolean = false,
    val error: String? = null,
)

sealed interface SessionUiState {
    data object Loading : SessionUiState
    data object LoggedOut : SessionUiState
    data class LoggedIn(val user: UserSession, val locked: Boolean) : SessionUiState
}

@HiltViewModel
class SessionViewModel @Inject constructor(
    private val restoreSession: RestoreSessionUseCase,
    private val login: LoginUseCase,
    private val logout: LogoutUseCase,
    private val hasStoredSession: HasStoredSessionUseCase,
) : ViewModel() {
    private val _session = MutableStateFlow<SessionUiState>(SessionUiState.Loading)
    val session = _session.asStateFlow()

    private val _form = MutableStateFlow(LoginUiState())
    val form = _form.asStateFlow()

    init {
        viewModelScope.launch {
            when (val result = restoreSession()) {
                is Outcome.Success -> {
                    val user = result.value
                    _session.value = if (user == null) {
                        SessionUiState.LoggedOut
                    } else {
                        SessionUiState.LoggedIn(user, locked = true)
                    }
                }

                is Outcome.Failure -> {
                    _session.value = SessionUiState.LoggedOut
                    _form.update { it.copy(error = result.error.message) }
                }
            }
        }
    }

    fun onEmailChange(value: String) {
        _form.update { it.copy(email = value, error = null) }
    }

    fun onPasswordChange(value: String) {
        _form.update { it.copy(password = value, error = null) }
    }

    fun submit() {
        val current = _form.value
        if (current.submitting) return
        _form.value = current.copy(submitting = true, error = null)
        viewModelScope.launch {
            when (val result = login(current.email, current.password)) {
                is Outcome.Success -> {
                    _session.value = SessionUiState.LoggedIn(result.value, locked = false)
                    _form.value = LoginUiState(email = current.email)
                }

                is Outcome.Failure -> {
                    _form.update { it.copy(submitting = false, error = result.error.message) }
                }
            }
        }
    }

    fun unlock() {
        val current = _session.value
        if (current is SessionUiState.LoggedIn && current.locked) {
            _session.value = current.copy(locked = false)
        }
    }

    fun signOut() {
        viewModelScope.launch {
            logout()
            _form.value = LoginUiState()
            _session.value = SessionUiState.LoggedOut
        }
    }

    fun hasStoredSession(): Boolean = hasStoredSession()
}

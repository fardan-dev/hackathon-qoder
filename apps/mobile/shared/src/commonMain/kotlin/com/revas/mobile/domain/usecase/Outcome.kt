package com.revas.mobile.domain.usecase

import com.revas.mobile.domain.model.MobileError

sealed interface Outcome<out T> {
    data class Success<T>(val value: T) : Outcome<T>
    data class Failure(val error: MobileError) : Outcome<Nothing>
}

inline fun <T> Outcome<T>.errorOrNull(): MobileError? =
    (this as? Outcome.Failure)?.error

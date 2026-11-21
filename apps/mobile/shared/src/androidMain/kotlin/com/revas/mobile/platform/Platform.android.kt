@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.platform

import android.util.Log
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

actual fun platformClock(): Clock = object : Clock {
    override fun now(): Instant = Instant.fromEpochMilliseconds(System.currentTimeMillis())
}

actual fun platformLogger(): MobileLogger = object : MobileLogger {
    override fun debug(event: String) {
        Log.d("RevasMobile", event)
    }

    override fun error(event: String) {
        Log.e("RevasMobile", event)
    }
}

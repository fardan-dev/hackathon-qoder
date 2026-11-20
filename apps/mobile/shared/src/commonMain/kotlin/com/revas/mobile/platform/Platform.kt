@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.platform

import kotlin.time.ExperimentalTime
import kotlin.time.Instant

interface Clock {
    fun now(): Instant
}

expect fun platformClock(): Clock

interface MobileLogger {
    fun debug(event: String)
    fun error(event: String)
}

expect fun platformLogger(): MobileLogger

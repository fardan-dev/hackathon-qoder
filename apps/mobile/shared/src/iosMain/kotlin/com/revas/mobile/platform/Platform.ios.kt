@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.platform

import platform.Foundation.NSDate
import platform.Foundation.NSLog
import platform.Foundation.timeIntervalSince1970
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

actual fun platformClock(): Clock = object : Clock {
    override fun now(): Instant =
        Instant.fromEpochMilliseconds((NSDate().timeIntervalSince1970 * 1_000).toLong())
}

actual fun platformLogger(): MobileLogger = object : MobileLogger {
    override fun debug(event: String) {
        NSLog("%@", event)
    }

    override fun error(event: String) {
        NSLog("%@", event)
    }
}

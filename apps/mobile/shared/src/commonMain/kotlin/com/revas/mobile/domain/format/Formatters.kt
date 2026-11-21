@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.domain.format

import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

object Formatters {
    private val jakarta = TimeZone.of("Asia/Jakarta")
    private val monthNames = listOf(
        "Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
        "Jul", "Agu", "Sep", "Okt", "Nov", "Des",
    )

    fun idr(amount: Long): String = "Rp ${groupThousands(amount)}"

    fun dateTime(instant: Instant): String {
        val local = instant.toLocalDateTime(jakarta)
        val day = local.day.toString().padStart(2, '0')
        val hour = local.hour.toString().padStart(2, '0')
        val minute = local.minute.toString().padStart(2, '0')
        return "$day ${monthNames[local.month.ordinal]} ${local.year} $hour:$minute WIB"
    }

    fun date(instant: Instant): String {
        val local = instant.toLocalDateTime(jakarta)
        val day = local.day.toString().padStart(2, '0')
        return "$day ${monthNames[local.month.ordinal]} ${local.year}"
    }

    fun relativeAge(instant: Instant, now: Instant): String {
        val minutes = (now - instant).inWholeMinutes
        return when {
            minutes < 1 -> "baru saja"
            minutes < 60 -> "$minutes menit lalu"
            minutes < 1_440 -> "${minutes / 60} jam lalu"
            else -> "${minutes / 1_440} hari lalu"
        }
    }

    private fun groupThousands(amount: Long): String {
        val digits = amount.toString()
        if (digits.length <= 3) return digits
        val builder = StringBuilder()
        val firstGroup = digits.length % 3
        if (firstGroup > 0) {
            builder.append(digits, 0, firstGroup)
        }
        var index = firstGroup
        while (index < digits.length) {
            if (builder.isNotEmpty()) builder.append('.')
            builder.append(digits, index, index + 3)
            index += 3
        }
        return builder.toString()
    }
}

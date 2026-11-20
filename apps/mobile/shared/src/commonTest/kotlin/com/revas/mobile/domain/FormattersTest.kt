package com.revas.mobile.domain

import com.revas.mobile.domain.format.Formatters
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.time.Instant

class FormattersTest {
    @Test
    fun formats_idr_with_thousand_separators() {
        assertEquals("Rp 12.500.000", Formatters.idr(12_500_000))
        assertEquals("Rp 0", Formatters.idr(0))
        assertEquals("Rp 999", Formatters.idr(999))
        assertEquals("Rp 1.000", Formatters.idr(1_000))
    }

    @Test
    fun formats_datetime_in_jakarta_time() {
        val instant = Instant.parse("2026-11-20T03:05:00Z")
        assertEquals("20 Nov 2026 10:05 WIB", Formatters.dateTime(instant))
    }

    @Test
    fun formats_relative_age() {
        val now = Instant.parse("2026-11-20T12:00:00Z")
        assertEquals("baru saja", Formatters.relativeAge(Instant.parse("2026-11-20T11:59:40Z"), now))
        assertEquals("5 menit lalu", Formatters.relativeAge(Instant.parse("2026-11-20T11:55:00Z"), now))
        assertEquals("2 jam lalu", Formatters.relativeAge(Instant.parse("2026-11-20T10:00:00Z"), now))
        assertEquals("1 hari lalu", Formatters.relativeAge(Instant.parse("2026-11-19T10:00:00Z"), now))
    }
}

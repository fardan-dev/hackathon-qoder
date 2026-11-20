package com.revas.mobile.data.cache

import app.cash.sqldelight.driver.jdbc.sqlite.JdbcSqliteDriver
import com.revas.mobile.db.MobileDatabase
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class SqlDelightCacheTest {
    private fun newCache(): SqlDelightCache {
        val driver = JdbcSqliteDriver(JdbcSqliteDriver.IN_MEMORY)
        MobileDatabase.Schema.create(driver)
        return SqlDelightCache(MobileDatabase(driver))
    }

    @Test
    fun stores_and_reads_alerts() {
        val cache = newCache()
        cache.storeAlerts(
            listOf(
                CachedAlertPayload("A1", """{"alert_id":"A1"}""", 1),
                CachedAlertPayload("A2", """{"alert_id":"A2"}""", 2),
            ),
            cachedAtEpochMillis = 2,
        )

        assertEquals(2, cache.alerts().size)
        assertEquals("""{"alert_id":"A1"}""", cache.alert("A1")?.payloadJson)
        assertNull(cache.alert("missing"))
    }

    @Test
    fun stores_cases_and_sync_state() {
        val cache = newCache()
        cache.storeCase("CASE-1", """{"case_id":"CASE-1"}""", 5)
        cache.storeSyncState("alerts", "\"v1\"", 5)

        assertEquals("""{"case_id":"CASE-1"}""", cache.case("CASE-1")?.payloadJson)
        assertEquals("\"v1\"", cache.syncState("alerts")?.etag)
        assertEquals(5, cache.syncState("alerts")?.lastSyncedAtEpochMillis)
    }

    @Test
    fun purge_removes_all_cached_data() {
        val cache = newCache()
        cache.storeAlerts(listOf(CachedAlertPayload("A1", "{}", 1)), 1)
        cache.storeCase("CASE-1", "{}", 1)
        cache.storeSyncState("alerts", "v1", 1)

        cache.purge()

        assertTrue(cache.alerts().isEmpty())
        assertNull(cache.case("CASE-1"))
        assertNull(cache.syncState("alerts"))
    }
}

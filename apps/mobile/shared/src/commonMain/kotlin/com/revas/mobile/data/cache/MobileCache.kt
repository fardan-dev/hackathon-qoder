package com.revas.mobile.data.cache

import com.revas.mobile.db.MobileDatabase

data class CachedAlertPayload(
    val alertId: String,
    val payloadJson: String,
    val cachedAtEpochMillis: Long,
)

data class CachedCasePayload(
    val caseId: String,
    val payloadJson: String,
    val cachedAtEpochMillis: Long,
)

data class CachedSyncState(
    val etag: String?,
    val lastSyncedAtEpochMillis: Long,
)

interface MobileCache {
    fun storeAlerts(items: List<CachedAlertPayload>, cachedAtEpochMillis: Long)
    fun alerts(): List<CachedAlertPayload>
    fun alert(alertId: String): CachedAlertPayload?
    fun storeCase(caseId: String, payloadJson: String, cachedAtEpochMillis: Long)
    fun case(caseId: String): CachedCasePayload?
    fun storeSyncState(cacheKey: String, etag: String?, syncedAtEpochMillis: Long)
    fun syncState(cacheKey: String): CachedSyncState?
    fun purge()
}

class SqlDelightCache(private val database: MobileDatabase) : MobileCache {
    private val queries get() = database.cacheQueries

    override fun storeAlerts(items: List<CachedAlertPayload>, cachedAtEpochMillis: Long) {
        database.transaction {
            items.forEach { item ->
                queries.upsertAlert(item.alertId, item.payloadJson, cachedAtEpochMillis)
            }
            queries.trimAlerts()
        }
    }

    override fun alerts(): List<CachedAlertPayload> =
        queries.selectAlerts().executeAsList().map {
            CachedAlertPayload(it.alertId, it.payloadJson, it.cachedAt)
        }

    override fun alert(alertId: String): CachedAlertPayload? =
        queries.selectAlertById(alertId).executeAsOneOrNull()?.let {
            CachedAlertPayload(it.alertId, it.payloadJson, it.cachedAt)
        }

    override fun storeCase(caseId: String, payloadJson: String, cachedAtEpochMillis: Long) {
        queries.upsertCase(caseId, payloadJson, cachedAtEpochMillis)
    }

    override fun case(caseId: String): CachedCasePayload? =
        queries.selectCaseById(caseId).executeAsOneOrNull()?.let {
            CachedCasePayload(it.caseId, it.payloadJson, it.cachedAt)
        }

    override fun storeSyncState(cacheKey: String, etag: String?, syncedAtEpochMillis: Long) {
        queries.upsertSyncState(cacheKey, etag, syncedAtEpochMillis)
    }

    override fun syncState(cacheKey: String): CachedSyncState? =
        queries.selectSyncState(cacheKey).executeAsOneOrNull()?.let {
            CachedSyncState(it.etag, it.lastSyncedAt)
        }

    override fun purge() {
        database.transaction {
            queries.deleteAlerts()
            queries.deleteCases()
            queries.deleteSyncState()
        }
    }
}

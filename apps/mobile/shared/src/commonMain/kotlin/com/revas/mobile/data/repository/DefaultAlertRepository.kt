@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.data.repository

import com.revas.mobile.data.cache.CachedAlertPayload
import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.mapper.toDomain
import com.revas.mobile.data.remote.AlertDto
import com.revas.mobile.data.remote.FetchResult
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFeed
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.repository.AlertRepository
import com.revas.mobile.domain.usecase.Outcome
import com.revas.mobile.platform.Clock
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

internal class DefaultAlertRepository(
    private val api: MobileApi,
    private val cache: MobileCache,
    private val clock: Clock,
    private val json: Json,
) : AlertRepository {
    override suspend fun loadFeed(filter: AlertFilter, forceRefresh: Boolean): Outcome<AlertFeed> {
        val cachedSync = cache.syncState(SYNC_KEY)
        val etag = if (!forceRefresh && filter.isEmpty) cachedSync?.etag else null
        return when (val result = api.listAlerts(filter, etag)) {
            is FetchResult.Success -> {
                val now = clock.now()
                val nowMillis = now.toEpochMilliseconds()
                val items = result.value.alerts.map {
                    CachedAlertPayload(it.alertId, json.encodeToString(it), nowMillis)
                }
                cache.storeAlerts(items, nowMillis)
                if (filter.isEmpty) {
                    cache.storeSyncState(SYNC_KEY, result.etag, nowMillis)
                }
                Outcome.Success(
                    AlertFeed(
                        alerts = result.value.alerts.map { it.toDomain() },
                        lastSyncedAt = now,
                        fromCache = false,
                    ),
                )
            }

            FetchResult.NotModified -> cachedFeed(filter, null)
            is FetchResult.Failure -> cachedFeed(filter, result.error)
        }
    }

    override suspend fun loadAlert(alertId: String): Outcome<Alert> {
        return when (val result = api.getAlert(alertId)) {
            is FetchResult.Success -> {
                cache.storeAlerts(
                    listOf(
                        CachedAlertPayload(
                            alertId = result.value.alertId,
                            payloadJson = json.encodeToString(result.value),
                            cachedAtEpochMillis = clock.now().toEpochMilliseconds(),
                        ),
                    ),
                    clock.now().toEpochMilliseconds(),
                )
                Outcome.Success(result.value.toDomain())
            }

            FetchResult.NotModified -> Outcome.Failure(MobileError.Unknown("Respons tidak valid"))
            is FetchResult.Failure -> {
                val cached = cache.alert(alertId)
                    ?.let { runCatching { json.decodeFromString<AlertDto>(it.payloadJson) }.getOrNull() }
                if (cached != null) Outcome.Success(cached.toDomain())
                else Outcome.Failure(result.error)
            }
        }
    }

    private fun cachedFeed(filter: AlertFilter, error: MobileError?): Outcome<AlertFeed> {
        val cached = cache.alerts()
            .mapNotNull { runCatching { json.decodeFromString<AlertDto>(it.payloadJson).toDomain() }.getOrNull() }
            .filter { filter.matches(it) }
        if (cached.isEmpty()) {
            return Outcome.Failure(error ?: MobileError.Network("Belum ada data alert tersimpan"))
        }
        val lastSyncedAt = cache.syncState(SYNC_KEY)?.lastSyncedAtEpochMillis
            ?.let { Instant.fromEpochMilliseconds(it) }
        return Outcome.Success(AlertFeed(cached, lastSyncedAt, fromCache = true))
    }

    private companion object {
        const val SYNC_KEY = "alerts"
    }
}

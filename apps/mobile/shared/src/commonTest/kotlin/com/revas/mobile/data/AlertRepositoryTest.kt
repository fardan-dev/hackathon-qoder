@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.data

import com.revas.mobile.data.cache.CachedAlertPayload
import com.revas.mobile.data.cache.CachedCasePayload
import com.revas.mobile.data.cache.CachedSyncState
import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.data.remote.MobileApiConfig
import com.revas.mobile.data.repository.DefaultAlertRepository
import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.ErrorKind
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.Severity
import com.revas.mobile.domain.usecase.Outcome
import com.revas.mobile.platform.Clock
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.MockRequestHandler
import io.ktor.client.engine.mock.respond
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

private val testJson = Json {
    ignoreUnknownKeys = true
    explicitNulls = false
}

private class InMemoryCache : MobileCache {
    val alertStore = linkedMapOf<String, CachedAlertPayload>()
    val caseStore = linkedMapOf<String, CachedCasePayload>()
    val syncStore = linkedMapOf<String, CachedSyncState>()
    var purged = false

    override fun storeAlerts(items: List<CachedAlertPayload>, cachedAtEpochMillis: Long) {
        items.forEach { alertStore[it.alertId] = it }
    }

    override fun alerts(): List<CachedAlertPayload> = alertStore.values.toList()
    override fun alert(alertId: String): CachedAlertPayload? = alertStore[alertId]

    override fun storeCase(caseId: String, payloadJson: String, cachedAtEpochMillis: Long) {
        caseStore[caseId] = CachedCasePayload(caseId, payloadJson, cachedAtEpochMillis)
    }

    override fun case(caseId: String): CachedCasePayload? = caseStore[caseId]

    override fun storeSyncState(cacheKey: String, etag: String?, syncedAtEpochMillis: Long) {
        syncStore[cacheKey] = CachedSyncState(etag, syncedAtEpochMillis)
    }

    override fun syncState(cacheKey: String): CachedSyncState? = syncStore[cacheKey]

    override fun purge() {
        purged = true
        alertStore.clear()
        caseStore.clear()
        syncStore.clear()
    }
}

private class FixedClock(private val instant: Instant) : Clock {
    override fun now(): Instant = instant
}

private fun api(handler: MockRequestHandler): MobileApi {
    val client = HttpClient(MockEngine(handler)) {
        expectSuccess = false
        install(ContentNegotiation) { json(testJson) }
    }
    return MobileApi(
        client = client,
        config = MobileApiConfig(
            identityBaseUrl = "https://identity.test",
            alertBaseUrl = "https://alert.test",
            caseBaseUrl = "https://case.test",
        ),
        tokenProvider = { "test-token" },
        json = testJson,
    )
}

private const val ALERT_LIST_JSON = """
{
  "alerts": [
    {
      "alert_id": "A1",
      "claim_id": "C1",
      "loss_type": "MOTOR",
      "severity": "High",
      "severity_rank": 1,
      "flag_count": 1,
      "claimed_amount": 12500000,
      "flags": [
        {
          "flag_id": "F1",
          "rule_id": "A1",
          "rule_version": 1,
          "rule_name": "Double claim",
          "family": "A",
          "severity": "High",
          "params_version": 1,
          "trigger_value": "2",
          "evidence_refs": ["claim:C1"],
          "explanation": "Indikasi klaim ganda",
          "reason_code": "RC1",
          "suppressed": false
        }
      ],
      "status": "OPEN",
      "suppressed": false,
      "sla_due_at": "2026-11-21T10:00:00Z",
      "sla_breached": false,
      "created_at": "2026-11-20T10:00:00Z"
    }
  ],
  "count": 1
}
"""

class AlertRepositoryTest {
    private val fixedNow = Instant.parse("2026-11-20T12:00:00Z")

    @Test
    fun maps_alert_response_to_domain() = runTest {
        val repository = DefaultAlertRepository(
            api = api { respond(ALERT_LIST_JSON, HttpStatusCode.OK, jsonHeaders()) },
            cache = InMemoryCache(),
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        val result = repository.loadFeed(AlertFilter(), forceRefresh = true)

        val feed = assertIs<Outcome.Success<*>>(result).value as com.revas.mobile.domain.model.AlertFeed
        assertEquals(1, feed.alerts.size)
        assertEquals(Severity.High, feed.alerts.single().severity)
        assertEquals(LossType.Motor, feed.alerts.single().lossType)
        assertEquals("Indikasi klaim ganda", feed.alerts.single().flags.single().explanation)
        assertEquals(false, feed.fromCache)
    }

    @Test
    fun maps_unauthorized_response() = runTest {
        val repository = DefaultAlertRepository(
            api = api {
                respond(
                    """{"code":"UNAUTHENTICATED","detail":"session required"}""",
                    HttpStatusCode.Unauthorized,
                    jsonHeaders(),
                )
            },
            cache = InMemoryCache(),
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        val result = repository.loadFeed(AlertFilter(), forceRefresh = true)

        val failure = assertIs<Outcome.Failure>(result)
        assertEquals(ErrorKind.Unauthorized, failure.error.kind)
    }

    @Test
    fun sends_etag_and_uses_cache_on_not_modified() = runTest {
        var etagSent: String? = null
        val cache = InMemoryCache()
        val repository = DefaultAlertRepository(
            api = api { request ->
                etagSent = request.headers[HttpHeaders.IfNoneMatch]
                if (etagSent != null) {
                    respond("", HttpStatusCode.NotModified)
                } else {
                    respond(
                        ALERT_LIST_JSON,
                        HttpStatusCode.OK,
                        headersOf(
                            HttpHeaders.ContentType to listOf(ContentType.Application.Json.toString()),
                            HttpHeaders.ETag to listOf("\"v1\""),
                        ),
                    )
                }
            },
            cache = cache,
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        repository.loadFeed(AlertFilter(), forceRefresh = false)
        val second = repository.loadFeed(AlertFilter(), forceRefresh = false)

        assertEquals("\"v1\"", etagSent)
        val feed = assertIs<Outcome.Success<*>>(second).value as com.revas.mobile.domain.model.AlertFeed
        assertTrue(feed.fromCache)
        assertEquals(1, feed.alerts.size)
    }

    @Test
    fun falls_back_to_cache_when_network_fails() = runTest {
        var offline = false
        val cache = InMemoryCache()
        val repository = DefaultAlertRepository(
            api = api {
                if (offline) throw RuntimeException("offline")
                respond(ALERT_LIST_JSON, HttpStatusCode.OK, jsonHeaders())
            },
            cache = cache,
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        repository.loadFeed(AlertFilter(), forceRefresh = true)
        offline = true
        val result = repository.loadFeed(AlertFilter(), forceRefresh = true)

        val feed = assertIs<Outcome.Success<*>>(result).value as com.revas.mobile.domain.model.AlertFeed
        assertTrue(feed.fromCache)
        assertEquals(1, feed.alerts.size)
    }

    @Test
    fun returns_failure_when_offline_and_cache_empty() = runTest {
        val repository = DefaultAlertRepository(
            api = api { throw RuntimeException("offline") },
            cache = InMemoryCache(),
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        val failure = assertIs<Outcome.Failure>(repository.loadFeed(AlertFilter(), true))
        assertEquals(ErrorKind.Network, failure.error.kind)
    }

    @Test
    fun filters_cached_alerts_locally() = runTest {
        var offline = false
        val cache = InMemoryCache()
        val repository = DefaultAlertRepository(
            api = api {
                if (offline) throw RuntimeException("offline")
                respond(ALERT_LIST_JSON, HttpStatusCode.OK, jsonHeaders())
            },
            cache = cache,
            clock = FixedClock(fixedNow),
            json = testJson,
        )

        repository.loadFeed(AlertFilter(), forceRefresh = true)
        offline = true

        val filtered = repository.loadFeed(AlertFilter(severity = Severity.Low), true)
        val failure = assertIs<Outcome.Failure>(filtered)
        assertEquals(ErrorKind.Network, failure.error.kind)
    }

    private fun jsonHeaders() = headersOf(
        HttpHeaders.ContentType to listOf(ContentType.Application.Json.toString()),
    )
}

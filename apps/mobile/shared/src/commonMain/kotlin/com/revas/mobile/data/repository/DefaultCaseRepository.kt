@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.data.repository

import com.revas.mobile.data.cache.MobileCache
import com.revas.mobile.data.mapper.toDomain
import com.revas.mobile.data.remote.CaseDto
import com.revas.mobile.data.remote.FetchResult
import com.revas.mobile.data.remote.MobileApi
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.repository.CaseRepository
import com.revas.mobile.domain.usecase.Outcome
import com.revas.mobile.platform.Clock
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.time.ExperimentalTime

internal class DefaultCaseRepository(
    private val api: MobileApi,
    private val cache: MobileCache,
    private val clock: Clock,
    private val json: Json,
) : CaseRepository {
    override suspend fun loadCase(caseId: String): Outcome<CaseDetail> {
        return when (val result = api.getCase(caseId)) {
            is FetchResult.Success -> {
                cache.storeCase(
                    result.value.caseId,
                    json.encodeToString(result.value),
                    clock.now().toEpochMilliseconds(),
                )
                Outcome.Success(result.value.toDomain())
            }

            FetchResult.NotModified -> Outcome.Failure(MobileError.Unknown("Respons tidak valid"))
            is FetchResult.Failure -> {
                val cached = cache.case(caseId)
                    ?.let { runCatching { json.decodeFromString<CaseDto>(it.payloadJson) }.getOrNull() }
                if (cached != null) Outcome.Success(cached.toDomain())
                else Outcome.Failure(result.error)
            }
        }
    }
}

@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.data.remote

import com.revas.mobile.domain.model.AlertFilter
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.MobileError
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.bearerAuth
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.parameter
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
import kotlin.coroutines.cancellation.CancellationException
import kotlin.time.ExperimentalTime

data class MobileApiConfig(
    val identityBaseUrl: String,
    val alertBaseUrl: String,
    val caseBaseUrl: String,
)

internal sealed interface FetchResult<out T> {
    data class Success<T>(val value: T, val etag: String?) : FetchResult<T>
    data object NotModified : FetchResult<Nothing>
    data class Failure(val error: MobileError) : FetchResult<Nothing>
}

internal class MobileApi(
    private val client: HttpClient,
    private val config: MobileApiConfig,
    private val tokenProvider: () -> String?,
    private val json: Json,
) {
    suspend fun login(email: String, password: String): FetchResult<LoginResponseDto> = call {
        val response = client.post("${config.identityBaseUrl}/v1/auth/login") {
            contentType(ContentType.Application.Json)
            setBody(LoginRequestDto(email, password))
        }
        if (!response.status.isSuccess()) return@call FetchResult.Failure(response.toProblem())
        FetchResult.Success(response.body<LoginResponseDto>(), null)
    }

    suspend fun currentUser(): FetchResult<UserDto> = call {
        val response = client.get("${config.identityBaseUrl}/v1/me") {
            authorize()
        }
        if (!response.status.isSuccess()) return@call FetchResult.Failure(response.toProblem())
        FetchResult.Success(response.body<UserDto>(), null)
    }

    suspend fun logout(): FetchResult<Unit> = call {
        val response = client.post("${config.identityBaseUrl}/v1/auth/logout") {
            authorize()
        }
        if (!response.status.isSuccess() && response.status != HttpStatusCode.Unauthorized) {
            return@call FetchResult.Failure(response.toProblem())
        }
        FetchResult.Success(Unit, null)
    }

    suspend fun listAlerts(filter: AlertFilter, etag: String?): FetchResult<AlertListResponseDto> = call {
        val response = client.get("${config.alertBaseUrl}/v1/alerts") {
            authorize()
            filter.status?.let { parameter("status", it.name) }
            filter.severity?.let { parameter("severity", it.name) }
            filter.lossType?.takeIf { it != LossType.Unknown }?.let { parameter("lob", it.name.uppercase()) }
            if (filter.slaBreachedOnly) parameter("sla", "breached")
            parameter("limit", 100)
            if (etag != null) header(HttpHeaders.IfNoneMatch, etag)
        }
        when {
            response.status == HttpStatusCode.NotModified -> FetchResult.NotModified
            response.status.isSuccess() ->
                FetchResult.Success(response.body<AlertListResponseDto>(), response.headers[HttpHeaders.ETag])
            else -> FetchResult.Failure(response.toProblem())
        }
    }

    suspend fun getAlert(alertId: String): FetchResult<AlertDto> = call {
        val response = client.get("${config.alertBaseUrl}/v1/alerts/$alertId") {
            authorize()
        }
        if (!response.status.isSuccess()) return@call FetchResult.Failure(response.toProblem())
        FetchResult.Success(response.body<AlertDto>(), null)
    }

    suspend fun getCase(caseId: String): FetchResult<CaseDto> = call {
        val response = client.get("${config.caseBaseUrl}/v1/cases/$caseId") {
            authorize()
        }
        if (!response.status.isSuccess()) return@call FetchResult.Failure(response.toProblem())
        FetchResult.Success(response.body<CaseDto>(), null)
    }

    fun close() = client.close()

    private fun io.ktor.client.request.HttpRequestBuilder.authorize() {
        tokenProvider()?.takeIf { it.isNotBlank() }?.let { bearerAuth(it) }
    }

    private suspend fun HttpResponse.toProblem(): MobileError {
        val problem = runCatching {
            json.decodeFromString<ProblemDto>(bodyAsText())
        }.getOrNull()
        return MobileError.Http(
            status = status.value,
            code = problem?.code,
            message = problem?.detail ?: "Permintaan gagal (${status.value})",
        )
    }

    private suspend fun <T> call(block: suspend () -> FetchResult<T>): FetchResult<T> = try {
        block()
    } catch (e: CancellationException) {
        throw e
    } catch (e: MobileError) {
        FetchResult.Failure(e)
    } catch (e: SerializationException) {
        FetchResult.Failure(MobileError.Unknown("Respons server tidak valid"))
    } catch (e: Exception) {
        FetchResult.Failure(MobileError.Network())
    }
}

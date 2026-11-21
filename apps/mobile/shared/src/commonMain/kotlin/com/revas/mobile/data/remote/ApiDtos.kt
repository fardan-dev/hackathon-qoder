package com.revas.mobile.data.remote

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
internal data class LoginRequestDto(
    val email: String,
    val password: String,
)

@Serializable
internal data class LoginResponseDto(
    val token: String,
    val user: UserDto,
    @SerialName("expires_at") val expiresAt: String? = null,
)

@Serializable
internal data class UserDto(
    @SerialName("user_id") val userId: String,
    @SerialName("tenant_id") val tenantId: String,
    val email: String,
    @SerialName("display_name") val displayName: String,
    val role: String,
)

@Serializable
internal data class ProblemDto(
    val code: String? = null,
    val detail: String? = null,
    val status: Int? = null,
    val trace_id: String? = null,
)

@Serializable
internal data class AlertListResponseDto(
    val alerts: List<AlertDto> = emptyList(),
    val count: Int = 0,
)

@Serializable
internal data class AlertDto(
    @SerialName("alert_id") val alertId: String,
    @SerialName("claim_id") val claimId: String,
    @SerialName("loss_type") val lossType: String = "",
    val severity: String = "Low",
    @SerialName("severity_rank") val severityRank: Int = 3,
    @SerialName("flag_count") val flagCount: Int = 0,
    @SerialName("claimed_amount") val claimedAmount: Long = 0,
    val flags: List<FlagDto> = emptyList(),
    val status: String = "OPEN",
    val suppressed: Boolean = false,
    @SerialName("sla_due_at") val slaDueAt: String = "",
    @SerialName("sla_breached") val slaBreached: Boolean = false,
    @SerialName("created_at") val createdAt: String = "",
)

@Serializable
internal data class FlagDto(
    @SerialName("flag_id") val flagId: String = "",
    @SerialName("rule_id") val ruleId: String = "",
    @SerialName("rule_version") val ruleVersion: Int = 0,
    @SerialName("rule_name") val ruleName: String = "",
    val family: String = "",
    val severity: String = "Low",
    @SerialName("params_version") val paramsVersion: Int = 0,
    @SerialName("trigger_value") val triggerValue: String = "",
    @SerialName("evidence_refs") val evidenceRefs: List<String> = emptyList(),
    val explanation: String = "",
    @SerialName("reason_code") val reasonCode: String = "",
    val suppressed: Boolean = false,
)

@Serializable
internal data class CaseDto(
    @SerialName("case_id") val caseId: String,
    val title: String = "",
    val summary: String = "",
    val status: String = "",
    val priority: Int = 0,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    @SerialName("claim_count") val claimCount: Int = 0,
    val claims: List<CaseClaimDto> = emptyList(),
    val evidence: List<EvidenceDto> = emptyList(),
    val decisions: List<DecisionDto> = emptyList(),
)

@Serializable
internal data class CaseClaimDto(
    @SerialName("claim_id") val claimId: String = "",
    @SerialName("alert_id") val alertId: String = "",
    @SerialName("loss_type") val lossType: String = "",
    val severity: String = "Low",
    @SerialName("claimed_amount") val claimedAmount: Long = 0,
    @SerialName("priority_rank") val priorityRank: Int = 0,
    val flags: List<FlagDto> = emptyList(),
)

@Serializable
internal data class EvidenceDto(
    @SerialName("evidence_id") val evidenceId: String = "",
    @SerialName("artifact_type") val artifactType: String = "",
    val source: String = "",
    @SerialName("content_sha256") val contentSha256: String = "",
    @SerialName("chain_of_custody") val chainOfCustody: JsonElement? = null,
    val notes: String = "",
    @SerialName("created_at") val createdAt: String = "",
)

@Serializable
internal data class DecisionDto(
    @SerialName("decision_id") val decisionId: String = "",
    val outcome: String = "",
    @SerialName("reason_code") val reasonCode: String = "",
    val notes: String = "",
    @SerialName("decided_at") val decidedAt: String = "",
)

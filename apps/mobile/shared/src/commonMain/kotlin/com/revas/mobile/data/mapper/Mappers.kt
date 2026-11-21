@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.data.mapper

import com.revas.mobile.data.remote.AlertDto
import com.revas.mobile.data.remote.CaseDto
import com.revas.mobile.data.remote.FlagDto
import com.revas.mobile.data.remote.UserDto
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFlag
import com.revas.mobile.domain.model.AlertStatus
import com.revas.mobile.domain.model.CaseClaim
import com.revas.mobile.domain.model.CaseDecision
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.EvidenceItem
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.Severity
import com.revas.mobile.domain.model.UserSession
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

private val unknownInstant = Instant.fromEpochMilliseconds(0)

internal fun String.toInstantOrUnknown(): Instant =
    runCatching { Instant.parse(this) }.getOrDefault(unknownInstant)

internal fun FlagDto.toDomain(): AlertFlag = AlertFlag(
    flagId = flagId.ifBlank { ruleId },
    ruleId = ruleId,
    ruleVersion = ruleVersion,
    ruleName = ruleName,
    family = family,
    severity = Severity.fromWire(severity),
    paramsVersion = paramsVersion,
    triggerValue = triggerValue,
    explanation = explanation,
    reasonCode = reasonCode,
    evidenceRefs = evidenceRefs,
    suppressed = suppressed,
)

internal fun AlertDto.toDomain(): Alert = Alert(
    alertId = alertId,
    claimId = claimId,
    lossType = LossType.fromWire(lossType),
    severity = Severity.fromWire(severity),
    severityRank = severityRank,
    flagCount = flagCount,
    claimedAmount = claimedAmount,
    flags = flags.map { it.toDomain() },
    status = AlertStatus.fromWire(status),
    suppressed = suppressed,
    slaDueAt = slaDueAt.toInstantOrUnknown(),
    slaBreached = slaBreached,
    createdAt = createdAt.toInstantOrUnknown(),
)

internal fun CaseDto.toDomain(): CaseDetail = CaseDetail(
    caseId = caseId,
    title = title,
    summary = summary,
    status = status,
    priority = priority,
    createdAt = createdAt.toInstantOrUnknown(),
    updatedAt = updatedAt.toInstantOrUnknown(),
    claimCount = claimCount,
    claims = claims.map {
        CaseClaim(
            claimId = it.claimId,
            alertId = it.alertId,
            lossType = LossType.fromWire(it.lossType),
            severity = Severity.fromWire(it.severity),
            claimedAmount = it.claimedAmount,
            priorityRank = it.priorityRank,
            flags = it.flags.map(FlagDto::toDomain),
        )
    },
    evidence = evidence.map {
        EvidenceItem(
            evidenceId = it.evidenceId,
            artifactType = it.artifactType,
            source = it.source,
            contentSha256 = it.contentSha256,
            notes = it.notes.ifBlank { null },
            createdAt = it.createdAt.toInstantOrUnknown(),
        )
    },
    decisions = decisions.map {
        CaseDecision(
            decisionId = it.decisionId,
            outcome = it.outcome,
            reasonCode = it.reasonCode,
            notes = it.notes,
            decidedAt = it.decidedAt.toInstantOrUnknown(),
        )
    },
)

internal fun UserDto.toDomain(expiresAt: String?): UserSession = UserSession(
    userId = userId,
    tenantId = tenantId,
    email = email,
    displayName = displayName,
    role = role,
    expiresAt = expiresAt?.let { runCatching { Instant.parse(it) }.getOrNull() },
)

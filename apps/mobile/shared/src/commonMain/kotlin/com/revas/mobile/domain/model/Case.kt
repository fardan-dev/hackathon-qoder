@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.domain.model

import kotlin.time.ExperimentalTime
import kotlin.time.Instant

data class CaseClaim(
    val claimId: String,
    val alertId: String,
    val lossType: LossType,
    val severity: Severity,
    val claimedAmount: Long,
    val priorityRank: Int,
    val flags: List<AlertFlag>,
)

data class EvidenceItem(
    val evidenceId: String,
    val artifactType: String,
    val source: String,
    val contentSha256: String,
    val notes: String?,
    val createdAt: Instant,
)

data class CaseDecision(
    val decisionId: String,
    val outcome: String,
    val reasonCode: String,
    val notes: String,
    val decidedAt: Instant,
)

data class CaseDetail(
    val caseId: String,
    val title: String,
    val summary: String,
    val status: String,
    val priority: Int,
    val createdAt: Instant,
    val updatedAt: Instant,
    val claimCount: Int,
    val claims: List<CaseClaim>,
    val evidence: List<EvidenceItem>,
    val decisions: List<CaseDecision>,
)

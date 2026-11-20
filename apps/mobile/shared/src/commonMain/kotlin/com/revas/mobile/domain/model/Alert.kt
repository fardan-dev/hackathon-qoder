@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.domain.model

import kotlin.time.ExperimentalTime
import kotlin.time.Instant

data class AlertFlag(
    val flagId: String,
    val ruleId: String,
    val ruleVersion: Int,
    val ruleName: String,
    val family: String,
    val severity: Severity,
    val paramsVersion: Int,
    val triggerValue: String,
    val explanation: String,
    val reasonCode: String,
    val evidenceRefs: List<String>,
    val suppressed: Boolean,
)

data class Alert(
    val alertId: String,
    val claimId: String,
    val lossType: LossType,
    val severity: Severity,
    val severityRank: Int,
    val flagCount: Int,
    val claimedAmount: Long,
    val flags: List<AlertFlag>,
    val status: AlertStatus,
    val suppressed: Boolean,
    val slaDueAt: Instant,
    val slaBreached: Boolean,
    val createdAt: Instant,
)

data class AlertFilter(
    val status: AlertStatus? = null,
    val severity: Severity? = null,
    val lossType: LossType? = null,
    val slaBreachedOnly: Boolean = false,
) {
    val isEmpty: Boolean
        get() = status == null && severity == null && lossType == null && !slaBreachedOnly

    fun matches(alert: Alert): Boolean =
        (status == null || alert.status == status) &&
            (severity == null || alert.severity == severity) &&
            (lossType == null || alert.lossType == lossType) &&
            (!slaBreachedOnly || alert.slaBreached)
}

data class AlertFeed(
    val alerts: List<Alert>,
    val lastSyncedAt: Instant?,
    val fromCache: Boolean,
)

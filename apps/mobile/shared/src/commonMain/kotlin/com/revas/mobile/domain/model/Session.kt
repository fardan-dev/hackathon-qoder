@file:OptIn(ExperimentalTime::class)

package com.revas.mobile.domain.model

import kotlin.time.ExperimentalTime
import kotlin.time.Instant

data class UserSession(
    val userId: String,
    val tenantId: String,
    val email: String,
    val displayName: String,
    val role: String,
    val expiresAt: Instant?,
)

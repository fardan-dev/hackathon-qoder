package com.revas.mobile.domain.model

enum class Severity {
    High,
    Medium,
    Low;

    companion object {
        fun fromWire(raw: String?): Severity = when (raw?.trim()?.lowercase()) {
            "high" -> High
            "medium" -> Medium
            else -> Low
        }
    }
}

enum class AlertStatus {
    OPEN,
    CLOSED;

    companion object {
        fun fromWire(raw: String?): AlertStatus =
            if (raw?.trim()?.uppercase() == "CLOSED") CLOSED else OPEN
    }
}

enum class LossType {
    Motor,
    Property,
    Unknown;

    companion object {
        fun fromWire(raw: String?): LossType = when (raw?.trim()?.uppercase()) {
            "MOTOR" -> Motor
            "PROPERTY" -> Property
            else -> Unknown
        }
    }
}

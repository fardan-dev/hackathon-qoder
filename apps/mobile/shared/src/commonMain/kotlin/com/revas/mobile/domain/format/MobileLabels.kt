package com.revas.mobile.domain.format

import com.revas.mobile.domain.model.AlertStatus
import com.revas.mobile.domain.model.LossType
import com.revas.mobile.domain.model.Severity

object MobileLabels {
    fun severity(severity: Severity): String = when (severity) {
        Severity.High -> "Tinggi"
        Severity.Medium -> "Sedang"
        Severity.Low -> "Rendah"
    }

    fun lossType(lossType: LossType): String = when (lossType) {
        LossType.Motor -> "Motor"
        LossType.Property -> "Properti"
        LossType.Unknown -> "Lainnya"
    }

    fun status(status: AlertStatus): String = when (status) {
        AlertStatus.OPEN -> "Terbuka"
        AlertStatus.CLOSED -> "Selesai"
    }

    fun severityCode(severity: Severity): String = severity.name.lowercase()
}

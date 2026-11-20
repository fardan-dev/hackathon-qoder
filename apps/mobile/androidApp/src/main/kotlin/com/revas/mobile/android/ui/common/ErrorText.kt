package com.revas.mobile.android.ui.common

import com.revas.mobile.domain.model.ErrorKind
import com.revas.mobile.domain.model.MobileError

fun MobileError.userMessage(): String = when (kind) {
    ErrorKind.Network -> "Tidak dapat terhubung ke server. Periksa koneksi lalu coba lagi."
    ErrorKind.Unauthorized -> "Sesi berakhir. Silakan masuk kembali."
    ErrorKind.Forbidden -> "Anda tidak memiliki akses ke data ini."
    ErrorKind.NotFound -> "Data tidak ditemukan."
    ErrorKind.Conflict -> "Data sudah berubah. Muat ulang."
    ErrorKind.Validation -> message ?: "Data yang dikirim tidak valid."
    ErrorKind.Server -> "Server sedang bermasalah. Coba lagi beberapa saat."
    ErrorKind.Unknown -> message ?: "Terjadi kesalahan tak terduga."
}

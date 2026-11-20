package com.revas.mobile.domain.model

enum class ErrorKind {
    Network,
    Unauthorized,
    Forbidden,
    NotFound,
    Conflict,
    Validation,
    Server,
    Unknown,
}

sealed class MobileError(message: String) : Exception(message) {
    abstract val kind: ErrorKind

    class Network(message: String = "Tidak ada koneksi ke server") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Network
    }

    class Unauthorized(message: String = "Sesi berakhir, silakan masuk kembali") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Unauthorized
    }

    class Forbidden(message: String = "Anda tidak memiliki akses") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Forbidden
    }

    class NotFound(message: String = "Data tidak ditemukan") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.NotFound
    }

    class Conflict(message: String = "Data sudah berubah, muat ulang") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Conflict
    }

    class Validation(message: String) : MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Validation
    }

    class Server(message: String = "Server sedang bermasalah") :
        MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Server
    }

    class Unknown(message: String = "Terjadi kesalahan") : MobileError(message) {
        override val kind: ErrorKind = ErrorKind.Unknown
    }

    class Http(
        val status: Int,
        val code: String?,
        message: String,
    ) : MobileError(message) {
        override val kind: ErrorKind = when (status) {
            401 -> ErrorKind.Unauthorized
            403 -> ErrorKind.Forbidden
            404 -> ErrorKind.NotFound
            409 -> ErrorKind.Conflict
            400, 422 -> ErrorKind.Validation
            429 -> ErrorKind.Server
            in 500..599 -> ErrorKind.Server
            else -> ErrorKind.Unknown
        }
    }
}

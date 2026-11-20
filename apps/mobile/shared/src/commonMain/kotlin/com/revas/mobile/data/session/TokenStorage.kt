package com.revas.mobile.data.session

interface TokenStorage {
    fun readToken(): String?
    fun writeToken(token: String)
    fun clearToken()
}

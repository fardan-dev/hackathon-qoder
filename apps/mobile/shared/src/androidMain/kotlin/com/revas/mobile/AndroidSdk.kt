package com.revas.mobile

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import app.cash.sqldelight.driver.android.AndroidSqliteDriver
import com.revas.mobile.data.cache.SqlDelightCache
import com.revas.mobile.data.remote.MobileApiConfig
import com.revas.mobile.data.session.TokenStorage
import com.revas.mobile.db.MobileDatabase
import com.revas.mobile.platform.platformClock
import com.revas.mobile.platform.platformLogger
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

private const val DATABASE_NAME = "revas-mobile.db"

fun createAndroidSdk(context: Context, config: MobileApiConfig): SharedSdk {
    val database = MobileDatabase(
        AndroidSqliteDriver(MobileDatabase.Schema, context, DATABASE_NAME),
    )
    return createSdk(
        config = config,
        cache = SqlDelightCache(database),
        tokenStorage = AndroidKeystoreTokenStorage(context),
        clock = platformClock(),
        logger = platformLogger(),
    )
}

class AndroidKeystoreTokenStorage(context: Context) : TokenStorage {
    private val prefs =
        context.applicationContext.getSharedPreferences("revas_session", Context.MODE_PRIVATE)

    override fun readToken(): String? {
        val stored = prefs.getString(KEY_TOKEN, null) ?: return null
        return runCatching { decrypt(stored) }.getOrNull()
    }

    override fun writeToken(token: String) {
        prefs.edit().putString(KEY_TOKEN, encrypt(token)).apply()
    }

    override fun clearToken() {
        prefs.edit().remove(KEY_TOKEN).apply()
        runCatching { keyStore().deleteEntry(KEY_ALIAS) }
    }

    private fun keyStore(): KeyStore =
        KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }

    private fun secretKey(): SecretKey {
        val existing = keyStore().getEntry(KEY_ALIAS, null) as? KeyStore.SecretKeyEntry
        if (existing != null) return existing.secretKey
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
        generator.init(
            KeyGenParameterSpec.Builder(
                KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT,
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setUserAuthenticationRequired(false)
                .build(),
        )
        return generator.generateKey()
    }

    private fun encrypt(value: String): String {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, secretKey())
        val ciphertext = cipher.doFinal(value.toByteArray(Charsets.UTF_8))
        return Base64.encodeToString(cipher.iv, Base64.NO_WRAP) + ":" +
            Base64.encodeToString(ciphertext, Base64.NO_WRAP)
    }

    private fun decrypt(stored: String): String {
        val parts = stored.split(":", limit = 2)
        require(parts.size == 2) { "Corrupt token payload" }
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(
            Cipher.DECRYPT_MODE,
            secretKey(),
            GCMParameterSpec(128, Base64.decode(parts[0], Base64.NO_WRAP)),
        )
        return cipher.doFinal(Base64.decode(parts[1], Base64.NO_WRAP)).toString(Charsets.UTF_8)
    }

    private companion object {
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val KEY_ALIAS = "revas_mobile_session_key"
        const val KEY_TOKEN = "session_token"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
    }
}

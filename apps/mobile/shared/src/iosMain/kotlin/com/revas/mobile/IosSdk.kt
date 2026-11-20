package com.revas.mobile

import app.cash.sqldelight.driver.native.NativeSqliteDriver
import com.revas.mobile.data.cache.SqlDelightCache
import com.revas.mobile.data.remote.MobileApiConfig
import com.revas.mobile.data.session.TokenStorage
import com.revas.mobile.db.MobileDatabase
import com.revas.mobile.platform.platformClock
import com.revas.mobile.platform.platformLogger

private const val DATABASE_NAME = "revas-mobile.db"

fun createIosSdk(config: MobileApiConfig, tokenStorage: TokenStorage): SharedSdk {
    val database = MobileDatabase(NativeSqliteDriver(MobileDatabase.Schema, DATABASE_NAME))
    return createSdk(
        config = config,
        cache = SqlDelightCache(database),
        tokenStorage = tokenStorage,
        clock = platformClock(),
        logger = platformLogger(),
    )
}

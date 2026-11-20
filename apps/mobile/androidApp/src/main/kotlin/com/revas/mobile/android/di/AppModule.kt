package com.revas.mobile.android.di

import android.content.Context
import com.revas.mobile.SharedSdk
import com.revas.mobile.android.BuildConfig
import com.revas.mobile.createAndroidSdk
import com.revas.mobile.data.remote.MobileApiConfig
import com.revas.mobile.domain.usecase.HasStoredSessionUseCase
import com.revas.mobile.domain.usecase.LoadAlertDetailUseCase
import com.revas.mobile.domain.usecase.LoadAlertsUseCase
import com.revas.mobile.domain.usecase.LoadCaseDetailUseCase
import com.revas.mobile.domain.usecase.LoginUseCase
import com.revas.mobile.domain.usecase.LogoutUseCase
import com.revas.mobile.domain.usecase.RestoreSessionUseCase
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppModule {
    @Provides
    @Singleton
    fun provideSdk(@ApplicationContext context: Context): SharedSdk = createAndroidSdk(
        context = context,
        config = MobileApiConfig(
            identityBaseUrl = BuildConfig.IDENTITY_BASE_URL,
            alertBaseUrl = BuildConfig.ALERT_BASE_URL,
            caseBaseUrl = BuildConfig.CASE_BASE_URL,
        ),
    )

    @Provides
    fun provideLoginUseCase(sdk: SharedSdk): LoginUseCase = sdk.useCases.login

    @Provides
    fun provideRestoreSessionUseCase(sdk: SharedSdk): RestoreSessionUseCase = sdk.useCases.restore

    @Provides
    fun provideLogoutUseCase(sdk: SharedSdk): LogoutUseCase = sdk.useCases.logout

    @Provides
    fun provideHasStoredSessionUseCase(sdk: SharedSdk): HasStoredSessionUseCase = sdk.useCases.hasStoredSession

    @Provides
    fun provideLoadAlertsUseCase(sdk: SharedSdk): LoadAlertsUseCase = sdk.useCases.loadAlerts

    @Provides
    fun provideLoadAlertDetailUseCase(sdk: SharedSdk): LoadAlertDetailUseCase = sdk.useCases.loadAlert

    @Provides
    fun provideLoadCaseDetailUseCase(sdk: SharedSdk): LoadCaseDetailUseCase = sdk.useCases.loadCase
}

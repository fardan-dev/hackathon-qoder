package com.revas.mobile.android.ui

import android.net.Uri
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavDeepLinkRequest
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import androidx.navigation.navDeepLink
import androidx.navigation.navOptions
import com.revas.mobile.android.ui.alert.AlertDetailScreen
import com.revas.mobile.android.ui.alert.AlertDetailViewModel
import com.revas.mobile.android.ui.case.CaseDetailScreen
import com.revas.mobile.android.ui.case.CaseDetailViewModel
import com.revas.mobile.android.ui.inbox.AlertInboxScreen
import com.revas.mobile.android.ui.inbox.AlertInboxViewModel
import com.revas.mobile.android.ui.session.LockScreen
import com.revas.mobile.android.ui.session.LoginScreen
import com.revas.mobile.android.ui.session.SessionUiState
import com.revas.mobile.android.ui.session.SessionViewModel
import com.revas.mobile.android.ui.session.SplashScreen

private object Routes {
    const val INBOX = "inbox"
    const val ALERT = "alert/{alertId}"
    const val CASE = "case/{caseId}"

    fun alert(alertId: String): String = "alert/${Uri.encode(alertId)}"
    fun case(caseId: String): String = "case/${Uri.encode(caseId)}"
}

@Composable
fun RevasRoot(
    deepLink: Uri? = null,
    sessionViewModel: SessionViewModel = hiltViewModel(),
) {
    val session by sessionViewModel.session.collectAsStateWithLifecycle()
    val form by sessionViewModel.form.collectAsStateWithLifecycle()

    Surface(modifier = Modifier.fillMaxSize(), color = MaterialTheme.colorScheme.background) {
        when (val current = session) {
            SessionUiState.Loading -> SplashScreen()
            SessionUiState.LoggedOut -> LoginScreen(
                form = form,
                onEmailChange = sessionViewModel::onEmailChange,
                onPasswordChange = sessionViewModel::onPasswordChange,
                onSubmit = sessionViewModel::submit,
            )

            is SessionUiState.LoggedIn -> {
                if (current.locked) {
                    LockScreen(
                        onUnlock = sessionViewModel::unlock,
                        onSignOut = sessionViewModel::signOut,
                    )
                } else {
                    AppNavHost(
                        userName = current.user.displayName.ifBlank { current.user.email },
                        deepLink = deepLink,
                        onSignOut = sessionViewModel::signOut,
                    )
                }
            }
        }
    }
}

@Composable
private fun AppNavHost(
    userName: String,
    deepLink: Uri?,
    onSignOut: () -> Unit,
) {
    val navController = rememberNavController()

    LaunchedEffect(deepLink) {
        if (deepLink != null) {
            navController.navigate(
                NavDeepLinkRequest.Builder.fromUri(deepLink).build(),
                navOptions { launchSingleTop = true },
            )
        }
    }

    NavHost(navController = navController, startDestination = Routes.INBOX) {
        composable(Routes.INBOX) {
            val viewModel: AlertInboxViewModel = hiltViewModel()
            val state by viewModel.state.collectAsStateWithLifecycle()
            AlertInboxScreen(
                state = state,
                userName = userName,
                onRefresh = { viewModel.refresh(force = true) },
                onSignOut = onSignOut,
                onSelectAlert = { navController.navigate(Routes.alert(it)) },
                onSelectSeverity = viewModel::selectSeverity,
                onToggleSla = viewModel::toggleSlaOnly,
            )
        }
        composable(
            route = Routes.ALERT,
            arguments = listOf(navArgument("alertId") { type = NavType.StringType }),
            deepLinks = listOf(navDeepLink { uriPattern = "revas://alert/{alertId}" }),
        ) {
            val viewModel: AlertDetailViewModel = hiltViewModel()
            val state by viewModel.state.collectAsStateWithLifecycle()
            AlertDetailScreen(
                state = state,
                onBack = { navController.popBackStack() },
                onReload = viewModel::reload,
            )
        }
        composable(
            route = Routes.CASE,
            arguments = listOf(navArgument("caseId") { type = NavType.StringType }),
            deepLinks = listOf(navDeepLink { uriPattern = "revas://case/{caseId}" }),
        ) {
            val viewModel: CaseDetailViewModel = hiltViewModel()
            val state by viewModel.state.collectAsStateWithLifecycle()
            CaseDetailScreen(
                state = state,
                onBack = { navController.popBackStack() },
                onReload = viewModel::reload,
                onSelectAlert = { navController.navigate(Routes.alert(it)) },
            )
        }
    }
}

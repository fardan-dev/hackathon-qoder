package com.revas.mobile.android.ui.session

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.fragment.app.FragmentActivity
import com.revas.mobile.android.security.BiometricGate
import com.revas.mobile.android.ui.findActivity

@Composable
fun LockScreen(
    onUnlock: () -> Unit,
    onSignOut: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val activity = LocalContext.current.findActivity() as? FragmentActivity
    var error by remember { mutableStateOf<String?>(null) }
    var requested by remember { mutableStateOf(false) }

    fun requestAuthentication() {
        if (activity == null) {
            onUnlock()
            return
        }
        BiometricGate.authenticate(
            activity = activity,
            onSuccess = onUnlock,
            onError = { message -> error = message },
        )
    }

    LaunchedEffect(activity) {
        if (activity == null || !BiometricGate.canAuthenticate(activity)) {
            onUnlock()
            return@LaunchedEffect
        }
        if (!requested) {
            requested = true
            requestAuthentication()
        }
    }

    Column(
        modifier = modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = "Sesi terkunci",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(8.dp))
        Text(
            text = "Verifikasi biometrik atau kredensial perangkat untuk melanjutkan.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (error != null) {
            Spacer(Modifier.height(12.dp))
            Text(
                text = error.orEmpty(),
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }
        Spacer(Modifier.height(24.dp))
        Button(onClick = { requestAuthentication() }) {
            Text("Verifikasi")
        }
        TextButton(onClick = onSignOut) {
            Text("Keluar dan masuk dengan kata sandi")
        }
    }
}

@Composable
fun SplashScreen(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.fillMaxSize(),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        CircularProgressIndicator()
        Spacer(Modifier.height(12.dp))
        Text("Memuat sesi", style = MaterialTheme.typography.bodyMedium)
    }
}

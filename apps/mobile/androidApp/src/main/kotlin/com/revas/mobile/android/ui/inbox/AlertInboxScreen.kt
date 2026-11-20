@file:OptIn(ExperimentalMaterial3Api::class, ExperimentalTime::class)

package com.revas.mobile.android.ui.inbox

import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.Card
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.revas.mobile.android.ui.components.ErrorCard
import com.revas.mobile.android.ui.components.LoadingPanel
import com.revas.mobile.android.ui.components.OfflineBanner
import com.revas.mobile.android.ui.components.ReadOnlyNotice
import com.revas.mobile.android.ui.components.SeverityBadge
import com.revas.mobile.domain.format.Formatters
import com.revas.mobile.domain.format.MobileLabels
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.Severity
import kotlin.time.ExperimentalTime

@Composable
fun AlertInboxScreen(
    state: AlertInboxUiState,
    userName: String,
    onRefresh: () -> Unit,
    onSignOut: () -> Unit,
    onSelectAlert: (String) -> Unit,
    onSelectSeverity: (Severity?) -> Unit,
    onToggleSla: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Scaffold(
        modifier = modifier,
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text("Alert")
                        Text(
                            text = "Halo, $userName",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                },
                actions = {
                    TextButton(onClick = onRefresh, enabled = !state.refreshing) { Text("Muat ulang") }
                    TextButton(onClick = onSignOut) { Text("Keluar") }
                },
            )
        },
    ) { innerPadding ->
        Column(
            Modifier
                .padding(innerPadding)
                .fillMaxSize(),
        ) {
            if (state.refreshing && state.alerts.isNotEmpty()) {
                LinearProgressIndicator(Modifier.fillMaxWidth())
            }
            FilterRow(state.filter.severity, state.filter.slaBreachedOnly, onSelectSeverity, onToggleSla)
            val error = state.error
            when {
                state.loading -> LoadingPanel()
                error != null && state.alerts.isEmpty() -> Column(Modifier.padding(16.dp)) {
                    ErrorCard(error, onRefresh)
                }
                state.isEmpty -> EmptyState()
                else -> AlertList(state, onSelectAlert)
            }
        }
    }
}

@Composable
private fun FilterRow(
    selected: Severity?,
    slaOnly: Boolean,
    onSelectSeverity: (Severity?) -> Unit,
    onToggleSla: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .horizontalScroll(rememberScrollState())
            .padding(horizontal = 12.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        FilterChip(
            selected = selected == null,
            onClick = { onSelectSeverity(null) },
            label = { Text("Semua") },
            modifier = Modifier.testTag("filter_all"),
        )
        Severity.entries.forEach { severity ->
            FilterChip(
                selected = selected == severity,
                onClick = { onSelectSeverity(if (selected == severity) null else severity) },
                label = { Text(MobileLabels.severity(severity)) },
                modifier = Modifier.testTag("filter_${severity.name}"),
            )
        }
        FilterChip(
            selected = slaOnly,
            onClick = onToggleSla,
            label = { Text("Lewat SLA") },
            modifier = Modifier.testTag("filter_sla"),
        )
    }
}

@Composable
private fun AlertList(state: AlertInboxUiState, onSelectAlert: (String) -> Unit) {
    val error = state.error
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .testTag("alert_list"),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (state.fromCache) {
            item("offline-banner") {
                OfflineBanner(state.lastSyncedAt)
            }
        }
        if (error != null) {
            item("error-banner") {
                ErrorCard(error, onRetry = {})
            }
        }
        items(state.alerts, key = { it.alertId }) { alert ->
            AlertRow(alert, onClick = { onSelectAlert(alert.alertId) })
        }
        item("read-only") {
            Spacer(Modifier.height(4.dp))
            ReadOnlyNotice()
        }
    }
}

@Composable
private fun AlertRow(alert: Alert, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Column(Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                SeverityBadge(alert.severity)
                Spacer(Modifier.width(8.dp))
                Text(
                    text = Formatters.idr(alert.claimedAmount),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                )
            }
            Spacer(Modifier.height(6.dp))
            Text("Klaim ${alert.claimId}", style = MaterialTheme.typography.bodyMedium)
            Text(
                text = listOfNotNull(
                    MobileLabels.status(alert.status),
                    MobileLabels.lossType(alert.lossType),
                    "${alert.flagCount} indikasi",
                    "Ditahan sementara".takeIf { alert.suppressed },
                ).joinToString(" · "),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.height(4.dp))
            Text(
                text = if (alert.slaBreached) {
                    "Lewat SLA · tenggat ${Formatters.dateTime(alert.slaDueAt)}"
                } else {
                    "SLA ${Formatters.dateTime(alert.slaDueAt)}"
                },
                style = MaterialTheme.typography.bodySmall,
                color = if (alert.slaBreached) {
                    MaterialTheme.colorScheme.error
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                },
            )
        }
    }
}

@Composable
private fun EmptyState() {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Tidak ada alert pada filter ini", style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(8.dp))
        Text(
            text = "Coba ubah filter atau muat ulang data.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}




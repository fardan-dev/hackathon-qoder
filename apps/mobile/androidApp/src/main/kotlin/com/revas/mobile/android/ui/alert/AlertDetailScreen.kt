@file:OptIn(ExperimentalMaterial3Api::class)

package com.revas.mobile.android.ui.alert

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Card
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.revas.mobile.android.ui.components.ErrorCard
import com.revas.mobile.android.ui.components.LoadingPanel
import com.revas.mobile.android.ui.components.ReadOnlyNotice
import com.revas.mobile.android.ui.components.SeverityBadge
import com.revas.mobile.domain.format.Formatters
import com.revas.mobile.domain.format.MobileLabels
import com.revas.mobile.domain.model.Alert
import com.revas.mobile.domain.model.AlertFlag

@Composable
fun AlertDetailScreen(
    state: AlertDetailUiState,
    onBack: () -> Unit,
    onReload: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Scaffold(
        modifier = modifier,
        topBar = {
            TopAppBar(
                title = { Text("Detail alert") },
                navigationIcon = { TextButton(onClick = onBack) { Text("Kembali") } },
                actions = { TextButton(onClick = onReload, enabled = !state.loading) { Text("Muat ulang") } },
            )
        },
    ) { innerPadding ->
        val alert = state.alert
        val error = state.error
        when {
            state.loading -> LoadingPanel(Modifier.padding(innerPadding))
            error != null && alert == null -> Column(Modifier.padding(innerPadding).padding(16.dp)) {
                ErrorCard(error, onReload)
            }
            alert != null -> AlertDetailContent(alert, error, Modifier.padding(innerPadding))
        }
    }
}

@Composable
private fun AlertDetailContent(alert: Alert, error: String?, modifier: Modifier = Modifier) {
    LazyColumn(
        modifier = modifier.fillMaxSize(),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (error != null) {
            item("error") { ErrorCard(error, onRetry = {}) }
        }
        item("summary") {
            Card(Modifier.fillMaxWidth()) {
                Column(Modifier.padding(16.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        SeverityBadge(alert.severity)
                        Spacer(Modifier.width(8.dp))
                        Text(
                            text = Formatters.idr(alert.claimedAmount),
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }
                    Spacer(Modifier.height(8.dp))
                    DetailLine("Klaim", alert.claimId)
                    DetailLine("Status", MobileLabels.status(alert.status))
                    DetailLine("Jenis kerugian", MobileLabels.lossType(alert.lossType))
                    DetailLine("Jumlah indikasi", alert.flagCount.toString())
                    DetailLine(
                        "SLA",
                        if (alert.slaBreached) {
                            "Lewat SLA · ${Formatters.dateTime(alert.slaDueAt)}"
                        } else {
                            Formatters.dateTime(alert.slaDueAt)
                        },
                    )
                    DetailLine("Dibuat", Formatters.dateTime(alert.createdAt))
                    if (alert.suppressed) {
                        Spacer(Modifier.height(8.dp))
                        Text(
                            text = "Alert ini ditahan sementara oleh aturan penahanan.",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.error,
                        )
                    }
                }
            }
        }
        item("flags-title") {
            Text("Indikasi aturan", style = MaterialTheme.typography.titleMedium)
        }
        items(alert.flags, key = { it.flagId }) { flag ->
            FlagCard(flag)
        }
        item("read-only") { ReadOnlyNotice() }
    }
}

@Composable
private fun FlagCard(flag: AlertFlag) {
    Card(Modifier.fillMaxWidth()) {
        Column(Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(flag.ruleName, style = MaterialTheme.typography.titleSmall)
                Spacer(Modifier.weight(1f))
                SeverityBadge(flag.severity)
            }
            Spacer(Modifier.height(4.dp))
            Text(
                text = "${flag.family} · ${flag.reasonCode} · v${flag.ruleVersion}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.height(8.dp))
            Text(flag.explanation, style = MaterialTheme.typography.bodyMedium)
            Spacer(Modifier.height(8.dp))
            HorizontalDivider()
            Spacer(Modifier.height(8.dp))
            DetailLine("Nilai pemicu", flag.triggerValue)
            DetailLine("Versi parameter", flag.paramsVersion.toString())
            DetailLine("Referensi bukti", flag.evidenceRefs.joinToString(", ").ifBlank { "-" })
            if (flag.suppressed) {
                Text(
                    text = "Ditahan sementara",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error,
                )
            }
        }
    }
}

@Composable
private fun DetailLine(label: String, value: String) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
        Text(
            text = label,
            modifier = Modifier.width(128.dp),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Text(value, style = MaterialTheme.typography.bodyMedium)
    }
}

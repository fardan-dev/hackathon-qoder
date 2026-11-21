@file:OptIn(ExperimentalMaterial3Api::class)

package com.revas.mobile.android.ui.case

import androidx.compose.foundation.clickable
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
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.revas.mobile.android.ui.components.ErrorCard
import com.revas.mobile.android.ui.components.KeepSecureScreen
import com.revas.mobile.android.ui.components.LoadingPanel
import com.revas.mobile.android.ui.components.ReadOnlyNotice
import com.revas.mobile.android.ui.components.SeverityBadge
import com.revas.mobile.domain.format.Formatters
import com.revas.mobile.domain.format.MobileLabels
import com.revas.mobile.domain.model.CaseClaim
import com.revas.mobile.domain.model.CaseDecision
import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.EvidenceItem

@Composable
fun CaseDetailScreen(
    state: CaseDetailUiState,
    onBack: () -> Unit,
    onReload: () -> Unit,
    onSelectAlert: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    KeepSecureScreen()
    Scaffold(
        modifier = modifier,
        topBar = {
            TopAppBar(
                title = { Text("Detail kasus") },
                navigationIcon = { TextButton(onClick = onBack) { Text("Kembali") } },
                actions = { TextButton(onClick = onReload, enabled = !state.loading) { Text("Muat ulang") } },
            )
        },
    ) { innerPadding ->
        val detail = state.case
        val error = state.error
        when {
            state.loading -> LoadingPanel(Modifier.padding(innerPadding))
            error != null && detail == null -> Column(Modifier.padding(innerPadding).padding(16.dp)) {
                ErrorCard(error, onReload)
            }
            detail != null -> CaseDetailContent(detail, error, onSelectAlert, Modifier.padding(innerPadding))
        }
    }
}

@Composable
private fun CaseDetailContent(
    detail: CaseDetail,
    error: String?,
    onSelectAlert: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    LazyColumn(
        modifier = modifier
            .fillMaxSize()
            .testTag("case_detail"),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (error != null) {
            item("error") { ErrorCard(error, onRetry = {}) }
        }
        item("summary") {
            Card(Modifier.fillMaxWidth()) {
                Column(Modifier.padding(16.dp)) {
                    Text(detail.title, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
                    Spacer(Modifier.height(8.dp))
                    Text(detail.summary, style = MaterialTheme.typography.bodyMedium)
                    Spacer(Modifier.height(12.dp))
                    DetailLine("Status", detail.status)
                    DetailLine("Prioritas", detail.priority.toString())
                    DetailLine("Jumlah klaim", detail.claimCount.toString())
                    DetailLine("Dibuat", Formatters.dateTime(detail.createdAt))
                    DetailLine("Diperbarui", Formatters.dateTime(detail.updatedAt))
                }
            }
        }
        item("claims-title") { SectionTitle("Klaim terkait") }
        items(detail.claims, key = { it.claimId }) { claim ->
            ClaimCard(claim, onClick = { onSelectAlert(claim.alertId) })
        }
        item("evidence-title") { SectionTitle("Bukti") }
        if (detail.evidence.isEmpty()) {
            item("evidence-empty") { MutedText("Belum ada bukti terlampir.") }
        } else {
            items(detail.evidence, key = { it.evidenceId }) { EvidenceCard(it) }
        }
        item("decisions-title") { SectionTitle("Riwayat keputusan") }
        if (detail.decisions.isEmpty()) {
            item("decisions-empty") { MutedText("Belum ada keputusan.") }
        } else {
            items(detail.decisions, key = { it.decisionId }) { DecisionCard(it) }
        }
        item("read-only") { ReadOnlyNotice() }
    }
}

@Composable
private fun ClaimCard(claim: CaseClaim, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Column(Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("Klaim ${claim.claimId}", style = MaterialTheme.typography.titleSmall)
                Spacer(Modifier.weight(1f))
                SeverityBadge(claim.severity)
            }
            Spacer(Modifier.height(4.dp))
            Text(
                text = "${Formatters.idr(claim.claimedAmount)} · ${MobileLabels.lossType(claim.lossType)} · ${claim.flags.size} indikasi",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.height(4.dp))
            Text(
                text = "Lihat detail alert",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.primary,
            )
        }
    }
}

@Composable
private fun EvidenceCard(evidence: EvidenceItem) {
    Card(Modifier.fillMaxWidth()) {
        Column(Modifier.padding(16.dp)) {
            Text(evidence.artifactType, style = MaterialTheme.typography.titleSmall)
            Spacer(Modifier.height(4.dp))
            Text(evidence.source, style = MaterialTheme.typography.bodySmall)
            Spacer(Modifier.height(4.dp))
            Text(
                text = "SHA-256 ${evidence.contentSha256.take(16)}…",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            evidence.notes?.takeIf { it.isNotBlank() }?.let {
                Spacer(Modifier.height(4.dp))
                Text(it, style = MaterialTheme.typography.bodySmall)
            }
            Spacer(Modifier.height(4.dp))
            Text(
                text = Formatters.dateTime(evidence.createdAt),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun DecisionCard(decision: CaseDecision) {
    Card(Modifier.fillMaxWidth()) {
        Column(Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(decision.outcome, style = MaterialTheme.typography.titleSmall)
                Spacer(Modifier.weight(1f))
                Text(
                    text = decision.reasonCode,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            if (decision.notes.isNotBlank()) {
                Spacer(Modifier.height(4.dp))
                Text(decision.notes, style = MaterialTheme.typography.bodySmall)
            }
            Spacer(Modifier.height(4.dp))
            Text(
                text = Formatters.dateTime(decision.decidedAt),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun SectionTitle(text: String) {
    Column {
        HorizontalDivider()
        Spacer(Modifier.height(8.dp))
        Text(text, style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(4.dp))
    }
}

@Composable
private fun MutedText(text: String) {
    Text(
        text = text,
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
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

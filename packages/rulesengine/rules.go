package rulesengine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func skip(reason, detail string) Outcome {
	return Outcome{SkippedReason: reason, SkippedDetail: detail}
}

func evalA1(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	invoiceNo := c.InvoiceNumber
	providerID := c.ProviderID
	if c.Invoice != nil {
		invoiceNo = c.Invoice.Number
		providerID = c.Invoice.ProviderID
	}
	if invoiceNo == "" {
		return skip("DATA_UNAVAILABLE", "invoice_number is missing")
	}
	var refs []string
	for _, other := range idx.byInvoice[invoiceNo] {
		if sameClaim(&other, c) {
			continue
		}
		otherProvider := other.ProviderID
		if other.Invoice != nil {
			otherProvider = other.Invoice.ProviderID
		}
		if otherProvider == providerID {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: invoiceNo + " @ " + providerID,
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Invoice %s (Provider %s) juga digunakan pada klaim %s. Perlu verifikasi apakah penagihan berulang ini sah.", invoiceNo, providerID, strings.Join(refs, ", ")),
	}
}

func evalA2(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	fp := c.Fingerprint()
	var refs []string
	for _, other := range idx.byFingerprint[fp] {
		if sameClaim(&other, c) {
			continue
		}
		refs = append(refs, other.ClaimID)
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fp,
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Kombinasi claimant %s, tanggal insiden %s, nominal %s, dan provider %s muncul pada klaim %s. Perlu verifikasi apakah ini klaim duplikat.", c.ClaimantID, c.IncidentDate.Format("2006-01-02"), rupiah(c.ClaimedAmount), c.ProviderID, strings.Join(refs, ", ")),
	}
}

func evalA3(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	doc := c.DocSHA256
	if doc == "" && c.Invoice != nil {
		doc = c.Invoice.DocSHA256
	}
	if doc == "" {
		return skip("DATA_UNAVAILABLE", "document hash is missing")
	}
	var refs []string
	for _, other := range idx.byDoc[doc] {
		if sameClaim(&other, c) {
			continue
		}
		if other.ClaimantID != c.ClaimantID || other.PolicyID != c.PolicyID {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: doc,
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Hash dokumen yang sama juga muncul pada klaim %s dengan claimant/polis berbeda. Periksa dokumen asli dan keterkaitannya.", strings.Join(refs, ", ")),
	}
}

func evalA4(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Invoice == nil || c.Invoice.Date.IsZero() {
		return skip("DATA_UNAVAILABLE", "invoice record is missing")
	}
	diff := c.Invoice.Amount - c.ClaimedAmount
	if diff == 0 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("invoice=%s claim=%s selisih=%s", rupiah(c.Invoice.Amount), rupiah(c.ClaimedAmount), rupiah(abs(diff))),
		EvidenceRefs: []string{c.Invoice.Number, c.ClaimID},
		Explanation:  fmt.Sprintf("Nominal invoice %s (%s) berbeda dari nominal klaim %s (%s). Perlu verifikasi nilai yang benar.", c.Invoice.Number, rupiah(c.Invoice.Amount), c.ClaimID, rupiah(c.ClaimedAmount)),
	}
}

func evalA5(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Invoice == nil || c.Invoice.Date.IsZero() {
		return skip("DATA_UNAVAILABLE", "invoice date is missing")
	}
	inv := c.Invoice.Date
	tooEarly := inv.Before(c.IncidentDate)
	tooLate := inv.After(c.ClaimDate.AddDate(0, 0, 30))
	if !tooEarly && !tooLate {
		return Outcome{}
	}
	reason := "sebelum tanggal insiden"
	if tooLate {
		reason = "lebih dari 30 hari setelah tanggal klaim"
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("invoice_date=%s claim_date=%s", inv.Format("2006-01-02"), c.ClaimDate.Format("2006-01-02")),
		EvidenceRefs: []string{c.Invoice.Number, c.ClaimID},
		Explanation:  fmt.Sprintf("Tanggal invoice %s (%s) berada %s. Periksa kronologi dokumen.", c.Invoice.Number, inv.Format("2006-01-02"), reason),
	}
}

func evalA6(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	if c.Invoice == nil {
		return skip("DATA_UNAVAILABLE", "invoice record is missing")
	}
	var refs []string
	for _, other := range idx.byInvoice[c.Invoice.Number] {
		if sameClaim(&other, c) || other.Invoice == nil {
			continue
		}
		if other.Invoice.ProviderID != c.Invoice.ProviderID || other.Invoice.Amount != c.Invoice.Amount {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("%s provider=%s amount=%s", c.Invoice.Number, c.Invoice.ProviderID, rupiah(c.Invoice.Amount)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Nomor invoice %s muncul dengan provider atau nominal berbeda pada klaim %s. Perlu verifikasi dokumen yang otentik.", c.Invoice.Number, strings.Join(refs, ", ")),
	}
}

func evalB1(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	limit := int64(0)
	switch c.LossType {
	case "MOTOR":
		limit = 50_000_000
	case "PROPERTY":
		limit = 100_000_000
	default:
		return skip("DATA_UNAVAILABLE", "loss_type is outside the thin-slice scope")
	}
	if c.ClaimedAmount <= limit {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%s > %s", rupiah(c.ClaimedAmount), rupiah(limit)),
		EvidenceRefs: []string{c.ClaimID},
		Explanation:  fmt.Sprintf("Nominal klaim %s melebihi batas absolut %s untuk %s. Perlu verifikasi kewajaran nominal.", rupiah(c.ClaimedAmount), rupiah(limit), c.LossType),
	}
}

func evalB2(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	stat, ok := idx.lossStats[c.LossType]
	if !ok || stat.count < 30 {
		return skip("POPULATION_TOO_SMALL", "batch population for the loss type is below 30")
	}
	if stat.mean <= 0 {
		return skip("DATA_UNAVAILABLE", "batch mean is not positive")
	}
	threshold := stat.mean * 3
	if float64(c.ClaimedAmount) <= threshold {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%s > 3x rata-rata %s", rupiah(c.ClaimedAmount), rupiah(int64(stat.mean))),
		EvidenceRefs: []string{c.ClaimID, "batch:" + c.BatchID},
		Explanation:  fmt.Sprintf("Nominal klaim %s berada di atas 3 kali rata-rata %s pada batch ini (%s). Perlu verifikasi apakah nominal wajar.", rupiah(c.ClaimedAmount), c.LossType, rupiah(int64(stat.mean))),
	}
}

func evalB3(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	limit := int64(0)
	switch c.LossType {
	case "MOTOR":
		limit = 50_000_000
	case "PROPERTY":
		limit = 100_000_000
	default:
		return skip("DATA_UNAVAILABLE", "loss_type is outside the thin-slice scope")
	}
	lower := limit * 85 / 100
	if c.ClaimedAmount < lower || c.ClaimedAmount > limit {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%s dalam rentang %s-%s", rupiah(c.ClaimedAmount), rupiah(lower), rupiah(limit)),
		EvidenceRefs: []string{c.ClaimID},
		Explanation:  fmt.Sprintf("Nominal klaim %s berada pada 85%%-100%% batas %s. Periksa indikasi pemecahan nominal untuk menghindari review.", rupiah(c.ClaimedAmount), rupiah(limit)),
	}
}

func evalB4(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.ClaimedAmount < 5_000_000 || c.ClaimedAmount%1_000_000 != 0 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: rupiah(c.ClaimedAmount),
		EvidenceRefs: []string{c.ClaimID},
		Explanation:  fmt.Sprintf("Nominal klaim %s adalah angka bulat besar (kelipatan 1.000.000). Perlu verifikasi kesesuaian dengan dokumen.", rupiah(c.ClaimedAmount)),
	}
}

func evalB5(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil {
		return skip("DATA_UNAVAILABLE", "policy record is missing")
	}
	if c.ClaimedAmount <= c.Policy.SumInsured {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%s > sum insured %s", rupiah(c.ClaimedAmount), rupiah(c.Policy.SumInsured)),
		EvidenceRefs: []string{c.ClaimID, c.PolicyID},
		Explanation:  fmt.Sprintf("Nominal klaim %s melebihi sum insured polis %s (%s). Periksa batas pertanggungan.", rupiah(c.ClaimedAmount), c.PolicyID, rupiah(c.Policy.SumInsured)),
	}
}

func evalB6(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil {
		return skip("DATA_UNAVAILABLE", "policy record is missing")
	}
	var total int64
	var refs []string
	for _, other := range idx.byClaimant[c.ClaimantID] {
		if other.PolicyID != c.PolicyID {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 365) {
			total += other.ClaimedAmount
			refs = append(refs, other.ClaimID)
		}
	}
	limit := c.Policy.SumInsured * 80 / 100
	if total <= limit {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("akumulasi %s > 80%% sum insured %s", rupiah(total), rupiah(c.Policy.SumInsured)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Akumulasi klaim claimant %s pada polis %s dalam 12 bulan mencapai %s, di atas 80%% sum insured. Perlu verifikasi sisa cover.", c.ClaimantID, c.PolicyID, rupiah(total)),
	}
}

func evalC1(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil || c.Policy.Inception.IsZero() {
		return skip("DATA_UNAVAILABLE", "policy inception date is missing")
	}
	days := daysBetween(c.Policy.Inception, c.IncidentDate)
	if days < 0 || days > 30 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("usia polis %d hari", days),
		EvidenceRefs: []string{c.ClaimID, c.PolicyID},
		Explanation:  fmt.Sprintf("Klaim terjadi %d hari setelah polis %s aktif (batas 30 hari). Periksa riwayat underwriting.", days, c.PolicyID),
	}
}

func evalC2(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil || c.Policy.Reinstatement == nil {
		return skip("DATA_UNAVAILABLE", "policy reinstatement date is missing")
	}
	days := daysBetween(*c.Policy.Reinstatement, c.IncidentDate)
	if days < 0 || days > 30 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("reinstatement %s, insiden +%d hari", c.Policy.Reinstatement.Format("2006-01-02"), days),
		EvidenceRefs: []string{c.ClaimID, c.PolicyID},
		Explanation:  fmt.Sprintf("Insiden terjadi %d hari setelah reinstatement polis %s (batas 30 hari). Perlu verifikasi kesinambungan cover.", days, c.PolicyID),
	}
}

func evalC3(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	days := daysBetween(c.IncidentDate, c.ClaimDate)
	if days <= 30 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%d hari setelah insiden", days),
		EvidenceRefs: []string{c.ClaimID},
		Explanation:  fmt.Sprintf("Klaim dilaporkan %d hari setelah insiden (batas 30 hari). Perlu verifikasi alasan keterlambatan.", days),
	}
}

func evalC4(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byClaimant[c.ClaimantID] {
		if other.LossType == c.LossType && withinDays(c.ClaimDate, other.ClaimDate, 90) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) < 2 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("%d klaim %s dalam 90 hari", len(refs), c.LossType),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant %s mengajukan %d klaim %s dalam 90 hari. Periksa polanya.", c.ClaimantID, len(refs), c.LossType),
	}
}

func evalC5(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byFingerprint[c.Fingerprint()] {
		if sameClaim(&other, c) || other.Status != "REJECTED" {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 180) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("fingerprint sama dengan klaim REJECTED %s", strings.Join(refs, ", ")),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Klaim ini memiliki fingerprint yang sama dengan klaim yang ditolak %s dalam 180 hari. Perlu verifikasi resubmisi.", strings.Join(refs, ", ")),
	}
}

func evalD1(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	hash := c.Identities["BANK_ACCOUNT"]
	if hash == "" {
		return skip("DATA_UNAVAILABLE", "bank account hash is missing")
	}
	var refs []string
	for _, other := range idx.byIdentity["BANK_ACCOUNT|"+hash] {
		if other.ClaimantID != c.ClaimantID {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: "BANK_ACCOUNT " + hash,
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Rekening bank yang sama digunakan oleh claimant berbeda pada klaim %s. Periksa keterkaitan identitas.", strings.Join(refs, ", ")),
	}
}

func evalD2(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	var matched []string
	for _, typ := range []string{"PHONE", "EMAIL", "ADDRESS"} {
		hash := c.Identities[typ]
		if hash == "" {
			continue
		}
		for _, other := range idx.byIdentity[typ+"|"+hash] {
			if other.ClaimantID != c.ClaimantID {
				refs = append(refs, other.ClaimID)
				matched = append(matched, typ)
			}
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	matched = sortedUnique(matched)
	return Outcome{
		TriggerValue: strings.Join(matched, ", "),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Kontak %s yang sama muncul pada claimant berbeda di klaim %s. Periksa apakah identitas memang bersama.", strings.Join(matched, ", "), strings.Join(refs, ", ")),
	}
}

func evalD3(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byClaimant[c.ClaimantID] {
		if other.ProviderID == c.ProviderID && inWindow(c.ClaimDate, other.ClaimDate, 180) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) < 3 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("%d klaim pada provider %s", len(refs), c.ProviderID),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant %s memiliki %d klaim pada provider %s dalam 180 hari. Periksa konsentrasi hubungan.", c.ClaimantID, len(refs), c.ProviderID),
	}
}

func evalD4(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	counts := idx.providerCount[c.LossType]
	if len(counts) < 10 {
		return skip("POPULATION_TOO_SMALL", "fewer than 10 providers in the batch for the loss type")
	}
	values := make([]float64, 0, len(counts))
	for _, n := range counts {
		values = append(values, float64(n))
	}
	median := medianOf(values)
	count := float64(counts[c.ProviderID])
	if count <= median*3 {
		return Outcome{}
	}
	refs := []string{c.ProviderID}
	for _, other := range idx.byProvider[c.ProviderID] {
		if other.LossType == c.LossType {
			refs = append(refs, other.ClaimID)
		}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("provider %s: %d klaim vs median %.1f", c.ProviderID, counts[c.ProviderID], median),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Volume klaim provider %s pada batch ini %d, lebih dari 3 kali median peer (%.1f). Periksa pola penagihan provider.", c.ProviderID, counts[c.ProviderID], median),
	}
}

func evalD5(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	totals := idx.providerAmount[c.LossType]
	counts := idx.providerCount[c.LossType]
	if len(counts) < 10 {
		return skip("POPULATION_TOO_SMALL", "fewer than 10 providers in the batch for the loss type")
	}
	avg := map[string]float64{}
	values := make([]float64, 0, len(totals))
	for provider, total := range totals {
		a := float64(total) / float64(counts[provider])
		avg[provider] = a
		values = append(values, a)
	}
	median := medianOf(values)
	if avg[c.ProviderID] <= median*3 {
		return Outcome{}
	}
	var refs []string
	for _, other := range idx.byProvider[c.ProviderID] {
		if other.LossType == c.LossType {
			refs = append(refs, other.ClaimID)
		}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("rata-rata provider %s = %s vs median peer %s", c.ProviderID, rupiah(int64(avg[c.ProviderID])), rupiah(int64(median))),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Rata-rata nominal klaim provider %s (%s) lebih dari 3 kali median peer (%s). Periksa kewajaran tagihan.", c.ProviderID, rupiah(int64(avg[c.ProviderID])), rupiah(int64(median))),
	}
}

func evalD6(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	component := idx.component(c.ClaimantID)
	if len(component) < 2 {
		return Outcome{}
	}
	claimsBy := map[string]struct{}{}
	var refs []string
	for _, other := range idx.all {
		if _, ok := component[other.ClaimantID]; !ok {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 180) {
			if _, seen := claimsBy[other.ClaimID]; !seen {
				claimsBy[other.ClaimID] = struct{}{}
				refs = append(refs, other.ClaimID)
			}
		}
	}
	if len(refs) < 3 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	claimants := make([]string, 0, len(component))
	for claimant := range component {
		claimants = append(claimants, claimant)
	}
	claimants = sortedUnique(claimants)
	return Outcome{
		TriggerValue: fmt.Sprintf("komponen %d claimant, %d klaim", len(claimants), len(refs)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant %s terhubung melalui atribut identitas bersama dengan %s dan memiliki %d klaim dalam 180 hari. Periksa indikasi jaringan.", c.ClaimantID, strings.Join(claimants, ", "), len(refs)),
	}
}

func evalD7(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil || c.Policy.AssetID == "" {
		return skip("DATA_UNAVAILABLE", "policy asset_id is missing")
	}
	var refs []string
	for _, other := range idx.byAsset[c.Policy.AssetID] {
		if other.PolicyID == c.PolicyID || other.Policy == nil {
			continue
		}
		if other.Policy.Status == "ACTIVE" && !other.IncidentDate.Before(other.Policy.Inception) && !other.IncidentDate.After(other.Policy.Expiry) {
			refs = append(refs, other.PolicyID+"/"+other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: "asset " + c.Policy.AssetID,
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Aset %s tercatat pada lebih dari satu polis aktif (%s). Periksa duplikasi pertanggungan.", c.Policy.AssetID, strings.Join(refs, ", ")),
	}
}

func evalE1(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil || c.Policy.CoverageAdded == nil {
		return skip("DATA_UNAVAILABLE", "coverage_added_date is missing")
	}
	days := daysBetween(*c.Policy.CoverageAdded, c.IncidentDate)
	if days < 0 || days > 30 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("cover ditambahkan %s, insiden +%d hari", c.Policy.CoverageAdded.Format("2006-01-02"), days),
		EvidenceRefs: []string{c.ClaimID, c.PolicyID},
		Explanation:  fmt.Sprintf("Cover pada polis %s ditambahkan %d hari sebelum insiden (batas 30 hari). Periksa motif penambahan cover.", c.PolicyID, days),
	}
}

func evalE2(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.Policy == nil {
		return skip("DATA_UNAVAILABLE", "policy record is missing")
	}
	reason := ""
	if c.Policy.Status != "ACTIVE" {
		reason = fmt.Sprintf("status polis %s", c.Policy.Status)
	} else if c.IncidentDate.Before(c.Policy.Inception) {
		reason = "insiden sebelum inception"
	} else if c.IncidentDate.After(c.Policy.Expiry) {
		reason = "insiden setelah expiry"
	}
	if reason == "" {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%s: insiden %s, polis %s..%s", reason, c.IncidentDate.Format("2006-01-02"), c.Policy.Inception.Format("2006-01-02"), c.Policy.Expiry.Format("2006-01-02")),
		EvidenceRefs: []string{c.ClaimID, c.PolicyID},
		Explanation:  fmt.Sprintf("Polis %s tidak aktif pada tanggal insiden (%s). Periksa status dan periode pertanggungan.", c.PolicyID, reason),
	}
}

func evalF1(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byClaimant[c.ClaimantID] {
		if withinDays(c.ClaimDate, other.ClaimDate, 90) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) < 3 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("%d klaim/90 hari", len(refs)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant %s mengajukan %d klaim lintas produk dalam 90 hari. Periksa kecepatan pengajuan.", c.ClaimantID, len(refs)),
	}
}

func evalF2(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byFingerprint[c.Fingerprint()] {
		if sameClaim(&other, c) || other.Status != "WITHDRAWN" {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 90) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: strings.Join(refs, ", "),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Klaim ini mengulang fingerprint klaim yang ditarik %s dalam 90 hari. Periksa siklus penarikan dan pengajuan ulang.", strings.Join(refs, ", ")),
	}
}

func evalF3(_ *EvalContext, _ *index, c *ClaimSnapshot) Outcome {
	if c.PriorNeedsInfo < 2 {
		return Outcome{}
	}
	return Outcome{
		TriggerValue: fmt.Sprintf("%d siklus NEEDS_INFO", c.PriorNeedsInfo),
		EvidenceRefs: []string{c.ClaimID},
		Explanation:  fmt.Sprintf("Klaim %s sudah melewati %d siklus NEEDS_INFO. Perlu verifikasi kelengkapan dokumen.", c.ClaimID, c.PriorNeedsInfo),
	}
}

func evalF4(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	if c.RepID == "" {
		return skip("DATA_UNAVAILABLE", "rep_id is missing")
	}
	claimants := map[string]struct{}{}
	var refs []string
	for _, other := range idx.all {
		if other.RepID != c.RepID || !inWindow(c.ClaimDate, other.ClaimDate, 180) {
			continue
		}
		claimants[other.ClaimantID] = struct{}{}
		refs = append(refs, other.ClaimID)
	}
	if len(claimants) < 5 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("representatif %s untuk %d claimant", c.RepID, len(claimants)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Representatif %s mengajukan klaim untuk %d claimant berbeda dalam 180 hari. Periksa keterkaitan pihak.", c.RepID, len(claimants)),
	}
}

func evalG1(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.all {
		if sameClaim(&other, c) || idx.isCurrent(other.ClaimID) || !other.PriorConfirmed {
			continue
		}
		if other.ClaimantID != c.ClaimantID && other.ProviderID != c.ProviderID {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 730) {
			refs = append(refs, other.ClaimID)
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: strings.Join(refs, ", "),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant atau provider pernah memiliki temuan CONFIRMED (%s) dalam 24 bulan. Perlu verifikasi riwayat.", strings.Join(refs, ", ")),
	}
}

func evalG2(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	var refs []string
	for _, other := range idx.byFingerprint[c.Fingerprint()] {
		if sameClaim(&other, c) || idx.isCurrent(other.ClaimID) || other.PriorClearedReason == "" {
			continue
		}
		if inWindow(c.ClaimDate, other.ClaimDate, 365) {
			refs = append(refs, other.ClaimID+" ("+other.PriorClearedReason+")")
		}
	}
	if len(refs) == 0 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: strings.Join(refs, ", "),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Fingerprint ini pernah diverifikasi CLEARED dalam 12 bulan (%s). Flag tetap dicatat untuk kelengkapan audit.", strings.Join(refs, ", ")),
	}
}

func evalG3(_ *EvalContext, idx *index, c *ClaimSnapshot) Outcome {
	policies := map[string]struct{}{}
	var refs []string
	for _, other := range idx.byClaimant[c.ClaimantID] {
		if withinDays(c.ClaimDate, other.ClaimDate, 365) {
			policies[other.PolicyID] = struct{}{}
			refs = append(refs, other.ClaimID)
		}
	}
	if len(policies) < 3 {
		return Outcome{}
	}
	refs = sortedUnique(refs)
	return Outcome{
		TriggerValue: fmt.Sprintf("%d polis berbeda", len(policies)),
		EvidenceRefs: refs,
		Explanation:  fmt.Sprintf("Claimant %s memiliki klaim pada %d polis berbeda dalam 12 bulan. Periksa riwayat lintas polis.", c.ClaimantID, len(policies)),
	}
}

func evalH(_ *EvalContext, _ *index, _ *ClaimSnapshot) Outcome {
	return skip("DATA_UNAVAILABLE", "document forensics analyzer is not part of the thin slice")
}

func (idx *index) component(start string) map[string]struct{} {
	seen := map[string]struct{}{start: {}}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for peer := range idx.claimantGraph[cur] {
			if _, ok := seen[peer]; ok {
				continue
			}
			seen[peer] = struct{}{}
			queue = append(queue, peer)
		}
	}
	return seen
}

func sameClaim(a, b *ClaimSnapshot) bool {
	return a.ClaimID == b.ClaimID && a.BatchID == b.BatchID
}

func sortedUnique(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func medianOf(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[mid]
	}
	return (cp[mid-1] + cp[mid]) / 2
}

func rupiah(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}
	if neg {
		return "-Rp " + b.String()
	}
	return "Rp " + b.String()
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

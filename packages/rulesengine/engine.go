package rulesengine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
)

type index struct {
	all            []ClaimSnapshot
	current        map[string]struct{}
	byClaimant     map[string][]ClaimSnapshot
	byProvider     map[string][]ClaimSnapshot
	byPolicy       map[string][]ClaimSnapshot
	byInvoice      map[string][]ClaimSnapshot
	byDoc          map[string][]ClaimSnapshot
	byIdentity     map[string][]ClaimSnapshot
	byAsset        map[string][]ClaimSnapshot
	byFingerprint  map[string][]ClaimSnapshot
	lossStats      map[string]lossStat
	providerCount  map[string]map[string]int
	providerAmount map[string]map[string]int64
	claimantGraph  map[string]map[string]struct{}
}

type lossStat struct {
	count int
	mean  float64
}

func buildIndex(history, current []ClaimSnapshot) *index {
	idx := &index{
		current:        map[string]struct{}{},
		byClaimant:     map[string][]ClaimSnapshot{},
		byProvider:     map[string][]ClaimSnapshot{},
		byPolicy:       map[string][]ClaimSnapshot{},
		byInvoice:      map[string][]ClaimSnapshot{},
		byDoc:          map[string][]ClaimSnapshot{},
		byIdentity:     map[string][]ClaimSnapshot{},
		byAsset:        map[string][]ClaimSnapshot{},
		byFingerprint:  map[string][]ClaimSnapshot{},
		lossStats:      map[string]lossStat{},
		providerCount:  map[string]map[string]int{},
		providerAmount: map[string]map[string]int64{},
		claimantGraph:  map[string]map[string]struct{}{},
	}
	for _, c := range history {
		idx.all = append(idx.all, c)
	}
	for _, c := range current {
		idx.all = append(idx.all, c)
		idx.current[c.ClaimID] = struct{}{}
	}
	sort.SliceStable(idx.all, func(i, j int) bool {
		if idx.all[i].ClaimID != idx.all[j].ClaimID {
			return idx.all[i].ClaimID < idx.all[j].ClaimID
		}
		return idx.all[i].BatchID < idx.all[j].BatchID
	})

	var currentAmounts = map[string][]int64{}
	for _, c := range idx.all {
		idx.byClaimant[c.ClaimantID] = append(idx.byClaimant[c.ClaimantID], c)
		idx.byProvider[c.ProviderID] = append(idx.byProvider[c.ProviderID], c)
		idx.byPolicy[c.PolicyID] = append(idx.byPolicy[c.PolicyID], c)
		if c.InvoiceNumber != "" {
			idx.byInvoice[c.InvoiceNumber] = append(idx.byInvoice[c.InvoiceNumber], c)
		}
		doc := c.DocSHA256
		if doc == "" && c.Invoice != nil {
			doc = c.Invoice.DocSHA256
		}
		if doc != "" {
			idx.byDoc[doc] = append(idx.byDoc[doc], c)
		}
		for typ, hash := range c.Identities {
			if hash == "" {
				continue
			}
			key := typ + "|" + hash
			idx.byIdentity[key] = append(idx.byIdentity[key], c)
		}
		if c.Policy != nil && c.Policy.AssetID != "" {
			idx.byAsset[c.Policy.AssetID] = append(idx.byAsset[c.Policy.AssetID], c)
		}
		idx.byFingerprint[c.Fingerprint()] = append(idx.byFingerprint[c.Fingerprint()], c)
		if _, ok := idx.current[c.ClaimID]; ok {
			currentAmounts[c.LossType] = append(currentAmounts[c.LossType], c.ClaimedAmount)
			if idx.providerCount[c.LossType] == nil {
				idx.providerCount[c.LossType] = map[string]int{}
				idx.providerAmount[c.LossType] = map[string]int64{}
			}
			idx.providerCount[c.LossType][c.ProviderID]++
			idx.providerAmount[c.LossType][c.ProviderID] += c.ClaimedAmount
		}
	}
	for lossType, amounts := range currentAmounts {
		var sum int64
		for _, a := range amounts {
			sum += a
		}
		idx.lossStats[lossType] = lossStat{count: len(amounts), mean: float64(sum) / float64(len(amounts))}
	}
	idx.buildClaimantGraph()
	return idx
}

func (idx *index) buildClaimantGraph() {
	for i := range idx.all {
		attrs := idx.all[i].Identities
		values := make([]string, 0, len(attrs))
		for typ, hash := range attrs {
			if hash == "" {
				continue
			}
			values = append(values, typ+"|"+hash)
		}
		sort.Strings(values)
		for _, v := range values {
			peers := idx.byIdentity[v]
			for _, peer := range peers {
				if peer.ClaimantID == idx.all[i].ClaimantID {
					continue
				}
				if idx.claimantGraph[idx.all[i].ClaimantID] == nil {
					idx.claimantGraph[idx.all[i].ClaimantID] = map[string]struct{}{}
				}
				idx.claimantGraph[idx.all[i].ClaimantID][peer.ClaimantID] = struct{}{}
			}
		}
	}
}

func (idx *index) others(c *ClaimSnapshot) []ClaimSnapshot {
	out := make([]ClaimSnapshot, 0, len(idx.all))
	for _, other := range idx.all {
		if other.ClaimID == c.ClaimID && other.BatchID == c.BatchID {
			continue
		}
		out = append(out, other)
	}
	return out
}

func (idx *index) isCurrent(claimID string) bool {
	_, ok := idx.current[claimID]
	return ok
}

func inWindow(anchor, t time.Time, days int) bool {
	start := anchor.AddDate(0, 0, -days)
	return !t.Before(start) && !t.After(anchor)
}

// withinDays treats rolling windows symmetrically so a pair of claims in the
// same batch is detected regardless of which claim is evaluated first.
func withinDays(a, b time.Time, days int) bool {
	gap := a.Sub(b)
	if gap < 0 {
		gap = -gap
	}
	return gap <= time.Duration(days)*24*time.Hour
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

type rule struct {
	meta     RuleMeta
	evaluate func(ec *EvalContext, idx *index, c *ClaimSnapshot) Outcome
}

// Evaluate runs the catalog deterministically. Current claims are sorted by
// claim_id and rules by catalog order, so identical input yields identical
// output regardless of caller ordering.
func Evaluate(ec EvalContext, rulePackVersion string) (BatchResult, error) {
	if rulePackVersion == "" {
		rulePackVersion = CatalogVersion
	}
	if rulePackVersion != CatalogVersion {
		return BatchResult{}, fmt.Errorf("unknown rule pack version %q (pinned catalog is %q)", rulePackVersion, CatalogVersion)
	}
	configHash, err := ec.configHash()
	if err != nil {
		return BatchResult{}, err
	}
	asOf := ec.AsOf.UTC().Truncate(time.Second)
	if asOf.IsZero() {
		return BatchResult{}, fmt.Errorf("as_of is required")
	}
	idx := buildIndex(ec.History, ec.Current)
	current := append([]ClaimSnapshot(nil), ec.Current...)
	sort.SliceStable(current, func(i, j int) bool { return current[i].ClaimID < current[j].ClaimID })

	rules := catalogRules()
	killSorted := append([]KillSwitch(nil), ec.KillSwitches...)
	sort.SliceStable(killSorted, func(i, j int) bool {
		if killSorted[i].Scope != killSorted[j].Scope {
			return killSorted[i].Scope < killSorted[j].Scope
		}
		return killSorted[i].Target < killSorted[j].Target
	})

	results := make([]ClaimResult, 0, len(current))
	for i := range current {
		c := current[i]
		cr := ClaimResult{ClaimID: c.ClaimID, BatchID: c.BatchID, Flags: []Flag{}, NotEvaluated: []NotEvaluated{}}
		for _, r := range rules {
			if r.meta.Status == StatusGatedOff && !ec.DocForensicsEnabled {
				cr.NotEvaluated = append(cr.NotEvaluated, NotEvaluated{
					RuleID: r.meta.ID, RuleVersion: r.meta.Version, Reason: "GATED_OFF", Detail: "document forensics feature flag is off",
				})
				continue
			}
			if killActive(killSorted, r.meta) {
				cr.NotEvaluated = append(cr.NotEvaluated, NotEvaluated{
					RuleID: r.meta.ID, RuleVersion: r.meta.Version, Reason: "KILL_SWITCH",
				})
				continue
			}
			out := r.evaluate(&ec, idx, &c)
			if out.SkippedReason != "" {
				cr.NotEvaluated = append(cr.NotEvaluated, NotEvaluated{
					RuleID: r.meta.ID, RuleVersion: r.meta.Version, Reason: out.SkippedReason, Detail: out.SkippedDetail,
				})
				continue
			}
			if out.TriggerValue == "" && out.Explanation == "" {
				continue
			}
			refs := cleanRefs(out.EvidenceRefs)
			cr.Flags = append(cr.Flags, Flag{
				FlagID:        flagID(ec.TenantID, c.ClaimID, r.meta.ID, r.meta.Version),
				RuleID:        r.meta.ID,
				RuleVersion:   r.meta.Version,
				RuleName:      r.meta.Name,
				Family:        r.meta.Family,
				Severity:      r.meta.Severity,
				ParamsVersion: r.meta.ParamsVersion,
				TriggerValue:  out.TriggerValue,
				EvidenceRefs:  refs,
				Explanation:   out.Explanation,
				ReasonCode:    r.meta.ReasonCode,
				Suppressed:    out.Suppressed,
			})
		}
		sort.SliceStable(cr.Flags, func(i, j int) bool {
			ri, rj := SeverityRank(cr.Flags[i].Severity), SeverityRank(cr.Flags[j].Severity)
			if ri != rj {
				return ri < rj
			}
			return cr.Flags[i].RuleID < cr.Flags[j].RuleID
		})
		sort.SliceStable(cr.NotEvaluated, func(i, j int) bool { return cr.NotEvaluated[i].RuleID < cr.NotEvaluated[j].RuleID })
		hash, err := events.Hash(struct {
			ClaimID      string         `json:"claim_id"`
			Flags        []Flag         `json:"flags"`
			NotEvaluated []NotEvaluated `json:"not_evaluated"`
		}{cr.ClaimID, cr.Flags, cr.NotEvaluated})
		if err != nil {
			return BatchResult{}, err
		}
		cr.OutputHash = hash
		results = append(results, cr)
	}

	batchHash, err := events.Hash(results)
	if err != nil {
		return BatchResult{}, err
	}
	return BatchResult{
		RulePackVersion: rulePackVersion,
		EngineVersion:   EngineVersion,
		ConfigVersion:   configHash,
		AsOf:            asOf,
		Results:         results,
		OutputHash:      batchHash,
	}, nil
}

func killActive(switches []KillSwitch, meta RuleMeta) bool {
	for _, sw := range switches {
		if !sw.Enabled {
			continue
		}
		switch sw.Scope {
		case "global":
			return true
		case "family":
			if sw.Target == meta.Family {
				return true
			}
		case "rule":
			if sw.Target == meta.ID {
				return true
			}
		}
	}
	return false
}

func flagID(tenantID, claimID, ruleID string, ruleVersion int) string {
	raw := fmt.Sprintf("v1|%s|%s|%s|%d", tenantID, claimID, ruleID, ruleVersion)
	sum := sha256.Sum256([]byte(raw))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func cleanRefs(refs []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	sort.Strings(out)
	return out
}

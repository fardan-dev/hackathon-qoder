package rulesengine

import (
	"sort"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/pii"
)

const (
	CatalogVersion = "2026.11.1"
	EngineVersion  = "rulesengine-1.0.0"
)

type Severity string

const (
	SeverityHigh   Severity = "High"
	SeverityMedium Severity = "Medium"
	SeverityLow    Severity = "Low"
)

func SeverityRank(s Severity) int {
	switch s {
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	default:
		return 3
	}
}

type RuleStatus string

const (
	StatusActive   RuleStatus = "ACTIVE"
	StatusGatedOff RuleStatus = "GATED_OFF"
)

type RuleMeta struct {
	ID            string         `json:"rule_id"`
	Version       int            `json:"version"`
	Family        string         `json:"family"`
	Name          string         `json:"name"`
	Severity      Severity       `json:"severity"`
	Params        map[string]any `json:"params"`
	ParamsVersion int            `json:"params_version"`
	Status        RuleStatus     `json:"status"`
	Owner         string         `json:"owner"`
	Typologies    []string       `json:"mapping_typology"`
	ReasonCode    string         `json:"reason_code"`
}

type Flag struct {
	FlagID        string   `json:"flag_id"`
	RuleID        string   `json:"rule_id"`
	RuleVersion   int      `json:"rule_version"`
	RuleName      string   `json:"rule_name"`
	Family        string   `json:"family"`
	Severity      Severity `json:"severity"`
	ParamsVersion int      `json:"params_version"`
	TriggerValue  string   `json:"trigger_value"`
	EvidenceRefs  []string `json:"evidence_refs"`
	Explanation   string   `json:"explanation"`
	ReasonCode    string   `json:"reason_code"`
	Suppressed    bool     `json:"suppressed"`
}

type NotEvaluated struct {
	RuleID      string `json:"rule_id"`
	RuleVersion int    `json:"rule_version"`
	Reason      string `json:"reason"`
	Detail      string `json:"detail,omitempty"`
}

type ClaimResult struct {
	ClaimID      string         `json:"claim_id"`
	BatchID      string         `json:"batch_id"`
	Flags        []Flag         `json:"flags"`
	NotEvaluated []NotEvaluated `json:"not_evaluated"`
	OutputHash   string         `json:"output_hash"`
}

type BatchResult struct {
	RulePackVersion string        `json:"rule_pack_version"`
	EngineVersion   string        `json:"engine_version"`
	ConfigVersion   string        `json:"config_version"`
	AsOf            time.Time     `json:"as_of"`
	Results         []ClaimResult `json:"results"`
	OutputHash      string        `json:"output_hash"`
}

type InvoiceSnapshot struct {
	Number     string    `json:"invoice_number"`
	Date       time.Time `json:"invoice_date"`
	ProviderID string    `json:"provider_id"`
	Amount     int64     `json:"invoice_amount"`
	ClaimID    string    `json:"claim_id"`
	DocSHA256  string    `json:"doc_sha256,omitempty"`
}

type PolicySnapshot struct {
	PolicyID      string     `json:"policy_id"`
	Product       string     `json:"product"`
	LossTypeScope string     `json:"loss_type_scope"`
	Status        string     `json:"status"`
	Inception     time.Time  `json:"inception_date"`
	CoverageAdded *time.Time `json:"coverage_added_date,omitempty"`
	Expiry        time.Time  `json:"expiry_date"`
	Reinstatement *time.Time `json:"reinstatement_date,omitempty"`
	SumInsured    int64      `json:"sum_insured"`
	Premium       int64      `json:"premium"`
	AssetID       string     `json:"asset_id,omitempty"`
}

type ProviderSnapshot struct {
	ProviderID string `json:"provider_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Region     string `json:"region"`
}

// ClaimSnapshot is the only data the rule engine may read. Ground-truth labels
// are intentionally absent; a leakage test enforces this boundary.
type ClaimSnapshot struct {
	TenantID      string            `json:"tenant_id"`
	ClaimID       string            `json:"claim_id"`
	BatchID       string            `json:"batch_id"`
	PolicyID      string            `json:"policy_id"`
	ClaimantID    string            `json:"claimant_id"`
	ProviderID    string            `json:"provider_id"`
	InvoiceNumber string            `json:"invoice_number,omitempty"`
	LossType      string            `json:"loss_type"`
	Status        string            `json:"status"`
	RepID         string            `json:"rep_id,omitempty"`
	ClaimDate     time.Time         `json:"claim_date"`
	IncidentDate  time.Time         `json:"incident_date"`
	ClaimedAmount int64             `json:"claimed_amount"`
	Currency      string            `json:"currency"`
	Invoice       *InvoiceSnapshot  `json:"invoice,omitempty"`
	Policy        *PolicySnapshot   `json:"policy,omitempty"`
	Provider      *ProviderSnapshot `json:"provider,omitempty"`
	Identities    map[string]string `json:"identity_attributes,omitempty"`
	DocSHA256     string            `json:"doc_sha256,omitempty"`

	PriorNeedsInfo     int    `json:"prior_needs_info,omitempty"`
	PriorClearedReason string `json:"prior_cleared_reason,omitempty"`
	PriorConfirmed     bool   `json:"prior_confirmed,omitempty"`
}

func (c ClaimSnapshot) Fingerprint() string {
	return pii.Fingerprint(c.TenantID, c.ClaimantID, c.IncidentDate.Format("2006-01-02"), c.ClaimedAmount, c.ProviderID)
}

type KillSwitch struct {
	Scope   string `json:"scope"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

type EvalContext struct {
	TenantID            string
	AsOf                time.Time
	Current             []ClaimSnapshot
	History             []ClaimSnapshot
	DocForensicsEnabled bool
	KillSwitches        []KillSwitch
}

type Outcome struct {
	TriggerValue  string
	EvidenceRefs  []string
	Explanation   string
	SkippedReason string
	SkippedDetail string
	Suppressed    bool
}

func (ec *EvalContext) configHash() (string, error) {
	switches := append([]KillSwitch(nil), ec.KillSwitches...)
	sort.SliceStable(switches, func(i, j int) bool {
		if switches[i].Scope != switches[j].Scope {
			return switches[i].Scope < switches[j].Scope
		}
		if switches[i].Target != switches[j].Target {
			return switches[i].Target < switches[j].Target
		}
		return !switches[i].Enabled && switches[j].Enabled
	})
	return events.Hash(struct {
		RulePack string       `json:"rule_pack"`
		Engine   string       `json:"engine"`
		Docs     bool         `json:"doc_forensics"`
		Switches []KillSwitch `json:"kill_switches"`
	}{CatalogVersion, EngineVersion, ec.DocForensicsEnabled, switches})
}

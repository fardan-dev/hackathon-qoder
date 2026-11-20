package pipeline

import (
	"os"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/rulesengine"
)

const (
	TopicClaimIngested       = "claim.ingested"
	TopicBatchIngested       = "batch.ingested"
	TopicClaimEvaluated      = "claim.evaluated"
	TopicAlertCreated        = "alert.created"
	TopicCaseDecided         = "case.decided"
	TopicSuppressionApproved = "suppression.approved"
	TopicSuppressionRevoked  = "suppression.revoked"
	TopicKillSwitchChanged   = "rule.kill_switch"
	TopicBillingExported     = "billing.exported"
)

// Topic appends the deployment's schema suffix (for example "-v1") so a
// breaking payload change can run alongside the old topic.
func Topic(base string) string {
	suffix := os.Getenv("KAFKA_TOPIC_SUFFIX")
	if suffix == "" {
		return base
	}
	if !strings.HasPrefix(suffix, "-") {
		suffix = "-" + suffix
	}
	return base + suffix
}

type ClaimIngested struct {
	BatchID       string    `json:"batch_id"`
	ClaimID       string    `json:"claim_id"`
	ClaimantID    string    `json:"claimant_id"`
	PolicyID      string    `json:"policy_id"`
	ProviderID    string    `json:"provider_id"`
	LossType      string    `json:"loss_type"`
	ClaimStatus   string    `json:"claim_status"`
	ClaimDate     time.Time `json:"claim_date"`
	IncidentDate  time.Time `json:"incident_date"`
	ClaimedAmount int64     `json:"claimed_amount"`
	Actor         string    `json:"actor"`
}

type BatchIngested struct {
	BatchID    string `json:"batch_id"`
	ClaimCount int    `json:"claim_count"`
	Actor      string `json:"actor"`
}

type ClaimEvaluated struct {
	BatchID         string                     `json:"batch_id"`
	ClaimID         string                     `json:"claim_id"`
	ClaimantID      string                     `json:"claimant_id"`
	PolicyID        string                     `json:"policy_id"`
	ProviderID      string                     `json:"provider_id"`
	LossType        string                     `json:"loss_type"`
	ClaimStatus     string                     `json:"claim_status"`
	ClaimDate       time.Time                  `json:"claim_date"`
	IncidentDate    time.Time                  `json:"incident_date"`
	ClaimedAmount   int64                      `json:"claimed_amount"`
	Fingerprint     string                     `json:"fingerprint"`
	Severity        rulesengine.Severity       `json:"severity"`
	Flags           []rulesengine.Flag         `json:"flags"`
	NotEvaluated    []rulesengine.NotEvaluated `json:"not_evaluated"`
	RulePackVersion string                     `json:"rule_pack_version"`
	EngineVersion   string                     `json:"engine_version"`
	ConfigVersion   string                     `json:"config_version"`
	AsOf            time.Time                  `json:"as_of"`
	OutputHash      string                     `json:"output_hash"`
	Actor           string                     `json:"actor"`
}

type AlertCreated struct {
	AlertID       string               `json:"alert_id"`
	ClaimID       string               `json:"claim_id"`
	BatchID       string               `json:"batch_id"`
	LossType      string               `json:"loss_type"`
	Severity      rulesengine.Severity `json:"severity"`
	SeverityRank  int                  `json:"severity_rank"`
	FlagCount     int                  `json:"flag_count"`
	ClaimedAmount int64                `json:"claimed_amount"`
	Fingerprint   string               `json:"fingerprint"`
	Flags         []rulesengine.Flag   `json:"flags"`
	Suppressed    bool                 `json:"suppressed"`
	SuppressionID string               `json:"suppression_id,omitempty"`
	SLADueAt      time.Time            `json:"sla_due_at"`
	CreatedAt     time.Time            `json:"created_at"`
	Actor         string               `json:"actor"`
}

type CaseDecided struct {
	CaseID         string    `json:"case_id"`
	ClaimIDs       []string  `json:"claim_ids"`
	Outcome        string    `json:"outcome"`
	ReasonCode     string    `json:"reason_code"`
	ReviewerID     string    `json:"reviewer_id"`
	ReviewerName   string    `json:"reviewer_name"`
	Notes          string    `json:"notes"`
	ValueProtected int64     `json:"value_protected"`
	ValueRecovered int64     `json:"value_recovered"`
	DecidedAt      time.Time `json:"decided_at"`
	Actor          string    `json:"actor"`
}

type SuppressionApproved struct {
	SuppressionID string    `json:"suppression_id"`
	ClaimID       string    `json:"claim_id"`
	Fingerprint   string    `json:"fingerprint"`
	Reason        string    `json:"reason"`
	ApprovedBy    string    `json:"approved_by"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	Actor         string    `json:"actor"`
}

type SuppressionRevoked struct {
	SuppressionID string    `json:"suppression_id"`
	ClaimID       string    `json:"claim_id"`
	Fingerprint   string    `json:"fingerprint"`
	RevokedBy     string    `json:"revoked_by"`
	RevokedAt     time.Time `json:"revoked_at"`
	Actor         string    `json:"actor"`
}

type KillSwitchChanged struct {
	SwitchID  string    `json:"switch_id"`
	Scope     string    `json:"scope"`
	Target    string    `json:"target"`
	Enabled   bool      `json:"enabled"`
	Reason    string    `json:"reason,omitempty"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
	Actor     string    `json:"actor"`
}

type BillingExported struct {
	ExportID  string    `json:"export_id"`
	Period    string    `json:"period"`
	Version   string    `json:"version"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	Actor     string    `json:"actor"`
}

// BatchClaims is the ingestion internal API response consumed by the rules
// service. It contains only engine input, never ground-truth labels.
type BatchClaims struct {
	BatchID string                      `json:"batch_id"`
	Claims  []rulesengine.ClaimSnapshot `json:"claims"`
}

package contracts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SchemaVersion is pinned per evaluation and stored with import batches.
const SchemaVersion = "phase01-v1"

type Date struct {
	time.Time
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return Date{}, fmt.Errorf("expected YYYY-MM-DD")
	}
	return Date{Time: t}, nil
}

func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02")
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.String())
}

func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

type Claim struct {
	Row           int    `json:"-"`
	ClaimID       string `json:"claim_id"`
	PolicyID      string `json:"policy_id"`
	ClaimantID    string `json:"claimant_id"`
	ClaimDate     Date   `json:"claim_date"`
	IncidentDate  Date   `json:"incident_date"`
	LossType      string `json:"loss_type"`
	ClaimedAmount int64  `json:"claimed_amount"`
	Currency      string `json:"currency"`
	ProviderID    string `json:"provider_id"`
	InvoiceNumber string `json:"invoice_number,omitempty"`
	Status        string `json:"status"`
	RepID         string `json:"rep_id,omitempty"`
}

type Invoice struct {
	Row           int    `json:"-"`
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   Date   `json:"invoice_date"`
	ProviderID    string `json:"provider_id"`
	InvoiceAmount int64  `json:"invoice_amount"`
	Currency      string `json:"currency"`
	ClaimID       string `json:"claim_id"`
	DocSHA256     string `json:"doc_sha256,omitempty"`
}

type Policy struct {
	Row           int    `json:"-"`
	PolicyID          string `json:"policy_id"`
	Product           string `json:"product"`
	LossTypeScope     string `json:"loss_type_scope"`
	InceptionDate     Date   `json:"inception_date"`
	CoverageAddedDate *Date  `json:"coverage_added_date,omitempty"`
	ExpiryDate        Date   `json:"expiry_date"`
	Status            string `json:"status"`
	SumInsured        int64  `json:"sum_insured"`
	Premium           int64  `json:"premium"`
	ReinstatementDate *Date  `json:"reinstatement_date,omitempty"`
	AssetID           string `json:"asset_id,omitempty"`
	AgentID           string `json:"agent_id,omitempty"`
}

type Provider struct {
	Row           int    `json:"-"`
	ProviderID string `json:"provider_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Region     string `json:"region"`
}

type IdentityAttribute struct {
	Row           int    `json:"-"`
	ClaimantID    string `json:"claimant_id"`
	AttributeType string `json:"attribute_type"`
	AttributeHash string `json:"attribute_hash"`
}

type Source struct {
	File   string
	SHA256 string
	Rows   int
}

type Batch struct {
	Version    string
	Claims     []Claim
	Invoices   []Invoice
	Policies   []Policy
	Providers  []Provider
	Identities []IdentityAttribute
	Sources    map[string]Source
}

func (b Batch) TotalRows() int {
	return len(b.Claims) + len(b.Invoices) + len(b.Policies) + len(b.Providers) + len(b.Identities)
}

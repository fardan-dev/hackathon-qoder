package contracts

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const (
	policyCSV = `policy_id,product,loss_type_scope,inception_date,coverage_added_date,expiry_date,status,sum_insured,premium,reinstatement_date,asset_id,agent_id
POL-1,MOTOR_COMP,MOTOR,2026-01-15,2026-01-15,2027-01-14,ACTIVE,150000000,4200000,,AST-1,AGT-1
`
	providerCSV = `provider_id,name,type,region
P-1,Bengkel Jaya,WORKSHOP,Bandung
`
	claimCSV = `claim_id,policy_id,claimant_id,claim_date,incident_date,loss_type,claimed_amount,currency,provider_id,invoice_number,status,rep_id
CLM-1,POL-1,C-1,2026-09-14,2026-09-10,MOTOR,18500000,IDR,P-1,INV-1,SUBMITTED,REP-1
`
	invoiceCSV = `invoice_number,invoice_date,provider_id,invoice_amount,currency,claim_id,doc_sha256
INV-1,2026-09-11,P-1,18500000,IDR,CLM-1,sha256:9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c9f2c
`
	identityCSV = `claimant_id,attribute_type,attribute_hash
C-1,BANK_ACCOUNT,sha256:ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12ab12
`
)

func validBatch(t *testing.T) Batch {
	t.Helper()
	var batch Batch
	batch.Version = SchemaVersion
	batch.Sources = map[string]Source{}
	for ds, raw := range map[string]string{
		DatasetPolicies: policyCSV, DatasetProviders: providerCSV, DatasetClaims: claimCSV,
		DatasetInvoices: invoiceCSV, DatasetIdentities: identityCSV,
	} {
		b, errs := ParseFile(ds, ds+".csv", []byte(raw))
		if len(errs) > 0 {
			t.Fatalf("fixture %s parse errors: %v", ds, errs)
		}
		batch.Policies = append(batch.Policies, b.Policies...)
		batch.Providers = append(batch.Providers, b.Providers...)
		batch.Claims = append(batch.Claims, b.Claims...)
		batch.Invoices = append(batch.Invoices, b.Invoices...)
		batch.Identities = append(batch.Identities, b.Identities...)
	}
	return batch
}

func TestParseAndValidateValidBatch(t *testing.T) {
	b := validBatch(t)
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if b.Claims[0].ClaimedAmount != 18500000 || b.Claims[0].Row != 2 {
		t.Fatalf("unexpected claim parse: %+v", b.Claims[0])
	}
}

func TestRejectOutOfScopeLOB(t *testing.T) {
	raw := strings.Replace(claimCSV, ",MOTOR,18500000", ",HEALTH,18500000", 1)
	_, errs := ParseFile(DatasetClaims, "claims.csv", []byte(raw))
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	b := validBatch(t)
	b.Claims[0].LossType = "HEALTH"
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(errs, CodeOutOfScopeLOB) {
		t.Fatalf("expected OUT_OF_SCOPE_LOB, got %v", errs)
	}
}

func TestRejectUnhashedIdentity(t *testing.T) {
	b := validBatch(t)
	b.Identities[0].AttributeHash = "3171234567890001"
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(errs, CodeUnhashedIdentity) {
		t.Fatalf("expected UNHASHED_IDENTITY, got %v", errs)
	}
}

func TestRejectMissingPolicyReference(t *testing.T) {
	b := validBatch(t)
	b.Policies = nil
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(errs, CodeReferential) {
		t.Fatalf("expected REFERENTIAL, got %v", errs)
	}
}

func TestRejectDateOrder(t *testing.T) {
	b := validBatch(t)
	b.Claims[0].IncidentDate = b.Claims[0].ClaimDate
	b.Claims[0].IncidentDate = mustDate(t, "2026-09-20")
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(errs, CodeDateOrder) {
		t.Fatalf("expected DATE_ORDER, got %v", errs)
	}
}

func TestCSVUnknownColumnRejected(t *testing.T) {
	raw := "claim_id,policy_id,claimant_id,claim_date,incident_date,loss_type,claimed_amount,currency,provider_id,status,raw_nik\n" +
		"CLM-1,POL-1,C-1,2026-09-14,2026-09-10,MOTOR,1,IDR,P-1,SUBMITTED,3171234567890001\n"
	_, errs := ParseFile(DatasetClaims, "claims.csv", []byte(raw))
	if !hasCode(errs, CodeUnknownField) {
		t.Fatalf("expected UNKNOWN_FIELD, got %v", errs)
	}
}

func TestDuplicateClaimInBatchRejected(t *testing.T) {
	b := validBatch(t)
	b.Claims = append(b.Claims, b.Claims[0])
	errs, err := Validate(context.Background(), b, NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(errs, CodeDuplicateKey) {
		t.Fatalf("expected DUPLICATE_KEY, got %v", errs)
	}
}

func TestJSONParse(t *testing.T) {
	raw := `[{"claim_id":"CLM-9","policy_id":"POL-9","claimant_id":"C-9","claim_date":"2026-09-14","incident_date":"2026-09-10","loss_type":"PROPERTY","claimed_amount":1000000,"currency":"IDR","provider_id":"P-9","status":"SUBMITTED"}]`
	b, errs := ParseFile(DatasetClaims, "claims.json", []byte(raw))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if len(b.Claims) != 1 || b.Claims[0].ClaimID != "CLM-9" || b.Claims[0].Row != 1 {
		t.Fatalf("unexpected JSON parse: %+v", b.Claims)
	}
}

func TestEscapeCSVCell(t *testing.T) {
	for _, in := range []string{"=cmd()", "+1", "-1", "@x"} {
		if got := EscapeCSVCell(in); !strings.HasPrefix(got, "'") {
			t.Fatalf("%q not escaped: %q", in, got)
		}
	}
	if got := EscapeCSVCell("safe"); got != "safe" {
		t.Fatalf("safe value changed: %q", got)
	}
}

func TestLookupErrorPropagates(t *testing.T) {
	b := validBatch(t)
	b.Policies = nil
	_, err := Validate(context.Background(), b, failingLookup{})
	if err == nil {
		t.Fatal("expected lookup error to abort validation")
	}
}

type failingLookup struct{}

func (failingLookup) PolicyExists(context.Context, string) (bool, error) {
	return false, fmt.Errorf("db down")
}
func (failingLookup) ProviderExists(context.Context, string) (bool, error) { return false, nil }
func (failingLookup) InvoiceExists(context.Context, string) (bool, error)  { return false, nil }
func (failingLookup) ClaimExists(context.Context, string) (bool, error)    { return false, nil }

func hasCode(errs []RowError, code string) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

func mustDate(t *testing.T, s string) Date {
	t.Helper()
	d, err := ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

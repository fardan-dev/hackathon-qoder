package contracts

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	CodeRequired         = "REQUIRED_FIELD"
	CodeInvalidType      = "INVALID_TYPE"
	CodeInvalidEnum      = "INVALID_ENUM"
	CodeOutOfScopeLOB    = "OUT_OF_SCOPE_LOB"
	CodeInvalidDate      = "INVALID_DATE"
	CodeDateOrder        = "DATE_ORDER"
	CodeDuplicateKey     = "DUPLICATE_KEY"
	CodeReferential      = "REFERENTIAL"
	CodeUnhashedIdentity = "UNHASHED_IDENTITY"
	CodeInvalidHash      = "INVALID_HASH"
	CodeCurrencyMismatch = "CURRENCY_MISMATCH"
	CodeUnknownField     = "UNKNOWN_FIELD"
	CodeMissingHeader    = "MISSING_HEADER"
	CodeDuplicateHeader  = "DUPLICATE_HEADER"
	CodeEmptyFile        = "EMPTY_FILE"
	CodeInvalidCSV       = "INVALID_CSV"
	CodeInvalidJSON      = "INVALID_JSON"
	CodeUnknownDataset   = "UNKNOWN_DATASET"
	CodeInvalidAmount    = "INVALID_AMOUNT"
	CodeInvalidStatus    = "INVALID_STATUS"
	CodePolicyScope      = "POLICY_SCOPE"
)

var (
	statusRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,31}$`)
	hashRe   = regexp.MustCompile(`^(sha256|hmac-sha256):[0-9a-f]{64}$`)
)

type RowError struct {
	Dataset string `json:"dataset"`
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e RowError) Error() string {
	return fmt.Sprintf("%s row %d field %s: %s (%s)", e.Dataset, e.Row, e.Field, e.Message, e.Code)
}

type Lookup interface {
	PolicyExists(ctx context.Context, policyID string) (bool, error)
	ProviderExists(ctx context.Context, providerID string) (bool, error)
	InvoiceExists(ctx context.Context, invoiceNumber string) (bool, error)
	ClaimExists(ctx context.Context, claimID string) (bool, error)
}

type NopLookup struct{}

func (NopLookup) PolicyExists(context.Context, string) (bool, error)   { return false, nil }
func (NopLookup) ProviderExists(context.Context, string) (bool, error) { return false, nil }
func (NopLookup) InvoiceExists(context.Context, string) (bool, error)  { return false, nil }
func (NopLookup) ClaimExists(context.Context, string) (bool, error)    { return false, nil }

// Validate applies the Phase 0/1 data contract. Any error rejects the whole
// batch (fail-fast); callers must not insert valid rows from a rejected batch.
func Validate(ctx context.Context, b Batch, lookup Lookup) ([]RowError, error) {
	if lookup == nil {
		lookup = NopLookup{}
	}
	var errs []RowError

	policyIDs := map[string]struct{}{}
	policies := map[string]Policy{}
	for _, p := range b.Policies {
		row := rowOf(p.Row)
		if err := requireFields(row, DatasetPolicies, map[string]string{
			"policy_id": p.PolicyID, "product": p.Product, "loss_type_scope": p.LossTypeScope,
			"status": p.Status,
		}, &errs); err != nil {
			continue
		}
		if !validLOB(p.LossTypeScope) {
			errs = append(errs, errRow(DatasetPolicies, row, "loss_type_scope", CodeOutOfScopeLOB, "thin slice accepts MOTOR and PROPERTY only"))
		}
		if p.InceptionDate.IsZero() || p.ExpiryDate.IsZero() {
			errs = append(errs, errRow(DatasetPolicies, row, "inception_date", CodeRequired, "inception_date and expiry_date are required"))
		} else if !p.InceptionDate.Before(p.ExpiryDate.Time) {
			errs = append(errs, errRow(DatasetPolicies, row, "expiry_date", CodeDateOrder, "inception_date must be before expiry_date"))
		}
		if !statusRe.MatchString(p.Status) {
			errs = append(errs, errRow(DatasetPolicies, row, "status", CodeInvalidStatus, "status must be an uppercase code"))
		}
		if p.SumInsured <= 0 {
			errs = append(errs, errRow(DatasetPolicies, row, "sum_insured", CodeInvalidAmount, "sum_insured must be positive"))
		}
		if p.Premium <= 0 {
			errs = append(errs, errRow(DatasetPolicies, row, "premium", CodeInvalidAmount, "premium must be positive"))
		}
		if _, dup := policyIDs[p.PolicyID]; dup {
			errs = append(errs, errRow(DatasetPolicies, row, "policy_id", CodeDuplicateKey, "policy_id appears more than once in the batch"))
		}
		policyIDs[p.PolicyID] = struct{}{}
		policies[p.PolicyID] = p
	}

	providerIDs := map[string]struct{}{}
	for _, p := range b.Providers {
		row := rowOf(p.Row)
		if err := requireFields(row, DatasetProviders, map[string]string{
			"provider_id": p.ProviderID, "name": p.Name, "type": p.Type, "region": p.Region,
		}, &errs); err != nil {
			continue
		}
		if _, dup := providerIDs[p.ProviderID]; dup {
			errs = append(errs, errRow(DatasetProviders, row, "provider_id", CodeDuplicateKey, "provider_id appears more than once in the batch"))
		}
		providerIDs[p.ProviderID] = struct{}{}
	}

	invoiceNumbers := map[string]struct{}{}
	seenInvoices := map[string]struct{}{}
	for _, in := range b.Invoices {
		row := rowOf(in.Row)
		if err := requireFields(row, DatasetInvoices, map[string]string{
			"invoice_number": in.InvoiceNumber, "provider_id": in.ProviderID,
			"currency": in.Currency, "claim_id": in.ClaimID,
		}, &errs); err != nil {
			continue
		}
		if in.InvoiceDate.IsZero() {
			errs = append(errs, errRow(DatasetInvoices, row, "invoice_date", CodeRequired, "invoice_date is required"))
		}
		if in.InvoiceAmount <= 0 {
			errs = append(errs, errRow(DatasetInvoices, row, "invoice_amount", CodeInvalidAmount, "invoice_amount must be positive"))
		}
		if in.Currency != "IDR" {
			errs = append(errs, errRow(DatasetInvoices, row, "currency", CodeCurrencyMismatch, "thin slice accepts IDR only"))
		}
		if in.DocSHA256 != "" && !hashRe.MatchString(in.DocSHA256) {
			errs = append(errs, errRow(DatasetInvoices, row, "doc_sha256", CodeInvalidHash, "doc_sha256 must be sha256:<64 hex>"))
		}
		key := in.InvoiceNumber + "|" + in.ProviderID
		if _, dup := seenInvoices[key]; dup {
			errs = append(errs, errRow(DatasetInvoices, row, "invoice_number", CodeDuplicateKey, "invoice_number/provider_id appears more than once in the batch"))
		}
		seenInvoices[key] = struct{}{}
		invoiceNumbers[in.InvoiceNumber] = struct{}{}
	}

	claimIDs := map[string]struct{}{}
	claims := map[string]Claim{}
	for _, c := range b.Claims {
		row := rowOf(c.Row)
		if err := requireFields(row, DatasetClaims, map[string]string{
			"claim_id": c.ClaimID, "policy_id": c.PolicyID, "claimant_id": c.ClaimantID,
			"loss_type": c.LossType, "currency": c.Currency, "provider_id": c.ProviderID, "status": c.Status,
		}, &errs); err != nil {
			continue
		}
		if !validLOB(c.LossType) {
			errs = append(errs, errRow(DatasetClaims, row, "loss_type", CodeOutOfScopeLOB, "thin slice accepts MOTOR and PROPERTY only"))
		}
		if c.Currency != "IDR" {
			errs = append(errs, errRow(DatasetClaims, row, "currency", CodeCurrencyMismatch, "thin slice accepts IDR only"))
		}
		if c.ClaimedAmount <= 0 {
			errs = append(errs, errRow(DatasetClaims, row, "claimed_amount", CodeInvalidAmount, "claimed_amount must be positive"))
		}
		if c.ClaimDate.IsZero() || c.IncidentDate.IsZero() {
			errs = append(errs, errRow(DatasetClaims, row, "claim_date", CodeRequired, "claim_date and incident_date are required"))
		} else if c.IncidentDate.After(c.ClaimDate.Time) {
			errs = append(errs, errRow(DatasetClaims, row, "incident_date", CodeDateOrder, "incident_date must be on or before claim_date"))
		}
		if !statusRe.MatchString(c.Status) {
			errs = append(errs, errRow(DatasetClaims, row, "status", CodeInvalidStatus, "status must be an uppercase code"))
		}
		if _, dup := claimIDs[c.ClaimID]; dup {
			errs = append(errs, errRow(DatasetClaims, row, "claim_id", CodeDuplicateKey, "claim_id appears more than once in the batch"))
		}
		claimIDs[c.ClaimID] = struct{}{}
		claims[c.ClaimID] = c

		if p, ok := policies[c.PolicyID]; ok {
			if p.LossTypeScope != "" && p.LossTypeScope != c.LossType {
				errs = append(errs, errRow(DatasetClaims, row, "policy_id", CodePolicyScope, "policy loss_type_scope does not match the claim"))
			}
		} else if ok, err := lookup.PolicyExists(ctx, c.PolicyID); err != nil {
			return nil, err
		} else if !ok {
			errs = append(errs, errRow(DatasetClaims, row, "policy_id", CodeReferential, "policy_id not found in the batch or master data"))
		}
		if _, ok := providerIDs[c.ProviderID]; !ok {
			if ok, err := lookup.ProviderExists(ctx, c.ProviderID); err != nil {
				return nil, err
			} else if !ok {
				errs = append(errs, errRow(DatasetClaims, row, "provider_id", CodeReferential, "provider_id not found in the batch or master data"))
			}
		}
		if c.InvoiceNumber != "" {
			if _, ok := invoiceNumbers[c.InvoiceNumber]; !ok {
				if ok, err := lookup.InvoiceExists(ctx, c.InvoiceNumber); err != nil {
					return nil, err
				} else if !ok {
					errs = append(errs, errRow(DatasetClaims, row, "invoice_number", CodeReferential, "invoice_number has no matching invoice record"))
				}
			}
		}
	}

	for _, in := range b.Invoices {
		row := rowOf(in.Row)
		if _, ok := claims[in.ClaimID]; !ok {
			if ok, err := lookup.ClaimExists(ctx, in.ClaimID); err != nil {
				return nil, err
			} else if !ok {
				errs = append(errs, errRow(DatasetInvoices, row, "claim_id", CodeReferential, "claim_id not found in the batch or master data"))
			}
		}
		if _, ok := providerIDs[in.ProviderID]; !ok {
			if ok, err := lookup.ProviderExists(ctx, in.ProviderID); err != nil {
				return nil, err
			} else if !ok {
				errs = append(errs, errRow(DatasetInvoices, row, "provider_id", CodeReferential, "provider_id not found in the batch or master data"))
			}
		}
	}

	seenIdentity := map[string]struct{}{}
	for _, ia := range b.Identities {
		row := rowOf(ia.Row)
		if err := requireFields(row, DatasetIdentities, map[string]string{
			"claimant_id": ia.ClaimantID, "attribute_type": ia.AttributeType, "attribute_hash": ia.AttributeHash,
		}, &errs); err != nil {
			continue
		}
		if !validAttributeType(ia.AttributeType) {
			errs = append(errs, errRow(DatasetIdentities, row, "attribute_type", CodeInvalidEnum, "attribute_type is not supported"))
		}
		if !isHashed(ia.AttributeHash) {
			errs = append(errs, errRow(DatasetIdentities, row, "attribute_hash", CodeUnhashedIdentity, "identity attributes must be sha256: or hmac-sha256: hashes; raw NIK is rejected"))
		}
		key := ia.ClaimantID + "|" + ia.AttributeType
		if _, dup := seenIdentity[key]; dup {
			errs = append(errs, errRow(DatasetIdentities, row, "attribute_type", CodeDuplicateKey, "claimant_id/attribute_type appears more than once in the batch"))
		}
		seenIdentity[key] = struct{}{}
	}

	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].Dataset != errs[j].Dataset {
			return datasetOrder(errs[i].Dataset) < datasetOrder(errs[j].Dataset)
		}
		if errs[i].Row != errs[j].Row {
			return errs[i].Row < errs[j].Row
		}
		if errs[i].Field != errs[j].Field {
			return errs[i].Field < errs[j].Field
		}
		return errs[i].Code < errs[j].Code
	})
	return errs, nil
}

func requireFields(row int, dataset string, fields map[string]string, errs *[]RowError) error {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	var failed bool
	for _, name := range names {
		if strings.TrimSpace(fields[name]) == "" {
			*errs = append(*errs, errRow(dataset, row, name, CodeRequired, name+" is required"))
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("missing fields")
	}
	return nil
}

func errRow(dataset string, row int, field, code, message string) RowError {
	return RowError{Dataset: dataset, Row: row, Field: field, Code: code, Message: message}
}

func rowOf(v int) int {
	if v <= 0 {
		return 0
	}
	return v
}

func validLOB(v string) bool { return v == "MOTOR" || v == "PROPERTY" }

func validAttributeType(v string) bool {
	switch v {
	case "BANK_ACCOUNT", "PHONE", "EMAIL", "ADDRESS", "NIK":
		return true
	}
	return false
}

func isHashed(v string) bool { return hashRe.MatchString(strings.ToLower(v)) }

func datasetOrder(ds string) int {
	switch ds {
	case DatasetPolicies:
		return 0
	case DatasetProviders:
		return 1
	case DatasetClaims:
		return 2
	case DatasetInvoices:
		return 3
	case DatasetIdentities:
		return 4
	}
	return 9
}

// EscapeCSVCell prevents spreadsheet formula injection on exports.
func EscapeCSVCell(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + v
	}
	return v
}

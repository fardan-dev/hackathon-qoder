package contracts

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	DatasetClaims     = "claims"
	DatasetInvoices   = "invoices"
	DatasetPolicies   = "policies"
	DatasetProviders  = "providers"
	DatasetIdentities = "identities"
)

var requiredColumns = map[string][]string{
	DatasetClaims:     {"claim_id", "policy_id", "claimant_id", "claim_date", "incident_date", "loss_type", "claimed_amount", "currency", "provider_id", "status"},
	DatasetInvoices:   {"invoice_number", "invoice_date", "provider_id", "invoice_amount", "currency", "claim_id"},
	DatasetPolicies:   {"policy_id", "product", "loss_type_scope", "inception_date", "expiry_date", "status", "sum_insured", "premium"},
	DatasetProviders:  {"provider_id", "name", "type", "region"},
	DatasetIdentities: {"claimant_id", "attribute_type", "attribute_hash"},
}

var optionalColumns = map[string]map[string]struct{}{
	DatasetClaims:     {"invoice_number": {}, "rep_id": {}},
	DatasetInvoices:   {"doc_sha256": {}},
	DatasetPolicies:   {"coverage_added_date": {}, "reinstatement_date": {}, "asset_id": {}, "agent_id": {}},
	DatasetProviders:  {},
	DatasetIdentities: {},
}

var aliases = map[string]string{
	"claim": "claims", "claims.csv": "claims", "claims.json": "claims",
	"invoice": "invoices", "invoices.csv": "invoices", "invoices.json": "invoices",
	"policy": "policies", "policies.csv": "policies", "policies.json": "policies",
	"provider": "providers", "providers.csv": "providers", "providers.json": "providers",
	"identity": "identities", "identities.csv": "identities", "identities.json": "identities",
}

// DatasetForFile resolves a filename to a dataset. The filename stem must start
// with claims/invoices/policies/providers/identities.
func DatasetForFile(name string) (string, bool) {
	base := strings.ToLower(strings.TrimSpace(name))
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	if ds, ok := aliases[base]; ok {
		return ds, true
	}
	for _, ds := range []string{DatasetClaims, DatasetInvoices, DatasetPolicies, DatasetProviders, DatasetIdentities} {
		if strings.HasPrefix(base, ds) {
			return ds, true
		}
	}
	return "", false
}

// ParseFile parses one uploaded file into batch records. Errors are returned as
// row errors so the batch can be rejected with a per-row report.
func ParseFile(dataset, filename string, data []byte) (Batch, []RowError) {
	var b Batch
	b.Version = SchemaVersion
	b.Sources = map[string]Source{dataset: {File: filename, Rows: 0}}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeEmptyFile, Message: "file is empty"}}
	}
	if isJSON(filename, trimmed) {
		return parseJSON(dataset, filename, trimmed)
	}
	return parseCSV(dataset, filename, trimmed)
}

func isJSON(filename string, data []byte) bool {
	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return true
	}
	return len(data) > 0 && (data[0] == '[' || data[0] == '{')
}

func parseJSON(dataset, filename string, data []byte) (Batch, []RowError) {
	b := Batch{Version: SchemaVersion, Sources: map[string]Source{dataset: {File: filename}}}
	trimmed := bytes.TrimSpace(data)
	if trimmed[0] == '{' {
		var wrapper struct {
			Records json.RawMessage `json:"records"`
		}
		if err := json.Unmarshal(trimmed, &wrapper); err != nil || len(wrapper.Records) == 0 {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: "expected a JSON array or an object with a records array"}}
		}
		trimmed = wrapper.Records
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.DisallowUnknownFields()

	switch dataset {
	case DatasetClaims:
		var rows []Claim
		if err := dec.Decode(&rows); err != nil {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: err.Error()}}
		}
		for i := range rows {
			rows[i].Row = i + 1
		}
		b.Claims = rows
	case DatasetInvoices:
		var rows []Invoice
		if err := dec.Decode(&rows); err != nil {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: err.Error()}}
		}
		for i := range rows {
			rows[i].Row = i + 1
		}
		b.Invoices = rows
	case DatasetPolicies:
		var rows []Policy
		if err := dec.Decode(&rows); err != nil {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: err.Error()}}
		}
		for i := range rows {
			rows[i].Row = i + 1
		}
		b.Policies = rows
	case DatasetProviders:
		var rows []Provider
		if err := dec.Decode(&rows); err != nil {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: err.Error()}}
		}
		for i := range rows {
			rows[i].Row = i + 1
		}
		b.Providers = rows
	case DatasetIdentities:
		var rows []IdentityAttribute
		if err := dec.Decode(&rows); err != nil {
			return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeInvalidJSON, Message: err.Error()}}
		}
		for i := range rows {
			rows[i].Row = i + 1
		}
		b.Identities = rows
	default:
		return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeUnknownDataset, Message: "unknown dataset"}}
	}
	b.Sources[dataset] = Source{File: filename, Rows: b.rowCount(dataset)}
	return b, nil
}

func (b Batch) rowCount(dataset string) int {
	switch dataset {
	case DatasetClaims:
		return len(b.Claims)
	case DatasetInvoices:
		return len(b.Invoices)
	case DatasetPolicies:
		return len(b.Policies)
	case DatasetProviders:
		return len(b.Providers)
	case DatasetIdentities:
		return len(b.Identities)
	}
	return 0
}

func parseCSV(dataset, filename string, data []byte) (Batch, []RowError) {
	b := Batch{Version: SchemaVersion, Sources: map[string]Source{dataset: {File: filename}}}
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err == io.EOF {
		return b, []RowError{{Dataset: dataset, Row: 0, Code: CodeEmptyFile, Message: "file is empty"}}
	}
	if err != nil {
		return b, []RowError{{Dataset: dataset, Row: 1, Code: CodeInvalidCSV, Message: err.Error()}}
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	cols := map[string]int{}
	var errs []RowError
	for i, h := range header {
		name := strings.ToLower(strings.TrimSpace(h))
		if name == "" {
			errs = append(errs, RowError{Dataset: dataset, Row: 1, Code: CodeMissingHeader, Message: "empty column name"})
			continue
		}
		if _, dup := cols[name]; dup {
			errs = append(errs, RowError{Dataset: dataset, Row: 1, Field: name, Code: CodeDuplicateHeader, Message: "duplicate column"})
			continue
		}
		if !isAllowedColumn(dataset, name) {
			errs = append(errs, RowError{Dataset: dataset, Row: 1, Field: name, Code: CodeUnknownField, Message: "unknown column"})
			continue
		}
		cols[name] = i
	}
	for _, req := range requiredColumns[dataset] {
		if _, ok := cols[req]; !ok {
			errs = append(errs, RowError{Dataset: dataset, Row: 1, Field: req, Code: CodeMissingHeader, Message: "required column is missing"})
		}
	}
	if len(errs) > 0 {
		return b, errs
	}

	rowNum := 1
	for {
		rowNum++
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs = append(errs, RowError{Dataset: dataset, Row: rowNum, Code: CodeInvalidCSV, Message: err.Error()})
			continue
		}
		values := make(map[string]string, len(cols))
		for name, idx := range cols {
			if idx < len(record) {
				values[name] = strings.TrimSpace(record[idx])
			}
		}
		switch dataset {
		case DatasetClaims:
			c, rowErr := claimFromMap(values, rowNum)
			if rowErr != nil {
				errs = append(errs, *rowErr)
				continue
			}
			c.Row = rowNum
			b.Claims = append(b.Claims, c)
		case DatasetInvoices:
			in, rowErr := invoiceFromMap(values, rowNum)
			if rowErr != nil {
				errs = append(errs, *rowErr)
				continue
			}
			in.Row = rowNum
			b.Invoices = append(b.Invoices, in)
		case DatasetPolicies:
			p, rowErr := policyFromMap(values, rowNum)
			if rowErr != nil {
				errs = append(errs, *rowErr)
				continue
			}
			p.Row = rowNum
			b.Policies = append(b.Policies, p)
		case DatasetProviders:
			p, rowErr := providerFromMap(values, rowNum)
			if rowErr != nil {
				errs = append(errs, *rowErr)
				continue
			}
			p.Row = rowNum
			b.Providers = append(b.Providers, p)
		case DatasetIdentities:
			ia, rowErr := identityFromMap(values, rowNum)
			if rowErr != nil {
				errs = append(errs, *rowErr)
				continue
			}
			ia.Row = rowNum
			b.Identities = append(b.Identities, ia)
		}
	}
	b.Sources[dataset] = Source{File: filename, Rows: b.rowCount(dataset)}
	return b, errs
}

func isAllowedColumn(dataset, name string) bool {
	for _, c := range requiredColumns[dataset] {
		if c == name {
			return true
		}
	}
	_, ok := optionalColumns[dataset][name]
	return ok
}

func claimFromMap(v map[string]string, row int) (Claim, *RowError) {
	var c Claim
	c.ClaimID = v["claim_id"]
	c.PolicyID = v["policy_id"]
	c.ClaimantID = v["claimant_id"]
	c.LossType = v["loss_type"]
	c.Currency = v["currency"]
	c.ProviderID = v["provider_id"]
	c.InvoiceNumber = v["invoice_number"]
	c.Status = v["status"]
	c.RepID = v["rep_id"]
	if err := setDate(&c.ClaimDate, v["claim_date"]); err != nil {
		return c, &RowError{Dataset: DatasetClaims, Row: row, Field: "claim_date", Code: CodeInvalidDate, Message: err.Error()}
	}
	if err := setDate(&c.IncidentDate, v["incident_date"]); err != nil {
		return c, &RowError{Dataset: DatasetClaims, Row: row, Field: "incident_date", Code: CodeInvalidDate, Message: err.Error()}
	}
	amount, err := parseAmount(v["claimed_amount"])
	if err != nil {
		return c, &RowError{Dataset: DatasetClaims, Row: row, Field: "claimed_amount", Code: CodeInvalidAmount, Message: err.Error()}
	}
	c.ClaimedAmount = amount
	return c, nil
}

func invoiceFromMap(v map[string]string, row int) (Invoice, *RowError) {
	var in Invoice
	in.InvoiceNumber = v["invoice_number"]
	in.ProviderID = v["provider_id"]
	in.Currency = v["currency"]
	in.ClaimID = v["claim_id"]
	in.DocSHA256 = v["doc_sha256"]
	if err := setDate(&in.InvoiceDate, v["invoice_date"]); err != nil {
		return in, &RowError{Dataset: DatasetInvoices, Row: row, Field: "invoice_date", Code: CodeInvalidDate, Message: err.Error()}
	}
	amount, err := parseAmount(v["invoice_amount"])
	if err != nil {
		return in, &RowError{Dataset: DatasetInvoices, Row: row, Field: "invoice_amount", Code: CodeInvalidAmount, Message: err.Error()}
	}
	in.InvoiceAmount = amount
	return in, nil
}

func policyFromMap(v map[string]string, row int) (Policy, *RowError) {
	var p Policy
	p.PolicyID = v["policy_id"]
	p.Product = v["product"]
	p.LossTypeScope = v["loss_type_scope"]
	p.Status = v["status"]
	p.AssetID = v["asset_id"]
	p.AgentID = v["agent_id"]
	if err := setDate(&p.InceptionDate, v["inception_date"]); err != nil {
		return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "inception_date", Code: CodeInvalidDate, Message: err.Error()}
	}
	if err := setDate(&p.ExpiryDate, v["expiry_date"]); err != nil {
		return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "expiry_date", Code: CodeInvalidDate, Message: err.Error()}
	}
	if v["coverage_added_date"] != "" {
		d, err := ParseDate(v["coverage_added_date"])
		if err != nil {
			return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "coverage_added_date", Code: CodeInvalidDate, Message: err.Error()}
		}
		p.CoverageAddedDate = &d
	}
	if v["reinstatement_date"] != "" {
		d, err := ParseDate(v["reinstatement_date"])
		if err != nil {
			return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "reinstatement_date", Code: CodeInvalidDate, Message: err.Error()}
		}
		p.ReinstatementDate = &d
	}
	sum, err := parseAmount(v["sum_insured"])
	if err != nil {
		return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "sum_insured", Code: CodeInvalidAmount, Message: err.Error()}
	}
	p.SumInsured = sum
	premium, err := parseAmount(v["premium"])
	if err != nil {
		return p, &RowError{Dataset: DatasetPolicies, Row: row, Field: "premium", Code: CodeInvalidAmount, Message: err.Error()}
	}
	p.Premium = premium
	return p, nil
}

func providerFromMap(v map[string]string, row int) (Provider, *RowError) {
	_ = row
	return Provider{
		ProviderID: v["provider_id"],
		Name:       v["name"],
		Type:       v["type"],
		Region:     v["region"],
	}, nil
}

func identityFromMap(v map[string]string, row int) (IdentityAttribute, *RowError) {
	_ = row
	return IdentityAttribute{
		ClaimantID:    v["claimant_id"],
		AttributeType: v["attribute_type"],
		AttributeHash: v["attribute_hash"],
	}, nil
}

func setDate(dst *Date, raw string) error {
	parsed, err := ParseDate(raw)
	if err != nil {
		return err
	}
	*dst = parsed
	return nil
}

func parseAmount(raw string) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("expected integer minor unit amount")
	}
	return n, nil
}

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fardanwardhana/qoder-hackathon/packages/contracts"
)

func TestDatasetIsDeterministicAndValid(t *testing.T) {
	dirA := writeDataset(t)
	dirB := writeDataset(t)
	for _, name := range []string{"claims.csv", "invoices.csv", "policies.csv", "providers.csv", "identities.csv", "ground_truth.csv"} {
		a, err := os.ReadFile(filepath.Join(dirA, name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(dirB, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Fatalf("%s differs between runs", name)
		}
	}

	var batch contracts.Batch
	for _, name := range []string{"claims.csv", "invoices.csv", "policies.csv", "providers.csv", "identities.csv"} {
		raw, err := os.ReadFile(filepath.Join(dirA, name))
		if err != nil {
			t.Fatal(err)
		}
		dataset, _ := contracts.DatasetForFile(name)
		parsed, errs := contracts.ParseFile(dataset, name, raw)
		if len(errs) > 0 {
			t.Fatalf("%s parse errors: %+v", name, errs)
		}
		batch.Claims = append(batch.Claims, parsed.Claims...)
		batch.Invoices = append(batch.Invoices, parsed.Invoices...)
		batch.Policies = append(batch.Policies, parsed.Policies...)
		batch.Providers = append(batch.Providers, parsed.Providers...)
		batch.Identities = append(batch.Identities, parsed.Identities...)
	}
	errs, err := contracts.Validate(context.Background(), batch, contracts.NopLookup{})
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) > 0 {
		t.Fatalf("dataset violates the data contract: %+v", errs)
	}
}

func writeDataset(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	g := &generator{salt: "test-salt"}
	g.build()
	g.save(dir)
	return dir
}

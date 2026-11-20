package main

import "testing"

func TestValidateDecision(t *testing.T) {
	valid := []decisionRequest{
		{Outcome: "CLEARED", ReasonCode: "FP_DUPLICATE", Notes: "same invoice"},
		{Outcome: "CLEARED", ReasonCode: "NI_OUT_OF_SCOPE"},
		{Outcome: "CLEARED", ReasonCode: "CF_CONTROL"},
		{Outcome: "CONFIRMED", ReasonCode: "DOCUMENT_FRAUD", Notes: "forged invoice"},
		{Outcome: "NEEDS_INFO", ReasonCode: "MISSING_DOCS"},
	}
	for _, req := range valid {
		if err := validateDecision(req); err != nil {
			t.Fatalf("%+v rejected: %v", req, err)
		}
	}
	invalid := []decisionRequest{
		{Outcome: "CONFIRMED", ReasonCode: "DOCUMENT_FRAUD"},
		{Outcome: "CLEARED", ReasonCode: "SOMETHING"},
		{Outcome: "CLEARED", ReasonCode: ""},
		{Outcome: "DENIED", ReasonCode: "FP_X"},
		{Outcome: "NEEDS_INFO", ReasonCode: "lower"},
	}
	for _, req := range invalid {
		if err := validateDecision(req); err == nil {
			t.Fatalf("%+v should be rejected", req)
		}
	}
}

func TestPackageHashDetectsTampering(t *testing.T) {
	base := []evidenceHash{
		{EvidenceID: "a", SHA256: "sha256:1111"},
		{EvidenceID: "b", SHA256: "sha256:2222"},
	}
	h1, err := packageHash(base)
	if err != nil {
		t.Fatal(err)
	}
	h2, _ := packageHash(base)
	if h1 != h2 {
		t.Fatal("package hash must be deterministic")
	}
	tampered := []evidenceHash{
		{EvidenceID: "a", SHA256: "sha256:1111"},
		{EvidenceID: "b", SHA256: "sha256:3333"},
	}
	h3, _ := packageHash(tampered)
	if h3 == h1 {
		t.Fatal("package hash must change when evidence changes")
	}
}

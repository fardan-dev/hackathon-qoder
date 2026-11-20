package main

import (
	"testing"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
)

func TestPeriodUsesJakartaMonth(t *testing.T) {
	utc := time.Date(2026, 10, 31, 18, 0, 0, 0, time.UTC)
	if got := periodOf(utc); got != "2026-11" {
		t.Fatalf("period = %s", got)
	}
	start, end, err := periodBounds("2026-11")
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 31, 17, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s want %s", start, want)
	}
	if want := time.Date(2026, 11, 30, 17, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s want %s", end, want)
	}
	if _, _, err := periodBounds("2026-13"); err == nil {
		t.Fatal("invalid period accepted")
	}
}

func TestBillingPayloadCanonical(t *testing.T) {
	a := billingPayload{Period: "2026-11", Version: "v1", FormulaVersion: formulaBilling,
		Currency: "IDR", Metrics: map[string]int64{"claims_evaluated": 2, "alerts_created": 1}, TotalEvents: 3}
	b := billingPayload{Period: "2026-11", Version: "v1", FormulaVersion: formulaBilling,
		Currency: "IDR", Metrics: map[string]int64{"alerts_created": 1, "claims_evaluated": 2}, TotalEvents: 3}
	rawA, err := events.Canonical(a)
	if err != nil {
		t.Fatal(err)
	}
	rawB, err := events.Canonical(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(rawA) != string(rawB) {
		t.Fatal("billing payload must be independent of map iteration order")
	}
}

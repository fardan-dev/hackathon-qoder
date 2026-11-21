package main

import (
	"testing"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/rulesengine"
)

func TestAlertIDIsDeterministicAndTenantScoped(t *testing.T) {
	a := alertID("tenant-a", "CLM-1")
	if a != alertID("tenant-a", "CLM-1") {
		t.Fatal("alert id must be deterministic")
	}
	if a == alertID("tenant-b", "CLM-1") {
		t.Fatal("alert id must be tenant scoped")
	}
}

func TestSLAClock(t *testing.T) {
	wib, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Skip("tzdata unavailable")
	}
	created := time.Date(2026, 11, 20, 10, 0, 0, 0, wib) // Friday
	if got := slaDue(created, rulesengine.SeverityHigh); !got.Equal(created.Add(24 * time.Hour)) {
		t.Fatalf("High SLA = %s", got)
	}
	medium := slaDue(created, rulesengine.SeverityMedium)
	wantMedium := time.Date(2026, 11, 25, 10, 0, 0, 0, wib)
	if !medium.Equal(wantMedium) {
		t.Fatalf("Medium SLA = %s want %s", medium, wantMedium)
	}
	low := slaDue(created, rulesengine.SeverityLow)
	wantLow := time.Date(2026, 12, 1, 10, 0, 0, 0, wib)
	if !low.Equal(wantLow) {
		t.Fatalf("Low SLA = %s want %s", low, wantLow)
	}
}

func TestHighestSeverity(t *testing.T) {
	flags := []rulesengine.Flag{
		{Severity: rulesengine.SeverityLow},
		{Severity: rulesengine.SeverityHigh},
		{Severity: rulesengine.SeverityMedium},
	}
	if got := highestSeverity(flags); got != rulesengine.SeverityHigh {
		t.Fatalf("highest severity = %s", got)
	}
	if got := highestSeverity(nil); got != rulesengine.SeverityLow {
		t.Fatalf("empty flags = %s", got)
	}
}

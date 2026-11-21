package main

import "testing"

func TestObjectOf(t *testing.T) {
	cases := []struct {
		event string
		key   string
		value string
		typ   string
	}{
		{"batch.ingested", "batch_id", "b1", "batch"},
		{"claim.evaluated", "claim_id", "c1", "claim"},
		{"alert.created", "alert_id", "a1", "alert"},
		{"case.decided", "case_id", "cs1", "case"},
		{"suppression.approved", "suppression_id", "s1", "suppression"},
		{"rule.kill_switch", "switch_id", "k1", "kill_switch"},
		{"billing.exported", "export_id", "e1", "billing_export"},
	}
	for _, tc := range cases {
		typ, id := objectOf(tc.event, map[string]any{tc.key: tc.value})
		if typ != tc.typ || id != tc.value {
			t.Fatalf("%s => %s/%s", tc.event, typ, id)
		}
	}
	typ, id := objectOf("unknown.event", nil)
	if typ != "event" || id != "" {
		t.Fatalf("fallback => %s/%s", typ, id)
	}
}

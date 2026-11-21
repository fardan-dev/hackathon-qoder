import { describe, expect, it } from "vitest";

import { mapAlert, parseAlertsResponse, type RawAlert } from "./live";

const RAW_ALERT: RawAlert = {
  alert_id: "sha256:335e1083e616ea97812124694f9fb9ec9c6c94bd4ce4c353e7e191b2a7c53d52",
  claim_id: "CLM-1015",
  loss_type: "MOTOR",
  claimed_amount: 35000000,
  status: "OPEN",
  suppressed: false,
  sla_due_at: "2026-11-21T12:39:34.259573Z",
  created_at: "2026-11-20T12:39:34.259573Z",
  flags: [
    {
      rule_id: "B5",
      rule_version: 1,
      rule_name: "Coverage Limit Breach",
      severity: "High",
      trigger_value: "Rp 35.000.000 > sum insured Rp 20.000.000",
      evidence_refs: ["CLM-1015", "POL-5015"],
      explanation: "Nominal klaim melebihi sum insured.",
    },
  ],
};

describe("live alert mapper", () => {
  it("normalizes severity, rule version, and SLA due date", () => {
    const alert = mapAlert(RAW_ALERT);

    expect(alert.alertId).toBe(RAW_ALERT.alert_id);
    expect(alert.groupId).toBeNull();
    expect(alert.status).toBe("OPEN");
    expect(alert.slaDueAt).toBe(RAW_ALERT.sla_due_at);
    expect(alert.flags[0]?.severity).toBe("HIGH");
    expect(alert.flags[0]?.ruleVersion).toBe("1");
  });

  it("surfaces suppressed alerts and closed status", () => {
    expect(mapAlert({ ...RAW_ALERT, suppressed: true }).status).toBe(
      "SUPPRESSED",
    );
    expect(
      mapAlert({ ...RAW_ALERT, suppressed: false, status: "CLOSED" }).status,
    ).toBe("CLOSED");
  });

  it("parses the queue envelope", () => {
    const alerts = parseAlertsResponse({ alerts: [RAW_ALERT], count: 1 });

    expect(alerts).toHaveLength(1);
    expect(alerts[0]?.claimId).toBe("CLM-1015");
  });
});

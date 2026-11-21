import { describe, expect, it } from "vitest";

import { parseDashboard } from "./live";

describe("dashboard parser", () => {
  it("normalizes severity buckets and sorts flag distribution", () => {
    const dashboard = parseDashboard({
      alerts: {
        avg_flags: 2.06,
        by_severity: { High: 11, Low: 2, Medium: 20 },
        sla_breached_open: 0,
        sla_compliance_bps: 10000,
        suppressed: 1,
        total: 33,
        total_flags: 68,
      },
      decisions: {
        cleared: 0,
        confirmed: 1,
        false_positive_bps: 0,
        needs_info: 0,
        total: 1,
        value_protected: 20000000,
        value_recovered: 0,
      },
      coverage: { covered: 10, denominator: 14 },
      flag_distribution: { B4: 20, A3: 4, D4: 10 },
      from: "2026-10-21T00:00:00Z",
      to: "2026-11-21T00:00:00Z",
      lob: "",
    });

    expect(dashboard.alerts.bySeverity).toEqual({
      HIGH: 11,
      MEDIUM: 20,
      LOW: 2,
    });
    expect(dashboard.flagDistribution).toEqual([
      { ruleId: "B4", count: 20 },
      { ruleId: "D4", count: 10 },
      { ruleId: "A3", count: 4 },
    ]);
    expect(dashboard.decisions.valueProtected).toBe(20000000);
  });
});

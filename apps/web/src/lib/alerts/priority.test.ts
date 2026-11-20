import { describe, expect, it } from "vitest";

import { createMockAlerts } from "./mock-data";
import {
  alertPriority,
  compareAlerts,
  highestSeverity,
  sortAlerts,
} from "./priority";
import type { Alert, Flag, Severity } from "./schema";

function makeFlag(severity: Severity, ruleId = "A1"): Flag {
  return {
    ruleId,
    ruleName: "Test rule",
    ruleVersion: "1.0.0",
    severity,
    triggerValue: "test",
    evidenceRefs: ["CLM-0000"],
    explanation: "Penjelasan uji.",
  };
}

function makeAlert(
  alertId: string,
  severity: Severity,
  flagCount: number,
  claimedAmount: number,
): Alert {
  return {
    alertId,
    claimId: `CLM-${alertId}`,
    groupId: null,
    lossType: "MOTOR",
    claimantId: "C-0001",
    providerId: "P-01",
    claimedAmount,
    status: "OPEN",
    createdAt: "2026-11-20T03:00:00.000Z",
    flags: Array.from({ length: flagCount }, (_, index) =>
      makeFlag(severity, `R${index}`),
    ),
  };
}

const NOW = new Date("2026-11-20T10:00:00+07:00");

describe("bundle severity", () => {
  it("uses the highest severity among flags regardless of order", () => {
    expect(
      highestSeverity([makeFlag("LOW"), makeFlag("HIGH"), makeFlag("MEDIUM")]),
    ).toBe("HIGH");
  });

  it("derives alert priority from the flag bundle", () => {
    const alert = makeAlert("ALT-1", "MEDIUM", 3, 100);
    alert.flags[2] = makeFlag("HIGH");
    expect(alertPriority(alert)).toBe("HIGH");
  });
});

describe("deterministic ordering (BRD §7.9 / D-07)", () => {
  it("sorts by severity, then flag count, then claim value", () => {
    const mediumThreeFlags = makeAlert("ALT-M3", "MEDIUM", 3, 1_000);
    const highOneFlag = makeAlert("ALT-H1", "HIGH", 1, 1_000);
    const highTwoFlagsSmall = makeAlert("ALT-H2S", "HIGH", 2, 100);
    const highTwoFlagsLarge = makeAlert("ALT-H2L", "HIGH", 2, 900);

    const sorted = sortAlerts([
      mediumThreeFlags,
      highOneFlag,
      highTwoFlagsSmall,
      highTwoFlagsLarge,
    ]);

    expect(sorted.map((alert) => alert.alertId)).toEqual([
      "ALT-H2L",
      "ALT-H2S",
      "ALT-H1",
      "ALT-M3",
    ]);
  });

  it("breaks full ties by alert id so ordering is stable", () => {
    const first = makeAlert("ALT-0002", "HIGH", 2, 500);
    const second = makeAlert("ALT-0001", "HIGH", 2, 500);

    expect(compareAlerts(first, second)).toBeGreaterThan(0);
    expect(sortAlerts([first, second]).map((a) => a.alertId)).toEqual([
      "ALT-0001",
      "ALT-0002",
    ]);
  });

  it("is idempotent for the seeded batch", () => {
    const alerts = createMockAlerts(NOW);
    const once = sortAlerts(alerts);
    const twice = sortAlerts(once);

    expect(twice.map((alert) => alert.alertId)).toEqual(
      once.map((alert) => alert.alertId),
    );
    expect(once[0]?.alertId).toBe("ALT-1010");
  });
});

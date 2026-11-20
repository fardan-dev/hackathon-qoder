import { describe, expect, it } from "vitest";

import type { Flag, Severity } from "./schema";
import {
  SLA_DUE_SOON_MS,
  addBusinessDays,
  slaDueAt,
  slaRemainingMs,
  slaState,
} from "./sla";

function flag(severity: Severity): Flag {
  return {
    ruleId: "A1",
    ruleName: "Test rule",
    ruleVersion: "1.0.0",
    severity,
    triggerValue: "test",
    evidenceRefs: ["CLM-0000"],
    explanation: "Penjelasan uji.",
  };
}

function createdAt(iso: string, severity: Severity) {
  return { createdAt: iso, flags: [flag(severity)] };
}

describe("business-day arithmetic", () => {
  it("skips weekends", () => {
    // Friday 20 Nov 2026, 10:00 WIB + 7 business days = Tuesday 1 Dec 2026.
    const start = new Date("2026-11-20T03:00:00.000Z");
    expect(addBusinessDays(start, 7).toISOString()).toBe(
      "2026-12-01T03:00:00.000Z",
    );
  });

  it("crosses a weekend for three business days", () => {
    // Wednesday 25 Nov 2026 + 3 business days = Monday 30 Nov 2026.
    const start = new Date("2026-11-25T03:00:00.000Z");
    expect(addBusinessDays(start, 3).toISOString()).toBe(
      "2026-11-30T03:00:00.000Z",
    );
  });
});

describe("SLA due dates (BRD §7.12)", () => {
  it("uses 24 hours for High", () => {
    const due = slaDueAt(createdAt("2026-11-20T03:00:00.000Z", "HIGH"));
    expect(due.toISOString()).toBe("2026-11-21T03:00:00.000Z");
  });

  it("uses 3 days for Medium", () => {
    const due = slaDueAt(createdAt("2026-11-20T03:00:00.000Z", "MEDIUM"));
    expect(due.toISOString()).toBe("2026-11-23T03:00:00.000Z");
  });

  it("uses 7 working days for Low", () => {
    const due = slaDueAt(createdAt("2026-11-20T03:00:00.000Z", "LOW"));
    expect(due.toISOString()).toBe("2026-12-01T03:00:00.000Z");
  });
});

describe("SLA state", () => {
  const now = new Date("2026-11-20T03:00:00.000Z");

  it("classifies overdue, due soon, and on track", () => {
    expect(slaState(new Date(now.getTime() - 1), now)).toBe("OVERDUE");
    expect(slaState(new Date(now.getTime() + SLA_DUE_SOON_MS), now)).toBe(
      "DUE_SOON",
    );
    expect(slaState(new Date(now.getTime() + SLA_DUE_SOON_MS + 1), now)).toBe(
      "ON_TRACK",
    );
  });

  it("reports remaining time as a signed duration", () => {
    expect(slaRemainingMs(new Date(now.getTime() + 3_600_000), now)).toBe(
      3_600_000,
    );
  });
});

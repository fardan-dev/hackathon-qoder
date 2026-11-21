import { describe, expect, it } from "vitest";

import {
  buildQueueQuery,
  matchesQueueFilters,
  parseQueueFilters,
  parseQueueSearch,
  type QueueFilters,
} from "./filters";
import { createMockAlerts } from "./mock-data";
import { buildQueue } from "./queue";

const NOW = new Date("2026-11-20T10:00:00+07:00");
const ALERTS = createMockAlerts(NOW);

function filters(overrides: Partial<QueueFilters> = {}): QueueFilters {
  return { ...parseQueueFilters({}), ...overrides };
}

function visibleIds(overrides: Partial<QueueFilters> = {}): string[] {
  return buildQueue(ALERTS, filters(overrides), NOW).rows.map(
    (row) => row.alert.alertId,
  );
}

describe("queue defaults", () => {
  it("shows High + Medium and hides suppressed and Low alerts", () => {
    const ids = visibleIds();

    expect(ids).toHaveLength(13);
    expect(ids).not.toContain("ALT-1010");
    expect(ids).not.toContain("ALT-1011");
    expect(ids).not.toContain("ALT-1012");
  });

  it("orders rows by priority, flag count, then claim value", () => {
    expect(visibleIds()).toEqual([
      "ALT-1003",
      "ALT-1015",
      "ALT-1013",
      "ALT-1002",
      "ALT-1006",
      "ALT-1001",
      "ALT-1007",
      "ALT-1016",
      "ALT-1014",
      "ALT-1008",
      "ALT-1004",
      "ALT-1005",
      "ALT-1009",
    ]);
  });

  it("shows every status and priority when both filters are opened", () => {
    const ids = visibleIds({ priority: "all", status: "all" });
    expect(ids).toHaveLength(16);
    expect(ids).toContain("ALT-1010");
    expect(ids).toContain("ALT-1012");
  });

  it("counts overdue alerts in the visible queue", () => {
    const { overdueCount, rows } = buildQueue(ALERTS, filters(), NOW);
    expect(overdueCount).toBe(3);
    expect(rows.filter((row) => row.sla === "OVERDUE")).toHaveLength(3);
  });
});

describe("queue filters", () => {
  it("filters by line of business", () => {
    const ids = visibleIds({ lossType: "PROPERTY" });
    expect(ids).toHaveLength(6);
    expect(ids).toContain("ALT-1002");
    expect(ids).not.toContain("ALT-1001");
  });

  it("filters by SLA state", () => {
    expect(visibleIds({ sla: "OVERDUE" })).toEqual([
      "ALT-1015",
      "ALT-1006",
      "ALT-1001",
    ]);
  });

  it("filters by status", () => {
    const ids = visibleIds({ status: "IN_REVIEW" });
    expect(ids).toEqual(["ALT-1003", "ALT-1009"]);
  });

  it("searches claim, claimant, provider, and evidence references", () => {
    expect(visibleIds({ q: "INV-7781" })).toHaveLength(3);
    expect(visibleIds({ q: "C-0209" })).toHaveLength(2);
    expect(visibleIds({ q: "p-19" })).toHaveLength(2);
  });

  it("keeps matching deterministic for the same input", () => {
    const alert = ALERTS[0];
    expect(alert).toBeDefined();
    if (!alert) return;

    const active = filters({ priority: "all", status: "all", q: "clm" });
    expect(matchesQueueFilters(alert, active, NOW)).toBe(true);
    expect(matchesQueueFilters(alert, active, NOW)).toBe(true);
  });
});

describe("URL round-trip", () => {
  it("parses known values and falls back on unknown ones", () => {
    expect(
      parseQueueFilters({
        priority: "HIGH",
        status: "SUPPRESSED",
        lossType: "MOTOR",
        sla: "OVERDUE",
        q: "  CLM-0001  ",
      }),
    ).toEqual({
      priority: "HIGH",
      status: "SUPPRESSED",
      lossType: "MOTOR",
      sla: "OVERDUE",
      q: "CLM-0001",
    });

    expect(parseQueueFilters({ priority: "URGENT", sla: "SOON" })).toEqual(
      filters(),
    );
  });

  it("serialises only non-default filters and restores them", () => {
    const active = filters({
      priority: "HIGH",
      status: "all",
      sla: "DUE_SOON",
      q: "C-0209",
    });
    const query = buildQueueQuery(active);

    expect(query).toBe("?priority=HIGH&status=all&sla=DUE_SOON&q=C-0209");
    expect(parseQueueSearch(query)).toEqual(active);
  });

  it("returns defaults for an empty or invalid back link", () => {
    expect(parseQueueSearch(undefined)).toEqual(filters());
    expect(parseQueueSearch("?priority=NOT_REAL")).toEqual(filters());
  });
});

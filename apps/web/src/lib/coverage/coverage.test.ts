import { describe, expect, it } from "vitest";

import {
  COVERAGE_ROWS,
  coverageCsv,
  summarizeCoverage,
} from "./coverage";

describe("coverage matrix", () => {
  it("covers the 14 in-scope typologies with 10 covered in Phase 0/1", () => {
    const summary = summarizeCoverage(COVERAGE_ROWS);
    expect(summary).toEqual({
      denominator: 14,
      covered: 10,
      residual: 4,
      coverageBps: 7143,
    });
  });

  it("maps every typology to at least one rule", () => {
    for (const row of COVERAGE_ROWS) {
      expect(row.ruleIds.length, row.key).toBeGreaterThan(0);
    }
  });

  it("records a residual risk reason for every uncovered typology", () => {
    for (const row of COVERAGE_ROWS.filter((item) => !item.covered)) {
      expect(row.residualRisk, row.key).toBeTruthy();
    }
  });

  it("exports one CSV row per typology", () => {
    const lines = coverageCsv(COVERAGE_ROWS).trimEnd().split("\n");
    expect(lines).toHaveLength(COVERAGE_ROWS.length + 1);
    expect(lines[0]).toBe(
      "tipologi,label,status,rule_ids,residual_risk",
    );
  });
});

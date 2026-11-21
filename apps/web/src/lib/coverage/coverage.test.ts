import { describe, expect, it } from "vitest";

import { coverageCsv, type CoverageRow } from "./coverage";
import { parseCoverage } from "./live";

const PAYLOAD = {
  summary: {
    coverage_bps: 7142,
    covered: 10,
    denominator: 14,
    residual: 4,
  },
  typologies: [
    {
      typology_key: "document_amount_edit",
      label: "Document amount edit (OCR mismatch)",
      in_scope: true,
      phase01_covered: false,
      rule_ids: ["H1", "H6"],
      sort_order: 11,
    },
    {
      typology_key: "invoice_reuse",
      label: "Invoice reuse / double billing",
      in_scope: true,
      phase01_covered: true,
      rule_ids: ["A1", "A3"],
      sort_order: 1,
    },
  ],
  residual_risk_register: [
    {
      typology_key: "document_amount_edit",
      residual_risk: "Document forensics analyzer is gated off.",
    },
  ],
};

const ROWS: CoverageRow[] = [
  {
    key: "invoice_reuse",
    label: "Invoice reuse / double billing",
    covered: true,
    ruleIds: ["A1", "A3"],
    residualRisk: null,
  },
  {
    key: "document_forgery",
    label: "Document metadata forgery, \"ELA\"",
    covered: false,
    ruleIds: ["H2", "H3"],
    residualRisk: "Analyzer gated off.",
  },
];

describe("coverage parser", () => {
  it("uses the API summary and sorts typologies by sort_order", () => {
    const view = parseCoverage(PAYLOAD);

    expect(view.summary).toEqual({
      coverageBps: 7142,
      covered: 10,
      denominator: 14,
      residual: 4,
    });
    expect(view.rows.map((row) => row.key)).toEqual([
      "invoice_reuse",
      "document_amount_edit",
    ]);
    expect(view.rows[1]?.residualRisk).toContain("gated off");
  });

  it("maps every typology to at least one rule", () => {
    for (const row of parseCoverage(PAYLOAD).rows) {
      expect(row.ruleIds.length, row.key).toBeGreaterThan(0);
    }
  });
});

describe("coverage CSV export", () => {
  it("exports one CSV row per typology", () => {
    const lines = coverageCsv(ROWS).trimEnd().split("\n");
    expect(lines).toHaveLength(ROWS.length + 1);
    expect(lines[0]).toBe("tipologi,label,status,rule_ids,residual_risk");
  });

  it("quotes commas and quotes in labels", () => {
    const csv = coverageCsv(ROWS);
    expect(csv).toContain('"Document metadata forgery, ""ELA"""');
    expect(csv).toContain("H2;H3");
    expect(csv).toContain(",residual,");
  });
});

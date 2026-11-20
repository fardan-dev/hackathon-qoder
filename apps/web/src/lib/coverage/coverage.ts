export interface CoverageRow {
  key: string;
  label: string;
  covered: boolean;
  ruleIds: string[];
  residualRisk: string | null;
}

export interface CoverageSummary {
  denominator: number;
  covered: number;
  residual: number;
  coverageBps: number;
}

export const COVERAGE_ROWS: CoverageRow[] = [
  {
    key: "invoice_reuse",
    label: "Invoice reuse / double billing",
    covered: true,
    ruleIds: ["A1", "A3", "A5", "A6"],
    residualRisk: null,
  },
  {
    key: "duplicate_claim",
    label: "Duplicate claim submission",
    covered: true,
    ruleIds: ["A2", "C3", "C5", "F2"],
    residualRisk: null,
  },
  {
    key: "amount_inflation",
    label: "Amount inflation vs invoice",
    covered: true,
    ruleIds: ["A4", "B1", "B2", "B3", "B4"],
    residualRisk: null,
  },
  {
    key: "structuring",
    label: "Structuring below approval threshold",
    covered: true,
    ruleIds: ["B3", "B4"],
    residualRisk: null,
  },
  {
    key: "cover_breach",
    label: "Claim outside cover / limit breach",
    covered: true,
    ruleIds: ["B5", "B6", "C1", "C2", "E1", "E2"],
    residualRisk: null,
  },
  {
    key: "new_policy_reinstatement",
    label: "New policy or reinstatement claim",
    covered: true,
    ruleIds: ["C1", "C2", "E1"],
    residualRisk: null,
  },
  {
    key: "repeat_claimant",
    label: "Repeat claimant within window",
    covered: true,
    ruleIds: ["C4", "F1", "F3", "G3"],
    residualRisk: null,
  },
  {
    key: "provider_overbilling",
    label: "Provider overbilling / volume anomaly",
    covered: true,
    ruleIds: ["D4", "D5"],
    residualRisk: null,
  },
  {
    key: "collusion_network",
    label: "Collusion network (shared identity/provider)",
    covered: true,
    ruleIds: ["D1", "D2", "D3", "D6", "D7", "F4"],
    residualRisk: null,
  },
  {
    key: "repeat_offender",
    label: "Repeat offender with prior confirmed flag",
    covered: true,
    ruleIds: ["G1", "G2", "G3"],
    residualRisk: null,
  },
  {
    key: "document_amount_edit",
    label: "Document amount edit (OCR mismatch)",
    covered: false,
    ruleIds: ["H1", "H6"],
    residualRisk:
      "Document forensics analyzer is gated off in Phase 0/1; relies on structured amount mismatch only.",
  },
  {
    key: "document_forgery",
    label: "Document metadata forgery",
    covered: false,
    ruleIds: ["H2", "H3", "H4", "H5", "H7"],
    residualRisk:
      "No document forensics analyzer in the thin slice; residual risk accepted until Phase 2.",
  },
  {
    key: "document_rescan",
    label: "Document re-scan / near-duplicate",
    covered: false,
    ruleIds: ["H5"],
    residualRisk:
      "Perceptual hashing is gated off; only exact doc_sha256 reuse is covered by A3.",
  },
  {
    key: "high_quality_forgery",
    label: "High-quality forgery (ELA/template)",
    covered: false,
    ruleIds: ["H8", "H9", "H10"],
    residualRisk:
      "ELA and template-deviation analysis are gated off; residual risk accepted.",
  },
];

export function summarizeCoverage(rows: CoverageRow[]): CoverageSummary {
  const denominator = rows.length;
  const covered = rows.filter((row) => row.covered).length;
  return {
    denominator,
    covered,
    residual: denominator - covered,
    coverageBps:
      denominator === 0 ? 0 : Math.round((covered * 10000) / denominator),
  };
}

function csvField(value: string): string {
  if (/[",\n]/.test(value)) return `"${value.replaceAll('"', '""')}"`;
  return value;
}

export function coverageCsv(rows: CoverageRow[]): string {
  const header = ["tipologi", "label", "status", "rule_ids", "residual_risk"];
  const lines = rows.map((row) =>
    [
      row.key,
      row.label,
      row.covered ? "covered" : "residual",
      row.ruleIds.join(";"),
      row.residualRisk ?? "",
    ]
      .map(csvField)
      .join(","),
  );
  return [header.join(","), ...lines].join("\n") + "\n";
}

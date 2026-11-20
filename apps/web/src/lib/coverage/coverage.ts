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

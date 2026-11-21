import { z } from "zod";

import type { CoverageRow, CoverageSummary } from "./coverage";

const rawCoverageSchema = z.object({
  summary: z.object({
    coverage_bps: z.number().int().min(0),
    covered: z.number().int().nonnegative(),
    denominator: z.number().int().nonnegative(),
    residual: z.number().int().nonnegative(),
  }),
  typologies: z.array(
    z.object({
      typology_key: z.string().min(1),
      label: z.string().min(1),
      in_scope: z.boolean(),
      phase01_covered: z.boolean(),
      rule_ids: z.array(z.string().min(1)),
      sort_order: z.number().int(),
    }),
  ),
  residual_risk_register: z.array(
    z.object({
      typology_key: z.string().min(1),
      residual_risk: z.string(),
    }),
  ),
});

export interface CoverageView {
  rows: CoverageRow[];
  summary: CoverageSummary;
}

export function parseCoverage(payload: unknown): CoverageView {
  const raw = rawCoverageSchema.parse(payload);
  const risks = new Map(
    raw.residual_risk_register.map((entry) => [
      entry.typology_key,
      entry.residual_risk,
    ]),
  );

  const rows = [...raw.typologies]
    .sort((a, b) => a.sort_order - b.sort_order)
    .map<CoverageRow>((item) => ({
      key: item.typology_key,
      label: item.label,
      covered: item.phase01_covered,
      ruleIds: item.rule_ids,
      residualRisk: risks.get(item.typology_key) ?? null,
    }));

  return {
    rows,
    summary: {
      coverageBps: raw.summary.coverage_bps,
      covered: raw.summary.covered,
      denominator: raw.summary.denominator,
      residual: raw.summary.residual,
    },
  };
}

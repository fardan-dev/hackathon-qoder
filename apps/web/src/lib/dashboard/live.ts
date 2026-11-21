import { z } from "zod";

import type { Severity } from "@/lib/alerts/schema";

const rawDashboardSchema = z.object({
  alerts: z.object({
    total: z.number().int().nonnegative(),
    suppressed: z.number().int().nonnegative(),
    total_flags: z.number().int().nonnegative(),
    avg_flags: z.number(),
    sla_breached_open: z.number().int().nonnegative(),
    sla_compliance_bps: z.number().int().min(0),
    by_severity: z.record(z.string(), z.number().int().nonnegative()),
  }),
  decisions: z.object({
    total: z.number().int().nonnegative(),
    confirmed: z.number().int().nonnegative(),
    cleared: z.number().int().nonnegative(),
    needs_info: z.number().int().nonnegative(),
    value_protected: z.number().int().nonnegative(),
    value_recovered: z.number().int().nonnegative(),
    false_positive_bps: z.number().int().min(0),
  }),
  coverage: z.object({
    covered: z.number().int().nonnegative(),
    denominator: z.number().int().nonnegative(),
  }),
  flag_distribution: z.record(z.string(), z.number().int().nonnegative()),
  from: z.string().min(1),
  to: z.string().min(1),
  lob: z.string(),
});

export interface DashboardData {
  alerts: {
    total: number;
    suppressed: number;
    totalFlags: number;
    avgFlags: number;
    slaBreachedOpen: number;
    slaComplianceBps: number;
    bySeverity: Record<Severity, number>;
  };
  decisions: {
    total: number;
    confirmed: number;
    cleared: number;
    needsInfo: number;
    valueProtected: number;
    valueRecovered: number;
    falsePositiveBps: number;
  };
  coverage: { covered: number; denominator: number };
  flagDistribution: Array<{ ruleId: string; count: number }>;
  from: string;
  to: string;
  lob: string;
}

export function parseDashboard(payload: unknown): DashboardData {
  const raw = rawDashboardSchema.parse(payload);

  return {
    alerts: {
      total: raw.alerts.total,
      suppressed: raw.alerts.suppressed,
      totalFlags: raw.alerts.total_flags,
      avgFlags: raw.alerts.avg_flags,
      slaBreachedOpen: raw.alerts.sla_breached_open,
      slaComplianceBps: raw.alerts.sla_compliance_bps,
      bySeverity: {
        HIGH: raw.alerts.by_severity.High ?? 0,
        MEDIUM: raw.alerts.by_severity.Medium ?? 0,
        LOW: raw.alerts.by_severity.Low ?? 0,
      },
    },
    decisions: {
      total: raw.decisions.total,
      confirmed: raw.decisions.confirmed,
      cleared: raw.decisions.cleared,
      needsInfo: raw.decisions.needs_info,
      valueProtected: raw.decisions.value_protected,
      valueRecovered: raw.decisions.value_recovered,
      falsePositiveBps: raw.decisions.false_positive_bps,
    },
    coverage: raw.coverage,
    flagDistribution: Object.entries(raw.flag_distribution)
      .map(([ruleId, count]) => ({ ruleId, count }))
      .sort((a, b) => b.count - a.count || a.ruleId.localeCompare(b.ruleId)),
    from: raw.from,
    to: raw.to,
    lob: raw.lob,
  };
}

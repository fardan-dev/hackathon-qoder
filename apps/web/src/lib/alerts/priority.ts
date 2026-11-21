import type { Alert, Flag, Severity } from "./schema";

const SEVERITY_WEIGHT: Record<Severity, number> = {
  HIGH: 0,
  MEDIUM: 1,
  LOW: 2,
};

export function highestSeverity(flags: Flag[]): Severity {
  return flags.reduce<Severity>(
    (highest, flag) =>
      SEVERITY_WEIGHT[flag.severity] < SEVERITY_WEIGHT[highest]
        ? flag.severity
        : highest,
    "LOW",
  );
}

export function alertPriority(alert: Pick<Alert, "flags">): Severity {
  return highestSeverity(alert.flags);
}

// BRD §7.9 / D-07: severity, lalu jumlah flag, lalu nilai klaim. No hidden score.
export function compareAlerts(a: Alert, b: Alert): number {
  const bySeverity =
    SEVERITY_WEIGHT[alertPriority(a)] - SEVERITY_WEIGHT[alertPriority(b)];
  if (bySeverity !== 0) return bySeverity;

  const byFlagCount = b.flags.length - a.flags.length;
  if (byFlagCount !== 0) return byFlagCount;

  const byAmount = b.claimedAmount - a.claimedAmount;
  if (byAmount !== 0) return byAmount;

  return a.alertId.localeCompare(b.alertId);
}

export function sortAlerts(alerts: Alert[]): Alert[] {
  return [...alerts].sort(compareAlerts);
}

import { matchesQueueFilters, type QueueFilters } from "./filters";
import { alertPriority, sortAlerts } from "./priority";
import type { Alert, Severity } from "./schema";
import {
  slaDueAt,
  slaRemainingMs,
  slaState,
  type SlaState,
} from "./sla";

export interface QueueRow {
  alert: Alert;
  priority: Severity;
  flagCount: number;
  dueAt: Date;
  sla: SlaState;
  remainingMs: number;
}

export interface QueueResult {
  rows: QueueRow[];
  overdueCount: number;
  highCount: number;
}

export function buildQueue(
  alerts: Alert[],
  filters: QueueFilters,
  now: Date,
): QueueResult {
  const rows = sortAlerts(alerts)
    .filter((alert) => matchesQueueFilters(alert, filters, now))
    .map<QueueRow>((alert) => {
      const dueAt = slaDueAt(alert);
      const remainingMs = slaRemainingMs(dueAt, now);

      return {
        alert,
        priority: alertPriority(alert),
        flagCount: alert.flags.length,
        dueAt,
        sla: slaState(dueAt, now),
        remainingMs,
      };
    });

  return {
    rows,
    overdueCount: rows.filter((row) => row.sla === "OVERDUE").length,
    highCount: rows.filter((row) => row.priority === "HIGH").length,
  };
}

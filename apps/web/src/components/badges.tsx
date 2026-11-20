import {
  PRIORITY_LABEL,
  STATUS_LABEL,
  formatSla,
} from "@/lib/alerts/format";
import type { AlertStatus, Severity } from "@/lib/alerts/schema";
import type { SlaState } from "@/lib/alerts/sla";

export function PriorityBadge({ priority }: { priority: Severity }) {
  return (
    <span className={`badge badge--priority-${priority.toLowerCase()}`}>
      {PRIORITY_LABEL[priority]}
    </span>
  );
}

export function StatusBadge({ status }: { status: AlertStatus }) {
  return (
    <span className={`badge badge--status-${status.toLowerCase()}`}>
      {STATUS_LABEL[status]}
    </span>
  );
}

export function SlaBadge({
  state,
  remainingMs,
}: {
  state: SlaState;
  remainingMs: number;
}) {
  return (
    <span className={`badge badge--sla-${state.toLowerCase()}`}>
      <span className="badge__dot" aria-hidden="true" />
      {formatSla(state, remainingMs)}
    </span>
  );
}

export function FlagCountBadge({ count }: { count: number }) {
  return (
    <span className="flag-count">
      <strong>{count}</strong> flag
    </span>
  );
}

export function HashBadge({ valid }: { valid: boolean }) {
  return (
    <span className={`badge badge--verify-${valid ? "ok" : "fail"}`}>
      {valid ? "Hash cocok" : "Hash tidak cocok"}
    </span>
  );
}

export function PackageStateBadge({ frozen }: { frozen: boolean }) {
  return (
    <span className={`badge badge--package-${frozen ? "frozen" : "open"}`}>
      {frozen ? "Dibekukan" : "Paket terbuka"}
    </span>
  );
}

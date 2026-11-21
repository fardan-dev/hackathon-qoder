import {
  PRIORITY_LABEL,
  STATUS_LABEL,
  formatSla,
} from "@/lib/alerts/format";
import type { AlertStatus, Severity } from "@/lib/alerts/schema";
import type { SlaState } from "@/lib/alerts/sla";
import { caseStatusLabel } from "@/lib/case/live";

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

export function CaseStatusBadge({ status }: { status: string }) {
  const slug = status.toLowerCase().replaceAll("_", "-");
  return (
    <span className={`badge badge--case-${slug}`}>
      {caseStatusLabel(status)}
    </span>
  );
}

export function DecisionBadge({ outcome }: { outcome: string }) {
  const slug = outcome.toLowerCase().replaceAll("_", "-");
  return (
    <span className={`badge badge--decision-${slug}`}>
      {outcome.replaceAll("_", " ")}
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

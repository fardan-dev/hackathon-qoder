import { alertPriority } from "./priority";
import type { AlertStatus, LossType, Severity } from "./schema";
import type { Alert } from "./schema";
import { slaDueAt, slaState, type SlaState } from "./sla";

export type PriorityFilter = "default" | "all" | Severity;
export type StatusFilter = "default" | "all" | AlertStatus;
export type LossTypeFilter = "all" | LossType;
export type SlaFilter = "all" | SlaState;

export interface QueueFilters {
  priority: PriorityFilter;
  status: StatusFilter;
  lossType: LossTypeFilter;
  sla: SlaFilter;
  q: string;
}

export type RawSearchParams = Record<
  string,
  string | string[] | undefined
>;

const DEFAULT_FILTERS: QueueFilters = {
  priority: "default",
  status: "default",
  lossType: "all",
  sla: "all",
  q: "",
};

const PRIORITY_VALUES = new Set<PriorityFilter>([
  "default",
  "all",
  "HIGH",
  "MEDIUM",
  "LOW",
]);

const STATUS_VALUES = new Set<StatusFilter>([
  "default",
  "all",
  "OPEN",
  "IN_REVIEW",
  "ESCALATED",
  "SUPPRESSED",
  "CLOSED",
]);

const LOSS_TYPE_VALUES = new Set<LossTypeFilter>([
  "all",
  "MOTOR",
  "PROPERTY",
]);

const SLA_VALUES = new Set<SlaFilter>([
  "all",
  "OVERDUE",
  "DUE_SOON",
  "ON_TRACK",
]);

function firstValue(value: string | string[] | undefined): string {
  if (Array.isArray(value)) return value[0] ?? "";
  return value ?? "";
}

export function parseQueueFilters(raw: RawSearchParams): QueueFilters {
  const priority = firstValue(raw.priority);
  const status = firstValue(raw.status);
  const lossType = firstValue(raw.lossType);
  const sla = firstValue(raw.sla);

  return {
    priority: PRIORITY_VALUES.has(priority as PriorityFilter)
      ? (priority as PriorityFilter)
      : DEFAULT_FILTERS.priority,
    status: STATUS_VALUES.has(status as StatusFilter)
      ? (status as StatusFilter)
      : DEFAULT_FILTERS.status,
    lossType: LOSS_TYPE_VALUES.has(lossType as LossTypeFilter)
      ? (lossType as LossTypeFilter)
      : DEFAULT_FILTERS.lossType,
    sla: SLA_VALUES.has(sla as SlaFilter)
      ? (sla as SlaFilter)
      : DEFAULT_FILTERS.sla,
    q: firstValue(raw.q).trim().slice(0, 120),
  };
}

export function parseQueueSearch(search: string | undefined): QueueFilters {
  if (!search) return { ...DEFAULT_FILTERS };

  const params = new URLSearchParams(
    search.startsWith("?") ? search.slice(1) : search,
  );
  const raw: RawSearchParams = {};

  for (const key of ["priority", "status", "lossType", "sla", "q"]) {
    const value = params.get(key);
    if (value !== null) raw[key] = value;
  }

  return parseQueueFilters(raw);
}

export function buildQueueQuery(filters: QueueFilters): string {
  const params = new URLSearchParams();

  if (filters.priority !== DEFAULT_FILTERS.priority) {
    params.set("priority", filters.priority);
  }
  if (filters.status !== DEFAULT_FILTERS.status) {
    params.set("status", filters.status);
  }
  if (filters.lossType !== DEFAULT_FILTERS.lossType) {
    params.set("lossType", filters.lossType);
  }
  if (filters.sla !== DEFAULT_FILTERS.sla) {
    params.set("sla", filters.sla);
  }
  if (filters.q) {
    params.set("q", filters.q);
  }

  const query = params.toString();
  return query ? `?${query}` : "";
}

function matchesSearch(alert: Alert, term: string): boolean {
  const haystack = [
    alert.alertId,
    alert.claimId,
    alert.claimantId,
    alert.providerId,
    alert.groupId ?? "",
    ...alert.flags.flatMap((flag) => [flag.ruleId, ...flag.evidenceRefs]),
  ]
    .join(" ")
    .toLowerCase();

  return haystack.includes(term.toLowerCase());
}

export function matchesQueueFilters(
  alert: Alert,
  filters: QueueFilters,
  now: Date,
): boolean {
  const priority = alertPriority(alert);

  if (filters.priority === "default") {
    if (priority === "LOW") return false;
  } else if (filters.priority !== "all" && priority !== filters.priority) {
    return false;
  }

  if (filters.status === "default") {
    if (alert.status === "SUPPRESSED") return false;
  } else if (filters.status !== "all" && alert.status !== filters.status) {
    return false;
  }

  if (filters.lossType !== "all" && alert.lossType !== filters.lossType) {
    return false;
  }

  if (
    filters.sla !== "all" &&
    slaState(slaDueAt(alert), now) !== filters.sla
  ) {
    return false;
  }

  if (filters.q && !matchesSearch(alert, filters.q)) {
    return false;
  }

  return true;
}

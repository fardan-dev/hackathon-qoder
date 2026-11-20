import type { AlertStatus, LossType, Severity } from "./schema";
import type { SlaState } from "./sla";

const idrFormatter = new Intl.NumberFormat("id-ID", {
  style: "currency",
  currency: "IDR",
  maximumFractionDigits: 0,
});

const dateTimeFormatter = new Intl.DateTimeFormat("id-ID", {
  timeZone: "Asia/Jakarta",
  dateStyle: "medium",
  timeStyle: "short",
});

export const PRIORITY_LABEL: Record<Severity, string> = {
  HIGH: "High",
  MEDIUM: "Medium",
  LOW: "Low",
};

export const STATUS_LABEL: Record<AlertStatus, string> = {
  OPEN: "Open",
  IN_REVIEW: "In review",
  ESCALATED: "Eskalasi",
  SUPPRESSED: "Suppressed",
};

export const LOSS_TYPE_LABEL: Record<LossType, string> = {
  MOTOR: "Motor",
  PROPERTY: "Properti",
};

export const SLA_LABEL: Record<SlaState, string> = {
  OVERDUE: "Lewat SLA",
  DUE_SOON: "Mendekati SLA",
  ON_TRACK: "Dalam SLA",
};

export function formatIdr(amount: number): string {
  return idrFormatter.format(amount);
}

export function formatDateTimeWib(value: string | Date): string {
  return `${dateTimeFormatter.format(new Date(value))} WIB`;
}

export function formatDuration(durationMs: number): string {
  const totalMinutes = Math.max(0, Math.round(Math.abs(durationMs) / 60_000));
  const days = Math.floor(totalMinutes / 1_440);
  const hours = Math.floor((totalMinutes % 1_440) / 60);
  const minutes = totalMinutes % 60;

  if (days > 0) {
    return hours > 0 ? `${days} hari ${hours} jam` : `${days} hari`;
  }
  if (hours > 0) {
    return minutes > 0 ? `${hours} jam ${minutes} menit` : `${hours} jam`;
  }
  return `${minutes} menit`;
}

export function formatSla(state: SlaState, remainingMs: number): string {
  if (state === "OVERDUE") {
    return `Lewat ${formatDuration(remainingMs)}`;
  }
  return `Sisa ${formatDuration(remainingMs)}`;
}

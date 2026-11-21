import { alertPriority } from "./priority";
import type { Flag } from "./schema";

const MINUTE_MS = 60_000;
const HOUR_MS = 60 * MINUTE_MS;
const WIB_OFFSET_MS = 7 * HOUR_MS;

export const SLA_DUE_SOON_MS = 4 * HOUR_MS;

export type SlaState = "OVERDUE" | "DUE_SOON" | "ON_TRACK";

export function addBusinessDays(from: Date, days: number): Date {
  const wib = new Date(from.getTime() + WIB_OFFSET_MS);
  let remaining = days;

  while (remaining > 0) {
    wib.setUTCDate(wib.getUTCDate() + 1);
    const weekday = wib.getUTCDay();
    if (weekday !== 0 && weekday !== 6) {
      remaining -= 1;
    }
  }

  return new Date(wib.getTime() - WIB_OFFSET_MS);
}

// BRD §7.12: High 24 jam, Medium 3 hari, Low 7 hari kerja.
export function slaDueAt(alert: {
  createdAt: string;
  flags: Flag[];
  slaDueAt?: string;
}): Date {
  if (alert.slaDueAt) {
    const provided = new Date(alert.slaDueAt);
    if (!Number.isNaN(provided.getTime())) return provided;
  }

  const created = new Date(alert.createdAt);

  switch (alertPriority(alert)) {
    case "HIGH":
      return new Date(created.getTime() + 24 * HOUR_MS);
    case "MEDIUM":
      return new Date(created.getTime() + 3 * 24 * HOUR_MS);
    case "LOW":
      return addBusinessDays(created, 7);
  }
}

export function slaRemainingMs(dueAt: Date, now: Date): number {
  return dueAt.getTime() - now.getTime();
}

export function slaState(dueAt: Date, now: Date): SlaState {
  const remaining = slaRemainingMs(dueAt, now);
  if (remaining < 0) return "OVERDUE";
  if (remaining <= SLA_DUE_SOON_MS) return "DUE_SOON";
  return "ON_TRACK";
}

import { createMockAlerts } from "./mock-data";
import type { Alert } from "./schema";

// Swap point for the Go alert service (`GET /v1/alerts`); thin slice runs on seeded data.
export async function listAlerts(now: Date = new Date()): Promise<Alert[]> {
  return createMockAlerts(now);
}

export async function findAlert(
  alertId: string,
  now: Date = new Date(),
): Promise<Alert | undefined> {
  const alerts = await listAlerts(now);
  return alerts.find((alert) => alert.alertId === alertId);
}

export async function findAlertsByGroup(
  groupId: string,
  now: Date = new Date(),
): Promise<Alert[]> {
  const alerts = await listAlerts(now);
  return alerts.filter((alert) => alert.groupId === groupId);
}

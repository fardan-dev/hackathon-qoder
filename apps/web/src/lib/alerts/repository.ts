import { apiGet, ApiError } from "@/lib/api/client";
import { alertApiUrl } from "@/lib/api/config";

import { mapAlert, parseAlertsResponse, type RawAlert } from "./live";
import type { Alert } from "./schema";

export async function listAlerts(token: string): Promise<Alert[]> {
  const payload = await apiGet<unknown>(
    alertApiUrl(),
    "/v1/alerts?limit=500&include_suppressed=true",
    token,
  );
  return parseAlertsResponse(payload);
}

export async function findAlert(
  token: string,
  rawAlertId: string,
): Promise<Alert | undefined> {
  let alertId: string;
  try {
    alertId = decodeURIComponent(rawAlertId);
  } catch {
    throw new ApiError(404, "NOT_FOUND", "Alert tidak ditemukan.");
  }

  // The Go router matches the literal `sha256:` prefix; percent-encoding it
  // makes the lookup miss. Keep the path segment restricted to ID characters.
  if (!/^[A-Za-z0-9:_-]+$/.test(alertId)) {
    throw new ApiError(404, "NOT_FOUND", "Alert tidak ditemukan.");
  }

  const payload = await apiGet<RawAlert>(
    alertApiUrl(),
    `/v1/alerts/${alertId}`,
    token,
  );
  return mapAlert(payload);
}

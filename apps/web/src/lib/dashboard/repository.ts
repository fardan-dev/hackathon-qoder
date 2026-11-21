import { apiGet } from "@/lib/api/client";
import { reportingApiUrl } from "@/lib/api/config";

import { parseDashboard, type DashboardData } from "./live";

export type { DashboardData } from "./live";

export async function getDashboard(token: string): Promise<DashboardData> {
  const payload = await apiGet<unknown>(
    reportingApiUrl(),
    "/v1/dashboard",
    token,
  );
  return parseDashboard(payload);
}

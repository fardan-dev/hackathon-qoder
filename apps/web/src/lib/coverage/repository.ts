import { apiGet } from "@/lib/api/client";
import { reportingApiUrl } from "@/lib/api/config";

import { parseCoverage, type CoverageView } from "./live";

export type { CoverageView } from "./live";

export async function getCoverage(token: string): Promise<CoverageView> {
  const payload = await apiGet<unknown>(
    reportingApiUrl(),
    "/v1/coverage",
    token,
  );
  return parseCoverage(payload);
}

import { apiGet, isUnauthorized } from "@/lib/api/client";
import { caseApiUrl } from "@/lib/api/config";

import { verifyPackage } from "./evidence";

export type { CaseEvidenceView } from "./evidence";
import {
  parseCaseDetail,
  parseCaseList,
  parsePackageVerification,
  type CaseDetail,
  type CaseSummary,
} from "./live";

export type { CaseDetail, CaseSummary } from "./live";

export async function listCases(token: string): Promise<CaseSummary[]> {
  const payload = await apiGet<unknown>(caseApiUrl(), "/v1/cases", token);
  return parseCaseList(payload);
}

export async function findCase(
  token: string,
  caseId: string,
): Promise<CaseDetail> {
  const payload = await apiGet<unknown>(
    caseApiUrl(),
    `/v1/cases/${encodeURIComponent(caseId)}`,
    token,
  );
  const detail = parseCaseDetail(payload, null);
  if (!detail.pkg) return detail;

  try {
    const verificationPayload = await apiGet<unknown>(
      caseApiUrl(),
      `/v1/cases/${encodeURIComponent(caseId)}/package/verify`,
      token,
    );
    return {
      ...detail,
      verification: parsePackageVerification(verificationPayload),
    };
  } catch (error) {
    if (isUnauthorized(error)) throw error;
    return {
      ...detail,
      verification: verifyPackage(detail.evidence, detail.pkg),
    };
  }
}

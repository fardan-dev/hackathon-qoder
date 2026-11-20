import {
  verifyPackage,
  type EvidenceArtifact,
  type EvidencePackage,
  type PackageVerification,
} from "./evidence";
import { createMockEvidence } from "./mock-evidence";

export interface CaseEvidenceView {
  artifacts: EvidenceArtifact[];
  pkg: EvidencePackage | null;
  verification: PackageVerification | null;
}

// Swap point for the Go case service (`GET /v1/cases/{caseID}`); thin slice
// runs on seeded evidence with the same hash semantics as the API.
export async function findCaseEvidence(
  groupId: string,
  now: Date = new Date(),
): Promise<CaseEvidenceView | undefined> {
  const entry = createMockEvidence(now).get(groupId);
  if (!entry) return undefined;

  return {
    artifacts: entry.artifacts,
    pkg: entry.pkg,
    verification: entry.pkg
      ? verifyPackage(entry.artifacts, entry.pkg)
      : null,
  };
}

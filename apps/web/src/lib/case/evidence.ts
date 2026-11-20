import { createHash } from "node:crypto";
import { z } from "zod";

const sha256Pattern = /^sha256:[0-9a-f]{64}$/;

export const chainOfCustodySchema = z.object({
  uploadedBy: z.string().min(1),
  sourceSystem: z.string().min(1),
  receivedAt: z.iso.datetime(),
});
export type ChainOfCustody = z.infer<typeof chainOfCustodySchema>;

export const evidenceArtifactSchema = z.object({
  evidenceId: z.string().min(1),
  artifactType: z.string().min(1),
  source: z.string().min(1),
  contentSha256: z.string().regex(sha256Pattern),
  chainOfCustody: chainOfCustodySchema,
  notes: z.string(),
  createdAt: z.iso.datetime(),
});
export type EvidenceArtifact = z.infer<typeof evidenceArtifactSchema>;

export const evidencePackageSchema = z.object({
  packageId: z.string().min(1),
  packageSha256: z.string().regex(sha256Pattern),
  artifactCount: z.number().int().positive(),
  engineVersion: z.string().min(1),
  rulePackVersion: z.string().min(1),
  frozenAt: z.iso.datetime(),
});
export type EvidencePackage = z.infer<typeof evidencePackageSchema>;

export interface PackageVerification {
  valid: boolean;
  stored: string;
  computed: string;
  artifactCount: number;
  checkedCount: number;
}

export function sha256Of(content: string): string {
  return `sha256:${createHash("sha256").update(content, "utf8").digest("hex")}`;
}

// Mirrors services/case packageHash: canonical JSON (sorted by evidence_id,
// no whitespace) over the evidence hashes.
export function packageHash(artifacts: EvidenceArtifact[]): string {
  const canonical = JSON.stringify(
    [...artifacts]
      .sort((a, b) =>
        a.evidenceId < b.evidenceId ? -1 : a.evidenceId > b.evidenceId ? 1 : 0,
      )
      .map((artifact) => ({
        evidence_id: artifact.evidenceId,
        sha256: artifact.contentSha256,
      })),
  );

  return sha256Of(canonical);
}

export function verifyPackage(
  artifacts: EvidenceArtifact[],
  pkg: EvidencePackage,
): PackageVerification {
  const computed = packageHash(artifacts);

  return {
    valid:
      computed === pkg.packageSha256 &&
      artifacts.length === pkg.artifactCount,
    stored: pkg.packageSha256,
    computed,
    artifactCount: pkg.artifactCount,
    checkedCount: artifacts.length,
  };
}

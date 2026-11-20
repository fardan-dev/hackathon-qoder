import { z } from "zod";

import { rawFlagSchema, mapRawFlag, normalizeSeverity } from "@/lib/alerts/live";
import { lossTypeSchema, type Flag, type LossType, type Severity } from "@/lib/alerts/schema";

import {
  evidenceArtifactSchema,
  evidencePackageSchema,
  type EvidenceArtifact,
  type EvidencePackage,
  type PackageVerification,
} from "./evidence";

const sha256Pattern = /^sha256:[0-9a-f]{64}$/;

const rawCaseSummarySchema = z.object({
  case_id: z.string().min(1),
  title: z.string(),
  summary: z.string(),
  status: z.string().min(1),
  priority: z.number().int(),
  has_open_evidence: z.boolean(),
  claim_count: z.number().int().nonnegative(),
  created_at: z.string().min(1),
  updated_at: z.string().min(1),
  closed_at: z.string().nullish(),
});

const rawClaimSchema = z.object({
  claim_id: z.string().min(1),
  alert_id: z.string().min(1),
  loss_type: z.enum(["MOTOR", "PROPERTY"]),
  severity: z.string().min(1),
  claimed_amount: z.number().int().nonnegative(),
  priority_rank: z.number().int(),
  flags: z.array(rawFlagSchema),
});

const rawEvidenceSchema = z.object({
  evidence_id: z.string().min(1),
  artifact_type: z.string().min(1),
  source: z.string().min(1),
  content_sha256: z.string().regex(sha256Pattern),
  chain_of_custody: z.object({
    collected_by: z.string().min(1),
    source_system: z.string().min(1),
  }),
  notes: z.string(),
  created_at: z.string().min(1),
});

const rawPackageSchema = z.object({
  package_id: z.string().min(1),
  package_sha256: z.string().regex(sha256Pattern),
  artifact_count: z.number().int().positive(),
  engine_version: z.string().min(1),
  rule_pack_version: z.string().min(1),
  frozen_at: z.string().min(1),
});

const rawDecisionSchema = z.object({
  decision_id: z.string().min(1),
  outcome: z.string().min(1),
  reason_code: z.string(),
  reviewer_id: z.string().min(1),
  notes: z.string(),
  value_protected: z.number().int().nonnegative(),
  value_recovered: z.number().int().nonnegative(),
  decided_at: z.string().min(1),
});

const rawCaseDetailSchema = rawCaseSummarySchema.extend({
  claims: z.array(rawClaimSchema),
  evidence: z.array(rawEvidenceSchema),
  package: rawPackageSchema.nullable(),
  decisions: z.array(rawDecisionSchema),
});

const rawVerificationSchema = z.object({
  valid: z.boolean(),
  stored_sha256: z.string(),
  computed_sha256: z.string(),
  artifact_count: z.number().int().nonnegative(),
  engine_version: z.string(),
  rule_pack_version: z.string(),
});

export const rawCaseListResponseSchema = z.object({
  cases: z.array(rawCaseSummarySchema),
  count: z.number().int().nonnegative(),
});

export interface CaseSummary {
  caseId: string;
  title: string;
  summary: string;
  status: string;
  priority: Severity;
  claimCount: number;
  hasOpenEvidence: boolean;
  createdAt: string;
  updatedAt: string;
  closedAt: string | null;
}

export interface CaseClaim {
  claimId: string;
  alertId: string;
  lossType: LossType;
  severity: Severity;
  claimedAmount: number;
  priorityRank: number;
  flags: Flag[];
}

export interface CaseDecision {
  decisionId: string;
  outcome: string;
  reasonCode: string;
  reviewerId: string;
  notes: string;
  valueProtected: number;
  valueRecovered: number;
  decidedAt: string;
}

export interface CaseDetail extends CaseSummary {
  claims: CaseClaim[];
  evidence: EvidenceArtifact[];
  pkg: EvidencePackage | null;
  decisions: CaseDecision[];
  verification: PackageVerification | null;
}

export function casePriority(rank: number): Severity {
  if (rank <= 1) return "HIGH";
  if (rank === 2) return "MEDIUM";
  return "LOW";
}

export function caseStatusLabel(status: string): string {
  const normalized = status.toUpperCase();
  const labels: Record<string, string> = {
    OPEN: "Open",
    CLOSED: "Closed",
    IN_REVIEW: "In review",
    INVESTIGATING: "Investigating",
  };
  return labels[normalized] ?? normalized.replaceAll("_", " ");
}

function mapSummary(
  raw: z.infer<typeof rawCaseSummarySchema>,
): CaseSummary {
  return {
    caseId: raw.case_id,
    title: raw.title,
    summary: raw.summary,
    status: raw.status.toUpperCase(),
    priority: casePriority(raw.priority),
    claimCount: raw.claim_count,
    hasOpenEvidence: raw.has_open_evidence,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
    closedAt: raw.closed_at ?? null,
  };
}

function mapClaim(raw: z.infer<typeof rawClaimSchema>): CaseClaim {
  return {
    claimId: raw.claim_id,
    alertId: raw.alert_id,
    lossType: lossTypeSchema.parse(raw.loss_type),
    severity: normalizeSeverity(raw.severity),
    claimedAmount: raw.claimed_amount,
    priorityRank: raw.priority_rank,
    flags: raw.flags.map(mapRawFlag),
  };
}

function mapEvidenceArtifact(
  raw: z.infer<typeof rawEvidenceSchema>,
): EvidenceArtifact {
  return evidenceArtifactSchema.parse({
    evidenceId: raw.evidence_id,
    artifactType: raw.artifact_type,
    source: raw.source,
    contentSha256: raw.content_sha256,
    chainOfCustody: {
      uploadedBy: raw.chain_of_custody.collected_by,
      sourceSystem: raw.chain_of_custody.source_system,
    },
    notes: raw.notes,
    createdAt: raw.created_at,
  });
}

function mapPackage(
  raw: z.infer<typeof rawPackageSchema> | null,
): EvidencePackage | null {
  if (!raw) return null;
  return evidencePackageSchema.parse({
    packageId: raw.package_id,
    packageSha256: raw.package_sha256,
    artifactCount: raw.artifact_count,
    engineVersion: raw.engine_version,
    rulePackVersion: raw.rule_pack_version,
    frozenAt: raw.frozen_at,
  });
}

function mapDecision(raw: z.infer<typeof rawDecisionSchema>): CaseDecision {
  return {
    decisionId: raw.decision_id,
    outcome: raw.outcome.toUpperCase(),
    reasonCode: raw.reason_code,
    reviewerId: raw.reviewer_id,
    notes: raw.notes,
    valueProtected: raw.value_protected,
    valueRecovered: raw.value_recovered,
    decidedAt: raw.decided_at,
  };
}

export function parseCaseList(payload: unknown): CaseSummary[] {
  return rawCaseListResponseSchema
    .parse(payload)
    .cases.map((item) => mapSummary(item));
}

export function parseCaseDetail(
  payload: unknown,
  verification: PackageVerification | null,
): CaseDetail {
  const raw = rawCaseDetailSchema.parse(payload);

  return {
    ...mapSummary(raw),
    claims: raw.claims.map(mapClaim),
    evidence: raw.evidence.map(mapEvidenceArtifact),
    pkg: mapPackage(raw.package),
    decisions: raw.decisions.map(mapDecision),
    verification,
  };
}

export function parsePackageVerification(payload: unknown): PackageVerification {
  const raw = rawVerificationSchema.parse(payload);
  return {
    valid: raw.valid,
    stored: raw.stored_sha256,
    computed: raw.computed_sha256,
    artifactCount: raw.artifact_count,
    checkedCount: raw.artifact_count,
  };
}

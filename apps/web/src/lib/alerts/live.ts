import { z } from "zod";

import {
  alertSchema,
  alertStatusSchema,
  lossTypeSchema,
  severitySchema,
  type Alert,
  type AlertStatus,
  type Flag,
} from "./schema";

export const rawFlagSchema = z.object({
  rule_id: z.string().min(1),
  rule_version: z.number().int().positive(),
  rule_name: z.string().min(1),
  severity: z.string().min(1),
  trigger_value: z.string().min(1),
  evidence_refs: z.array(z.string().min(1)).min(1),
  explanation: z.string().min(1),
});

const rawAlertSchema = z.object({
  alert_id: z.string().min(1),
  claim_id: z.string().min(1),
  loss_type: z.enum(["MOTOR", "PROPERTY"]),
  claimed_amount: z.number().int().nonnegative(),
  status: z.string().min(1),
  suppressed: z.boolean().optional(),
  sla_due_at: z.string().nullish(),
  created_at: z.string().min(1),
  flags: z.array(rawFlagSchema).min(1),
});

export const rawAlertsResponseSchema = z.object({
  alerts: z.array(rawAlertSchema),
  count: z.number().int().nonnegative(),
});

export type RawAlert = z.infer<typeof rawAlertSchema>;

export function normalizeSeverity(value: string) {
  return severitySchema.parse(value.trim().toUpperCase());
}

function normalizeStatus(value: string, suppressed?: boolean): AlertStatus {
  if (suppressed) return "SUPPRESSED";
  return alertStatusSchema.parse(value.trim().toUpperCase());
}

export function mapRawFlag(raw: z.infer<typeof rawFlagSchema>): Flag {
  return {
    ruleId: raw.rule_id,
    ruleName: raw.rule_name,
    ruleVersion: String(raw.rule_version),
    severity: normalizeSeverity(raw.severity),
    triggerValue: raw.trigger_value,
    evidenceRefs: raw.evidence_refs,
    explanation: raw.explanation,
  };
}

export function mapAlert(raw: RawAlert): Alert {
  return alertSchema.parse({
    alertId: raw.alert_id,
    claimId: raw.claim_id,
    groupId: null,
    lossType: lossTypeSchema.parse(raw.loss_type),
    claimedAmount: raw.claimed_amount,
    status: normalizeStatus(raw.status, raw.suppressed),
    createdAt: raw.created_at,
    slaDueAt: raw.sla_due_at ?? undefined,
    flags: raw.flags.map(mapRawFlag),
  });
}

export function parseAlertsResponse(payload: unknown): Alert[] {
  return rawAlertsResponseSchema
    .parse(payload)
    .alerts.map((alert) => mapAlert(alert));
}

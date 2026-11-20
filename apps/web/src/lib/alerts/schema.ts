import { z } from "zod";

export const severitySchema = z.enum(["HIGH", "MEDIUM", "LOW"]);
export type Severity = z.infer<typeof severitySchema>;

export const alertStatusSchema = z.enum([
  "OPEN",
  "IN_REVIEW",
  "ESCALATED",
  "SUPPRESSED",
]);
export type AlertStatus = z.infer<typeof alertStatusSchema>;

export const lossTypeSchema = z.enum(["MOTOR", "PROPERTY"]);
export type LossType = z.infer<typeof lossTypeSchema>;

export const flagSchema = z.object({
  ruleId: z.string().min(1),
  ruleName: z.string().min(1),
  ruleVersion: z.string().min(1),
  severity: severitySchema,
  triggerValue: z.string().min(1),
  evidenceRefs: z.array(z.string().min(1)).min(1),
  explanation: z.string().min(1),
});
export type Flag = z.infer<typeof flagSchema>;

export const alertSchema = z.object({
  alertId: z.string().min(1),
  claimId: z.string().min(1),
  groupId: z.string().nullable(),
  lossType: lossTypeSchema,
  claimantId: z.string().min(1),
  providerId: z.string().min(1),
  claimedAmount: z.number().int().positive(),
  status: alertStatusSchema,
  createdAt: z.iso.datetime(),
  flags: z.array(flagSchema).min(1),
});
export type Alert = z.infer<typeof alertSchema>;

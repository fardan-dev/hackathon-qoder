import { describe, expect, it } from "vitest";

import { casePriority, parseCaseDetail, parseCaseList } from "./live";

const DETAIL = {
  case_id: "01a1f8f8-4d57-718c-891d-6ec6edabad12",
  title: "Invoice reuse investigation",
  summary: "Synthetic demo case.",
  status: "CLOSED",
  priority: 1,
  owner_id: "90df4bbb-1d42-56fc-9dfc-ac43a0fa5a55",
  version: 2,
  created_at: "2026-11-20T12:39:44.727182Z",
  updated_at: "2026-11-20T12:39:44.824411Z",
  closed_at: "2026-11-20T12:39:44.824411Z",
  claims: [
    {
      claim_id: "CLM-1001",
      alert_id: "sha256:1766438e89ea989b77bbc5456eb65cb25b43651955eba7c5084bbb1854799d62",
      loss_type: "MOTOR",
      severity: "High",
      claimed_amount: 20000000,
      priority_rank: 1,
      flags: [
        {
          rule_id: "A3",
          rule_version: 1,
          rule_name: "Reused Document Hash",
          severity: "High",
          trigger_value: "sha256:0000",
          evidence_refs: ["CLM-1002"],
          explanation: "Hash dokumen yang sama muncul pada klaim lain.",
        },
      ],
    },
  ],
  evidence: [
    {
      evidence_id: "01a1f8f8-4d7d-767e-925f-b3822f94d778",
      artifact_type: "TIMELINE",
      source: "demo-script",
      content_sha256:
        "sha256:8a9d8a27c5442e10b5a2e01a87168820e78f263b4e1df9611a86323d049e4df4",
      chain_of_custody: {
        collected_by: "demo-script",
        source_system: "synthetic dataset",
      },
      notes: "",
      created_at: "2026-11-20T12:39:44.765467Z",
    },
  ],
  package: {
    package_id: "01a1f8f8-4d8e-7725-a57f-a3f8588b0e6e",
    package_sha256:
      "sha256:597745205c218b0e6b8a606dbc44689ce11ba8557a33ed2c1cdd02decc394f85",
    artifact_count: 1,
    engine_version: "rulesengine-1.0.0",
    rule_pack_version: "2026.11.1",
    frozen_at: "2026-11-20T12:39:44.782468Z",
  },
  decisions: [
    {
      decision_id: "01a1f8f8-4db8-7549-9e4b-5e2533a2ecd9",
      outcome: "CONFIRMED",
      reason_code: "INVOICE_REUSE_CONFIRMED",
      reviewer_id: "90df4bbb-1d42-56fc-9dfc-ac43a0fa5a55",
      notes: "",
      value_protected: 20000000,
      value_recovered: 0,
      decided_at: "2026-11-20T12:39:44.825009Z",
    },
  ],
  has_open_evidence: false,
  claim_count: 1,
};

describe("live case mapper", () => {
  it("maps case detail, claims, evidence, and decisions", () => {
    const detail = parseCaseDetail(DETAIL, {
      valid: true,
      stored: DETAIL.package.package_sha256,
      computed: DETAIL.package.package_sha256,
      artifactCount: 1,
      checkedCount: 1,
    });

    expect(detail.priority).toBe("HIGH");
    expect(detail.claims[0]?.flags[0]?.ruleId).toBe("A3");
    expect(detail.evidence[0]?.chainOfCustody.uploadedBy).toBe("demo-script");
    expect(detail.evidence[0]?.chainOfCustody.receivedAt).toBeUndefined();
    expect(detail.pkg?.artifactCount).toBe(1);
    expect(detail.verification?.valid).toBe(true);
    expect(detail.decisions[0]?.outcome).toBe("CONFIRMED");
  });

  it("maps the case list envelope", () => {
    const cases = parseCaseList({ cases: [DETAIL], count: 1 });

    expect(cases).toHaveLength(1);
    expect(cases[0]?.title).toBe("Invoice reuse investigation");
  });

  it("maps priority ranks", () => {
    expect(casePriority(1)).toBe("HIGH");
    expect(casePriority(2)).toBe("MEDIUM");
    expect(casePriority(3)).toBe("LOW");
  });
});

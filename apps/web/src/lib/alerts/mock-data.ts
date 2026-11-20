import { alertSchema, type Alert, type Flag, type Severity } from "./schema";

const RULES = {
  A1: { name: "Reused Invoice", version: "1.4.0", severity: "HIGH" },
  A2: {
    name: "Duplicate Claim Fingerprint",
    version: "1.2.0",
    severity: "HIGH",
  },
  A4: {
    name: "Invoice–Claim Amount Mismatch",
    version: "1.1.0",
    severity: "MEDIUM",
  },
  A5: { name: "Invoice Date Anomaly", version: "1.0.0", severity: "LOW" },
  B1: { name: "High Amount", version: "1.3.0", severity: "MEDIUM" },
  B2: { name: "Batch Outlier", version: "1.1.0", severity: "MEDIUM" },
  B3: {
    name: "Structured Below-Threshold",
    version: "1.0.0",
    severity: "MEDIUM",
  },
  B4: { name: "Round Number", version: "1.0.0", severity: "LOW" },
  B5: {
    name: "Coverage Limit Breach",
    version: "1.2.0",
    severity: "HIGH",
  },
  C1: { name: "Early Claim", version: "1.1.0", severity: "MEDIUM" },
  C3: { name: "Late Reporting", version: "1.0.0", severity: "LOW" },
  C4: {
    name: "Repeat Claims Window",
    version: "1.2.0",
    severity: "MEDIUM",
  },
  C5: {
    name: "Resubmission After Rejection",
    version: "1.1.0",
    severity: "HIGH",
  },
  D1: { name: "Shared Bank Account", version: "1.3.0", severity: "HIGH" },
  D2: {
    name: "Shared Contact Identity",
    version: "1.2.0",
    severity: "MEDIUM",
  },
  D6: { name: "Collusion Ring", version: "1.1.0", severity: "HIGH" },
  D7: {
    name: "Same Asset Multi-Policy",
    version: "1.2.0",
    severity: "HIGH",
  },
  E2: {
    name: "Policy Not Active at Incident",
    version: "1.3.0",
    severity: "HIGH",
  },
  F1: { name: "Submission Velocity", version: "1.1.0", severity: "MEDIUM" },
  G1: { name: "Prior Confirmed Flag", version: "1.2.0", severity: "HIGH" },
} as const satisfies Record<
  string,
  { name: string; version: string; severity: Severity }
>;

type RuleId = keyof typeof RULES;

function flag(
  ruleId: RuleId,
  triggerValue: string,
  evidenceRefs: string[],
  explanation: string,
): Flag {
  const rule = RULES[ruleId];
  return {
    ruleId,
    ruleName: rule.name,
    ruleVersion: rule.version,
    severity: rule.severity,
    triggerValue,
    evidenceRefs,
    explanation,
  };
}

function minutesAgo(now: Date, minutes: number): string {
  return new Date(now.getTime() - minutes * 60_000).toISOString();
}

// Deterministic synthetic batch (BRD §10) so every thin-slice screen works without an API.
export function createMockAlerts(now: Date = new Date()): Alert[] {
  const alerts: Alert[] = [
    {
      alertId: "ALT-1001",
      claimId: "CLM-0001",
      groupId: "CG-0001",
      lossType: "MOTOR",
      claimantId: "C-0142",
      providerId: "P-22",
      claimedAmount: 18_500_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 30 * 60),
      flags: [
        flag(
          "A1",
          "invoice=INV-7781; provider=P-22",
          ["INV-7781", "CLM-0311"],
          "Invoice INV-7781 (Provider P-22, Rp 18.500.000) juga digunakan pada klaim CLM-0311 (12 Sep 2026).",
        ),
        flag(
          "C4",
          "claims_90d=3; limit=2",
          ["C-0142"],
          "Claimant C-0142 mengajukan 3 klaim MOTOR dalam 90 hari (batas: 2).",
        ),
      ],
    },
    {
      alertId: "ALT-1002",
      claimId: "CLM-0002",
      groupId: "CG-0002",
      lossType: "PROPERTY",
      claimantId: "C-0209",
      providerId: "P-31",
      claimedAmount: 185_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 20 * 60),
      flags: [
        flag(
          "B1",
          "amount=185000000; limit=100000000",
          ["CLM-0002"],
          "Nilai klaim Rp 185.000.000 melewati batas absolut PROPERTY Rp 100.000.000.",
        ),
        flag(
          "D1",
          "bank_account_hash=ACC-4471; claimants=3",
          ["ACC-4471", "C-0209", "C-0331"],
          "Rekening BCA ****4471 juga dipakai claimant C-0209 dan C-0331.",
        ),
      ],
    },
    {
      alertId: "ALT-1003",
      claimId: "CLM-0003",
      groupId: null,
      lossType: "MOTOR",
      claimantId: "C-0331",
      providerId: "P-22",
      claimedAmount: 81_500_000,
      status: "IN_REVIEW",
      createdAt: minutesAgo(now, 10 * 60),
      flags: [
        flag(
          "B5",
          "amount=81500000; sum_insured=75000000",
          ["POL-7712"],
          "Nilai klaim Rp 81.500.000 melewati sum insured polis Rp 75.000.000.",
        ),
        flag(
          "A4",
          "invoice=8150000; claim=81500000",
          ["INV-9002"],
          "Invoice INV-9002 terbaca Rp 8.150.000 sedangkan klaim menyebut Rp 81.500.000 (selisih 10 kali).",
        ),
        flag(
          "C1",
          "policy_age_days=12; limit=30",
          ["POL-7712"],
          "Polis baru aktif 12 hari sebelum tanggal kejadian (ambang 30 hari).",
        ),
      ],
    },
    {
      alertId: "ALT-1004",
      claimId: "CLM-0004",
      groupId: null,
      lossType: "PROPERTY",
      claimantId: "C-0456",
      providerId: "P-44",
      claimedAmount: 120_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 50 * 60),
      flags: [
        flag(
          "B1",
          "amount=120000000; limit=100000000",
          ["CLM-0004"],
          "Nilai klaim Rp 120.000.000 melewati batas absolut PROPERTY Rp 100.000.000.",
        ),
        flag(
          "C1",
          "policy_age_days=8; limit=30",
          ["POL-8104"],
          "Polis aktif 8 hari sebelum tanggal kejadian (ambang 30 hari).",
        ),
      ],
    },
    {
      alertId: "ALT-1005",
      claimId: "CLM-0005",
      groupId: null,
      lossType: "MOTOR",
      claimantId: "C-0512",
      providerId: "P-07",
      claimedAmount: 52_400_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 30 * 60),
      flags: [
        flag(
          "B1",
          "amount=52400000; limit=50000000",
          ["CLM-0005"],
          "Nilai klaim Rp 52.400.000 melewati batas absolut MOTOR Rp 50.000.000.",
        ),
        flag(
          "B3",
          "ratio=0.92; band=0.85-1.00",
          ["CLM-0005"],
          "Nilai klaim berada di 92% batas B1 (rentang 85–100%).",
        ),
      ],
    },
    {
      alertId: "ALT-1006",
      claimId: "CLM-0006",
      groupId: "CG-0003",
      lossType: "PROPERTY",
      claimantId: "C-0188",
      providerId: "P-19",
      claimedAmount: 96_000_000,
      status: "ESCALATED",
      createdAt: minutesAgo(now, 70 * 60),
      flags: [
        flag(
          "A2",
          "fingerprint=claimant+incident_date+amount+provider",
          ["CLM-0077"],
          "Fingerprint klaim (claimant, tanggal kejadian, nominal, provider) identik dengan CLM-0077.",
        ),
        flag(
          "D6",
          "linked_claims=4; window_180d",
          ["ACC-5520", "PHONE-3344"],
          "Terdapat 4 klaim terhubung lewat rekening dan kontak yang sama dalam 180 hari.",
        ),
      ],
    },
    {
      alertId: "ALT-1007",
      claimId: "CLM-0007",
      groupId: null,
      lossType: "MOTOR",
      claimantId: "C-0664",
      providerId: "P-22",
      claimedAmount: 17_950_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 5 * 60),
      flags: [
        flag(
          "A1",
          "invoice=INV-7781; provider=P-22",
          ["INV-7781", "CLM-0001"],
          "Invoice INV-7781 juga tercatat pada klaim CLM-0001 dengan provider yang sama.",
        ),
        flag(
          "G1",
          "prior_confirmed_within_24m",
          ["P-22"],
          "Provider P-22 memiliki flag CONFIRMED dalam 24 bulan terakhir.",
        ),
      ],
    },
    {
      alertId: "ALT-1008",
      claimId: "CLM-0008",
      groupId: null,
      lossType: "PROPERTY",
      claimantId: "C-0721",
      providerId: "P-55",
      claimedAmount: 67_500_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 40 * 60),
      flags: [
        flag(
          "C4",
          "claims_90d=3; limit=2",
          ["C-0721"],
          "Claimant C-0721 mengajukan 3 klaim PROPERTY dalam 90 hari (batas: 2).",
        ),
        flag(
          "F1",
          "claims_cross_product_90d=3; limit=2",
          ["C-0721"],
          "Claimant C-0721 mengajukan 3 klaim dalam 90 hari lintas produk (batas: 2).",
        ),
        flag(
          "B3",
          "ratio=0.98; band=0.85-1.00",
          ["CLM-0008"],
          "Nilai klaim berada di 98% batas B1 (rentang 85–100%).",
        ),
      ],
    },
    {
      alertId: "ALT-1009",
      claimId: "CLM-0009",
      groupId: null,
      lossType: "MOTOR",
      claimantId: "C-0803",
      providerId: "P-41",
      claimedAmount: 46_000_000,
      status: "IN_REVIEW",
      createdAt: minutesAgo(now, 2 * 60),
      flags: [
        flag(
          "B1",
          "amount=46000000; limit=50000000",
          ["CLM-0009"],
          "Nilai klaim Rp 46.000.000 berada dekat batas absolut MOTOR Rp 50.000.000.",
        ),
      ],
    },
    {
      alertId: "ALT-1010",
      claimId: "CLM-0010",
      groupId: "CG-0002",
      lossType: "PROPERTY",
      claimantId: "C-0914",
      providerId: "P-12",
      claimedAmount: 230_000_000,
      status: "SUPPRESSED",
      createdAt: minutesAgo(now, 55 * 60),
      flags: [
        flag(
          "B5",
          "amount=230000000; sum_insured=200000000",
          ["POL-6612"],
          "Nilai klaim Rp 230.000.000 melewati sum insured polis Rp 200.000.000.",
        ),
        flag(
          "B1",
          "amount=230000000; limit=100000000",
          ["CLM-0010"],
          "Nilai klaim Rp 230.000.000 melewati batas absolut PROPERTY Rp 100.000.000.",
        ),
        flag(
          "D1",
          "bank_account_hash=ACC-4471; claimants=3",
          ["ACC-4471", "C-0209", "C-0914"],
          "Rekening BCA ****4471 juga dipakai claimant C-0209 dan C-0914.",
        ),
      ],
    },
    {
      alertId: "ALT-1011",
      claimId: "CLM-0011",
      groupId: null,
      lossType: "MOTOR",
      claimantId: "C-1011",
      providerId: "P-63",
      claimedAmount: 12_300_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 100 * 60),
      flags: [
        flag(
          "C3",
          "report_lag_days=42; limit=30",
          ["CLM-0011"],
          "Klaim dilaporkan 42 hari setelah tanggal kejadian (ambang 30 hari).",
        ),
      ],
    },
    {
      alertId: "ALT-1012",
      claimId: "CLM-0012",
      groupId: null,
      lossType: "PROPERTY",
      claimantId: "C-1122",
      providerId: "P-28",
      claimedAmount: 88_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 200 * 60),
      flags: [
        flag(
          "A5",
          "invoice_date_before_incident_days=5",
          ["INV-3320"],
          "Tanggal invoice mendahului tanggal kejadian 5 hari.",
        ),
        flag(
          "B4",
          "amount=88000000; multiple_of=1000000",
          ["CLM-0012"],
          "Nominal klaim Rp 88.000.000 merupakan kelipatan bulat Rp 1.000.000 di atas Rp 5.000.000.",
        ),
      ],
    },
    {
      alertId: "ALT-1013",
      claimId: "CLM-0013",
      groupId: "CG-0001",
      lossType: "MOTOR",
      claimantId: "C-1234",
      providerId: "P-22",
      claimedAmount: 62_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 12 * 60),
      flags: [
        flag(
          "B5",
          "amount=62000000; sum_insured=55000000",
          ["POL-5544"],
          "Nilai klaim Rp 62.000.000 melewati sum insured polis Rp 55.000.000.",
        ),
        flag(
          "E2",
          "incident_date=2026-10-03; policy_end=2026-10-03",
          ["POL-5544"],
          "Tanggal kejadian berada di luar periode aktif polis (berakhir 3 Okt 2026).",
        ),
        flag(
          "A1",
          "invoice=INV-7781; provider=P-22",
          ["INV-7781", "CLM-0001"],
          "Invoice INV-7781 juga tercatat pada klaim CLM-0001 dengan provider yang sama.",
        ),
      ],
    },
    {
      alertId: "ALT-1014",
      claimId: "CLM-0014",
      groupId: null,
      lossType: "PROPERTY",
      claimantId: "C-1345",
      providerId: "P-77",
      claimedAmount: 150_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 27 * 60),
      flags: [
        flag(
          "B1",
          "amount=150000000; limit=100000000",
          ["CLM-0014"],
          "Nilai klaim Rp 150.000.000 melewati batas absolut PROPERTY Rp 100.000.000.",
        ),
        flag(
          "B2",
          "amount_ratio=3.4; batch_population=36",
          ["BATCH-0042"],
          "Nilai klaim 3,4 kali rata-rata PROPERTY pada batch ini (populasi 36 klaim).",
        ),
        flag(
          "D2",
          "phone_hash=PHONE-3344; claimants=3",
          ["PHONE-3344", "C-0209", "C-1345"],
          "Nomor telepon +62 812-****-3344 juga muncul pada claimant C-0209 dan C-1345.",
        ),
      ],
    },
    {
      alertId: "ALT-1015",
      claimId: "CLM-0015",
      groupId: "CG-0003",
      lossType: "MOTOR",
      claimantId: "C-1456",
      providerId: "P-19",
      claimedAmount: 75_000_000,
      status: "ESCALATED",
      createdAt: minutesAgo(now, 32 * 60),
      flags: [
        flag(
          "C5",
          "similar_to=CLM-0042; rejected_days_ago=63",
          ["CLM-0042"],
          "Klaim mirip dengan klaim CLM-0042 yang ditolak 63 hari lalu (fingerprint sama).",
        ),
        flag(
          "G1",
          "prior_confirmed_within_24m",
          ["P-19"],
          "Provider P-19 memiliki flag CONFIRMED dalam 24 bulan terakhir.",
        ),
        flag(
          "D6",
          "linked_claims=4; window_180d",
          ["ACC-5520", "PHONE-3344"],
          "Terdapat 4 klaim terhubung lewat rekening dan kontak yang sama dalam 180 hari.",
        ),
      ],
    },
    {
      alertId: "ALT-1016",
      claimId: "CLM-0016",
      groupId: null,
      lossType: "PROPERTY",
      claimantId: "C-1567",
      providerId: "P-31",
      claimedAmount: 45_000_000,
      status: "OPEN",
      createdAt: minutesAgo(now, 3 * 60),
      flags: [
        flag(
          "D7",
          "asset=AST-3301; active_policies=2",
          ["AST-3301", "POL-3301", "POL-3302"],
          "Aset AST-3301 terdaftar pada 2 polis aktif berbeda.",
        ),
      ],
    },
  ];

  return alerts.map((alert) => alertSchema.parse(alert));
}

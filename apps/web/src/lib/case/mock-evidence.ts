import {
  evidenceArtifactSchema,
  evidencePackageSchema,
  packageHash,
  sha256Of,
  type EvidenceArtifact,
  type EvidencePackage,
} from "./evidence";

interface SeedArtifact {
  evidenceId: string;
  artifactType: string;
  source: string;
  content: string;
  notes: string;
  uploadedBy: string;
  sourceSystem: string;
  daysAgo: number;
}

interface SeedGroup {
  artifacts: SeedArtifact[];
  frozen: {
    packageId: string;
    engineVersion: string;
    rulePackVersion: string;
    daysAgo: number;
  } | null;
}

export interface CaseEvidence {
  artifacts: EvidenceArtifact[];
  pkg: EvidencePackage | null;
}

const GROUPS: Record<string, SeedGroup> = {
  "CG-0001": {
    frozen: {
      packageId: "PKG-20261118-0001",
      engineVersion: "rulesengine-1.0.0",
      rulePackVersion: "2026.11.1",
      daysAgo: 1,
    },
    artifacts: [
      {
        evidenceId: "EV-0001",
        artifactType: "policy_schedule",
        source: "claim:CLM-0001/policy-schedule.pdf",
        content: "demo-evidence|EV-0001|policy-schedule|v1",
        notes:
          "Jadwal polis motor dengan tanggal efektif dan nilai pertanggungan yang diklaim.",
        uploadedBy: "reviewer.demo",
        sourceSystem: "claims-ingest",
        daysAgo: 3,
      },
      {
        evidenceId: "EV-0002",
        artifactType: "payment_proof",
        source: "claim:CLM-0001/payment-transfers.csv",
        content: "demo-evidence|EV-0002|payment-transfers|v1",
        notes: "Riwayat pembayaran premi yang perlu dicocokkan dengan polis.",
        uploadedBy: "reviewer.demo",
        sourceSystem: "claims-ingest",
        daysAgo: 2,
      },
      {
        evidenceId: "EV-0003",
        artifactType: "workshop_estimate",
        source: "claim:CLM-0013/repair-estimate.pdf",
        content: "demo-evidence|EV-0003|repair-estimate|v1",
        notes:
          "Estimasi bengkel rekanan; nilai komponen perlu dibandingkan dengan klaim terkait.",
        uploadedBy: "supervisor.demo",
        sourceSystem: "document-vault",
        daysAgo: 2,
      },
    ],
  },
  "CG-0002": {
    frozen: null,
    artifacts: [
      {
        evidenceId: "EV-0004",
        artifactType: "photo_evidence",
        source: "claim:CLM-0002/photos.zip",
        content: "demo-evidence|EV-0004|photos|v1",
        notes:
          "Foto kerusakan properti; metadata waktu unggah masih perlu verifikasi.",
        uploadedBy: "reviewer.demo",
        sourceSystem: "document-vault",
        daysAgo: 1,
      },
      {
        evidenceId: "EV-0005",
        artifactType: "invoice",
        source: "claim:CLM-0010/invoice.pdf",
        content: "demo-evidence|EV-0005|invoice|v1",
        notes: "Invoice perbaikan properti yang dikaitkan dengan klaim terkait.",
        uploadedBy: "reviewer.demo",
        sourceSystem: "document-vault",
        daysAgo: 1,
      },
    ],
  },
  "CG-0003": {
    frozen: {
      packageId: "PKG-20261119-0003",
      engineVersion: "rulesengine-1.0.0",
      rulePackVersion: "2026.11.1",
      daysAgo: 0.5,
    },
    artifacts: [
      {
        evidenceId: "EV-0006",
        artifactType: "repair_quote",
        source: "claim:CLM-0006/repair-quote.pdf",
        content: "demo-evidence|EV-0006|repair-quote|v1",
        notes:
          "Penawaran perbaikan properti dengan rincian material dan upah.",
        uploadedBy: "reviewer.demo",
        sourceSystem: "claims-ingest",
        daysAgo: 2,
      },
      {
        evidenceId: "EV-0007",
        artifactType: "inspection_report",
        source: "claim:CLM-0015/inspection-report.pdf",
        content: "demo-evidence|EV-0007|inspection-report|v1",
        notes:
          "Laporan inspeksi kendaraan; pola kerusakan perlu dicocokkan dengan kronologi klaim.",
        uploadedBy: "supervisor.demo",
        sourceSystem: "document-vault",
        daysAgo: 1,
      },
    ],
  },
};

function isoDaysAgo(now: Date, days: number): string {
  return new Date(now.getTime() - days * 24 * 60 * 60 * 1000).toISOString();
}

export function createMockEvidence(now: Date = new Date()): Map<string, CaseEvidence> {
  const result = new Map<string, CaseEvidence>();

  for (const [groupId, seed] of Object.entries(GROUPS)) {
    const artifacts = seed.artifacts.map((artifact) =>
      evidenceArtifactSchema.parse({
        evidenceId: artifact.evidenceId,
        artifactType: artifact.artifactType,
        source: artifact.source,
        contentSha256: sha256Of(artifact.content),
        chainOfCustody: {
          uploadedBy: artifact.uploadedBy,
          sourceSystem: artifact.sourceSystem,
          receivedAt: isoDaysAgo(now, artifact.daysAgo),
        },
        notes: artifact.notes,
        createdAt: isoDaysAgo(now, artifact.daysAgo),
      }),
    );

    const pkg = seed.frozen
      ? evidencePackageSchema.parse({
          packageId: seed.frozen.packageId,
          packageSha256: packageHash(artifacts),
          artifactCount: artifacts.length,
          engineVersion: seed.frozen.engineVersion,
          rulePackVersion: seed.frozen.rulePackVersion,
          frozenAt: isoDaysAgo(now, seed.frozen.daysAgo),
        })
      : null;

    result.set(groupId, { artifacts, pkg });
  }

  return result;
}

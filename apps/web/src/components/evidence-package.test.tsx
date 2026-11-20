import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { EvidencePackage } from "./EvidencePackage";
import { findCaseEvidence } from "@/lib/case/repository";

const NOW = new Date("2026-11-20T10:00:00+07:00");

describe("evidence package UI", () => {
  it("shows a verified frozen package with artifact hashes", async () => {
    const evidence = await findCaseEvidence("CG-0001", NOW);
    if (!evidence) throw new Error("missing mock evidence");

    render(
      <EvidencePackage
        evidence={evidence}
        claimIds={["CLM-0001", "CLM-0013"]}
      />,
    );

    expect(
      screen.getByRole("region", { name: "Evidence package" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Hash cocok")).toBeInTheDocument();
    expect(screen.getByText("PKG-20261118-0001")).toBeInTheDocument();
    expect(screen.getByText("CLM-0001 · CLM-0013")).toBeInTheDocument();
    expect(screen.getAllByText(/^sha256:[0-9a-f]{64}$/)).toHaveLength(
      evidence.artifacts.length + 1,
    );
    expect(
      screen.getByText("claim:CLM-0001/policy-schedule.pdf"),
    ).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(
      /cocok dengan rekomputasi/,
    );
  });

  it("shows the open package state before freezing", async () => {
    const evidence = await findCaseEvidence("CG-0002", NOW);
    if (!evidence) throw new Error("missing mock evidence");

    render(<EvidencePackage evidence={evidence} claimIds={["CLM-0002"]} />);

    expect(screen.getByText("Paket terbuka")).toBeInTheDocument();
    expect(screen.getByText("Belum dibekukan")).toBeInTheDocument();
    expect(screen.queryByText("Hash cocok")).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/masih terbuka/);
  });
});

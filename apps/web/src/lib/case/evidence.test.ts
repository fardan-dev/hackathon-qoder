import { describe, expect, it } from "vitest";

import { packageHash, sha256Of, verifyPackage } from "./evidence";
import { createMockEvidence } from "./mock-evidence";
import { findCaseEvidence } from "./repository";

const NOW = new Date("2026-11-20T10:00:00+07:00");

function group(groupId: string) {
  const entry = createMockEvidence(NOW).get(groupId);
  if (!entry) throw new Error(`missing mock group ${groupId}`);
  return entry;
}

describe("sha256Of", () => {
  it("matches the published SHA-256 vector", () => {
    expect(sha256Of("abc")).toBe(
      "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
    );
  });
});

describe("packageHash", () => {
  it("is deterministic regardless of artifact order", () => {
    const { artifacts } = group("CG-0001");
    const reversed = [...artifacts].reverse();

    expect(packageHash(artifacts)).toBe(packageHash(reversed));
  });

  it("changes when an artifact hash changes", () => {
    const { artifacts } = group("CG-0001");
    const tampered = artifacts.map((artifact, index) =>
      index === 0
        ? { ...artifact, contentSha256: sha256Of("tampered") }
        : artifact,
    );

    expect(packageHash(tampered)).not.toBe(packageHash(artifacts));
  });
});

describe("verifyPackage", () => {
  it("verifies frozen mock packages", () => {
    for (const groupId of ["CG-0001", "CG-0003"]) {
      const { artifacts, pkg } = group(groupId);
      if (!pkg) throw new Error(`expected frozen package for ${groupId}`);

      const result = verifyPackage(artifacts, pkg);

      expect(result.valid).toBe(true);
      expect(result.computed).toBe(pkg.packageSha256);
      expect(result.checkedCount).toBe(pkg.artifactCount);
    }
  });

  it("detects tampered evidence", () => {
    const { artifacts, pkg } = group("CG-0001");
    if (!pkg) throw new Error("expected frozen package");

    const tampered = artifacts.map((artifact, index) =>
      index === 0
        ? { ...artifact, contentSha256: sha256Of("tampered") }
        : artifact,
    );
    const result = verifyPackage(tampered, pkg);

    expect(result.valid).toBe(false);
    expect(result.computed).not.toBe(result.stored);
  });

  it("detects missing evidence", () => {
    const { artifacts, pkg } = group("CG-0001");
    if (!pkg) throw new Error("expected frozen package");

    const result = verifyPackage(artifacts.slice(1), pkg);

    expect(result.valid).toBe(false);
    expect(result.checkedCount).toBe(pkg.artifactCount - 1);
  });
});

describe("findCaseEvidence", () => {
  it("returns an open package before freezing", async () => {
    const view = await findCaseEvidence("CG-0002", NOW);

    expect(view).toBeDefined();
    expect(view?.pkg).toBeNull();
    expect(view?.verification).toBeNull();
    expect(view?.artifacts.length).toBeGreaterThan(0);
  });

  it("returns a verified package for a frozen case", async () => {
    const view = await findCaseEvidence("CG-0003", NOW);

    expect(view?.pkg).not.toBeNull();
    expect(view?.verification?.valid).toBe(true);
  });

  it("returns undefined for an unknown case", async () => {
    await expect(findCaseEvidence("CG-9999", NOW)).resolves.toBeUndefined();
  });
});

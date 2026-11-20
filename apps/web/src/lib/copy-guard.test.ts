import { readdirSync, readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, it } from "vitest";

const SRC = join(process.cwd(), "src");
const BANNED = /\b(terbukti|palsu|fraud)\b/i;

function sourceFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) return sourceFiles(path);
    if (!/\.(ts|tsx)$/.test(entry.name) || /\.test\.(ts|tsx)$/.test(entry.name)) {
      return [];
    }
    return [path];
  });
}

describe("UI copy guardrail", () => {
  it("does not use fraud-conclusion wording", () => {
    const violations = sourceFiles(SRC).flatMap((file) =>
      readFileSync(file, "utf8")
        .split("\n")
        .flatMap((line, index) =>
          BANNED.test(line)
            ? [`${relative(SRC, file)}:${index + 1}: ${line.trim()}`]
            : [],
        ),
    );

    expect(violations).toEqual([]);
  });
});

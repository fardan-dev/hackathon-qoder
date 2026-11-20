import { COVERAGE_ROWS, type CoverageRow } from "./coverage";

// Swap point for the Go reporting service (`GET /v1/coverage`); thin slice
// runs on the same seeded matrix as `cmd/seed`.
export async function listCoverage(): Promise<CoverageRow[]> {
  return COVERAGE_ROWS;
}

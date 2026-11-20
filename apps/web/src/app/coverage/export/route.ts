import { coverageCsv } from "@/lib/coverage/coverage";
import { listCoverage } from "@/lib/coverage/repository";

export async function GET() {
  const csv = coverageCsv(await listCoverage());
  return new Response(csv, {
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition": 'attachment; filename="coverage-matrix.csv"',
      "Cache-Control": "no-store",
    },
  });
}

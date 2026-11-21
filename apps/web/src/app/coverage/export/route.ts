import { redirect } from "next/navigation";

import { isUnauthorized } from "@/lib/api/client";
import { getSessionToken } from "@/lib/auth/session";
import { coverageCsv } from "@/lib/coverage/coverage";
import { getCoverage } from "@/lib/coverage/repository";

export async function GET() {
  const token = await getSessionToken();
  if (!token) redirect("/login?next=%2Fcoverage%2Fexport");

  let rows;
  try {
    ({ rows } = await getCoverage(token));
  } catch (error) {
    if (isUnauthorized(error)) {
      redirect("/login?next=%2Fcoverage%2Fexport");
    }
    throw error;
  }

  return new Response(coverageCsv(rows), {
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition": 'attachment; filename="coverage-matrix.csv"',
      "Cache-Control": "no-store",
    },
  });
}

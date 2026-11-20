import type { Metadata } from "next";

import { summarizeCoverage } from "@/lib/coverage/coverage";
import { listCoverage } from "@/lib/coverage/repository";

export const metadata: Metadata = {
  title: "Coverage Tipologi",
};

function formatRate(bps: number): string {
  return `${(bps / 100).toFixed(1).replace(".", ",")}%`;
}

function Stat({
  value,
  label,
  detail,
  tone,
}: {
  value: string | number;
  label: string;
  detail?: string;
  tone?: "warn";
}) {
  return (
    <div className={`stat${tone ? ` stat--${tone}` : ""}`}>
      <span className="stat__value">{value}</span>
      <span className="stat__label">{label}</span>
      {detail ? <span className="stat__detail">{detail}</span> : null}
    </div>
  );
}

export default async function CoveragePage() {
  const rows = await listCoverage();
  const summary = summarizeCoverage(rows);

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <p className="eyebrow">Coverage Phase 0/1</p>
          <h1>Coverage 14 Tipologi</h1>
          <p className="lede">
            Matrix read-only terhadap tipologi in-scope. Status menunjukkan
            cakupan rule, bukan penilaian atas klaim tertentu.
          </p>
        </div>
        <a className="button button--primary" href="/coverage/export">
          Unduh CSV
        </a>
      </div>

      <section className="stats" aria-label="Ringkasan coverage">
        <Stat value={summary.denominator} label="Tipologi in-scope" />
        <Stat value={summary.covered} label="Ter-cover Phase 0/1" />
        <Stat value={summary.residual} label="Residual risk" tone="warn" />
        <Stat
          value={formatRate(summary.coverageBps)}
          label="Coverage"
          detail={`${summary.covered} dari ${summary.denominator} tipologi`}
        />
      </section>

      <div className="table-wrap">
        <table className="queue-table coverage-table">
          <caption className="sr-only">
            Matrix coverage 14 tipologi beserta rule dan residual risk
          </caption>
          <thead>
            <tr>
              <th scope="col">Tipologi</th>
              <th scope="col">Status</th>
              <th scope="col">Rule</th>
              <th scope="col">Residual risk</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.key}>
                <td data-label="Tipologi">
                  <span className="claim-id">{row.label}</span>
                  <span className="cell-sub mono">{row.key}</span>
                </td>
                <td data-label="Status">
                  {row.covered ? (
                    <span className="badge badge--verify-ok">Ter-cover</span>
                  ) : (
                    <span className="badge badge--priority-medium">
                      Residual
                    </span>
                  )}
                </td>
                <td data-label="Rule">
                  <div className="rule-list">
                    {row.ruleIds.map((ruleId) => (
                      <span className="chip" key={ruleId}>
                        {ruleId}
                      </span>
                    ))}
                  </div>
                </td>
                <td data-label="Residual risk">
                  <span className="cell-sub coverage-risk">
                    {row.residualRisk ?? "—"}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <aside className="notice coverage-notice" aria-label="Catatan coverage">
        <strong>Indikatif, bukan kesimpulan.</strong> Tipologi residual dicatat
        sebagai risiko yang diterima sampai analyzer dokumen diaktifkan pada
        fase berikutnya.
      </aside>
    </div>
  );
}

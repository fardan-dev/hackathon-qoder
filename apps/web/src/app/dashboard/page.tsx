import type { Metadata } from "next";
import Link from "next/link";

import { formatDateTimeWib, formatIdr } from "@/lib/alerts/format";
import { redirectOnUnauthorized } from "@/lib/api/guard";
import { requireSession } from "@/lib/auth/session";
import { getDashboard } from "@/lib/dashboard/repository";

export const metadata: Metadata = {
  title: "Dashboard",
};

function formatRate(bps: number): string {
  return `${(bps / 100).toFixed(2).replace(".", ",")}%`;
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
  tone?: "danger" | "warn";
}) {
  return (
    <div className={`stat${tone ? ` stat--${tone}` : ""}`}>
      <span className="stat__value">{value}</span>
      <span className="stat__label">{label}</span>
      {detail ? <span className="stat__detail">{detail}</span> : null}
    </div>
  );
}

export default async function DashboardPage() {
  const { token } = await requireSession("/dashboard");
  const data = await redirectOnUnauthorized(
    getDashboard(token),
    "/dashboard",
  );

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <p className="eyebrow">Reporting projection</p>
          <h1>Dashboard</h1>
          <p className="lede">
            Agregat alert, keputusan, dan coverage untuk periode{" "}
            {formatDateTimeWib(data.from)} — {formatDateTimeWib(data.to)}.
            Tanpa ID klaim atau PII.
          </p>
        </div>
      </div>

      <section className="stats" aria-label="Ringkasan dashboard">
        <Stat
          value={data.alerts.total}
          label="Total alert"
          detail={`${data.alerts.totalFlags} flag · rata-rata ${data.alerts.avgFlags.toFixed(2)}`}
        />
        <Stat
          value={data.alerts.slaBreachedOpen}
          label="SLA terlampaui (open)"
          tone={data.alerts.slaBreachedOpen > 0 ? "danger" : undefined}
        />
        <Stat
          value={data.decisions.total}
          label="Keputusan reviewer"
          detail={`${data.decisions.confirmed} confirmed · ${data.decisions.cleared} cleared`}
        />
        <Stat
          value={formatIdr(data.decisions.valueProtected)}
          label="Nilai dilindungi"
        />
      </section>

      <div className="dashboard-grid">
        <section className="panel" aria-labelledby="alerts-heading">
          <h2 id="alerts-heading">Alert</h2>
          <dl className="meta-grid">
            <div>
              <dt>High</dt>
              <dd>{data.alerts.bySeverity.HIGH}</dd>
            </div>
            <div>
              <dt>Medium</dt>
              <dd>{data.alerts.bySeverity.MEDIUM}</dd>
            </div>
            <div>
              <dt>Low</dt>
              <dd>{data.alerts.bySeverity.LOW}</dd>
            </div>
            <div>
              <dt>Suppressed</dt>
              <dd>{data.alerts.suppressed}</dd>
            </div>
            <div>
              <dt>SLA compliance</dt>
              <dd>{formatRate(data.alerts.slaComplianceBps)}</dd>
            </div>
          </dl>
        </section>

        <section className="panel" aria-labelledby="decisions-heading">
          <h2 id="decisions-heading">Keputusan</h2>
          <dl className="meta-grid">
            <div>
              <dt>Confirmed</dt>
              <dd>{data.decisions.confirmed}</dd>
            </div>
            <div>
              <dt>Cleared</dt>
              <dd>{data.decisions.cleared}</dd>
            </div>
            <div>
              <dt>Needs info</dt>
              <dd>{data.decisions.needsInfo}</dd>
            </div>
            <div>
              <dt>False-positive rate</dt>
              <dd>{formatRate(data.decisions.falsePositiveBps)}</dd>
            </div>
            <div>
              <dt>Nilai dipulihkan</dt>
              <dd>{formatIdr(data.decisions.valueRecovered)}</dd>
            </div>
          </dl>
        </section>

        <section className="panel" aria-labelledby="coverage-heading">
          <h2 id="coverage-heading">Coverage</h2>
          <p className="stat__value">
            {data.coverage.covered}/{data.coverage.denominator}
          </p>
          <p className="hint">
            Tipologi in-scope yang tercakup rule Phase 0/1.
          </p>
          <Link className="button button--ghost" href="/coverage">
            Buka matrix coverage
          </Link>
        </section>
      </div>

      <section className="panel" aria-labelledby="flags-heading">
        <h2 id="flags-heading">Distribusi flag</h2>
        {data.flagDistribution.length > 0 ? (
          <div className="table-wrap">
            <table className="queue-table dashboard-table">
              <caption className="sr-only">
                Jumlah flag per rule pada periode laporan
              </caption>
              <thead>
                <tr>
                  <th scope="col">Rule</th>
                  <th scope="col" className="num">
                    Jumlah flag
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.flagDistribution.map((entry) => (
                  <tr key={entry.ruleId}>
                    <td data-label="Rule">
                      <span className="chip">{entry.ruleId}</span>
                    </td>
                    <td data-label="Jumlah flag" className="num">
                      {entry.count}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="hint">Belum ada flag pada periode ini.</p>
        )}
      </section>

      <aside className="notice" aria-label="Catatan dashboard">
        <strong>Agregat, bukan kesimpulan.</strong> Angka berasal dari
        proyeksi reporting dan tidak mengubah status klaim di sistem sumber.
      </aside>
    </div>
  );
}

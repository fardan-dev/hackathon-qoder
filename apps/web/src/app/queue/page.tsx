import type { Metadata } from "next";

import { AlertTable } from "@/components/AlertTable";
import { EmptyState } from "@/components/EmptyState";
import { FilterBar } from "@/components/FilterBar";
import {
  buildQueueQuery,
  parseQueueFilters,
  type RawSearchParams,
} from "@/lib/alerts/filters";
import { buildQueue } from "@/lib/alerts/queue";
import { listAlerts } from "@/lib/alerts/repository";
import { redirectOnUnauthorized } from "@/lib/api/guard";
import { requireSession } from "@/lib/auth/session";

export const metadata: Metadata = {
  title: "Antrean Review",
};

type SearchParams = Promise<RawSearchParams>;

function Stat({
  value,
  label,
  detail,
  tone,
}: {
  value: number;
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

export default async function QueuePage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  const filters = parseQueueFilters(await searchParams);
  const { token } = await requireSession("/queue");
  const now = new Date();
  const alerts = await redirectOnUnauthorized(
    listAlerts(token),
    "/queue",
  );
  const { rows, overdueCount, highCount } = buildQueue(alerts, filters, now);
  const queueQuery = buildQueueQuery(filters);

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <p className="eyebrow">Reviewer workspace</p>
          <h1>Antrean Review</h1>
          <p className="lede">
            Alert bersifat indikatif dan perlu verifikasi manusia. Urutan
            default: prioritas, lalu jumlah flag, lalu nilai klaim.
          </p>
        </div>
      </div>

      <section className="stats" aria-label="Ringkasan antrean">
        <Stat
          value={rows.length}
          label="Alert tampil"
          detail={`dari ${alerts.length} alert`}
        />
        <Stat value={overdueCount} label="Lewat SLA" tone="danger" />
        <Stat value={highCount} label="Prioritas High" tone="warn" />
        <Stat value={alerts.length} label="Total alert" />
      </section>

      <FilterBar filters={filters} />

      <div className="result-bar">
        <p className="result-count" aria-live="polite">
          Menampilkan <strong>{rows.length}</strong> dari {alerts.length} alert.
        </p>
        {queueQuery ? <span className="chip">Filter aktif</span> : null}
      </div>

      {rows.length > 0 ? (
        <AlertTable rows={rows} queueQuery={queueQuery} />
      ) : (
        <EmptyState />
      )}
    </div>
  );
}

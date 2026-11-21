import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import {
  FlagCountBadge,
  PriorityBadge,
  SlaBadge,
  StatusBadge,
} from "@/components/badges";
import { ApiError } from "@/lib/api/client";
import { redirectOnUnauthorized } from "@/lib/api/guard";
import { requireSession } from "@/lib/auth/session";
import {
  LOSS_TYPE_LABEL,
  formatDateTimeWib,
  formatIdr,
  shortAlertId,
} from "@/lib/alerts/format";
import {
  buildQueueQuery,
  parseQueueSearch,
  type RawSearchParams,
} from "@/lib/alerts/filters";
import { alertPriority } from "@/lib/alerts/priority";
import { findAlert } from "@/lib/alerts/repository";
import { slaDueAt, slaRemainingMs, slaState } from "@/lib/alerts/sla";

type Params = Promise<{ alertId: string }>;
type SearchParams = Promise<RawSearchParams>;

export async function generateMetadata({
  params,
}: {
  params: Params;
}): Promise<Metadata> {
  await params;
  return { title: "Detail alert" };
}

export default async function AlertDetailPage({
  params,
  searchParams,
}: {
  params: Params;
  searchParams: SearchParams;
}) {
  const { alertId } = await params;
  const raw = await searchParams;
  const back = Array.isArray(raw.back) ? raw.back[0] : raw.back;
  const backQuery = buildQueueQuery(parseQueueSearch(back));
  const backHref = `/queue${backQuery}`;

  const { token } = await requireSession(`/alerts/${alertId}`);
  const now = new Date();
  const alert = await redirectOnUnauthorized(
    findAlert(token, alertId),
    `/alerts/${alertId}`,
  ).catch((error: unknown) => {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  });
  if (!alert) notFound();

  const priority = alertPriority(alert);
  const dueAt = slaDueAt(alert);
  const remainingMs = slaRemainingMs(dueAt, now);
  const state = slaState(dueAt, now);
  const flagCount = alert.flags.length;

  return (
    <div className="page">
      <Link href={backHref} className="back-link">
        <span aria-hidden="true">←</span> Kembali ke antrean
      </Link>

      <div className="page-head detail-head">
        <div>
          <p className="eyebrow">
            {shortAlertId(alert.alertId)} ·{" "}
            {LOSS_TYPE_LABEL[alert.lossType]}
          </p>
          <h1>{alert.claimId}</h1>
          <p className="lede">
            Alert indikatif dari {flagCount} rule. Perlu verifikasi reviewer
            sebelum keputusan apa pun.
          </p>
        </div>

        <div className="detail-head__side">
          <div className="badge-row">
            <PriorityBadge priority={priority} />
            <StatusBadge status={alert.status} />
            <SlaBadge state={state} remainingMs={remainingMs} />
          </div>
          {alert.groupId ? (
            <Link
              className="button button--primary"
              href={`/cases/${alert.groupId}${
                backQuery ? `?back=${encodeURIComponent(backQuery)}` : ""
              }`}
            >
              Buka kasus {alert.groupId}
            </Link>
          ) : (
            <Link className="button button--ghost" href="/cases">
              Lihat daftar kasus
            </Link>
          )}
        </div>
      </div>

      <section className="panel" aria-labelledby="summary-heading">
        <h2 id="summary-heading">Ringkasan klaim</h2>
        <dl className="meta-grid">
          <div>
            <dt>Line of business</dt>
            <dd>{LOSS_TYPE_LABEL[alert.lossType]}</dd>
          </div>
          <div>
            <dt>Claimant</dt>
            <dd className="mono">{alert.claimantId ?? "—"}</dd>
          </div>
          <div>
            <dt>Provider</dt>
            <dd className="mono">{alert.providerId ?? "—"}</dd>
          </div>
          <div>
            <dt>Nilai klaim</dt>
            <dd>{formatIdr(alert.claimedAmount)}</dd>
          </div>
          <div>
            <dt>Masuk</dt>
            <dd>
              <time dateTime={alert.createdAt}>
                {formatDateTimeWib(alert.createdAt)}
              </time>
            </dd>
          </div>
          <div>
            <dt>Batas SLA</dt>
            <dd>
              <time dateTime={dueAt.toISOString()}>
                {formatDateTimeWib(dueAt)}
              </time>
            </dd>
          </div>
          <div>
            <dt>Case group</dt>
            <dd className="mono">{alert.groupId ?? "—"}</dd>
          </div>
          <div>
            <dt>Status alert</dt>
            <dd>
              <StatusBadge status={alert.status} />
            </dd>
          </div>
        </dl>
      </section>

      <section className="panel" aria-labelledby="priority-heading">
        <div className="section-head">
          <h2 id="priority-heading">Dasar prioritas</h2>
          <PriorityBadge priority={priority} />
        </div>
        <ol className="rationale">
          <li>
            <span className="rationale__factor">Severity bundle</span>
            <span className="rationale__value">
              Severity tertinggi dari {flagCount} flag
            </span>
          </li>
          <li>
            <span className="rationale__factor">Jumlah flag</span>
            <span className="rationale__value">{flagCount} flag</span>
          </li>
          <li>
            <span className="rationale__factor">Nilai klaim</span>
            <span className="rationale__value">
              {formatIdr(alert.claimedAmount)}
            </span>
          </li>
        </ol>
        <p className="hint">
          Urutan ini deterministik untuk input dan versi rule yang sama; tidak
          ada skor tersembunyi.
        </p>
      </section>

      <section aria-labelledby="flags-heading">
        <div className="section-head">
          <h2 id="flags-heading">Flag dan penjelasan</h2>
          <FlagCountBadge count={flagCount} />
        </div>
        <p className="hint">
          Setiap flag memuat rule dan versi, nilai pemicu, referensi record,
          serta penjelasan bahasa manusia.
        </p>

        <ol className="flag-list">
          {alert.flags.map((flag) => (
            <li key={`${flag.ruleId}-${flag.ruleVersion}`} className="flag-card">
              <div className="flag-card__head">
                <h3>
                  <span className="mono flag-card__rule">{flag.ruleId}</span>
                  {flag.ruleName}
                </h3>
                <span className="flag-card__meta">
                  v{flag.ruleVersion} · <PriorityBadge priority={flag.severity} />
                </span>
              </div>

              <dl className="flag-card__details">
                <div>
                  <dt>Nilai pemicu</dt>
                  <dd className="mono">{flag.triggerValue}</dd>
                </div>
                <div>
                  <dt>Referensi</dt>
                  <dd className="mono">{flag.evidenceRefs.join(" · ")}</dd>
                </div>
                <div>
                  <dt>Penjelasan</dt>
                  <dd>{flag.explanation}</dd>
                </div>
              </dl>
            </li>
          ))}
        </ol>
      </section>

      <aside className="notice" aria-label="Catatan guardrail">
        <strong>Indikatif, bukan kesimpulan.</strong> Alert ini menandai pola
        yang perlu diverifikasi. Status klaim di sistem sumber tidak berubah.
      </aside>
    </div>
  );
}

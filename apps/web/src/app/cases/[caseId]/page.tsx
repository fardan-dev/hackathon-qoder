import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { EvidencePackage } from "@/components/EvidencePackage";
import {
  CaseStatusBadge,
  DecisionBadge,
  FlagCountBadge,
  PriorityBadge,
} from "@/components/badges";
import {
  LOSS_TYPE_LABEL,
  formatDateTimeWib,
  formatIdr,
  shortAlertId,
} from "@/lib/alerts/format";
import { ApiError } from "@/lib/api/client";
import { redirectOnUnauthorized } from "@/lib/api/guard";
import { requireSession } from "@/lib/auth/session";
import { findCase } from "@/lib/case/repository";

type Params = Promise<{ caseId: string }>;

export async function generateMetadata({
  params,
}: {
  params: Params;
}): Promise<Metadata> {
  await params;
  return { title: "Detail kasus" };
}

export default async function CasePage({ params }: { params: Params }) {
  const { caseId } = await params;
  const { token } = await requireSession(`/cases/${caseId}`);
  const detail = await redirectOnUnauthorized(
    findCase(token, caseId),
    `/cases/${caseId}`,
  ).catch((error: unknown) => {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  });

  const totalAmount = detail.claims.reduce(
    (sum, claim) => sum + claim.claimedAmount,
    0,
  );

  return (
    <div className="page">
      <Link href="/cases" className="back-link">
        <span aria-hidden="true">←</span> Kembali ke daftar kasus
      </Link>

      <div className="page-head">
        <div>
          <p className="eyebrow">Case {detail.caseId}</p>
          <h1>{detail.title || `Kasus ${detail.caseId}`}</h1>
          <p className="lede">
            {detail.summary || "Kasus tanpa ringkasan."} {detail.claims.length}{" "}
            klaim terkait · total nilai {formatIdr(totalAmount)}.
          </p>
        </div>

        <div className="detail-head__side">
          <div className="badge-row">
            <CaseStatusBadge status={detail.status} />
            <PriorityBadge priority={detail.priority} />
          </div>
        </div>
      </div>

      <section className="panel" aria-labelledby="case-summary-heading">
        <h2 id="case-summary-heading">Ringkasan kasus</h2>
        <dl className="meta-grid">
          <div>
            <dt>Status</dt>
            <dd>
              <CaseStatusBadge status={detail.status} />
            </dd>
          </div>
          <div>
            <dt>Jumlah klaim</dt>
            <dd>{detail.claimCount}</dd>
          </div>
          <div>
            <dt>Dibuka</dt>
            <dd>
              <time dateTime={detail.createdAt}>
                {formatDateTimeWib(detail.createdAt)}
              </time>
            </dd>
          </div>
          <div>
            <dt>Diperbarui</dt>
            <dd>
              <time dateTime={detail.updatedAt}>
                {formatDateTimeWib(detail.updatedAt)}
              </time>
            </dd>
          </div>
          <div>
            <dt>Ditutup</dt>
            <dd>
              {detail.closedAt ? (
                <time dateTime={detail.closedAt}>
                  {formatDateTimeWib(detail.closedAt)}
                </time>
              ) : (
                "—"
              )}
            </dd>
          </div>
          <div>
            <dt>Bukti terbuka</dt>
            <dd>{detail.hasOpenEvidence ? "Ya" : "Tidak"}</dd>
          </div>
        </dl>
      </section>

      <section aria-labelledby="case-claims-heading">
        <div className="section-head">
          <h2 id="case-claims-heading">Klaim terkait</h2>
        </div>
        <div className="table-wrap">
          <table className="queue-table">
            <caption className="sr-only">
              Klaim pada kasus ini beserta alert dan flag terkait
            </caption>
            <thead>
              <tr>
                <th scope="col">Prioritas</th>
                <th scope="col">Klaim</th>
                <th scope="col">Alert</th>
                <th scope="col">Flag</th>
                <th scope="col" className="num">
                  Nilai klaim
                </th>
              </tr>
            </thead>
            <tbody>
              {detail.claims.map((claim) => (
                <tr key={claim.claimId}>
                  <td data-label="Prioritas">
                    <PriorityBadge priority={claim.severity} />
                  </td>
                  <td data-label="Klaim">
                    <span className="claim-id">{claim.claimId}</span>
                    <span className="cell-sub">
                      {LOSS_TYPE_LABEL[claim.lossType]}
                    </span>
                  </td>
                  <td data-label="Alert">
                    <Link
                      className="row-link"
                      href={`/alerts/${claim.alertId}`}
                    >
                      Buka {shortAlertId(claim.alertId)}
                    </Link>
                  </td>
                  <td data-label="Flag">
                    <FlagCountBadge count={claim.flags.length} />
                    <span className="cell-sub">
                      {claim.flags.map((flag) => flag.ruleId).join(" · ")}
                    </span>
                  </td>
                  <td data-label="Nilai klaim" className="num">
                    {formatIdr(claim.claimedAmount)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {detail.evidence.length > 0 || detail.pkg ? (
        <EvidencePackage
          evidence={{
            artifacts: detail.evidence,
            pkg: detail.pkg,
            verification: detail.verification,
          }}
          claimIds={detail.claims.map((claim) => claim.claimId)}
        />
      ) : (
        <section aria-labelledby="evidence-heading">
          <h2 id="evidence-heading">Evidence package</h2>
          <p className="hint">
            Belum ada artefak bukti yang tercatat untuk kasus ini.
          </p>
        </section>
      )}

      <section aria-labelledby="decisions-heading">
        <div className="section-head">
          <h2 id="decisions-heading">Keputusan reviewer</h2>
          <span className="chip">{detail.decisions.length} keputusan</span>
        </div>

        {detail.decisions.length > 0 ? (
          <ol className="evidence-list">
            {detail.decisions.map((decision) => (
              <li key={decision.decisionId} className="evidence-card">
                <div className="evidence-card__head">
                  <h3>
                    <DecisionBadge outcome={decision.outcome} />{" "}
                    <span className="mono flag-card__rule">
                      {decision.reasonCode || "—"}
                    </span>
                  </h3>
                  <span className="flag-card__meta">
                    <time dateTime={decision.decidedAt}>
                      {formatDateTimeWib(decision.decidedAt)}
                    </time>
                  </span>
                </div>

                <dl className="flag-card__details">
                  <div>
                    <dt>Reviewer</dt>
                    <dd className="mono">{decision.reviewerId}</dd>
                  </div>
                  <div>
                    <dt>Nilai dilindungi</dt>
                    <dd>{formatIdr(decision.valueProtected)}</dd>
                  </div>
                  <div>
                    <dt>Nilai dipulihkan</dt>
                    <dd>{formatIdr(decision.valueRecovered)}</dd>
                  </div>
                  <div>
                    <dt>Catatan</dt>
                    <dd>{decision.notes || "—"}</dd>
                  </div>
                </dl>
              </li>
            ))}
          </ol>
        ) : (
          <p className="hint">Belum ada keputusan tercatat pada kasus ini.</p>
        )}
      </section>

      <aside className="notice" aria-label="Catatan kasus">
        <strong>Indikatif, bukan kesimpulan.</strong> Klaim pada kasus ini masih
        memerlukan verifikasi reviewer; halaman ini tidak mengubah status klaim
        di sistem sumber.
      </aside>
    </div>
  );
}

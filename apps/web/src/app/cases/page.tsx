import type { Metadata } from "next";
import Link from "next/link";

import {
  CaseStatusBadge,
  PriorityBadge,
} from "@/components/badges";
import { formatDateTimeWib } from "@/lib/alerts/format";
import { redirectOnUnauthorized } from "@/lib/api/guard";
import { requireSession } from "@/lib/auth/session";
import { listCases } from "@/lib/case/repository";

export const metadata: Metadata = {
  title: "Kasus",
};

function Stat({ value, label }: { value: number; label: string }) {
  return (
    <div className="stat">
      <span className="stat__value">{value}</span>
      <span className="stat__label">{label}</span>
    </div>
  );
}

export default async function CasesPage() {
  const { token } = await requireSession("/cases");
  const cases = await redirectOnUnauthorized(listCases(token), "/cases");

  const openCount = cases.filter((item) => item.status === "OPEN").length;

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <p className="eyebrow">Case workspace</p>
          <h1>Kasus</h1>
          <p className="lede">
            Kasus mengelompokkan klaim terkait beserta bukti dan keputusan
            reviewer. Semua tampilan read-only.
          </p>
        </div>
      </div>

      <section className="stats" aria-label="Ringkasan kasus">
        <Stat value={cases.length} label="Total kasus" />
        <Stat value={openCount} label="Masih open" />
        <Stat value={cases.length - openCount} label="Closed" />
      </section>

      {cases.length > 0 ? (
        <div className="table-wrap">
          <table className="queue-table">
            <caption className="sr-only">
              Daftar kasus beserta status, prioritas, dan jumlah klaim
            </caption>
            <thead>
              <tr>
                <th scope="col">Kasus</th>
                <th scope="col">Status</th>
                <th scope="col">Prioritas</th>
                <th scope="col" className="num">
                  Klaim
                </th>
                <th scope="col">Bukti terbuka</th>
                <th scope="col">Dibuka</th>
                <th scope="col">
                  <span className="sr-only">Aksi</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {cases.map((item) => (
                <tr key={item.caseId}>
                  <td data-label="Kasus">
                    <Link
                      className="claim-id"
                      href={`/cases/${item.caseId}`}
                    >
                      {item.title || item.caseId}
                    </Link>
                    <span className="cell-sub">{item.summary}</span>
                  </td>
                  <td data-label="Status">
                    <CaseStatusBadge status={item.status} />
                  </td>
                  <td data-label="Prioritas">
                    <PriorityBadge priority={item.priority} />
                  </td>
                  <td data-label="Klaim" className="num">
                    {item.claimCount}
                  </td>
                  <td data-label="Bukti terbuka">
                    {item.hasOpenEvidence ? "Ya" : "Tidak"}
                  </td>
                  <td data-label="Dibuka">
                    <time dateTime={item.createdAt}>
                      {formatDateTimeWib(item.createdAt)}
                    </time>
                  </td>
                  <td data-label="Aksi" className="cell-action">
                    <Link
                      className="row-link"
                      href={`/cases/${item.caseId}`}
                    >
                      Buka kasus
                      <span className="sr-only"> {item.title}</span>
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="empty-state">
          <h2>Belum ada kasus</h2>
          <p>
            Kasus dibuat setelah reviewer menindaklanjuti alert. Mulai dari{" "}
            <Link href="/queue">antrean review</Link>.
          </p>
        </div>
      )}
    </div>
  );
}

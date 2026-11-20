import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { AlertTable } from "@/components/AlertTable";
import { EvidencePackage } from "@/components/EvidencePackage";
import { PriorityBadge } from "@/components/badges";
import {
  buildQueueQuery,
  parseQueueSearch,
  type QueueFilters,
  type RawSearchParams,
} from "@/lib/alerts/filters";
import { formatIdr } from "@/lib/alerts/format";
import { buildQueue } from "@/lib/alerts/queue";
import { findAlertsByGroup } from "@/lib/alerts/repository";
import { findCaseEvidence } from "@/lib/case/repository";

const CASE_FILTERS: QueueFilters = {
  priority: "all",
  status: "all",
  lossType: "all",
  sla: "all",
  q: "",
};

type Params = Promise<{ caseId: string }>;
type SearchParams = Promise<RawSearchParams>;

export async function generateMetadata({
  params,
}: {
  params: Params;
}): Promise<Metadata> {
  const { caseId } = await params;
  return { title: `Kasus ${caseId}` };
}

export default async function CasePage({
  params,
  searchParams,
}: {
  params: Params;
  searchParams: SearchParams;
}) {
  const { caseId } = await params;
  const raw = await searchParams;
  const back = Array.isArray(raw.back) ? raw.back[0] : raw.back;
  const backQuery = buildQueueQuery(parseQueueSearch(back));

  const now = new Date();
  const [alerts, evidence] = await Promise.all([
    findAlertsByGroup(caseId, now),
    findCaseEvidence(caseId, now),
  ]);
  if (alerts.length === 0) notFound();

  const { rows } = buildQueue(alerts, CASE_FILTERS, now);
  const totalAmount = alerts.reduce(
    (sum, alert) => sum + alert.claimedAmount,
    0,
  );
  const topPriority = rows[0]?.priority ?? "LOW";

  return (
    <div className="page">
      <Link href={`/queue${backQuery}`} className="back-link">
        <span aria-hidden="true">←</span> Kembali ke antrean
      </Link>

      <div className="page-head">
        <div>
          <p className="eyebrow">Case group</p>
          <h1>Kasus {caseId}</h1>
          <p className="lede">
            {alerts.length} klaim terkait · total nilai {formatIdr(totalAmount)}.
            Seluruh status klaim sumber tetap read-only.
          </p>
        </div>
        <PriorityBadge priority={topPriority} />
      </div>

      <AlertTable rows={rows} queueQuery={backQuery} />

      {evidence ? (
        <EvidencePackage
          evidence={evidence}
          claimIds={alerts.map((alert) => alert.claimId)}
        />
      ) : null}

      <aside className="notice" aria-label="Catatan kasus">
        <strong>Indikatif, bukan kesimpulan.</strong> Klaim pada kasus ini masih
        memerlukan verifikasi reviewer; tidak ada perubahan status otomatis.
      </aside>
    </div>
  );
}

import Link from "next/link";

import {
  FlagCountBadge,
  PriorityBadge,
  SlaBadge,
  StatusBadge,
} from "@/components/badges";
import {
  LOSS_TYPE_LABEL,
  formatDateTimeWib,
  formatIdr,
} from "@/lib/alerts/format";
import type { QueueRow } from "@/lib/alerts/queue";

export function AlertTable({
  rows,
  queueQuery = "",
}: {
  rows: QueueRow[];
  queueQuery?: string;
}) {
  const backSuffix = queueQuery
    ? `?back=${encodeURIComponent(queueQuery)}`
    : "";

  return (
    <div className="table-wrap">
      <table className="queue-table">
        <caption className="sr-only">
          Daftar alert terurut prioritas, lalu jumlah flag, lalu nilai klaim
        </caption>
        <thead>
          <tr>
            <th scope="col">Prioritas</th>
            <th scope="col">Klaim</th>
            <th scope="col">Flag</th>
            <th scope="col" className="num">
              Nilai klaim
            </th>
            <th scope="col">SLA</th>
            <th scope="col">Status</th>
            <th scope="col">Masuk</th>
            <th scope="col">
              <span className="sr-only">Aksi</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map(({ alert, priority, flagCount, dueAt, sla, remainingMs }) => (
            <tr key={alert.alertId}>
              <td data-label="Prioritas">
                <PriorityBadge priority={priority} />
              </td>
              <td data-label="Klaim">
                <span className="claim-id">{alert.claimId}</span>
                <span className="cell-sub">
                  {LOSS_TYPE_LABEL[alert.lossType]} · {alert.alertId}
                </span>
              </td>
              <td data-label="Flag">
                <FlagCountBadge count={flagCount} />
              </td>
              <td data-label="Nilai klaim" className="num">
                {formatIdr(alert.claimedAmount)}
              </td>
              <td data-label="SLA">
                <SlaBadge state={sla} remainingMs={remainingMs} />
                <span className="cell-sub">
                  Batas{" "}
                  <time dateTime={dueAt.toISOString()}>
                    {formatDateTimeWib(dueAt)}
                  </time>
                </span>
              </td>
              <td data-label="Status">
                <StatusBadge status={alert.status} />
              </td>
              <td data-label="Masuk">
                <time dateTime={alert.createdAt}>
                  {formatDateTimeWib(alert.createdAt)}
                </time>
              </td>
              <td data-label="Aksi" className="cell-action">
                <Link
                  className="row-link"
                  href={`/alerts/${alert.alertId}${backSuffix}`}
                >
                  Buka detail
                  <span className="sr-only"> {alert.claimId}</span>
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

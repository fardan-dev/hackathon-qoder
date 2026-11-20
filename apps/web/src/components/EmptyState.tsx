import Link from "next/link";

export function EmptyState() {
  return (
    <div className="empty-state" role="status">
      <h2>Tidak ada alert yang cocok</h2>
      <p>
        Ubah kombinasi filter atau kosongkan pencarian untuk melihat antrean
        default.
      </p>
      <Link href="/queue" className="button button--primary">
        Reset filter
      </Link>
    </div>
  );
}

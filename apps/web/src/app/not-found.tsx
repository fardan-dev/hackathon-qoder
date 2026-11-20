import Link from "next/link";

export default function NotFound() {
  return (
    <div className="page">
      <div className="empty-state">
        <h1>Alert atau kasus tidak ditemukan</h1>
        <p>Data mungkin sudah berubah atau tautan tidak valid.</p>
        <Link href="/queue" className="button button--primary">
          Kembali ke antrean
        </Link>
      </div>
    </div>
  );
}

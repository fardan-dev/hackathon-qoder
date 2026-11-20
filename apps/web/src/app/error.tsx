"use client";

export default function AppError({ reset }: { reset: () => void }) {
  return (
    <div className="page">
      <div className="empty-state" role="alert">
        <h1>Antrean gagal dimuat</h1>
        <p>
          Data belum bisa diambil. Coba muat ulang; filter yang dipilih tetap
          tersimpan di URL.
        </p>
        <button type="button" className="button button--primary" onClick={reset}>
          Coba lagi
        </button>
      </div>
    </div>
  );
}

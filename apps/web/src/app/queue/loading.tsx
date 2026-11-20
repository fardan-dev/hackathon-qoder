export default function QueueLoading() {
  return (
    <div className="page" aria-busy="true">
      <div className="page-head">
        <div>
          <p className="eyebrow">Reviewer workspace</p>
          <h1>Antrean Review</h1>
          <p className="lede">Memuat antrean alert…</p>
        </div>
      </div>

      <div className="stats" aria-hidden="true">
        {[0, 1, 2, 3].map((item) => (
          <div key={item} className="stat">
            <span className="skeleton skeleton--value" />
            <span className="skeleton skeleton--label" />
          </div>
        ))}
      </div>

      <div className="skeleton skeleton--filters" aria-hidden="true" />

      <div className="table-wrap" aria-hidden="true">
        <div className="skeleton-table">
          {Array.from({ length: 6 }, (_, index) => (
            <div key={index} className="skeleton skeleton--row" />
          ))}
        </div>
      </div>
    </div>
  );
}

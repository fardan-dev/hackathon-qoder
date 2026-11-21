import Link from "next/link";

import type { QueueFilters } from "@/lib/alerts/filters";

const PRIORITY_OPTIONS = [
  { value: "default", label: "High + Medium (default)" },
  { value: "all", label: "Semua prioritas" },
  { value: "HIGH", label: "High" },
  { value: "MEDIUM", label: "Medium" },
  { value: "LOW", label: "Low" },
] as const;

const STATUS_OPTIONS = [
  { value: "default", label: "Aktif (tanpa suppressed)" },
  { value: "all", label: "Semua status" },
  { value: "OPEN", label: "Open" },
  { value: "IN_REVIEW", label: "In review" },
  { value: "ESCALATED", label: "Eskalasi" },
  { value: "SUPPRESSED", label: "Suppressed" },
  { value: "CLOSED", label: "Closed" },
] as const;

const LOSS_TYPE_OPTIONS = [
  { value: "all", label: "Semua lini" },
  { value: "MOTOR", label: "Motor" },
  { value: "PROPERTY", label: "Properti" },
] as const;

const SLA_OPTIONS = [
  { value: "all", label: "Semua SLA" },
  { value: "OVERDUE", label: "Lewat SLA" },
  { value: "DUE_SOON", label: "Jatuh tempo ≤ 4 jam" },
  { value: "ON_TRACK", label: "Dalam SLA" },
] as const;

export function FilterBar({ filters }: { filters: QueueFilters }) {
  return (
    <form className="filter-bar" action="/queue" method="get" role="search">
      <fieldset>
        <legend className="sr-only">Filter antrean</legend>

        <div className="field field--grow">
          <label htmlFor="filter-q">Cari</label>
          <input
            id="filter-q"
            name="q"
            type="search"
            defaultValue={filters.q}
            placeholder="ID klaim, claimant, provider, invoice"
            autoComplete="off"
          />
        </div>

        <div className="field">
          <label htmlFor="filter-priority">Prioritas</label>
          <select
            id="filter-priority"
            name="priority"
            defaultValue={filters.priority}
          >
            {PRIORITY_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="filter-status">Status</label>
          <select id="filter-status" name="status" defaultValue={filters.status}>
            {STATUS_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="filter-lob">Line of business</label>
          <select
            id="filter-lob"
            name="lossType"
            defaultValue={filters.lossType}
          >
            {LOSS_TYPE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="filter-sla">SLA</label>
          <select id="filter-sla" name="sla" defaultValue={filters.sla}>
            {SLA_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="filter-bar__actions">
          <button type="submit" className="button button--primary">
            Terapkan filter
          </button>
          <Link href="/queue" className="button button--ghost">
            Reset
          </Link>
        </div>
      </fieldset>

      <p className="filter-bar__hint">
        Default menampilkan High + Medium; alert suppressed dan Low
        disembunyikan sampai difilter.
      </p>
    </form>
  );
}

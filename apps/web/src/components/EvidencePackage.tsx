import { HashBadge, PackageStateBadge } from "@/components/badges";
import { formatDateTimeWib } from "@/lib/alerts/format";
import type { CaseEvidenceView } from "@/lib/case/repository";

const ARTIFACT_LABEL: Record<string, string> = {
  policy_schedule: "Jadwal polis",
  payment_proof: "Bukti transfer",
  workshop_estimate: "Estimasi bengkel",
  photo_evidence: "Foto kerusakan",
  invoice: "Invoice perbaikan",
  repair_quote: "Penawaran perbaikan",
  inspection_report: "Laporan inspeksi",
};

function artifactLabel(artifactType: string): string {
  return ARTIFACT_LABEL[artifactType] ?? artifactType.replaceAll("_", " ");
}

export function EvidencePackage({
  evidence,
  claimIds,
}: {
  evidence: CaseEvidenceView;
  claimIds: string[];
}) {
  const { artifacts, pkg, verification } = evidence;

  return (
    <section aria-labelledby="evidence-heading">
      <div className="section-head">
        <h2 id="evidence-heading">Evidence package</h2>
        <PackageStateBadge frozen={pkg !== null} />
      </div>
      <p className="hint">
        Setiap artefak menyimpan hash SHA-256 konten dan jejak chain of
        custody. Setelah dibekukan, paket tidak dapat diubah tanpa jejak audit.
      </p>

      <div className="package-card">
        <div className="package-card__head">
          <div>
            <p className="eyebrow">Paket bukti</p>
            <h3>{pkg ? <span className="mono">{pkg.packageId}</span> : "Belum dibekukan"}</h3>
          </div>
          {pkg && verification ? <HashBadge valid={verification.valid} /> : null}
        </div>

        <dl className="meta-grid">
          <div>
            <dt>Package SHA-256</dt>
            <dd>
              {pkg ? (
                <code className="hash">{pkg.packageSha256}</code>
              ) : (
                "—"
              )}
            </dd>
          </div>
          <div>
            <dt>Jumlah artefak</dt>
            <dd>
              {artifacts.length}
              {pkg ? ` dari ${pkg.artifactCount} tercatat` : ""}
            </dd>
          </div>
          <div>
            <dt>Engine</dt>
            <dd className="mono">{pkg?.engineVersion ?? "—"}</dd>
          </div>
          <div>
            <dt>Rule pack</dt>
            <dd className="mono">{pkg?.rulePackVersion ?? "—"}</dd>
          </div>
          <div>
            <dt>Dibekukan</dt>
            <dd>
              {pkg ? (
                <time dateTime={pkg.frozenAt}>
                  {formatDateTimeWib(pkg.frozenAt)}
                </time>
              ) : (
                "—"
              )}
            </dd>
          </div>
          <div>
            <dt>Klaim terkait</dt>
            <dd className="mono">{claimIds.join(" · ")}</dd>
          </div>
        </dl>

        {pkg && verification ? (
          <p
            className={`verify-note verify-note--${verification.valid ? "ok" : "fail"}`}
            role="status"
          >
            {verification.valid
              ? "Hash paket cocok dengan rekomputasi dari hash artefak tersimpan."
              : "Hash paket tidak cocok dengan rekomputasi. Package perlu ditinjau sebelum dipakai."}
          </p>
        ) : (
          <p className="verify-note" role="status">
            Paket masih terbuka: bukti dapat ditambahkan sampai dibekukan.
            Setelah dibekukan, penambahan bukti ditolak dan perubahan meninggalkan
            jejak audit.
          </p>
        )}
      </div>

      <ol className="evidence-list">
        {artifacts.map((artifact) => (
          <li key={artifact.evidenceId} className="evidence-card">
            <div className="evidence-card__head">
              <h3>
                <span className="mono flag-card__rule">
                  {artifact.evidenceId}
                </span>
                {artifactLabel(artifact.artifactType)}
              </h3>
              <span className="flag-card__meta">
                <time dateTime={artifact.createdAt}>
                  {formatDateTimeWib(artifact.createdAt)}
                </time>
              </span>
            </div>

            <dl className="flag-card__details">
              <div>
                <dt>SHA-256 konten</dt>
                <dd>
                  <code className="hash">{artifact.contentSha256}</code>
                </dd>
              </div>
              <div>
                <dt>Sumber</dt>
                <dd className="mono">{artifact.source}</dd>
              </div>
              <div>
                <dt>Chain of custody</dt>
                <dd>
                  {artifact.chainOfCustody.uploadedBy} ·{" "}
                  {artifact.chainOfCustody.sourceSystem} · diterima{" "}
                  {formatDateTimeWib(artifact.chainOfCustody.receivedAt)}
                </dd>
              </div>
              <div>
                <dt>Catatan</dt>
                <dd>{artifact.notes}</dd>
              </div>
            </dl>
          </li>
        ))}
      </ol>
    </section>
  );
}

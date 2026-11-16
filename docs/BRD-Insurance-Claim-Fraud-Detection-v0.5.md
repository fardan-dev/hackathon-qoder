# BRD — Insurance Claim Fraud Detection

**Proyek:** FSI Mini-Hackathon / Take-home using Sonus
**Skenario:** 04 — Insurance (complexity: Medium)
**Versi:** 0.5 (Draft)
**Tanggal:** 16 November 2026
**Status:** Untuk direview tim
**Dokumen terkait:** PRD v0.2, [TRD v0.1](./TRD-Platform-Deteksi-Fraud-Asuransi-2026.md), QA Report (di bawah payung "Ship & Pitch")

### Changelog

| Versi | Perubahan |
|-------|-----------|
| 0.1 | Versi awal: 5 rule (reused invoice, repeat claims, high amount), antrean review, explainability |
| 0.2 | Katalog rule diperluas ke 7 famili / 27 rule; kerangka cakupan (coverage matrix); data polis/provider/identitas; siklus hidup rule; governance; residual risk; roadmap |
| 0.3 | **Tambah kapabilitas OCR & forensik dokumen** (famili H: 10 rule, pipeline dokumen async, chain of custody, dataset manipulasi sintetis). Modul bergerbang: nonaktif pada demo hackathon (constraint "No OCR"), aktif Fase 2 |
| 0.4 | H5 diperjelas: kandidat dHash256 (≤ 16 bit) + konfirmasi kemiripan teks OCR; dataset dokumen sintetis digenerate (`datasets/doc_forgery/`) beserta generator (`tools/generate_doc_dataset.js`) |
| 0.5 | Sinkronisasi dengan PRD v0.1: document control, phase map, default conflict decisions, §3A business case & revenue model, data dictionary, coverage denominator, reason code taxonomy, priority model, revenue metering FR/AC, traceability matrix, N/A register. Koreksi hitungan rule: A–G = 33, total A–H = 43. |
| 0.5 (rev. 2) | Referensi PRD diperbarui ke v0.2 (thin slice, koreksi regulasi, sync register). Tidak ada perubahan keputusan atau scope BRD. |
| 0.5 (rev. 3) | Tambah referensi TRD v0.1 dan pemetaan NFR/acceptance di TRD §23. Tidak ada perubahan keputusan atau scope BRD. |

> **Koreksi hitungan:** changelog 0.2 menyebut "27 rule"; katalog aktual famili A–G berisi 33 rule, dan total A–H berisi 43 rule. Angka pada changelog lama dipertahankan sebagai sejarah, bukan acuan.

---

## 0. Document Control & Scope Contract

### 0.1 Status dokumen

| Item | Nilai |
|---|---|
| Dokumen | BRD — Insurance Claim Fraud Detection |
| Versi | 0.5 |
| Tanggal | 16 November 2026 |
| Status | Draft untuk review tim |
| Pemilik | Product Management |
| Dokumen terkait | PRD Sentinel v0.2, [TRD Sentinel v0.1](./TRD-Platform-Deteksi-Fraud-Asuransi-2026.md), QA Report |
| Hubungan dengan PRD | **Default v0.5 (rev. 2):** BRD ini adalah *Phase 0/1 thin slice* dari PRD v0.2. Konfirmasi final: `TBD` |
| Perubahan utama v0.5 | Sinkronisasi PRD, §3A business case, data dictionary, traceability, default decisions, placeholder revenue |

### 0.2 Apa yang diikat dokumen ini

Dokumen ini mengikat **Phase 0/1**:

- Rules-only, deterministik, tanpa ML.
- Data sintetis; tanpa PII nyata.
- P&C auto dan properti; health/travel/life tidak diaktifkan.
- Pemrosesan lokal/offline; tanpa API pihak ketiga.
- Keputusan klaim tetap di luar sistem; status klaim read-only.

Dokumen ini **bukan**:

- Platform komersial penuh multi-tenant.
- Spesifikasi ML, graph, GenAI, atau integrasi core system.
- Pengganti PRD, TRD, atau QA Report.
- Justifikasi untuk mengklaim kesiapan produksi.

### 0.3 Source of truth per domain

| Domain | Pemilik | Dokumen acuan |
|---|---|---|
| Visi, pasar, model bisnis, GTM jangka panjang | Product | PRD §0–3, §14 |
| Roadmap & fase | Product | PRD §15; BRD memetakan |
| Tipologi, rule, threshold, coverage Phase 0/1 | Fraud/Data Science | BRD §7, §12 |
| Data contract Phase 0/1 | Data Engineering | BRD §10 |
| Target data model | Data/Architecture | PRD §9.2, §18.2 |
| NFR Phase 0/1 | Engineering Lead | BRD §9 |
| NFR target | CTO/SRE | PRD §8 |
| Kepatuhan, keamanan, privasi produksi | Legal/Compliance | PRD §12 |
| Metrik & formula | Product + Data Science | Metric dictionary bersama (`TBD` dibuat) |
| Glossary & ID | Product Ops | Registry bersama (`TBD` dibuat) |
| Acceptance Phase 0/1 | QA | BRD §13 |
| Acceptance pilot/produksi | QA | PRD §18.3 |
| Risiko | Product + Compliance | Risk register bersama (`TBD` dibuat) |

### 0.4 Phase map

| BRD | PRD | Isi |
|---|---|---|
| **Fase 1 — Hackathon MVP** | PRD Phase 0/1 | Rules A–G, queue, flag bundle, evidence, dashboard, audit, coverage matrix |
| **Fase 2 — 1–3 bulan** | PRD Phase 1b/2 | OCR & forensik H1–H7, chain of custody, identity/graph minimal, document viewer |
| **Fase 3 — 3–6 bulan** | PRD Phase 2/3 | ML scoring, CV forensics, SIU case management, external data, consortium |
| — | PRD Phase 1b — Pilot | Backtest → shadow → live pilot dengan design partner |

> Pemetaan ini **bukan 1:1**. Fase 2 BRD memetakan ke sebagian PRD Phase 1b/2; Fase 3 BRD memetakan ke PRD Phase 2/3. Detail kapabilitas ada di §18.2.

### 0.5 Default conflict decisions (v0.5)

| # | Topik | Keputusan default v0.5 | Alasan | Status |
|---|---|---|---|---|
| D-01 | ML | Rules-first di Phase 0/1; ML di Fase 2/3 | Menjaga explainability dan menghindari klaim ML yang belum terbukti | Default — perlu konfirmasi |
| D-02 | Batas produk | Triage-only; status klaim read-only; tanpa write-back | Menjaga batas "flag bukan bukti" dan mengurangi risiko consumer protection | Default — perlu konfirmasi |
| D-03 | Retensi | Per kategori; data demo 12 bulan; produksi `TBD` (PRD: 5–10 tahun) | Menghindari retensi seragam yang tidak punya dasar hukum | Default — Legal `TBD` |
| D-04 | Identitas | Phase 0/1 hash sintetis; produksi tokenization/vault; raw NIK dilarang di aplikasi/log | Menjaga privasi tanpa mematahkan matching | Default — skema `TBD` |
| D-05 | Lini | P&C auto/properti; HEALTH/TRAVEL diterima skema tetapi ditolak Phase 0/1; life N/A | Menyelaraskan dengan PRD MVP P&C | Default — perlu konfirmasi |
| D-06 | Pricing | Platform fee + per-claim + implementation + pilot credit; tanpa per-alert; success fee opsional | Selaras dengan PRD §14.2 | Default — angka `TBD` |
| D-07 | Prioritas | Deterministik: severity → jumlah flag → nilai klaim; expected loss/capacity di Fase 2 | Menjaga explainability Phase 0/1 | Default — perlu konfirmasi |
| D-08 | Semantik | BRD "case group"/"review outcome"; PRD "investigation case"/"claim decision"; tidak dicampur | Menghindari metrik dan UI yang menyesatkan | Default — perlu konfirmasi |

---

## 1. Ringkasan Eksekutif

Tool triase klaim asuransi berbasis aturan tetap: mengimpor data klaim, invoice, polis, provider, dan (opsional) dokumen bukti sintetis, lalu **menandai** klaim yang perlu investigasi dari beberapa sisi — penyalahgunaan dokumen, anomali nominal, pola temporal, relasi antar-pihak, pelanggaran cover, perilaku pengajuan, riwayat, dan **forensik dokumen (deteksi invoice palsu / gambar yang dimanipulasi)**.

Tiga prinsip yang mengikat seluruh dokumen:

1. **Flag bukan bukti.** Sistem tidak menolak klaim, tidak mengubah status klaim, tidak memberi skor probabilitas. Keputusan *confirm* / *clear* selalu oleh reviewer manusia. Ini berlaku lebih keras lagi untuk temuan forensik — anomali metadata atau ELA adalah **indikasi**, bukan kesimpulan.
2. **Cakupan diukur, bukan diklaim.** Tidak ada sistem rule-based yang menangkap semua fraud. BRD mewajibkan **coverage matrix** dan **residual risk register**.
3. **Kapabilitas bertahap.** Lapisan dokumen (OCR + forensik) terdokumentasi penuh di dokumen ini tetapi **nonaktif pada demo hackathon** karena constraint resmi "No ML, OCR". Aktivasinya via feature flag pada Fase 2 sehingga arsitektur sudah siap tanpa mengubah alur inti.

**Tambahan v0.5 — business case:** BRD ini juga menetapkan cara produk menghasilkan revenue dari masalah yang diselesaikan (§3A), dengan formula, unit economics, dan placeholder `TBD` yang harus divalidasi bersama design partner. Angka revenue di dokumen ini **bukan target atau klaim**, melainkan kerangka perhitungan.

---

## 2. Latar Belakang & Problem Statement

### 2.1 Konteks

Tim claims/fraud unit menerima ribuan klaim per periode. Sebagian klaim membawa bukti berupa scan/foto invoice, kuitansi, dan estimasi biaya. Modusnya beragam: dokumen dipakai ulang, nominal digelembungkan, metadata dokumen menunjukkan penyuntingan, teks invoice tidak cocok dengan data klaim, klaim berulang, relasi antar-pihak, hingga pola pengajuan yang menghindari review. Saat ini penemuan hanya mengandalkan sampling manual — dokumen hampir tidak pernah diverifikasi secara sistematis.

### 2.2 Pain points

| # | Masalah | Dampak |
|---|---------|--------|
| P1 | Klaim berulang tidak terdeteksi konsisten | Pola lolos, tergantung keberuntungan sampling |
| P2 | Dokumen dipakai ulang lintas klaim | Potensi pembayaran ganda atas dokumen yang sama |
| P3 | Nominal tidak wajar tidak terprioritaskan | Waktu reviewer salah alokasi |
| P4 | Tidak ada catatan mengapa sebuah klaim dicurigai | Review tidak dapat dipertanggungjawabkan |
| P5 | Keputusan clear/confirm tidak tercatat | Tidak ada pembelajaran dari false positive |
| P6 | Cakupan deteksi tidak diketahui | Tidak diketahui modus mana yang tak terpantau |
| P7 | Modus lintas-klaim (jaringan) tak terlihat | Kolusi tidak pernah muncul di antrean |
| P8 | Dokumen tidak diverifikasi sistematis | Invoice yang diedit/dimanipulasi lolos kecuali kebetulan terlihat |
| P9 | Tidak ada jejak integritas dokumen | Temuan tidak bisa dipakai SIU karena tidak ada chain of custody |

### 2.3 Akar masalah

Tidak ada lapisan triase deterministik yang memunculkan kandidat investigasi **beserta alasannya**, bekerja lintas batch, lintas record, dan lintas modalitas (data terstruktur + dokumen).

### 2.4 Problem–solution fit

> Untuk **perusahaan asuransi P&C mid-market di Indonesia** yang tim SIU-nya masih mengandalkan sampling manual dan tidak memverifikasi dokumen secara sistematis, tool ini memberikan **flag yang dapat dijelaskan, evidence yang dapat diperiksa, serta coverage map yang jujur** pada beban review <15%, berbeda dari **rules engine generik atau sampling manual**, karena **rule pack dipetakan ke tipologi fraud lokal dan setiap flag membawa nilai pemicu serta referensi record/dokumen**.

**Biaya inaction** (formula lengkap di §3A):

| Komponen | Formula | Placeholder |
|---|---|---|
| Jam reviewer terbuang | `alerts × time_wasted_per_alert × loaded_hourly_cost` | `TBD` |
| Fraud lolos | `claims × leakage_rate × avg_claim × non_detection_rate` | `TBD` |
| False positive | `alerts × fp_rate × cost_per_fp` | `TBD` |
| Temuan audit | `findings × cost_per_finding` | `TBD` |
| Dokumentasi tidak terverifikasi | `documents × exposure_per_document` | `TBD` |

---

## 3. Tujuan Bisnis

| ID | Tujuan | Ukuran keberhasilan |
|----|--------|---------------------|
| OBJ-1 | Keluasan deteksi terukur | ≥ 60% tipologi fraud in-scope Phase 0/1 tercakup ≥ 1 rule aktif; ≥ 75% setelah Fase 2; sisanya di residual risk register |
| OBJ-2 | Kasus tanaman terdeteksi | 100% kasus tanaman ter-flag oleh rule yang relevan |
| OBJ-3 | Beban review terkelola | Klaim ter-flag < 15% dari total klaim per batch |
| OBJ-4 | Setiap flag dapat dijelaskan | 100% flag memiliki penjelasan, nilai pemicu, referensi record/dokumen terkait |
| OBJ-5 | Keputusan selalu oleh manusia | 0 status klaim diubah sistem; 100% flag berakhir dengan keputusan teraudit |
| OBJ-6 | Perbaikan berbasis feedback | Precision per rule terukur, dipakai tuning threshold (manual, bulanan) |
| OBJ-7 | Integritas bukti | 100% dokumen asli tersimpan immutable dengan hash; setiap analisis dokumen dapat direproduksi dengan versi engine tercatat |

> Objektif revenue dan unit economics ada di **§3A**. OBJ-1..7 adalah objektif kualitas deteksi, bukan revenue.

---

## 3A. Business Case & Revenue Model

> **Status angka:** seluruh angka di section ini adalah **placeholder `TBD`** sampai divalidasi dengan design partner. Contoh perhitungan bersifat ilustratif untuk menguji formula, bukan target, bukan klaim, dan tidak boleh dikutip sebagai komitmen.

### 3A.1 Nilai yang diselesaikan

Produk ini menghasilkan nilai melalui empat tuas:

| Tuas | Formula | Placeholder | Owner validasi |
|---|---|---|---|
| Waktu reviewer | `alerts × time_saved_per_alert × loaded_hourly_cost` | `TBD` | Product + SIU |
| Fraud dicegah/dipulihkan | `confirmed_cases × avg_claim_value × realization_rate` | `TBD` | Data Science + Finance |
| Friksi false positive | `false_positives_avoided × cost_per_fp` | `TBD` | Product |
| Risiko audit | `findings_avoided × cost_per_finding` | `TBD` | Compliance |

**Aturan atribusi:** nilai fraud dicegah tidak boleh diklaim tanpa baseline dan kontrol. Gunakan holdout/control group atau pilot comparison; selisih yang tidak terverifikasi dicatat sebagai `TBD`, bukan revenue.

### 3A.2 Value proposition

- **Bagi reviewer:** flag yang bisa diperiksa dalam hitungan menit, bukan jam.
- **Bagi Head of Claims/SIU:** coverage map + residual risk, bukan klaim "100% fraud terdeteksi".
- **Bagi CFO:** ROI yang bisa dihitung dari jam reviewer dan recovery, dengan asumsi yang terbuka.
- **Bagi Compliance:** jejak audit, reason code, dan integritas dokumen.

**Customer ROI:**

```text
Customer ROI = (V_time + V_fraud + V_friction + V_audit − Price) / Price
```

Jika ROI belum bisa dihitung karena baseline belum ada, tulis `TBD` dan jangan mengarang angka.

### 3A.3 Revenue stream & pricing model

| Stream | Metrik | Mekanisme | Fase | Catatan |
|---|---|---|---|---|
| Paid pilot | Fixed fee per pilot | Dibayar di muka; dikreditkan ke kontrak tahunan jika lanjut | Phase 0/1 | `TBD` |
| Platform subscription | Volume band klaim/tahun | Tahunan | Phase 1+ | `TBD` |
| Per-claim scoring | Jumlah klaim di-score | Bulanan di belakang | Phase 1+ | Bukan per-alert |
| Document forensics add-on | Jumlah dokumen diproses | Bulanan | Fase 2 | `TBD` |
| Implementation & rule tuning | Proyek one-time | Milestone | Phase 0/1+ | `TBD` |
| Success fee opsional | % verified fraud protected/recovered | Tahunan, capped | Fase 2/3 | Hanya dengan baseline disepakati |

**Prinsip pricing:**

1. Jangan menagih per alert — menghukum deteksi.
2. Per-claim scoring dapat diterima; transparan terhadap overage.
3. Pilot fee dikreditkan ke kontrak tahunan.
4. Success fee opsional, capped, dan diverifikasi dengan kontrol.
5. Harga akhir = `min(value_ceiling, max(cost_floor, competitive_anchor))`.
   - `value_ceiling` = persentase dari quantified annual value (`TBD`, indikatif 10–20%).
   - `cost_floor` = COGS + target gross margin.
   - `competitive_anchor` = `TBD`.

### 3A.4 Formula revenue

```text
ARR = Σ_tenant [F + (C × p) + (D × d) + A] + S

Year-1 cash = ARR + Σ I

Gross profit = ARR − COGS
Gross margin = Gross profit / ARR

ARPA = ARR / jumlah tenant
CAC payback (bulan) = CAC / (ARPA × Gross margin / 12)
LTV = ARPA × Gross margin / churn
LTV/CAC ≥ 3
Break-even tenant = fixed_opex / (ARPA × Gross margin)

Customer ROI = (V_time + V_fraud + V_friction + V_audit − Price) / Price
```

| Variabel | Arti | Placeholder |
|---|---|---|
| `F` | Platform fee tahunan per tenant | `TBD` |
| `C` | Klaim di-score per tahun | `TBD` |
| `p` | Harga per klaim | `TBD` |
| `D` | Dokumen diproses per tahun | `TBD` |
| `d` | Harga per dokumen | `TBD` |
| `A` | Add-on/layanan lain | `TBD` |
| `S` | Success fee (opsional) | `TBD` |
| `I` | Implementation fee one-time | `TBD` |
| `COGS` | Cloud, OCR, data, support, delivery | `TBD` |
| `CAC` | Sales & marketing cost per tenant | `TBD` |
| `churn` | Annual logo/revenue churn | `TBD` |

### 3A.5 Unit economics

**Komponen COGS:**

- Cloud/compute dan storage.
- OCR engine/lisensi dan document processing.
- Data acquisition (bila ada).
- Support & customer success.
- Implementation/professional services.
- Compliance/audit (proporsional).

**Guardrail dari PRD §14.4:**

| Metrik | Target | Status BRD |
|---|---|---|
| Gross margin SaaS | ≥ 70% | `TBD` — hitung setelah pilot |
| CAC payback | ≤ 18 bulan | `TBD` |
| Net revenue retention | ≥ 110% | `TBD` — Fase 2+ |
| LTV/CAC | ≥ 3 | `TBD` |

### 3A.6 Contoh perhitungan (ilustratif, bukan target)

| Komponen | Formula | Contoh |
|---|---|---|
| Klaim/tahun | `C` | 150.000 |
| Flag rate | — | 12% → 18.000 alert |
| Platform fee | `F` | IDR 180 jt |
| Per-claim | `C × p` | 150.000 × IDR 600 = IDR 90 jt |
| Document add-on | `D × d` | 15.000 × IDR 3.000 = IDR 45 jt |
| **ARR** | `F + Cp + Dd` | **IDR 315 jt** |
| Implementation | `I` | IDR 90 jt |
| **Year-1 cash** | `ARR + I` | **IDR 405 jt** |
| COGS | 30% ARR | IDR 94,5 jt |
| Gross margin | 70% | IDR 220,5 jt |
| CAC | — | IDR 250 jt |
| CAC payback | `250 / (315 × 0,7 / 12)` | ±13,6 bulan |
| Churn | — | 15% |
| LTV | `315 × 0,7 / 0,15` | IDR 1,47 M |
| LTV/CAC | — | ±5,9x |

> Contoh di atas hanya menguji konsistensi formula. Ganti semua angka dengan data pilot sebelum dipakai untuk keputusan harga.

### 3A.7 Revenue objectives

| ID | Objektif | Target | Waktu |
|---|---|---|---|
| R-1 | Paid pilot | 2 pilot berbayar | Phase 0/1 |
| R-2 | Konversi pilot | ≥ 1 pilot menjadi kontrak tahunan | ≤ 6 bulan setelah pilot |
| R-3 | ARR awal | `TBD` | Tahun pertama |
| R-4 | Gross margin | ≥ 70% | Steady state |
| R-5 | CAC payback | ≤ 18 bulan | Setelah 2 tenant |
| R-6 | Customer ROI | ≥ 3x (dengan kontrol) | Setelah pilot |
| R-7 | Validasi WTP | ≥ 10 wawancara | Phase 0 |
| R-8 | NRR | ≥ 110% | Fase 2+ |

### 3A.8 Skenario & sensitivitas

| Skenario | Tenant | Klaim/tenant | ARR | Gross margin | CAC payback | LTV/CAC |
|---|---|---|---|---|---|---|
| Conservative | 1 | `TBD` | `TBD` | `TBD` | `TBD` | `TBD` |
| Base | 2 | `TBD` | `TBD` | `TBD` | `TBD` | `TBD` |
| Aggressive | 4 | `TBD` | `TBD` | `TBD` | `TBD` | `TBD` |

Sensitivitas wajib diuji untuk: harga per klaim, flag rate, biaya OCR/dokumen, churn, dan biaya sales.

### 3A.9 Asumsi, validasi, dan risiko revenue

**Asumsi yang harus divalidasi:**

| ID | Asumsi | Placeholder | Sumber | Owner | Validate by |
|---|---|---|---|---|---|
| A-01 | Profil tenant awal (P&C mid-market) | `TBD` | Design partner | Product | `TBD` |
| A-02 | Volume klaim/tahun | `TBD` | Design partner | Product | `TBD` |
| A-03 | Flag rate | 15% max (OBJ-3) | Dataset/pilot | Data | `TBD` |
| A-04 | Waktu review per flag | `TBD` | Time study | Product | `TBD` |
| A-05 | Loaded cost reviewer | `TBD` | HR/Finance | Finance | `TBD` |
| A-06 | Leakage fraud | `TBD` | Audit/claims | Compliance | `TBD` |
| A-07 | Nilai klaim rata-rata | `TBD` | Design partner | Finance | `TBD` |
| A-08 | Harga/WTP | `TBD` | 10–15 interviews | Product | `TBD` |
| A-09 | Cost to serve | `TBD` | Pilot | Engineering | `TBD` |
| A-10 | Churn | `TBD` | Pilot cohort | Product | `TBD` |
| A-11 | CAC | `TBD` | Sales pipeline | Sales | `TBD` |
| A-12 | Konversi pilot | `TBD` | Pilot | Product | `TBD` |
| A-13 | Sales cycle | `TBD` | Pipeline | Sales | `TBD` |
| A-14 | Volume dokumen | `TBD` | Design partner | Product | `TBD` |
| A-15 | Retensi legal | `TBD` | Legal | Compliance | `TBD` |

**Rencana validasi:**

1. 10–15 wawancara willingness-to-pay dengan Head of Claims/SIU, CFO, dan Compliance.
2. Dua paid pilot dengan success criteria dan harga pilot yang disepakati.
3. Uji tiga paket harga (floor/base/ceiling) pada pilot.
4. Pilot scorecard: flag rate, confirmation rate, review time, value protected/recovered.
5. Recalculate unit economics setelah pilot; jangan mempertahankan angka pra-pilot.
6. Review Legal/Compliance untuk kontrak, DPA, dan retensi.

**Risiko revenue:**

| ID | Risiko | Mitigasi | Owner |
|---|---|---|---|
| R-11 | Atribusi fraud dicegah diperdebatkan | Baseline + control group + formula disepakati | Product + Finance |
| R-12 | Budget owner tidak jelas (claims vs risk vs IT) | Multi-stakeholder mapping; pilot sponsor eksekutif | Sales |
| R-13 | Sensitivitas harga pasar Indonesia | Tiered packaging; pilot credit; value-based proof | Product |
| R-14 | Biaya OCR/dokumen membengkak | Kuota dokumen; confidence routing; local OCR | Engineering |
| R-15 | Pilot tidak konversi karena integrasi | Scope pilot tanpa integrasi; success criteria jelas | Product |
| R-16 | Churn setelah pilot | Onboarding, training, quarterly value review | Customer Success |

**Revenue recognition:**

- Subscription: ratable sepanjang periode.
- Implementation: milestone.
- Usage/per-claim: bulanan di belakang.
- Pilot: fixed fee, dikreditkan bila lanjut.
- Success fee: tahunan, setelah verifikasi.

---

## 4. Ruang Lingkup

### 4.1 In Scope

- Impor file **CSV/JSON sintetis**: klaim, invoice, polis, provider, (opsional) peta identitas.
- Unggah dokumen bukti sintetis (JPEG/PNG/PDF): invoice scan, kuitansi, estimasi.
- Validasi skema & data quality saat impor.
- Rule engine aturan tetap (deterministik, tanpa ML) — **8 famili, 43 rule** (§7).
- **Forensik dokumen (Fase 2, bergerbang):** OCR lokal, ekstraksi metadata (EXIF/PDF), deteksi near-duplicate gambar, deteksi jejak penyuntingan, cross-check teks OCR vs field terstruktur, laporan forensik per dokumen.
- Flag bundle per klaim dengan evidence + penjelasan bahasa manusia.
- Antrean review + pengelompokan klaim terkait menjadi *case group*.
- Keputusan reviewer, suppression terbatas dengan kedaluwarsa.
- Dashboard, coverage matrix, ekspor, audit trail, **evidence package** (termasuk laporan forensik & hash dokumen).
- Rule registry dengan versi + backtesting.
- **Usage metering, value report, dan billing export** untuk mendukung revenue (§8.3).
- Web app responsif; opsional aplikasi mobile pada fase pitch.

### 4.2 Out of Scope

- **Machine learning / scoring model / prediksi** (termasuk model CV berbasis deep learning untuk forensik) — Fase 2/3, bukan Phase 0/1.
- **Penolakan klaim otomatis / perubahan status klaim / write-back ke claims system** — tetap di luar sistem.
- Rekonstruksi 3D, analisis pola noise kamera tingkat lanjut, deteksi copy-move berbasis model — Fase 3.
- Investigasi lanjutan (wawancara, survei, keputusan bayar) — di luar sistem; output diteruskan ke SIU.
- Integrasi core system asuransi / API pihak ketiga / OCR cloud (semua pemrosesan **offline lokal**) — Phase 0/1.
- Data eksternal (blacklist industri, data kepolisian).
- Audit internal (fraud oleh pegawai/agen dari sisi akses sistem).
- Underwriting analytics (non-disclosure saat polis diterbitkan).
- Data produksi / PII nyata — hanya data sintetis.
- Multi-tenant, SSO/MFA, data residency, DSR/DPIA — N/A Phase 0/1; produksi mengacu PRD §12.

> Catatan: OCR & forensik dokumen **tidak lagi out-of-scope** — statusnya *in scope, fase 2, default nonaktif*. Constraint hackathon "No ML, OCR" dihormati lewat feature flag (§4.4).

### 4.3 Asumsi

- Riwayat batch tersimpan dan diperhitungkan rule lintas-batch.
- Setiap klaim merujuk maksimal satu invoice utama; satu polis merujuk satu aset.
- Atribut identitas (rekening, telepon, email) tersedia dalam bentuk **hash** pada data sintetis.
- Untuk klaim yang menyertakan dokumen: scan JPEG/PNG/PDF satu halaman, ≤ 10 MB per dokumen, maks. 3 dokumen per klaim.
- Volume tipikal: 5.000–50.000 klaim per batch; dokumen dianalisis untuk klaim yang memenuhi kriteria aktivasi (default: nominal ≥ 20 jt atau ter-flag ≥ 1 rule famili A–G), maks. 1.000 dokumen/batch.
- Mata uang tunggal (IDR) untuk MVP.
- **Phase 0/1 hanya menerima `loss_type` MOTOR dan PROPERTY.** HEALTH/TRAVEL tetap ada di skema untuk kompatibilitas, tetapi ditolak dengan pesan "unsupported line in Phase 0/1". Life tidak ada di skema.
- **Tidak ada PII nyata.** Semua identitas di-hash; tidak ada raw NIK, telepon, atau rekening di data demo.

### 4.4 Constraint

- Data CSV/JSON + dokumen sintetis; fixed rules; **flags are not proof**.
- Tidak ada akses internet/API pihak ketiga; OCR berjalan lokal (mis. Tesseract, pinned version).
- **Constraint hackathon: "No ML, OCR"** → modul dokumen (famili H + pipeline OCR) **nonaktif** pada demo MVP kecuali tim memutuskan lain secara eksplisit. Aktivasi = konfigurasi, bukan perubahan kode.
- Budget & waktu mengikuti mini-hackathon (Diagnose → Build → Stress-test).
- **Tidak ada integrasi claims system dan tidak ada write-back status klaim di Phase 0/1.**

### 4.5 Filosofi Cakupan (mengikat)

1. **Bertahan berlapis.** Setiap modus diusahakan tertangkap minimal dua sudut berbeda (contoh: invoice ganda terlihat dari rule dokumen A1 *dan* rule relasi provider D3). Satu rule meleset tidak membuat modus hilang total.
2. **Tidak ada cakupan 100%.** Rule bersifat retrospektif dan hanya melihat data/bukti yang tersedia. Fraud yang memerlukan forensik tingkat lanjut, data eksternal, atau pola baru tidak mungkin tertangkap — dan **harus didaftarkan**, bukan disembunyikan.
3. **Cakupan diukur per tipologi.** Setiap rule dipetakan ke tipologi fraud (§12.1). Coverage matrix adalah deliverable demo.
4. **Presisi dijaga agar cakupan tidak menjadi noise.** Setiap rule baru wajib lolos uji false-positive pada klaim/dokumen kontrol sebelum diaktifkan.
5. **Temuan forensik bersifat indikatif.** Metadata hilang bukan bukti manipulasi (banyak screenshot sah menghapus EXIF); anomali ELA punya false positive tinggi. Karena itu severity forensik diatur konservatif dan selalu disertai pembanding visual bagi reviewer.

### 4.6 Definisi cakupan & denominator

- **Coverage denominator Phase 0/1** = tipologi in-scope untuk P&C auto/properti yang dapat dideteksi dari data Phase 0/1. Tipologi out-of-scope (staged accident/arson, fraud internal, non-disclosure underwriting, modus baru) tetap dicantumkan di §12.1 tetapi **tidak dihitung** dalam denominator.
- **Covered** = memiliki ≥ 1 rule aktif yang dipetakan; **partial** = cakupan terbatas dan dicatat; **not covered** = tidak ada rule.
- Dengan definisi ini: 14 tipologi in-scope; Phase 0/1 covered = 10/14 (71%) → memenuhi OBJ-1 ≥60%; Fase 2 covered = 13/14 (93%) → memenuhi OBJ-1 ≥75%.
- Tanpa definisi ini, angka bisa turun ke 10/18 (56%) dan tampak gagal. Denominator harus selalu ditulis bersama metrik.

---

## 5. Stakeholder & Persona

| Peran | Deskripsi | Kebutuhan utama | Mapping PRD |
|---|---|---|---|
| Claims Reviewer | Pengguna harian tim claims | Melihat flag + alasannya, memutuskan confirm/clear cepat | Claims Adjuster / Investigator SIU |
| Claims/Fraud Supervisor | Pengawas | Oversight, ekspor, reopen batch, persetujuan suppression | Kepala SIU / Claims Integrity Lead |
| Admin | Pengelola sistem | Konfigurasi threshold, aktivasi rule, cakupan analisis dokumen, manajemen user | Platform Admin |
| SIU Investigator (downstream) | Penerima output, di luar sistem | Evidence package lengkap dengan chain of custody | Investigator SIU |
| Reviewer/Penguji (juri) | Evaluator eksternal | Bukti kerja end-to-end, demo, QA report, coverage matrix | Auditor/Regulator (read-only) |
| Finance/Management (read-only) | Melihat usage, value report, billing export | Revenue dan ROI terukur | Finance |
| Compliance/Legal | Review kepatuhan | Retensi, DPA, bahasa UI, audit trail | Risk & Compliance Officer |
| Data/IT Engineer | Integrasi masa depan | API, mapping, observability | Data/IT Engineer |
| Underwriter | N/A Phase 0/1 | N/A | Underwriter |
| Provider Network Manager | N/A Phase 0/1 | N/A | Provider Network Manager |

> Persona PRD yang belum punya peran di Phase 0/1 ditandai N/A dan tidak boleh dianggap sudah didukung.

---

## 6. Proses Bisnis

### 6.1 As-Is

1. Klaim masuk dan diproses di sistem inti (di luar tool).
2. Fraud unit sampling manual dari spreadsheet; dokumen diperiksa sekilas.
3. Kecurigaan dicatat informal tanpa struktur.
4. Klaim lanjut tanpa jejak analisis; modus lintas-klaim tidak pernah terlihat.

### 6.2 To-Be

1. Admin/Reviewer mengunggah klaim, invoice, polis, provider; reviewer mengunggah dokumen bukti.
2. Sistem memvalidasi data dan menolak re-impor file identik.
3. Rule engine aktif menandai klaim terhadap batch berjalan + riwayat (famili A–G).
4. *(Fase 2)* Pipeline dokumen berjalan async: hash → metadata → OCR → cross-check → laporan forensik; flag famili H ditambahkan ke bundle.
5. Flag digabung per klaim; klaim terkait dikelompokkan menjadi *case group*.
6. Reviewer memutuskan per flag/bundle: **CONFIRMED**, **CLEARED** (reason code), **NEEDS_INFO**; dapat membuka penampil dokumen side-by-side.
7. Supervisor memantau dashboard & coverage matrix, mengekspor evidence package, menutup batch.
8. Reason code terkumpul → tuning threshold bulanan → rule diaktifkan/dipensiunkan dengan versi baru.
9. *(v0.5)* Sistem mencatat usage dan menghasilkan value report untuk review revenue/ROI.

---

## 7. Katalog Rule (Rule Engine)

Semua rule deterministik: input sama + versi rule sama → output identik. Tidak ada skor risiko gabungan.

**Hitungan katalog:** A–G = 33 rule; H = 10 rule; total = 43 rule.

### 7.1 Famili A — Dokumen & Duplikasi (data terstruktur)

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| A1 | Reused Invoice | `invoice_number` + `provider_id` sama pada ≥ 2 `claim_id` berbeda, termasuk riwayat batch | — | High | Must |
| A2 | Duplicate Claim Fingerprint | `claimant_id` + `incident_date` + `claimed_amount` + `provider_id` muncul > 1 kali | — | High | Must |
| A3 | Reused Document Hash | hash dokumen sama pada claimant/polis berbeda | — | High | Should |
| A4 | Invoice–Claim Amount Mismatch | `invoice_amount` ≠ `claimed_amount` (toleransi 0) | 0 | Medium | Should |
| A5 | Invoice Date Anomaly | `invoice_date` < `incident_date` atau > `claim_date` + 30 hari | 30 hari | Low | Could |
| A6 | Invoice Identity Conflict | `invoice_number` sama tetapi `provider_id`/amount berbeda antar file | — | High | Should |

### 7.2 Famili B — Anomali Nominal

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| B1 | High Amount | `claimed_amount` > batas absolut per `loss_type` | MOTOR 50 jt; PROPERTY 100 jt; HEALTH/TRAVEL N/A Phase 0/1 | Medium | Must |
| B2 | Batch Outlier | `claimed_amount` > 3 × rata-rata `loss_type` dalam batch; aktif bila populasi ≥ 30 klaim | 3×, min 30 | Medium | Should |
| B3 | Structured Below-Threshold | nominal dalam rentang 85%–100% dari batas B1 | 85–100% | Medium | Should |
| B4 | Round Number | nominal ≥ 5 jt dan kelipatan 1.000.000 tepat | 5 jt | Low | Could |
| B5 | Coverage Limit Breach | `claimed_amount` > `sum_insured` polis | — | High | Must |
| B6 | Aggregate Near Limit | akumulasi klaim claimant per polis 12 bulan > 80% limit | 80%, 12 bln | Medium | Should |

### 7.3 Famili C — Pola Temporal

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| C1 | Early Claim | klaim pada polis dengan usia ≤ 30 hari | 30 hari | Medium | Should |
| C2 | Post-Reinstatement Claim | klaim ≤ 30 hari setelah reinstatement polis yang pernah lapsed | 30 hari | High | Should |
| C3 | Late Reporting | `claim_date` − `incident_date` > 30 hari | 30 hari | Low | Could |
| C4 | Repeat Claims Window | claimant + `loss_type` sama: ≥ 2 klaim dalam 90 hari bergulir | 90 hari | Medium | Must |
| C5 | Resubmission After Rejection | klaim baru berfingerprint mirip klaim `REJECTED` ≤ 180 hari | 180 hari | High | Should |

### 7.4 Famili D — Jaringan & Relasi

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| D1 | Shared Bank Account | `bank_account_hash` sama pada ≥ 2 claimant berbeda | — | High | Should |
| D2 | Shared Contact Identity | `phone_hash`/`email_hash`/`address_hash` sama antar claimant | — | Medium | Should |
| D3 | Claimant–Provider Concentration | pasangan claimant–provider ≥ 3 klaim dalam 180 hari | 3, 180 hari | Medium | Could |
| D4 | Provider Volume Anomaly | jumlah klaim provider > 3 × median peer `loss_type` (min 10 provider) | 3×, min 10 | Medium | Could |
| D5 | Provider Billing Outlier | rata-rata nominal klaim per provider > 3 × rata-rata peer | 3× | Medium | Could |
| D6 | Collusion Ring | komponen terhubung via atribut identitas bersama dengan ≥ 3 klaim dalam 180 hari | 3, 180 hari | High | Could |
| D7 | Same Asset Multi-Policy | `asset_id` sama pada ≥ 2 polis aktif | — | High | Should |

### 7.5 Famili E — Polis & Cover

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| E1 | Claim on Recently Added Coverage | cover ditambahkan ≤ 30 hari sebelum `incident_date` | 30 hari | Medium | Could |
| E2 | Policy Not Active at Incident | `incident_date` di luar periode aktif polis | — | High | Must |

### 7.6 Famili F — Perilaku & Proses

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| F1 | Submission Velocity | > 2 klaim per claimant per 90 hari lintas produk | 2, 90 hari | Medium | Should |
| F2 | Withdraw–Resubmit Cycle | klaim ditarik lalu resubmit fingerprint sama ≤ 90 hari | 90 hari | Medium | Could |
| F3 | Repeated NEEDS_INFO Stall | ≥ 2 siklus NEEDS_INFO pada klaim yang sama | 2 | Low | Could |
| F4 | Representative Concentration | pengaju/representatif sama untuk ≥ 5 claimant tidak berelasi dalam 180 hari | 5, 180 hari | Medium | Could |

### 7.7 Famili G — Historis

| ID | Rule | Kondisi pemicu | Parameter default | Severity | MVP |
|----|------|----------------|-------------------|----------|-----|
| G1 | Prior Confirmed Flag | claimant/provider dengan flag `CONFIRMED` ≤ 24 bulan | 24 bln | High | Should |
| G2 | Prior Cleared Suppression | fingerprint pernah `CLEARED` dengan reason code valid → flag diturunkan (tidak dihapus) | berlaku 12 bln | — | Should |
| G3 | Cross-Policy Claim History | claimant klaim pada ≥ 3 polis berbeda dalam 12 bulan | 3, 12 bln | Medium | Could |

### 7.8 Famili H — Forensik Dokumen (OCR & manipulasi gambar)

Berlaku untuk dokumen bukti: invoice scan, kuitansi, estimasi biaya. Kolom **Aktivasi**: MVP (default **Off** karena constraint hackathon), Fase 2, atau Fase 3.

| ID | Rule | Kondisi pemicu | Parameter default | Severity | Aktivasi |
|----|------|----------------|-------------------|----------|----------|
| H1 | OCR–Field Mismatch | teks hasil OCR (nomor invoice, nominal, tanggal, nama provider) berbeda dari field terstruktur melebihi toleransi | amount tolerance 0; string similarity < 90% | High | Fase 2 |
| H2 | Metadata Tampering Trace | EXIF/PDF metadata memuat software editor (Photoshop/GIMP/Canva/dll.), ATAU thumbnail EXIF tidak cocok dengan gambar utama, ATAU metadata hilang padahal diperlukan (mis. scan scanner) | daftar software configurable | High | Fase 2 |
| H3 | Metadata Timeline Contradiction | `metadata.created/modified` > `claim_date`, ATAU `invoice_date` (OCR) < `metadata.created` pada dokumen yang diklaim dicetak sebelum klaim | — | High | Fase 2 |
| H4 | Re-save / Double Compression | indikasi kompresi ulang: tabel kuantisasi JPEG tidak konsisten, pola blok 8×8 ganda, dimensi/ukuran tidak wajar untuk scan baru | — | Medium | Fase 2 |
| H5 | Near-Duplicate Document | Kandidat: dHash256 dokumen ≈ dokumen lain (Hamming ≤ 16) meski di-resave/resize; **wajib dikonfirmasi** kemiripan teks OCR ≥ 98% atau `invoice_number`+amount identik | ≤ 16 (256-bit); similarity ≥ 0.98 | High | Fase 2 |
| H6 | OCR Arithmetic Mismatch | jumlah line item hasil OCR ≠ total yang tertera (toleransi 1% nominal) | 1% | Medium | Fase 2 |
| H7 | OCR Quality Failure (routing) | confidence OCR < 60% / dokumen tidak terbaca → **wajib review manual**, tidak boleh di-clear otomatis | 60% | Medium (routing) | Fase 2 |
| H8 | Rasterized / Overlaid Text | region teks berupa gambar atau font tidak konsisten dalam satu lapisan logis dokumen | — | Medium | Fase 3 |
| H9 | ELA Local Anomaly | anomali error-level terkonsentrasi menandakan area ditempel/diedit | ambang per-engine | Low | Fase 3 (eksperimental) |
| H10 | Template Deviation | layout dokumen menyimpang dari template provider yang dikenal (similarity < 80%) | 80% | Low | Fase 3 |

Catatan penerapan:

- H1 membandingkan **dua kali**: OCR↔field terstruktur dan OCR↔angka pada invoice referensi; kedua mismatch dilaporkan terpisah.
- H2 tidak boleh berdiri sendiri sebagai High bila metadata memang seragam hilang (mis. semua dokumen dari aplikasi mobile) — lihat uji false-positive §12.3.
- H9/H10 hanya menambah konteks; tidak pernah menjadi dasar tunggal prioritas.
- **H5 dua tahap:** pada dokumen berbasis teks, hash gambar didominasi layout (halaman putih luas) sehingga dokumen berbeda bisa berjarak hash kecil. Karena itu H5 = kandidat hash (dHash256 ≤ 16 bit) **+ konfirmasi** kemiripan teks OCR. Kalibrasi pada dataset sintetis: near-duplikat berjarak 5–7 bit; dokumen berbeda bisa 0–12 bit → hash tunggal tanpa konfirmasi akan false positive.
- Semua temuan forensik menyertakan **artefak evidence**: potongan teks OCR + lokasi region, diff metadata, gambar side-by-side (original vs highlighted), versi engine.

### 7.9 Cara rule bekerja bersama

- **Bundling:** semua flag pada satu klaim digabung menjadi satu *bundle*; severity bundle = tertinggi.
- **Prioritas antrean (deterministik):** High → Medium → Low; tie-break: jumlah flag, lalu nilai klaim. SLA: High 24 jam, Medium 3 hari, Low 7 hari kerja.
- **Antrean default** hanya High + Medium; Low dapat difilter.
- **Case grouping:** klaim yang terhubung lewat A1/A3/H5/D1/D6 dibungkus menjadi *investigation case* (di BRD disebut **case group**).
- **Suppression (G2):** fingerprint yang sudah di-clear dengan reason code valid dikecualikan dari antrean default **selama masa berlaku**; flag tetap tercatat dan teraudit. Tidak ada penghapusan permanen.
- **Doc pipeline scope:** analisis dokumen hanya berjalan untuk klaim yang memenuhi kriteria aktivasi (default: nominal ≥ 20 jt atau ter-flag ≥ 1 rule A–G) dan maks. 1.000 dokumen/batch — kontrol volume, biaya, dan waktu demo.

### 7.10 Aturan keterjelasan (explainability) — wajib bagi setiap flag

Setiap flag **harus** menyimpan dan menampilkan:

1. `rule_id` + nama rule + versi rule (untuk H: + versi engine OCR/forensik).
2. Nilai pemicu (invoice number, jumlah klaim, nominal, batas, tanggal, metadata key/value).
3. Referensi record/dokumen terkait (daftar `claim_id`/`provider_id`/`claimant_id`/`doc_id`).
4. Penjelasan template bahasa manusia, contoh:
   - *"Invoice INV-7781 (Provider P-22, Rp 18.500.000) juga digunakan pada klaim CLM-0311 (12 Sep 2026)."*
   - *"Claimant C-0142 mengajukan 3 klaim HEALTH dalam 90 hari (batas: 2)."*
   - *"Rekening BCA ****4471 juga dipakai claimant C-0209 dan C-0331."*
   - *"Teks invoice terbaca Rp 8.150.000 sedangkan klaim menyebut Rp 81.500.000 (selisih 10×)."*
   - *"Metadata invoice menunjukkan penyuntingan oleh Adobe Photoshop pada 14 Sep 2026, sehari sebelum klaim diajukan (CLM-0456)."*
   - *"Gambar invoice mirip 96% dengan dokumen INV-6620 pada klaim CLM-0311 setelah di-resave."*

Flag tanpa keempat elemen dianggap cacat data (AC-1).

**Guardrail bahasa UI:** dilarang menampilkan kata "terbukti", "palsu", atau "fraud" sebagai kesimpulan. Gunakan "perlu verifikasi", "indikasi", "periksa dokumen". Ini berlaku untuk semua layar dan ekspor.

### 7.11 Siklus hidup rule

| Tahap | Aturan main |
|-------|-------------|
| Registrasi | Setiap rule: ID, versi, pemilik, parameter, mapping tipologi (§12.1), status aktivasi |
| Aktivasi | Setelah lolos uji false-positive pada klaim/dokumen kontrol (target: 0 false positive untuk Must; < 5% untuk Should/Could; < 8% untuk H9–H10 yang eksperimental) |
| Feature flag | Famili H dikontrol global (`doc_forensics_enabled`) + per-rule; default off di MVP |
| Backtest | Replay batch historis dengan versi rule/engine tertentu; hasil dibandingkan sebelum aktivasi |
| Perubahan | Parameter/threshold via Admin, tercatat nilai lama → baru; perubahan engine OCR memicu re-analisis berversi baru |
| Pensiun | Rule dengan confirm rate < 30% setelah 3 bulan ditinjau; dinonaktifkan dengan catatan |
| Data tidak tersedia | Rule yang butuh field/dokumen tidak tersedia otomatis nonaktif dan ditandai di coverage matrix |
| Kill switch | Admin dapat menonaktifkan rule global/per-rule tanpa deploy; aksi tercatat di audit log |

### 7.12 Model prioritas

| Komponen | Phase 0/1 (BRD) | PRD target | Catatan |
|---|---|---|---|
| Sort utama | Severity High → Medium → Low | Expected loss × capacity | Deterministik vs optimasi |
| Tie-break | Jumlah flag, lalu nilai klaim | Expected loss | Tidak ada skor tersembunyi di Phase 0/1 |
| SLA | High 24 jam; Medium 3 hari; Low 7 hari kerja | Configurable | `TBD` untuk hari kerja |
| Queue default | High + Medium | High + Medium | Low dapat difilter |
| Kapasitas | Tidak dimodelkan | Capacity-aware, simulasi threshold | Phase 2 |

> Jangan mencampur metrik prioritas BRD dengan expected loss PRD dalam satu laporan. Jika skor ditambahkan, reason code dan driver harus tetap terlihat.

### 7.13 Reason code taxonomy (starter library)

| Kategori | Kode | Arti |
|---|---|---|
| False positive | `FP_DATA_ERROR` | Data sumber salah/tidak lengkap |
| False positive | `FP_LEGITIMATE_REPEAT` | Klaim berulang yang sah |
| False positive | `FP_FAMILY_SHARED_ACCOUNT` | Rekening/identitas keluarga bersama |
| False positive | `FP_INSTALLMENT_PLAN` | Klaim cicilan sah dengan invoice berulang |
| False positive | `FP_DOCUMENT_REISSUE` | Dokumen diterbitkan ulang secara sah |
| False positive | `FP_PROVIDER_BULK_BILLING` | Pola billing provider yang sah |
| False positive | `FP_POLICY_REINSTATEMENT_VALID` | Reinstatement sah |
| False positive | `FP_OTHER` | Lainnya, wajib catatan |
| Needs info | `NI_MISSING_DOCUMENT` | Dokumen belum lengkap |
| Needs info | `NI_UNVERIFIABLE_IDENTITY` | Identitas belum dapat diverifikasi |
| Needs info | `NI_PROVIDER_CONFIRMATION` | Perlu konfirmasi provider |
| Needs info | `NI_POLICY_RECORD` | Perlu klarifikasi polis |
| Needs info | `NI_OTHER` | Lainnya, wajib catatan |
| Confirmed | `CF_INVOICE_REUSE` | Invoice dipakai ulang |
| Confirmed | `CF_DUPLICATE_CLAIM` | Klaim duplikat |
| Confirmed | `CF_AMOUNT_INFLATION` | Nominal digelembungkan |
| Confirmed | `CF_DOCUMENT_EDIT` | Dokumen diedit/dimanipulasi |
| Confirmed | `CF_PROVIDER_OVERBILLING` | Provider overbilling |
| Confirmed | `CF_NETWORK_COLLUSION` | Jaringan kolusi |
| Confirmed | `CF_COVER_BREACH` | Pelanggaran cover |
| Confirmed | `CF_REPEAT_OFFENDER` | Repeat offender |
| Confirmed | `CF_OTHER` | Lainnya, wajib catatan |

> Library ini adalah starter. Final library dan mapping ke §12.1 harus disepakati dengan design partner (`TBD`). Tanpa library, precision proxy tidak bermakna.

---

## 8. Kebutuhan Fungsional

### 8.1 Umum & data

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| FR-01 | Unggah file klaim CSV/JSON (maks. 20 MB; satu batch = satu periode) | Must |
| FR-02 | Unggah file invoice referensi | Must |
| FR-03 | Unggah master polis | Must |
| FR-04 | Unggah master provider | Must |
| FR-05 | Unggah peta identitas opsional (hash) | Should |
| FR-06 | Validasi skema & data quality; file invalid ditolak utuh dengan laporan per baris | Must |
| FR-07 | Tolak re-impor file identik (hash konten + periode/sumber) | Must |
| FR-08 | Jalankan rule aktif terhadap batch berjalan + riwayat | Must |
| FR-09 | Rule MVP Must: A1, A2, B1, B5, C4, E2 | Must |
| FR-10 | Rule Should: A3, A4, A6, B2, B3, B6, C1, C2, C5, D1, D2, D7, F1, G1, G2 | Should |
| FR-11 | Rule Could dapat diaktifkan via konfigurasi: A5, B4, C3, D3–D6, E1, F2–F4, G3 | Could |
| FR-12 | Setiap flag menyimpan evidence + penjelasan (§7.10) | Must |
| FR-13 | Flag bundle per klaim; severity = tertinggi; antrean default High+Medium dengan SLA | Must |
| FR-14 | Case grouping + ekspor *evidence package* per case | Should |
| FR-15 | Antrean review: filter per famili/rule/severity/status; pencarian `claim_id`/`claimant_id`/`provider_id`/`doc_id` | Must |
| FR-16 | Keputusan reviewer `CONFIRMED`/`CLEARED`/`NEEDS_INFO`; catatan wajib; `CLEARED` wajib reason code | Must |
| FR-17 | Suppression berbasis fingerprint dengan masa berlaku (Supervisor), teraudit | Should |
| FR-18 | Status klaim read-only bagi sistem | Must |
| FR-19 | Dashboard batch: total klaim, % ter-flag, sebaran per famili/rule, keputusan, precision proxy per rule | Must |
| FR-20 | Halaman coverage matrix: tipologi → rule → status → residual risk | Should |
| FR-21 | Ekspor CSV daftar flag + penjelasan + keputusan | Must |
| FR-22 | Audit trail: impor, flag, keputusan, suppression, perubahan rule/threshold, akses dokumen, login | Must |
| FR-23 | RBAC 3 peran (§11) | Must |
| FR-24 | Rule registry & versi; aktivasi/deaktivasi; approval perubahan parameter | Should |
| FR-25 | Backtesting batch historis dengan versi rule terpilih | Should |
| FR-26 | Batas beban kerja reviewer + indikator backlog | Should |
| FR-27 | Riwayat batch; tutup (read-only) & reopen oleh Supervisor | Should |
| FR-28 | UI responsif (min. 375 px) | Should |
| FR-29 | Ringkasan PDF untuk manajemen | Could |
| FR-30 | Dry-run impor: validasi + preview jumlah flag sebelum commit | Could |

### 8.2 OCR & forensik dokumen (Fase 2, bergerbang)

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| DR-01 | Unggah dokumen per klaim (JPEG/PNG/PDF, ≤ 10 MB, maks. 3 dokumen/klaim) dengan status per dokumen | Fase 2 — Must |
| DR-02 | Validasi file: magic bytes vs ekstensi, batas ukuran/dimensi, penolakan file rusak; simpan **immutable** + SHA-256 | Fase 2 — Must |
| DR-03 | Pipeline async: `UPLOADED → HASHING → METADATA → OCR → CROSSCHECK → DONE/FAILED`, progres per batch, UI tetap responsif | Fase 2 — Must |
| DR-04 | Ekstraksi metadata EXIF/PDF (software, timestamps, thumbnail, producer, revisi PDF) | Fase 2 — Must |
| DR-05 | OCR lokal dengan template/regex per jenis dokumen; simpan teks, region, confidence, versi engine | Fase 2 — Must |
| DR-06 | Cross-check H1: OCR vs field terstruktur vs invoice referensi; tampilkan diff berdampingan | Fase 2 — Must |
| DR-07 | pHash near-duplicate (H5) lintas batch dengan cache hash gambar | Fase 2 — Must |
| DR-08 | Deteksi jejak penyuntingan (H2, H3) dan re-save (H4) | Fase 2 — Must |
| DR-09 | H6 arithmetic check line item vs total (bila line item terbaca) | Fase 2 — Should |
| DR-10 | H7 routing: OCR gagal → flag review manual, **tidak pernah auto-clear** | Fase 2 — Must |
| DR-11 | Penampil dokumen side-by-side (asli vs overlay highlight region mismatch/anomali) | Fase 2 — Should |
| DR-12 | Laporan forensik per dokumen (JSON/PDF): check yang dijalankan, temuan, skor, artefak evidence | Fase 2 — Must |
| DR-13 | Evidence package: sertakan laporan forensik + hash dokumen + log pemrosesan | Fase 2 — Should |
| DR-14 | Konfigurasi cakupan analisis (kriteria klaim, kuota dokumen/batch) oleh Admin | Fase 2 — Should |
| DR-15 | Versioning engine + re-analisis dokumen saat engine berubah (hasil berversi, tidak menimpa) | Fase 2 — Should |
| DR-16 | Chain of custody: originals write-once, verifikasi hash saat dibuka, akses dokumen teraudit | Fase 2 — Must |
| DR-17 | H8–H10 (rasterized text, ELA, template deviation) eksperimental, default off | Fase 3 — Could |
| DR-18 | Retention & pembersihan dokumen sesuai kebijakan (lihat §10.10; default demo 12 bulan) | Fase 2 — Should |

### 8.3 Revenue & value metering (Phase 0/1)

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| FR-31 | Usage metering per batch/periode: klaim di-score, dokumen diproses, flag, case group, reviewer aktif; immutable; ekspor CSV/JSON | Must |
| FR-32 | Value report per batch/tenant: confirmed cases, estimasi value protected/recovered, time saved; formula version; confidence; tanpa double counting; ekspor | Should |
| FR-33 | Konfigurasi plan/entitlement: volume band, feature flag, peringatan overage 80%/100%; audit; tanpa silent blocking | Should |
| FR-34 | Billing export siap-finance: periode, plan, usage, overage, unit price; rekonsiliasi 100% dengan metering; tanpa PII | Must |
| FR-35 | Pilot scorecard: baseline vs pilot; rekomendasi konversi; sign-off Supervisor/Admin | Should |

---

## 9. Kebutuhan Non-Fungsional

| ID | Kategori | Kebutuhan | Cara verifikasi |
|----|----------|-----------|-----------------|
| NFR-01 | Performance | 50.000 klaim + master diproses < 90 dtk (rule A–G, termasuk historis) | Performance test |
| NFR-02 | Performance | Antrean & dashboard termuat < 3 dtk pada 50.000 klaim | Performance test |
| NFR-03 | Determinisme | Input & versi rule sama → output identik; backtest 2× diff = 0 | Automated test |
| NFR-04 | Explainability | 0 flag tanpa penjelasan/evidence lengkap | Automated test + review juri |
| NFR-05 | Integritas keputusan | Tidak ada jalur kode yang menulis status klaim / menolak klaim | Code review + test |
| NFR-06 | Fairness | Rule tidak memakai atribut terlindungi; rule berkorelasi demografi wajib dicatat & dipantau | Rule review + dokumentasi |
| NFR-07 | Privacy | Atribut identitas hanya hash; PII dimasking; minimisasi; log bebas PII | Code review + pentest |
| NFR-08 | Security | Password hash (bcrypt/argon2); sesi ≤ 30 menit idle | Pentest |
| NFR-09 | Security | Upload hanya `.csv`/`.json`/`.jpg`/`.png`/`.pdf`; validasi MIME + magic bytes + ukuran; file di luar webroot | Pentest |
| NFR-10 | Security | Escaping CSV injection pada ekspor | Pentest |
| NFR-11 | Security | OWASP Top 10 (SQLi, XSS, IDOR, broken access control) | Pentest report |
| NFR-12 | Auditability | Audit log append-only; tidak dapat diubah lewat aplikasi | Test + review |
| NFR-13 | Reliability | Impor atomik; pipeline dokumen idempoten (dokumen sama tidak diproses dua kali kecuali re-analisis eksplisit) | Test file rusak + retry |
| NFR-14 | Usability | Median waktu keputusan per flag High < 5 menit | UAT |
| NFR-15 | Portability | Berjalan lokal via Docker Compose (1 perintah) | Demo environment |
| NFR-16 | Doc performance | OCR 1 halaman ≤ 15 dtk pada laptop standar; 200 dokumen diproses background < 20 menit; antrean tetap responsif | Performance test |
| NFR-17 | Storage integrity | Dokumen asli immutable (write-once), hash diverifikasi saat dibuka, akses teraudit | Test + code review |
| NFR-18 | Encryption | Dokumen & laporan forensik terenkripsi at-rest (min. AES-256 / OS-level), akses terbatas per peran | Config review + pentest |
| NFR-19 | Offline | Semua pemrosesan dokumen lokal; tidak ada API eksternal; engine version pinned & tercatat | Code review |
| NFR-20 | Reproducibility | Hasil OCR/forensik menyertakan versi engine + parameter; re-analisis menghasilkan record berversi | Test |
| NFR-21 | Volume control | Kuota dokumen/batch + kriteria aktivasi dipatuhi; melebihi kuota → ditolak dengan pesan jelas | Test |
| NFR-22 | Metering accuracy | 100% rekonsiliasi usage record dengan event sumber; audit | Automated test + reconciliation report |
| NFR-23 | Billing export integrity | Export versioned, checksum, tanpa PII; dapat direproduksi | Test + review |
| NFR-24 | Value report reproducibility | Input & formula version sama → output identik; formula terdokumentasi | Automated test |

---

## 10. Kebutuhan Data

### 10.1 Skema klaim

```csv
claim_id,policy_id,claimant_id,claim_date,incident_date,loss_type,claimed_amount,currency,provider_id,invoice_number,status,rep_id
CLM-0001,POL-7712,C-0142,2026-09-14,2026-09-10,MOTOR,18500000,IDR,P-22,INV-7781,SUBMITTED,REP-07
```

### 10.2 Skema invoice

```csv
invoice_number,invoice_date,provider_id,invoice_amount,currency,claim_id
INV-7781,2026-09-11,P-22,18500000,IDR,CLM-0001
```

### 10.3 Skema polis

```csv
policy_id,product,loss_type_scope,inception_date,expiry_date,status,sum_insured,premium,reinstatement_date,asset_id,agent_id
POL-7712,MOTOR_COMP,MOTOR,2026-01-15,2027-01-14,ACTIVE,150000000,4200000,,AST-3301,AGT-05
```

### 10.4 Skema provider

```csv
provider_id,name,type,region
P-22,Bengkel Jaya Motor,WORKSHOP,Bandung
```

### 10.5 Skema peta identitas (opsional, hashed)

```csv
claimant_id,attribute_type,attribute_hash
C-0142,BANK_ACCOUNT,sha256:ab12...
C-0142,PHONE,sha256:cd34...
```

### 10.6 Metadata dokumen (Fase 2)

```csv
doc_id,claim_id,doc_type,file_name,mime,size_bytes,sha256,uploaded_at
DOC-001,CLM-0001,INVOICE,inv_7781.jpg,image/jpeg,482113,sha256:9f2c...,2026-09-15T08:12:00Z
```

Nilai `sha256` dihitung server saat unggah; `phash`, hasil metadata, OCR, dan laporan forensik disimpan sebagai record turunan berversi (tidak menimpa asli).

### 10.7 Data dictionary entitas inti

| Entitas | Field wajib | Field opsional | Catatan |
|---|---|---|---|
| Alert | `alert_id`, `claim_id`, `rule_id`, `rule_version`, `severity`, `trigger_value`, `evidence_ref`, `status`, `created_at` | `group_id` | Satu rule = satu alert; bundle per klaim |
| Case | `case_id`, `alert_ids`, `owner`, `status`, `priority`, `created_at` | `closed_at`, `summary` | Disebut **case group** di Phase 0/1 |
| Evidence | `evidence_id`, `case_id`, `type`, `source`, `hash`, `chain_of_custody`, `created_at` | `notes` | Hash wajib untuk dokumen |
| Decision | `decision_id`, `case_id`, `outcome`, `reason_code`, `reviewer_id`, `notes`, `decided_at` | `confidence` | Outcome: CONFIRMED/CLEARED/NEEDS_INFO; bukan status klaim |
| Rule | `rule_id`, `version`, `name`, `family`, `params`, `status`, `owner`, `activated_at` | `retired_at`, `mapping_typology` | Versi immutable |
| Label | `label_id`, `case_id`, `disposition`, `confidence`, `reviewer_id`, `label_source`, `created_at` | `notes` | Hanya untuk evaluasi/tuning; bukan input rule |
| AuditEvent | `event_id`, `actor`, `action`, `object_type`, `object_id`, `timestamp`, `hash` | `metadata` | Append-only |
| UsageRecord | `period`, `tenant_id`, `claims_scored`, `documents_processed`, `flags_created`, `case_groups`, `active_users`, `confirmed_cases`, `value_protected_estimate`, `value_recovered` | `notes` | Tanpa PII; dasar billing |
| Tenant | `tenant_id` | — | N/A Phase 0/1; reserved |
| Consent | `consent_id`, `subject_id`, `purpose`, `source`, `expiry` | — | N/A Phase 0/1; produksi PRD §12 |
| Asset | `asset_id`, `type`, `identifier`, `policy_id` | `attributes` | Diperlukan D7/B5 |
| Payment | `payment_id`, `claim_id`, `bank_account_hash`, `amount` | `status` | N/A Phase 0/1; hash rekening ada di identity map |

### 10.8 Aturan data quality

- `claim_id`, `policy_id`: wajib, unik; `policy_id` harus ada di master polis.
- `claimed_amount`/`invoice_amount`: wajib, > 0, maks. 2 desimal.
- `loss_type`: enum tetap (`MOTOR`, `HEALTH`, `PROPERTY`, `TRAVEL`); **Phase 0/1 hanya menerima `MOTOR` dan `PROPERTY`**.
- Tanggal ISO-8601; `incident_date` ≤ `claim_date`; `inception_date` < `expiry_date`.
- `currency` seragam per batch (IDR untuk MVP).
- `invoice_number` wajib bila klaim menyertakan invoice; harus ada di invoice referensi.
- Atribut identitas wajib hash (`sha256:…`); nilai mentah ditolak.
- Dokumen: magic bytes harus cocok dengan ekstensi; dimensi minimum 800×600 untuk OCR yang layak; PDF maks. 5 halaman (halaman pertama dianalisis).
- Baris gagal = impor gagal utuh (fail-fast) dengan laporan per baris.

### 10.9 Dataset evaluasi (sintetis, berlabel)

- Label kebenaran `is_fraud_ground_truth` **hanya untuk evaluasi** — rule engine tidak boleh membacanya (dijamin test).
- Dataset terstruktur (§10.1–10.5): kasus tanaman per famili + klaim kontrol bersih yang mirip pola fraud (keluarga satu rekening, klaim cicilan sah, nominal besar wajar).
- **Dataset dokumen (Fase 2):** generator invoice sintetis (template provider) dengan varian berikut:

| Varian | Contoh | Label |
|---|---|---|
| Authentic | invoice utuh dari generator | kontrol |
| Amount-edited | nominal diedit di area angka | fraud |
| Number-edited | nomor invoice diedit | fraud |
| Re-saved | JPEG di-resave 2–3× | fraud (H4) |
| Near-duplicate | invoice sama di-resave/resize untuk klaim lain | fraud (H5) |
| EXIF-edited | metadata diubah/disisipi software editor | fraud (H2) |
| EXIF-stripped (sah) | screenshot/kirim WhatsApp yang sah | **kontrol** (uji false positive) |
| Timeline-contradiction | metadata dibuat setelah klaim | fraud (H3) |
| Rasterized text | area nominal ditempel sebagai gambar | fraud (H8, Fase 3) |

- Dataset contoh sudah digenerate di `datasets/doc_forgery/` (28 dokumen: 9 kontrol, 19 fraud; `manifest.csv` memuat `label`, `expected_rules`, `sha256`, `dhash256`; generator: `tools/generate_doc_dataset.js`).

### 10.10 Retensi & klasifikasi data

| Data | Phase 0/1 (sintetis) | Produksi (target PRD) | Catatan |
|---|---|---|---|
| Klaim/invoice/polis/provider | 12 bulan setelah batch ditutup | 5–10 tahun `TBD` | Legal |
| Dokumen & turunan forensik | 12 bulan | 5–10 tahun `TBD` | DR-18 |
| Audit log | 12 bulan | ≥ 5 tahun `TBD` | Immutable |
| Usage/billing record | 24 bulan | 7 tahun `TBD` | Finance/audit |
| Label/keputusan | 12 bulan setelah batch ditutup | `TBD` | Feedback model |
| PII | Hanya hash/token | Tokenization/vault | Raw NIK dilarang |
| Backup | 30 hari | `TBD` | Uji restore |

> Retensi produksi **belum final**. Angka PRD 5–10 tahun adalah guardrail, bukan keputusan. Legal harus menetapkan per kategori sebelum pilot produksi.

### 10.11 Data untuk revenue metering

- `UsageRecord` dibuat per periode (batch/bulan) dan tidak dapat diubah setelah ditutup.
- Sumber: event `claim_scored`, `document_processed`, `flag_created`, `case_group_created`, `decision_made`.
- Rekonsiliasi: total usage = jumlah event sumber; selisih > 0 adalah defect.
- Export billing tidak boleh memuat PII; hanya ID agregat dan angka.
- Value report memakai formula version; perubahan formula menghasilkan record baru, bukan menimpa.

---

## 11. Matriks Akses (RBAC)

| Aksi | Reviewer | Supervisor | Admin | Finance/Management (read-only) |
|------|:---:|:---:|:---:|:---:|
| Upload file data & dokumen | Ya | Tidak | Ya | Tidak |
| Lihat antrean, penjelasan, case | Ya | Ya | Ya | Tidak |
| Lihat/buka dokumen asli | Ya | Ya | Ya | Tidak |
| Putuskan CONFIRMED/CLEARED/NEEDS_INFO | Ya | Tidak | Ya | Tidak |
| Buat/cabut suppression | Tidak | Ya | Tidak | Tidak |
| Ekspor laporan & evidence package | Ya | Ya | Ya | Ya |
| Tutup/reopen batch | Tidak | Ya | Ya | Tidak |
| Aktivasi rule, threshold, cakupan dokumen, backtest | Tidak | Tidak | Ya | Tidak |
| Lihat coverage matrix | Ya | Ya | Ya | Ya |
| Manajemen user | Tidak | Tidak | Ya | Tidak |
| Lihat usage & value report | Tidak | Ya | Ya | Ya |
| Ekspor billing/usage | Tidak | Tidak | Ya | Ya |

Akses dokumen (buka, unduh) selalu tercatat di audit log (NFR-17/FR-22).

> Finance/Management adalah peran read-only. Di Phase 0/1, akses dapat diberikan lewat ekspor, bukan login sistem (`TBD`).

---

## 12. Evaluasi & Metrik Deteksi

### 12.1 Coverage matrix (tipologi fraud → rule)

Deliverable wajib demo; disimpan di aplikasi dan lampiran BRD. Denominator mengikuti §4.6: tipologi in-scope saja.

| Tipologi fraud | Rule | Status |
|---|---|---|
| Invoice/dokumen dipakai ulang | A1, A3, A6, H5 | Covered |
| Klaim duplikat / resubmisi | A2, C5, F2 | Covered |
| Nominal digelembungkan (struktur) | A4, B1, B2 | Covered (parsial) |
| **Nominal diedit di dokumen** | **H1, H6** | Covered Fase 2 |
| **Dokumen diedit / dipalsukan** | **H2, H3, H4, H5** | Covered parsial Fase 2 |
| **Dokumen dipindai-ulang untuk klaim lain** | **H5 (+A3)** | Covered Fase 2 |
| Nominal dipecah menghindari review (structuring) | B3, B4 | Covered |
| Klaim melebihi cover / polis tidak aktif | B5, B6, E2 | Covered |
| Polis baru / reinstatement lalu klaim | C1, C2, E1 | Covered |
| Klaim berulang claimant | C4, F1, G3 | Covered |
| Provider overbilling | D4, D5 | Covered (parsial) |
| Jaringan kolusi / identitas bersama | D1, D2, D3, D6, D7, F4 | Covered (parsial) |
| Repeat offender | G1 | Covered |
| Kecelakaan diatur (staged), arson | — | **Not covered** → §16.2 |
| Fraud internal (agen/pegawai) | — | **Not covered** → §16.2 |
| Non-disclosure underwriting | — | **Not covered** (di luar claims) |
| Pemalsuan berkualitas tinggi tanpa jejak metadata/CV | H8–H10 eksperimental | **Sebagian** → §16.2 |
| Modus baru | — | **Not covered** (rule retrospektif) |

**Rekap:** 14 tipologi in-scope; Phase 0/1 covered 10/14 = 71% (memenuhi ≥60%); Fase 2 covered 13/14 = 93% (memenuhi ≥75%). Tipologi out-of-scope (4 baris) tidak masuk denominator.

### 12.2 Metrik & target

| Metrik | Target | Cara ukur |
|--------|--------|-----------|
| Recall kasus tanaman (A–G) | 100% | Dataset §10.9 |
| False flag klaim kontrol | 0 (Must), ≤ 5% (Should/Could) | Dataset §10.9 |
| Recall manipulasi dokumen (Fase 2) | ≥ 100% untuk metadata/near-dupe; ≥ 80% untuk amount/number edit; ≥ 60% re-save | Dataset dokumen §10.9 |
| False positive dokumen sah | ≤ 10% (EXIF-stripped sah ≤ 15% **Low only**) | Dataset dokumen §10.9 |
| Precision proxy per rule | ≥ 60% | Simulasi keputusan reviewer |
| Kelengkapan evidence | 100% | Automated test |
| Cakupan tipologi | ≥ 60% Phase 0/1; ≥ 75% Fase 2 (denominator §4.6) | Coverage matrix |
| Beban review | < 15% klaim ter-flag | Dashboard |
| Metering reconciliation | 100% | Usage report vs event source |
| Value report completeness | 100% confirmed cases punya value fields | Report check |

### 12.3 Uji adversarial (red-team)

- Memecah invoice menjadi tiga nominal kecil (B3/B4 tidak memicu? catat).
- Invoice ganda dengan provider berbeda (A6).
- Keluarga dengan rekening bersama (D1 false positive → clear + reason code).
- Klaim cicilan sah dengan invoice berulang (A1 false positive).
- **Laundering dokumen:** foto ulang invoice dari layar → metadata baru, kompresi baru (uji H2/H4; catat batas).
- **Forgery rapi:** edit nominal + EXIF konsisten + tanpa artefak → harus jujur dicatat sebagai tidak terdeteksi (residual).
- Screenshot sah tanpa EXIF → harus tetap terkendali (Low saja).

---

## 13. Kriteria Keberhasilan (Hackathon Pass Criteria)

Sesuai slide — *"Pass: Explain each flag; a human reviewer confirms or clears it"*.

| # | Kriteria | Terkait |
|---|----------|---------|
| AC-1 | **Explain each flag.** 100% flag memiliki rule + versi, nilai pemicu, referensi record/dokumen, dan penjelasan tampil di UI | FR-12, NFR-04 |
| AC-2 | **Human reviewer confirms or clears.** Tidak ada flag keluar dari `OPEN` tanpa keputusan eksplisit; `CLEARED` wajib reason code; `CONFIRMED` masuk daftar investigasi | FR-16 |
| AC-3 | **No automatic rejection.** Setelah batch diproses, `status` seluruh klaim identik dengan file sumber | FR-18, NFR-05 |
| AC-4 | **Deteksi kasus tanaman.** Recall 100% (A–G); klaim kontrol bersih 0 flag dari rule Must | FR-08–FR-11, §12.2 |
| AC-5 | **Deteksi lintas batch.** Invoice dari batch lama memicu A1; CONFIRMED lama memicu G1; riwayat masuk C4 | FR-08 |
| AC-6 | **Cakupan terdokumentasi.** Coverage matrix + residual risk register terisi | FR-20, §12.1, §16.2 |
| AC-7 | **Determinisme.** Backtest 2× per versi rule → output identik | NFR-03, FR-25 |
| AC-8 | Keputusan & suppression teraudit | FR-22 |
| AC-9 | **Deteksi manipulasi dokumen (Fase 2).** Setiap dokumen fraud yang ditanam memunculkan flag famili H dengan evidence (potongan OCR vs field, diff metadata, pHash match); recall sesuai target §12.2 | DR-04–DR-09 |
| AC-10 | **Chain of custody.** Hash tersimpan = hash file unggahan; evidence package memuat hash + laporan forensik + versi engine | DR-12, DR-13, DR-16, NFR-17 |
| AC-11 | **OCR gagal ≠ auto-clear.** Dokumen dengan confidence < 60% dirutekan ke review manual (H7), tidak pernah di-clear sistem | DR-10 |
| AC-12 | **Kontrol false positive dokumen.** Screenshot sah tanpa EXIF tidak menghasilkan flag High (maksimal Low) | §10.9, §12.3 |
| AC-13 | **Metering accuracy.** Usage record rekonsiliasi 100% dengan event sumber; tidak ada double count | FR-31, NFR-22 |
| AC-14 | **Value report.** Setiap CONFIRMED case punya value fields; formula version tercatat; ekspor tanpa PII | FR-32, NFR-24 |
| AC-15 | **Entitlement.** Volume band dipatuhi; overage diberi peringatan, bukan silent block | FR-33 |
| AC-16 | **Billing export.** Export siap-finance, checksum, rekonsiliasi dengan metering | FR-34, NFR-23 |

Format uji utama (Given/When/Then):

- **AC-1:** Given batch dengan 12 klaim ter-flag — When antrean dibuka — Then 12/12 menampilkan penjelasan yang bisa dilacak ke nilai pemicu + record/dokumen terkait.
- **AC-2/AC-3:** Given batch selesai — When reviewer meng-clear flag — Then status klaim tidak berubah; hanya record keputusan + audit bertambah; `CLEARED` tanpa reason code ditolak.
- **AC-4:** Given 20 kasus tanaman lintas 7 famili + 200 klaim kontrol — When diproses — Then recall 100%, false flag (Must) 0.
- **AC-9:** Given invoice dengan total diubah 10× (Rp 2.442.000 menjadi Rp 24.420.000) sementara line item tidak berubah — When pipeline selesai — Then flag H1 muncul dengan diff "OCR: 24.420.000 vs referensi invoice: 2.442.000 (selisih 10×)" dan H6 menandai ketidaksesuaian line item dengan total.
- **AC-10:** Given dokumen yang diunggah — When evidence package diekspor — Then SHA-256 di laporan = SHA-256 file asli, dan versi engine tercantum.
- **AC-13:** Given satu batch — When metering ditutup — Then jumlah klaim/dokumen/flag sama dengan event sumber; selisih 0.
- **AC-14:** Given case CONFIRMED — When value report dibuat — Then value fields terisi atau ditandai `TBD`, formula version terlihat, dan tidak ada PII.

### 13.1 Mapping AC → PRD AT

| BRD | PRD | Status |
|---|---|---|
| AC-1..AC-8 | PRD AT-01..AT-12 (sebagian) | Partial — perlu mapping rinci |
| AC-9..AC-12 | PRD AT-06, AT-11 | Partial — PRD belum punya AC forensik sedetail BRD |
| AC-13..AC-16 | `AT-13..AT-16` (belum ada di PRD) | **Missing** — tambahkan ke PRD |

---

## 14. Rencana Pengujian (ringkas — detail di QA Report)

| Jenis | Cakupan minimum |
|-------|-----------------|
| Functional test | Semua FR Must; batas per rule: nominal tepat di threshold vs +0.01, klaim ke-1 vs ke-2, invoice tanpa referensi, tanggal invalid, file kosong/header salah, 1 baris rusak di 10.000, re-impor identik |
| Rule accuracy | Dataset §10.9: recall per famili, false flag pada kontrol |
| Doc forensics (Fase 2) | Dataset dokumen: recall per tipe manipulasi, false positive pada screenshot sah, H7 routing, hash immutable, re-analisis berversi |
| Determinisme | Backtest 2× per versi rule/engine; diff = 0 |
| Performance test | 50.000 klaim < 90 dtk; OCR 1 halaman ≤ 15 dtk; 200 dokumen < 20 menit background; dashboard < 3 dtk |
| Metering/billing | Rekonsiliasi 100%; double-count; overage 80%/100%; export checksum; tanpa PII |
| Penetration test | Auth bypass, IDOR antar-batch/case/dokumen, upload file non-image disamarkan, decompression bomb, XSS di catatan, CSV injection, SQLi pencarian, path traversal pada nama file dokumen |
| Automated test | Unit test per rule (positif/negatif/boundary), label-leakage test, integration impor → flag → keputusan → ekspor, imutabilitas status klaim, idempotensi impor & pipeline dokumen |

---

## 15. Governance, Privasi & Fairness

| Aspek | Ketentuan |
|-------|-----------|
| Human-in-the-loop | Tidak ada konsekuensi terhadap claimant tanpa keputusan reviewer |
| Flags are not proof | Temuan forensik dipresentasikan sebagai indikasi; UI dilarang menampilkan kata "terbukti/palsu" — gunakan "perlu verifikasi" |
| Non-diskriminasi | Rule tidak memakai atribut terlindungi; rule berbasis wilayah/provider wajib dicatat sebagai risky-proxy & dipantau |
| Explainability | Penjelasan dapat dipahami reviewer tanpa pengetahuan teknis; tidak ada skor tersembunyi |
| Privasi data | Identitas di-hash; dokumen = PII → akses per peran, terenkripsi, retensi sesuai §10.10, log bebas isi dokumen |
| Chain of custody | Original dokumen write-once + hash diverifikasi; setiap akses tercatat; laporan forensik mencantumkan versi engine |
| Perubahan rule | Via Admin + audit log; perubahan besar wajib backtest |
| Eskalasi sengketa | Reviewer dapat ajukan `NEEDS_INFO`/catatan; Supervisor meninjau suppression |
| Pelatihan reviewer | Reviewer wajib memahami interpretasi metadata/ELA sebelum membaca flag famili H (onboarding singkat di demo) |
| **Compliance applicability** | Phase 0/1 hanya data sintetis; kewajiban UU PDP/OJK untuk data nyata ditunda ke produksi dan mengacu PRD §12 |
| **Koreksi regulasi PRD** | `POJK 23/POJK.05/2015` adalah Produk Asuransi & Pemasaran (dicabut POJK 8/2024); perlindungan konsumen = POJK 22/2023; UU ITE mencakup UU 1/2024. Legal wajib memvalidasi |
| **Revenue governance** | Value report tidak boleh mengklaim fraud dicegah tanpa baseline/kontrol; formula versioned; Finance dan Compliance meninjau |
| **Billing governance** | Usage record immutable; billing export checksum; rekonsiliasi bulanan; tanpa PII |

---

## 16. Risiko, Residual Risk & Roadmap

### 16.1 Risiko proyek

| ID | Risiko | Mitigasi | Owner |
|----|--------|----------|-------|
| R1 | False positive tinggi (keluarga satu rekening, klaim cicilan) | Reason code + suppression; uji kontrol "mirip fraud"; tuning bulanan | Product + Data Science |
| R2 | Rule terlalu banyak → noise | Aktivasi bertahap; precision proxy per rule; pensiun rule < 30% | Product |
| R3 | Rule lintas-batch lambat | Index `invoice_number`, `claimant_id`, hash atribut; cache pHash; uji riwayat ≥ 3 batch | Engineering |
| R4 | Juri menilai tool black-box | Kalimat penjelasan + evidence; tanpa skor risiko | Product + UX |
| R5 | Scope creep (ML/auto-reject) | Out-of-scope §4.2; NFR-05 | Product |
| R6 | Data sintetis tidak realistis | Dataset dirancang per tipologi + kontrol | Data |
| R7 | OCR noisy pada scan berkualitas rendah → H1 false positive | H7 routing; toleransi amount; tampilkan confidence ke reviewer; threshold per jenis dokumen | Data + Engineering |
| R8 | Volume dokumen membebani demo | Kuota 1.000 dokumen/batch, kriteria aktivasi, processing background, engine lokal ringan | Engineering |
| R9 | Temuan forensik disalahgunakan sebagai "bukti" | Bahasa UI "perlu verifikasi"; governance §15; severity konservatif | Compliance + Product |
| R10 | Dokumen PII bocor / snapshot demo | Hanya data sintetis; enkripsi; RBAC; masking pada screenshot demo | Security + Compliance |
| R11 | Atribusi fraud dicegah diperdebatkan | Baseline + control group + formula disepakati | Product + Finance |
| R12 | Budget owner tidak jelas | Multi-stakeholder mapping; pilot sponsor eksekutif | Sales |
| R13 | Sensitivitas harga pasar Indonesia | Tiered packaging; pilot credit; value-based proof | Product |
| R14 | Biaya OCR/dokumen membengkak | Kuota dokumen; confidence routing; local OCR | Engineering |
| R15 | Pilot tidak konversi karena integrasi | Scope pilot tanpa integrasi; success criteria jelas | Product |
| R16 | Churn setelah pilot | Onboarding, training, quarterly value review | Customer Success |

### 16.2 Residual risk register (wajib dipertahankan)

| Fraud yang tidak terjangkau penuh | Alasan | Kontrol saat ini | Rencana |
|---|---|---|---|
| Forgery berkualitas tinggi (nominal diedit + metadata konsisten + tanpa artefak) | Butuh forensik tingkat lanjut / model CV | H1/H2/H4 menangkap sebagian | Fase 3: model forensik + review spesialis |
| Foto ulang layar (metadata & kompresi baru) | Laundering dokumen menghapus jejak asli | Cross-check OCR (H1) tetap bekerja pada isi | Catat sebagai batas; edukasi reviewer |
| Copy-move / splicing halus | H9 ELA eksperimental, false positive tinggi | Tidak dipakai sebagai dasar prioritas | Fase 3: model CV |
| Staged accident, arson, self-inflicted | Butuh investigasi lapangan & data eksternal | Tidak ada | SIU process; data eksternal Fase 3 |
| Fraud internal (agen/pegawai) | Butuh data audit sistem internal | Tidak ada | Integrasi audit internal Fase 3 |
| Non-disclosure underwriting | Di luar siklus klaim | Tidak ada | Underwriting analytics terpisah |
| Modus baru (unknown unknowns) | Rule retrospektif | Review coverage bulanan | Fase 3: ML scoring dari label reviewer |
| Kolusi tanpa atribut bersama | Graph buta tanpa data identitas | D-family parsial | Perkaya graf relasi Fase 2 |
| Cross-insurer fraud | Butuh data sharing industri | Tidak ada | Konsorsium Fase 3 |

### 16.3 Roadmap

1. **Fase 1 (hackathon MVP):** rule A–G (6 Must + minimal 6 Should), antrean, explainability, dashboard, ekspor, audit. Modul dokumen **off** (constraint "No OCR"). Usage metering + billing export minimal.
2. **Fase 2 (1–3 bulan):** aktifkan pipeline dokumen (H1–H7, H9 routing eksperimental off), chain of custody, penampil side-by-side; perluas data identitas & graf relasi; tuning bulanan berbasis reason code; pilot scorecard.
3. **Fase 3 (3–6 bulan):** forensik berbasis model CV (copy-move, ELA terkalibrasi), ML scoring di atas rules dengan label reviewer (rules tetap baseline yang explainable), integrasi SIU case management, data eksternal bila legally feasible, pelaporan fraud committee.

> Roadmap ini memetakan ke PRD §15: Fase 1 → Phase 0/1; Fase 2 → Phase 1b/2; Fase 3 → Phase 2/3.

---

## 17. Milestone (Workflow Sonus: Diagnose → Build → Stress-test)

| Fase | Deliverable |
|------|-------------|
| Diagnose | BRD v0.5, PRD, dataset sintetis berlabel (terstruktur + dokumen), coverage matrix awal, test scenario, business case §3A |
| Build (MVP) | Web app: impor 4 file, rule engine A–G (6 Must + ≥ 6 Should), antrean + penjelasan, keputusan reviewer, suppression, dashboard, ekspor, audit, usage metering |
| Build (track dokumen) | Pipeline dokumen + famili H (H1–H7), laporan forensik, chain of custody — dikembangkan paralel, **default off** untuk demo |
| Stress-test | Performance & pentest, red-team §12.3, QA report, tuning threshold, coverage matrix final, rekonsiliasi metering |
| Revenue validation | 10–15 WTP interviews, 2 paid pilot, pilot scorecard, recalculation unit economics (§3A.9) |
| Show | Demo: flag → penjelasan → keputusan manusia → coverage matrix + residual risk + value report; modul dokumen didemokan hanya jika tim memutuskan mengaktifkan flag |

---

## 18. Sync Register & Traceability

### 18.1 Hubungan dokumen

- **Default v0.5 (rev. 2):** BRD ini adalah Phase 0/1 thin slice dari PRD v0.2. PRD memiliki target-state; BRD memiliki Phase 0/1.
- Jika stakeholder memutuskan ini produk terpisah, status register berubah menjadi *reference only* dan keputusan harus dicatat di decision log. Status: `TBD`.
- BRD tidak boleh mengklaim kapabilitas PRD yang belum diimplementasikan.
- **TRD v0.1** memetakan NFR-01..24, AC-1..16, dan arsitektur HA; TRD tidak mengubah scope/keputusan BRD. Traceability ada di TRD §23.

### 18.2 Capability traceability (BRD → PRD)

| BRD | PRD v0.2 | Status | Aksi |
|---|---|---|---|
| §0 document control | §0A | Match | Phase map & source of truth |
| §3A business case | §13.4, §13.5, §14 | Partial | PRD pricing/ROI; BRD pilot economics |
| §7 rule catalog | E3, E6, E17 | Partial | BRD §7 detail; PRD E17 mengikat thin slice |
| §8 FR-01..30 | E1, E3, E7, E8, E10, E12, E17 | Partial | Map FR ke epic; N/A register |
| §8.2 DR-01..18 | E6, Fase 2 | Partial | Detail forensik tetap di BRD |
| §8.3 FR-31..35 | FR-1204, FR-1207–1210 | Match | Usage, value report, entitlement, billing |
| §9 NFR-01..24 | NFR-1..20 | Partial | Determinism, reproducibility, metering |
| §10 data contract | §9.7, §18.2 | Partial | Field-level mapping `TBD` |
| §11 RBAC | FR-1202/1301, §4 | Partial | 3 peran thin slice vs 10 persona target |
| §12 metrics/coverage | §3.3, §13.5 | Partial | Denominator & planted vs production |
| §13 AC | §18.3 AT + mapping | Partial | AC-1..8 mapping rinci `TBD` |
| §15 governance | §12 | Partial | Retensi/tokenization `TBD` Legal |
| §16.2 residual risk | §16.2 | Match | Register diadopsi |
| §16.3 roadmap | §15.0 | Partial | Phase map |
| §17 milestones | §15 | Partial | Milestone map |
| §19 open questions | §17.3, §19.6 | Partial | Decision log |
| §20 glossary | §18.1 | Partial | Glossary digabung |
| §18 sync register | §19 | Match | Register paralel |

### 18.3 Kapabilitas PRD yang N/A di Phase 0/1

| Kapabilitas | Status BRD | Referensi PRD |
|---|---|---|
| Multi-tenant, SSO/MFA | N/A | E12, E13 |
| API/SFTP/real-time | N/A — CSV/JSON upload | E1 |
| Entity resolution/golden record | N/A — hash matching saja | E2 |
| Graph analytics/explorer | N/A — case grouping saja | E5 |
| Case management workspace | N/A — evidence package saja | E8 |
| ML scoring/model governance | N/A sampai Fase 3 | E4, §9 |
| GenAI assistant | N/A | E15 |
| External data/consortium | N/A | E14 |
| Notifications/collaboration | N/A | E16 |
| Underwriting fraud | N/A | J5 |
| Data residency/DSR/DPIA | N/A data sintetis; produksi PRD §12 | §12 |
| Availability/DR/observability | N/A | §8 |
| Localization/WCAG | IDR only; WCAG N/A | §11 |
| Commercial/GTM | N/A — pilot economics saja | §14 |
| Sertifikasi ISO/SOC 2, vendor risk | N/A | §12 |

### 18.4 BRD items yang diadopsi ke PRD v0.2

Semua item berikut sudah diadopsi ke PRD v0.2; detail implementasi tetap di BRD.

1. Rule catalog + typology mapping (§7, §12.1).
2. Coverage matrix + denominator definition (§4.6, §12.1).
3. Residual risk register (§16.2).
4. Control-claim false-positive gates + rule retirement (§7.11).
5. Label leakage test + deterministic backtest/reproducibility (NFR-03, NFR-20).
6. UI language guardrail "flag bukan bukti" (§7.10, §15).
7. Concrete data contract + data quality rules (§10).
8. Doc forensics thresholds + chain of custody detail (§7.8, §8.2).
9. Adversarial/red-team scenarios (§12.3).
10. Usage metering/value report/billing export (FR-31..35).

### 18.5 Decision log

| ID | Keputusan | Alasan | Owner | Status | Review |
|---|---|---|---|---|---|
| D-01 | ML Phase 2/3, bukan Phase 0/1 | Explainability dan bukti | Product + Data Science | Default v0.5 (rev. 2) | `TBD` |
| D-02 | Triage-only; tanpa write-back | Batas "flag bukan bukti" | Product + Legal | Default v0.5 (rev. 2) | `TBD` |
| D-03 | Retensi per kategori; produksi `TBD` | Legal basis | Compliance | Default v0.5 (rev. 2) | `TBD` |
| D-04 | Tokenization/vault di produksi | Privasi + matching | CTO + Compliance | Default v0.5 (rev. 2) | `TBD` |
| D-05 | P&C auto/properti saja | Selaras PRD MVP | Product | Default v0.5 (rev. 2) | `TBD` |
| D-06 | Platform + per-claim + implementation + pilot credit | Selaras PRD §14.2 | Product + Finance | Default v0.5 (rev. 2) | `TBD` |
| D-07 | Prioritas deterministik Phase 0/1 | Explainability | Product | Default v0.5 (rev. 2) | `TBD` |
| D-08 | Pisahkan "case group"/"review outcome" dari "investigation case"/"claim decision" | Menghindari metrik menyesatkan | Product | Default v0.5 (rev. 2) | `TBD` |

### 18.6 Sync checklist

1. Setiap item punya source of truth (§0.3).
2. Setiap item punya status: Match / Partial / N/A / Conflict.
3. N/A punya alasan + referensi PRD (§18.3).
4. Conflict punya keputusan, owner, tanggal (§18.5).
5. Metrik punya satu definisi + formula (metric dictionary `TBD`).
6. Entitas punya satu skema + owner (§10.7).
7. Nama fase sama di kedua dokumen (§0.4).
8. Setiap versi saling mereferensikan (changelog + doc control).

### 18.7 Change control

- Perubahan BRD yang memengaruhi PRD harus dicatat di changelog kedua dokumen.
- Satu change set memperbarui BRD, PRD, dan decision log.
- Review sync dilakukan setiap akhir fase dan sebelum pilot.

---

## 19. Open Questions

1. Ketersediaan atribut identitas (hash) dan dokumen contoh — disediakan panitia atau dibangun sendiri?
2. Threshold default per `loss_type` mengikuti contoh §7.2 atau policy aktual?
3. Jendela 90/30/180 hari — sesuai kebijakan fraud unit?
4. Klaim `REJECTED`/`WITHDRAWN` historis dihitung dalam A1/C4/C5/G1?
5. Otorisasi suppression: cukup Supervisor atau empat mata (Admin + Supervisor)?
6. Apakah modul OCR/forensik **diaktifkan saat demo** meskipun slide menyebut "No OCR"? (rekomendasi: tetap off; tampilkan sebagai roadmap + arsitektur siap)
7. Engine OCR pilihan: Tesseract vs PaddleOCR lokal — mana yang paling andal untuk invoice sintetis?
8. Volume & format dokumen di dataset panitia; apakah PDF multi-halaman perlu didukung penuh?
9. Kebijakan retensi dokumen & kebutuhan pernyataan chain of custody formal dari legal?
10. IDR saja cukup untuk MVP?
11. Apakah BRD ini thin slice PRD atau produk terpisah?
12. Siapa target tenant pertama dan berapa volume klaim/tahun untuk revenue model?
13. Berapa range willingness-to-pay dan harga pilot?
14. Berapa cost to serve (cloud/OCR/support) per klaim dan per dokumen?
15. Berapa durasi pilot, conversion target, dan churn assumption?
16. Apakah success fee dipakai? Bagaimana verifikasinya?
17. Skema tokenization NIK seperti apa yang disetujui Legal?
18. Kapan health masuk roadmap, dan apakah TRAVEL dihapus dari enum Phase 0/1?

---

## 20. Glosarium

| Istilah | Arti |
|---------|------|
| Tipologi fraud | Kategori modus kecurangan |
| Flag | Penanda klaim cocok dengan rule; bukan tuduhan, bukan bukti |
| Flag bundle | Semua flag pada satu klaim; severity = tertinggi |
| Case group | Kumpulan klaim terkait di BRD Phase 0/1; bukan investigation case PRD |
| Investigation case | Kumpulan klaim terkait dilihat sebagai satu modus (PRD) |
| Evidence | Nilai pemicu + record/dokumen terkait yang membuat flag dapat diperiksa |
| Confirm / Clear | Keputusan reviewer: lanjut investigasi / false positive (+ reason code) |
| Review outcome | CONFIRMED/CLEARED/NEEDS_INFO di BRD; bukan keputusan klaim |
| Suppression | Pengecualian sementara fingerprint yang sudah di-clear dari antrean default; teraudit |
| Coverage matrix | Pemetaan tipologi fraud → rule → status cakupan |
| Residual risk | Fraud yang tidak terjangkau sistem beserta alasannya |
| Precision proxy | `CONFIRMED / (CONFIRMED + CLEARED)` per rule |
| Structuring | Memecah nominal di bawah batas review |
| OCR | Optical Character Recognition — ekstraksi teks dari gambar/PDF |
| EXIF | Metadata tertanam pada foto (perangkat, waktu, software, thumbnail) |
| pHash | Perceptual hash — sidik gambar yang tahan terhadap re-save/resize |
| ELA | Error Level Analysis — teknik visual indikatif area gambar yang diedit |
| Chain of custody | Jejak integritas bukti: hash, akses, versi pemrosesan tercatat |
| Fixed rules | Aturan deterministik tanpa ML/prediksi |
| Idempoten | Impor/proses ulang tidak mengubah state data |
| ARR | Annual Recurring Revenue |
| ARPA | Average Revenue Per Account |
| COGS | Cost of Goods Sold |
| Gross margin | `(Revenue − COGS) / Revenue` |
| CAC | Customer Acquisition Cost |
| CAC payback | Waktu untuk menutup CAC dari gross profit |
| LTV | Lifetime Value |
| NRR | Net Revenue Retention |
| WTP | Willingness to Pay |
| Value protected | Nilai klaim berisiko yang tidak dibayar/dikurangi |
| Value recovered | Recovery dari investigasi |
| Pilot fee | Biaya pilot tetap; dikreditkan bila lanjut |
| Success fee | Persentase nilai fraud terverifikasi; opsional dan capped |
| Source of truth | Dokumen pemilik definitif untuk suatu domain |
| Traceability matrix | Peta item BRD ↔ PRD ↔ status ↔ fase |
| N/A register | Daftar kapabilitas yang sengaja tidak berlaku di Phase 0/1 |
| Decision log | Catatan keputusan konflik beserta owner dan tanggal |
| PII / PHI | Personally Identifiable Information / Protected Health Information |
| DSR | Data Subject Request |
| DPIA | Data Protection Impact Assessment |
| PAS | Policy Administration System |
| Multi-tenant | Satu platform melayani banyak tenant terisolasi |
| Entity resolution | Proses mencocokkan dan menggabungkan entitas yang sama |
| Golden record | Representasi tunggal entitas setelah entity resolution |
| Expected loss | Estimasi kerugian yang diharapkan; dasar prioritas PRD |
| Drift | Perubahan distribusi data/performa model |

---

**Akhir dokumen.**
Dokumen ini adalah draft v0.5 untuk review. Angka revenue, retensi, tokenization, dan keputusan konflik bertanda `TBD`/default dan belum final. Perubahan harus dicatat di changelog dan disinkronkan dengan PRD.

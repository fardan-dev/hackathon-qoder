# PRD — Platform Deteksi Fraud Asuransi Multi-Lini

**Nama kerja:** Sentinel Fraud Platform (placeholder)  
**Versi:** 0.2 — Draft untuk review, disinkronkan dengan BRD v0.5  
**Tanggal:** 16 November 2026  
**Pemilik:** Product Management  
**Status:** Draft  
**Pasar awal:** Indonesia (OJK)  
**Lini produk:** P&C pada MVP; jiwa dan kesehatan pada Phase 2  
**Model bisnis:** B2B SaaS multi-tenant, dengan opsi private cloud/on-prem untuk enterprise  
**Dokumen terkait:** BRD Insurance Claim Fraud Detection v0.5, TRD, QA Report  
**Hubungan dokumen:** PRD ini adalah target-state 24 bulan. BRD v0.5 adalah Phase 0/1 thin slice (rules-only, data sintetis, P&C auto/properti) yang menjadi baseline implementasi awal.

**Reviewer wajib:** Head of Claims/SIU, Chief Compliance Officer, Legal, CTO, Head of Data Science, UX Lead, Head of Customer Success.

> Dokumen ini adalah product requirements document, bukan nasihat hukum. Seluruh interpretasi regulasi wajib divalidasi oleh Legal/Compliance sebelum implementasi.

### Changelog

| Versi | Perubahan |
|---|---|
| 0.1 | Draft awal: visi multi-line, 16 epic, NFR, model komersial, pilot |
| 0.2 | Sinkronisasi dengan BRD v0.5: dokumen kontrol & phase map, thin slice rules-only, koreksi regulasi (POJK 22/2023, POJK 8/2024, UU 1/2024), read-only default, data dictionary, reason code, coverage denominator, usage/value/billing requirements, residual risk, sync register, acceptance mapping |
| 0.2 (rev. 2) | Tambah referensi TRD v0.1 di §18.4 dan §19.2; tidak ada perubahan scope atau requirement |
| 0.2 (rev. 3) | Keputusan pra-MVP: mobile reviewer/supervisor read-only + push (M-01..M-10), aksi tulis di web; GA gate M9; tim 13 engineer (11 core + 2 mobile); WCAG baseline thin slice vs AA MVP/GA. Sinkron dengan BRD rev. 4 dan TRD rev. 4 |

---

## Daftar Isi

0. Ringkasan Eksekutif
0A. Dokumen Kontrol, Phase Map & Source of Truth
1. Latar Belakang dan Pernyataan Masalah
2. Visi, Misi, dan Prinsip Produk
3. Tujuan dan Metrik Keberhasilan
4. Persona dan Peran Pengguna
5. Alur Kerja Utama
6. Ruang Lingkup dan Strategi Rilis
7. Kebutuhan Fungsional
8. Kebutuhan Non-Fungsional
9. Data, AI/ML, dan Tata Kelola Model
10. Integrasi dan Arsitektur
11. UX/UI dan Lokalisasi
12. Kepatuhan, Keamanan, dan Privasi
13. Analitik, Pelaporan, dan Eksperimen
14. Model Komersial dan GTM
15. Rencana Rilis dan Milestone
16. Risiko dan Mitigasi
17. Asumsi, Ketergantungan, dan Open Questions
18. Lampiran
19. Sync Register & Traceability (BRD v0.5)

---

## 0. Ringkasan Eksekutif

Perusahaan asuransi multi-lini di Indonesia menghadapi masalah yang sama: data klaim, polis, pihak, dokumen, dan jaringan relasi tersebar di banyak sistem; deteksi fraud masih bergantung pada rules dan intuisi investigator; dan alert yang dihasilkan melebihi kapasitas tim SIU. Akibatnya, fraud lolos, klaim sah tertunda, dan keputusan sulit dipertahankan saat audit.

Sentinel adalah platform multi-tenant yang menyatukan **data, aturan, model ML, analitik jaringan, forensik dokumen, dan alur investigasi** dalam satu tempat. Outputnya bukan sekadar skor: setiap alert menjelaskan alasannya, setiap kasus memiliki bukti yang tertaut, dan setiap keputusan memiliki jejak audit.

**Rekomendasi strategi rilis:**

- **MVP (bulan 0–6):** shared core + P&C auto dan properti. Fokus pada kualitas deteksi, triage, case management, dan auditability.
- **Phase 2 (bulan 9–15):** modul jiwa dan kesehatan, analitik provider, graph lanjutan, forensik dokumen, konektor data eksternal.
- **Phase 3 (bulan 15–24):** asisten investigasi GenAI, orkestrasi bukti, data konsorsium opt-in, ekspansi ASEAN.

**Multi-line adalah arsitektur produk, bukan janji bahwa semua lini tersedia di MVP.** Data model, rule pack, model, dan template investigasi dirancang multi-line sejak awal, tetapi rollout dilakukan per lini agar kualitas tidak dangkal.

**Target hasil MVP:**

- Deteksi fraud terkonfirmasi meningkat **≥25%** pada volume alert yang sama dibanding baseline rules-only.
- Waktu triage turun **≥30%**.
- **≥95%** alert risiko tinggi tercakup dalam investigasi dengan bukti.
- **100%** keputusan memiliki reason code dan jejak audit.
- **Nol** keputusan penolakan klaim otomatis tanpa manusia.

---

## 0A. Dokumen Kontrol, Phase Map & Source of Truth

### 0A.1 Status dan hubungan dokumen

| Item | Nilai |
|---|---|
| Dokumen | PRD — Platform Deteksi Fraud Asuransi Multi-Lini |
| Versi | 0.2 |
| Tanggal | 16 November 2026 |
| Status | Draft untuk review |
| Target-state | 24 bulan (M0–M24) |
| Thin slice | BRD v0.5 — Phase 0/1, rules-only, data sintetis, P&C auto/properti |
| Dokumen acuan | BRD v0.5, TRD, QA Report |
| Perubahan v0.2 | Sinkronisasi BRD v0.5; koreksi regulasi; thin slice; revenue metering; traceability |

**Aturan dasar:** PRD ini tidak boleh mengklaim kapabilitas yang belum ada di thin slice. Sebaliknya, thin slice tidak boleh menambah scope yang tidak ada di PRD tanpa keputusan tertulis.

### 0A.2 Phase map

| BRD v0.5 | PRD | Isi |
|---|---|---|
| Fase 1 — Hackathon MVP | Phase 0/1 **thin slice** | Rules A–G, CSV/JSON, queue, evidence, coverage, audit; tanpa ML/OCR/integrasi/write-back |
| Fase 2 — 1–3 bulan | Phase 1b/2 | OCR & forensik dokumen, identity/graph minimal, pilot |
| Fase 3 — 3–6 bulan | Phase 2/3 | ML scoring, CV forensics, SIU case management, external data |
| — | Commercial MVP (M2–M9) | P&C auto/properti, rules + baseline ML, OCR basic, 1 integrasi, case management, mobile read-only + push |

> **Thin slice ≠ commercial MVP.** Thin slice adalah bukti implementasi Phase 0/1; commercial MVP tetap memiliki ML baseline, OCR, dan integrasi sesuai PRD.

### 0A.3 Source of truth

| Domain | Pemilik | Acuan |
|---|---|---|
| Visi, pasar, model bisnis, GTM | Product | PRD §0–3, §14 |
| Roadmap & fase | Product | PRD §15; BRD memetakan |
| Tipologi, rule, threshold Phase 0/1 | Fraud/Data Science | BRD §7, §12 |
| Data contract Phase 0/1 | Data Engineering | BRD §10 |
| Target data model | Data/Architecture | PRD §9.2, §18.2 |
| NFR Phase 0/1 | Engineering Lead | BRD §9 |
| NFR target | CTO/SRE | PRD §8 |
| Kepatuhan produksi | Legal/Compliance | PRD §12 |
| Acceptance Phase 0/1 | QA | BRD §13 |
| Acceptance pilot/produksi | QA | PRD §18.3 |

### 0A.4 Default decisions

| ID | Keputusan | Status |
|---|---|---|
| D-01 | ML di Phase 2/3; thin slice rules-only | Default v0.2 |
| D-02 | Thin slice triage-only/read-only; write-back Phase 2, opt-in | Default v0.2 |
| D-03 | Retensi per kategori; produksi TBD Legal | Default v0.2 |
| D-04 | Tokenization/vault di produksi; raw NIK dilarang di aplikasi/log | Default v0.2 |
| D-05 | P&C auto/properti MVP; health/life Phase 2 | Default v0.2 |
| D-06 | Platform + per-claim + implementation + pilot credit; tanpa per-alert | Default v0.2 |
| D-07 | Prioritas deterministik di thin slice; expected loss Phase 2 | Default v0.2 |
| D-08 | "Case group"/"review outcome" (BRD) ≠ "investigation case"/"claim decision" (PRD) | Default v0.2 |

### 0A.5 Definisi "synced"

Dokumen dianggap sinkron jika:

1. Setiap domain punya source of truth.
2. Setiap item punya status Match/Partial/N/A/Conflict.
3. N/A punya alasan dan referensi.
4. Conflict punya keputusan, owner, dan tanggal.
5. Metrik punya satu definisi dan formula.
6. Entitas punya satu skema dan owner.
7. Nama fase konsisten.
8. Setiap versi saling mereferensikan.

---

## 1. Latar Belakang dan Pernyataan Masalah

### 1.1 Konteks pasar

Pasar global deteksi fraud asuransi berada di kisaran **USD 8 miliar pada 2026**, dengan estimasi pertumbuhan tahunan yang sangat bervariasi (11–25%) tergantung cakupan riset. Yang lebih penting: adopsi masih dangkal. Survei global FRISS 2026 menemukan hanya **9% insurer** yang telah mengimplementasikan AI/ML untuk fraud secara penuh; 36% masih pilot, 36% merencanakan, dan 19% belum memiliki rencana. Sebagian besar alat yang terpasang masih berupa rules dan database lookup. Lihat [Market & Vendor Landscape 2026](./insurance-fraud-detection-market-landscape-2026.md) untuk analisis pasar dan vendor.

Di Indonesia, konteksnya memiliki ciri sendiri:

- Penetrasi asuransi masih tumbuh, sementara kanal digital dan klaim cashless meningkat.
- Data fraud teragregasi lintas perusahaan terbatas; banyak insurer masih mengandalkan data internal dan investigasi manual.
- Core system beragam: sebagian besar menggunakan vendor lokal, sebagian menggunakan platform global, dan banyak proses masih berbasis spreadsheet, email, dan dokumen PDF.
- Regulasi perlindungan data (UU PDP) dan pengawasan OJK meningkatkan tuntutan atas transparansi, keadilan, dan auditability keputusan.
- Lini P&C, jiwa, dan kesehatan memiliki tipologi fraud, data, dan pola investigasi yang berbeda; solusi generik sering gagal karena tidak memahami konteks lokal.

### 1.2 Masalah operasional

| Masalah | Gejala di lapangan | Dampak | Peluang produk |
|---|---|---|---|
| Data terfragmentasi | Klaim, polis, pihak, dokumen, dan provider berada di sistem berbeda | Investigator menghabiskan waktu mengumpulkan data | Entity resolution dan satu case view |
| Rules-based false positive | Banyak alert karena aturan kaku; konfirmasi fraud rendah | Alert fatigue; klaim sah tertunda | Hybrid rules + ML + capacity-aware triage |
| Investigation bottleneck | Investigasi manual 14+ hari; kasus menumpuk | Fraud tidak tertangani; recovery rendah | Case management, otomasi bukti, prioritisasi |
| Document fraud | Invoice, kuitansi, surat keterangan, foto, dan dokumen medis dipalsukan | Klaim palsu lolos | OCR, tamper detection, cross-document consistency |
| Network/ring fraud | Klaim terpisah tetapi berbagi telepon, alamat, rekening, bengkel, atau provider | Fraud terorganisir tidak terlihat | Graph analytics dan ring detection |
| Tidak ada feedback loop | Keputusan investigator tidak kembali ke model | Model tidak belajar; false positive berulang | Label store, active learning, retraining |
| Beban audit | Alasan keputusan tidak terdokumentasi | Risiko regulator, sengketa konsumen, reputasi | Explainability, evidence trail, audit log |
| Friksi ke nasabah | Klaim sah diperiksa berlebihan | NPS turun; keluhan meningkat | Precision, SLA, customer-friendly workflow |
| Inkonsistensi multi-line | P&C, jiwa, dan kesehatan memakai alat berbeda | Biaya lisensi dan operasional tinggi | Shared core, line-specific configuration |

### 1.3 Dampak jika tidak ada tindakan

- **Leakage** klaim meningkat dan rasio klaim memburuk.
- **Waktu penyelesaian klaim** meningkat, terutama saat catastrophic event atau lonjakan klaim.
- **Biaya investigasi per kasus** naik karena manual effort dan data hunting.
- **Risiko regulasi** meningkat karena keputusan tidak dapat dijelaskan atau didokumentasikan.
- **Pengalaman nasabah** memburuk karena false positive dan permintaan dokumen berulang.
- **Fraud terorganisir** tidak terdeteksi karena tidak ada analitik jaringan lintas klaim.

**Kerangka biaya inaction** (placeholder `TBD`; formula dan validasi detail di BRD §3A):

| Komponen | Formula | Placeholder |
|---|---|---|
| Jam investigator terbuang | `alerts × time_wasted_per_alert × loaded_hourly_cost` | `TBD` |
| Fraud lolos | `claims × leakage_rate × avg_claim × non_detection_rate` | `TBD` |
| False positive | `alerts × fp_rate × cost_per_fp` | `TBD` |
| Temuan audit | `findings × cost_per_finding` | `TBD` |
| Dokumentasi tidak terverifikasi | `documents × exposure_per_document` | `TBD` |

> Angka biaya inaction tidak boleh diklaim sebagai revenue. Validasi dengan baseline design partner dan control group.

### 1.4 Mengapa sekarang

1. **Digitalisasi klaim** menciptakan data yang lebih kaya, tetapi juga memperbesar volume dan kecepatan fraud.
2. **UU PDP** berlaku penuh sejak Oktober 2024; platform harus privacy-by-design, bukan retrofit.
3. **OJK** semakin menekankan tata kelola, manajemen risiko, perlindungan konsumen, dan akuntabilitas keputusan.
4. **Biaya AI turun** dan model open-source/Managed AI memungkinkan deteksi yang lebih baik tanpa tim riset besar.
5. **Agentic AI** membuka peluang otomasi pengumpulan bukti dan penyusunan laporan investigasi, bukan hanya scoring.
6. **Kompetisi insurer** mendorong efisiensi klaim dan diferenciasi layanan.

### 1.5 Problem statement

> Perusahaan asuransi multi-lini di Indonesia tidak memiliki satu platform yang menghubungkan data klaim, polis, pihak, dokumen, dan jaringan relasi untuk mendeteksi fraud secara akurat, menjelaskan alasannya, dan mengelola investigasi secara audit-ready.

---

## 2. Visi, Misi, dan Prinsip Produk

### 2.1 Visi

> Menjadi lapisan kecerdasan fraud standar industri asuransi Indonesia: setiap klaim berisiko tinggi terdeteksi, dijelaskan, dan diinvestigasi dengan bukti yang dapat diaudit.

### 2.2 Misi

Menyatukan data, aturan, model, dan alur investigasi agar tim SIU dapat bekerja lebih cepat, lebih tepat, dan lebih adil—tanpa mengganti core system yang sudah ada.

### 2.3 Prinsip produk

1. **Human-in-the-loop.** AI mendeteksi dan merekomendasikan; manusia memutuskan. Tidak ada penolakan klaim otomatis.
2. **Evidence-first.** Setiap alert harus memiliki alasan, data pendukung, dan tautan ke bukti.
3. **Explainable by default.** Investigator, compliance, dan regulator harus dapat memahami mengapa klaim ditandai.
4. **Capacity-aware.** Prioritisasi mempertimbangkan kapasitas SIU, nilai klaim, SLA, dan expected loss—bukan hanya skor.
5. **Privacy-by-design.** Minimisasi data, consent/legal basis, masking, retensi, dan data residency Indonesia.
6. **Multi-line, configurable.** Shared core, tetapi rule pack, model, dokumen, dan template investigasi berbeda per lini.
7. **Interoperable.** API-first, tidak mengganti PAS/claims system, dan mendukung batch/CSV untuk lingkungan yang belum mature.
8. **Measurable.** Setiap fitur harus terhubung ke metrik deteksi, efisiensi, atau kepatuhan.

---

## 3. Tujuan dan Metrik Keberhasilan

### 3.1 Tujuan produk

| ID | Tujuan | Target | Waktu |
|---|---|---|---|
| G1 | Meningkatkan deteksi fraud terkonfirmasi pada volume alert yang sama | +25% vs baseline rules-only | 6 bulan setelah produksi |
| G2 | Mengurangi waktu triage | -30% | 3 bulan setelah produksi |
| G3 | Meningkatkan cakupan investigasi alert risiko tinggi | ≥95% | 6 bulan setelah produksi |
| G4 | Mengurangi false positive pada klaim bersih | -20% tanpa menurunkan recall | 6 bulan setelah produksi |
| G5 | Memastikan setiap keputusan audit-ready | 100% reason code + evidence + audit log | MVP |
| G6 | Menyediakan platform multi-line | P&C live di MVP; jiwa & kesehatan live ≤12 bulan setelah GA | Roadmap |

> **Definisi pengukuran (v0.2):**
> - **G1** diukur pada **fixed review capacity** (jumlah alert yang benar-benar dapat direview tim), bukan pada volume alert mentah. Jika kapasitas berubah, G1 tidak comparable.
> - **G4** diukur sebagai false positive rate pada klaim kontrol/bersih dengan **recall guardrail**; G1 dan G4 harus dilaporkan berpasangan, bukan sebagai bukti independen.
> - **Planted-test recall** (BRD §12.2) mengukur kemampuan rule pada dataset berlabel; **bukan** business lift dan tidak boleh dibandingkan langsung dengan G1.
> - **Typology coverage** adalah metrik terpisah dari North Star. Denominator mengikuti BRD §4.6: 14 tipologi in-scope; Phase 0/1 covered 10/14 (71%), Fase 2 13/14 (93%). Jangan menulis "coverage" tanpa denominator.

### 3.2 North Star Metric

**Investigation coverage of high-risk alerts** — persentase alert risiko tinggi yang diinvestigasi tuntas dengan bukti dalam periode tertentu.

Metrik ini dipilih karena:
- Menyeimbangkan deteksi dengan outcome; alert yang tidak pernah dikerjakan tidak bernilai.
- Langsung terhubung ke nilai bisnis: fraud yang tercegah/dipulihkan.
- Mendorong desain kapasitas dan otomasi, bukan hanya menambah aturan.

**Metrik pendukung:** **Fraud value protected per tahun per tenant** (fraud dicegah + dipulihkan + penghematan operasional).

### 3.3 KPI dan guardrails

| Kategori | Metrik | Baseline | Target |
|---|---|---|---|
| Deteksi | Recall@k pada fixed review capacity (k = kapasitas reviewer) | Baseline rules | +25% |
| Deteksi | Referral precision | Baseline tenant | ≥40% atau +10 poin |
| Deteksi | Lift over rules-only | — | ≥20% |
| Operasional | Time-to-flag (real-time) | — | <5 menit p95 |
| Operasional | Triage time | Baseline | -30% |
| Operasional | Case cycle time | Baseline | -25% |
| Operasional | Investigation coverage | ~25% (indikatif; benchmark vendor) | ≥95% |
| Operasional | Cases per investigator per month | Baseline | +30% |
| Bisnis | Fraud value protected | Baseline | Target tenant-specific |
| Bisnis | False positive rate | Baseline | -20% |
| Adopsi | % klaim yang di-scored | — | ≥80% |
| Adopsi | Weekly active investigators | — | ≥70% |
| Adopsi | % alert yang dikerjakan | — | ≥90% |
| Kepatuhan | Keputusan dengan audit trail | — | 100% |
| Kepatuhan | Temuan kritis audit | — | 0 |
| Guardrail | Complaint rate klaim | Baseline | Tidak naik |
| Guardrail | Fairness parity antar segmen | — | Dalam toleransi yang disetujui |
| Guardrail | Model drift unresolved | — | 0 >30 hari |
| Guardrail | Auto-denial | — | 0 |

> Catatan: baseline “benchmark vendor” bersifat indikatif dan tidak menggantikan pengukuran baseline tenant. Semua target harus dikunci bersama design partner sebelum pilot. Recall@k bukan recall pada volume alert tetap; k harus didefinisikan sebagai kapasitas review riil. Untuk thin slice, gunakan metrik BRD §12.2 (planted recall + false positive kontrol) dan laporkan terpisah dari business lift.

---

## 4. Persona dan Peran Pengguna

| Persona | Tujuan utama | Pain point | Kebutuhan produk | Ukuran sukses |
|---|---|---|---|---|
| **Kepala SIU / Claims Integrity Lead** | Menurunkan leakage dan meningkatkan produktivitas SIU | Tidak tahu ROI; alert menumpuk; sulit membuktikan hasil | Dashboard eksekutif, kapasitas tim, laporan audit, ROI | Leakage turun; coverage naik; audit bersih |
| **Investigator SIU** | Menyelesaikan kasus dengan bukti kuat | Data tersebar; dokumen manual; tidak ada timeline | Case workspace, evidence links, graph, document viewer | Waktu kasus turun; kasus selesai naik |
| **Analis Fraud** | Menemukan pola dan menyetel aturan/model | Tidak bisa menguji aturan; data tidak konsisten | Rule builder, backtesting, graph explorer, model monitoring | Deteksi naik; false positive turun |
| **Claims Adjuster / Handler** | Memutuskan klaim dengan cepat dan adil | Alert tidak jelas; bolak-balik minta dokumen | Alert reason codes, next-best-action, SLA | Waktu klaim turun; friksi turun |
| **Underwriter (jiwa/kesehatan)** | Menilai non-disclosure dan risiko aplikasi | Data aplikasi terbatas; sulit melihat riwayat | Underwriting risk score, referral, audit | Risiko terpilih lebih baik |
| **Risk & Compliance Officer** | Memastikan kepatuhan dan auditability | Keputusan black-box; data retention tidak jelas | Explainability, audit log, regulatory reports, DSR | Temuan audit turun; kepatuhan terjaga |
| **Data/IT Engineer** | Mengintegrasikan platform dengan aman | Mapping data rumit; API tidak jelas | API docs, connector framework, observability, SSO | Integrasi <8 minggu; insiden turun |
| **Platform Admin** | Mengelola tenant, peran, dan konfigurasi | Konfigurasi tersebar; sulit mengontrol akses | Admin console, RBAC, audit, usage metering | Onboarding <1 hari; akses terkontrol |
| **Finance/Management (read-only)** | Melihat usage, value report, dan billing export | Sulit menghitung ROI dan biaya per tenant | Usage metering, value report, billing export, audit | Rekonsiliasi 100%; ROI terukur |
| **Provider Network Manager (kesehatan)** | Mengelola risiko provider | Tidak ada benchmark billing; audit manual | Provider peer analytics, case management | Provider anomali turun |
| **Auditor/Regulator (read-only)** | Memeriksa keputusan dan bukti | Tidak bisa merekonstruksi keputusan | Read-only export, audit trail, evidence index | Pemeriksaan lebih cepat |

> **Phase 0/1 thin slice** hanya mengaktifkan peran **Reviewer**, **Supervisor**, dan **Admin** (BRD §11). Persona lain adalah target-state; jangan mengklaim sudah didukung di thin slice. Finance/Management dapat dilayani lewat ekspor read-only, bukan login sistem.

---

## 5. Alur Kerja Utama

> **Fase:** J1–J2 commercial MVP; J3–J5 Phase 2; J6 Phase 2; J7 Phase 2/3. **Thin slice** hanya mencakup subset J1–J2 (CSV/JSON, rules-only, tanpa ML/OCR/integrasi).

### 5.1 J1 — Klaim P&C auto (FNOL sampai keputusan)

1. Klaim masuk melalui aplikasi, agen, bengkel, atau call center.
2. Platform mengonsumsi data klaim, polis, pihak, kendaraan, dan dokumen.
3. Entity resolution mencocokkan tertanggung, pengemudi, kendaraan, bengkel, dan rekening.
4. Rules dan model menghasilkan skor, reason codes, dan expected loss.
5. Alert masuk ke queue; triage memprioritaskan berdasarkan risiko, nilai, SLA, dan kapasitas.
6. Investigator membuka kasus: melihat timeline, dokumen, graph relasi, dan riwayat klaim.
7. Investigator memutuskan: bayar, investigasi lanjut, tolak, rujuk ke penegak hukum, atau recovery.
8. Keputusan dan outcome dikembalikan ke model sebagai feedback.

**Momen produk yang menentukan:** skor real-time saat FNOL, deteksi duplikasi lintas polis, checklist dokumen otomatis, graph relasi, dan ringkasan kasus.

### 5.2 J2 — Klaim properti (termasuk CAT surge)

1. Klaim properti masuk dengan foto/video, estimasi kerusakan, dan invoice.
2. Platform memvalidasi lokasi, waktu kejadian, polis, dan riwayat klaim.
3. Document AI memeriksa konsistensi foto, invoice, dan estimasi.
4. Graph mendeteksi vendor/contractor yang berulang atau terkait pihak internal.
5. Mode surge mengaktifkan prioritisasi dan kapasitas tambahan.
6. Investigator menyelesaikan kasus dengan bukti yang tertaut.

### 5.3 J3 — Klaim jiwa (death claim)

1. Klaim meninggal masuk dengan dokumen kematian, identitas ahli waris, dan polis.
2. Platform memverifikasi identitas, hubungan ahli waris, dan riwayat polis.
3. Document AI memeriksa keaslian dokumen dan konsistensi tanggal.
4. Model menilai risiko non-disclosure, klaim dini, dan pola agen/broker.
5. Investigator meninjau rekam medis (dengan legal basis), timeline, dan graph relasi.
6. Keputusan dan alasan tercatat; kasus ditutup dengan audit trail.

### 5.4 J4 — Klaim kesehatan (cashless dan reimbursement)

1. Klaim masuk dari provider atau nasabah.
2. Platform mencocokkan peserta, polis, provider, diagnosis, prosedur, dan tarif.
3. Model mendeteksi upcoding, phantom billing, duplikasi, dan layanan tidak perlu.
4. Provider peer benchmarking membandingkan pola billing dengan peer.
5. Alert membuat kasus provider atau kasus peserta.
6. Investigator/auditor provider menyelesaikan kasus dan memberi feedback.

### 5.5 J5 — Underwriting fraud (jiwa/kesehatan)

1. Aplikasi masuk dengan data medis, finansial, dan agen.
2. Platform menilai risiko non-disclosure dan indikasi misrepresentasi.
3. Underwriter menerima skor, alasan, dan rekomendasi pemeriksaan tambahan.
4. Keputusan underwriting tercatat; outcome menjadi label untuk model.

### 5.6 J6 — Investigasi jaringan/portofolio

1. Analis membuka graph explorer.
2. Mencari entitas: telepon, alamat, rekening, kendaraan, provider, agen.
3. Platform menampilkan klaster, centrality, dan pola berulang.
4. Analis membuat kasus jaringan dan menautkan klaim terkait.
5. Investigator menyelesaikan kasus; outcome memperbarui model dan rule.

### 5.7 J7 — Tata kelola model

1. Tim data memantau performa, drift, dan fairness.
2. Investigator memberi label true/false positive dan outcome.
3. Tim data menjalankan retraining di lingkungan sandbox.
4. Model divalidasi, disetujui, dan dipromosikan dengan model card.
5. Audit mencatat versi model, threshold, dan perubahan.

---

## 6. Ruang Lingkup dan Strategi Rilis

### 6.1 In scope

- **Thin slice (BRD v0.5):** rules A–G (33 rule; famili H bergerbang off), CSV/JSON upload, flag bundle, queue, evidence, coverage matrix, audit, usage metering; P&C auto/properti; data sintetis; lokal/offline; **tanpa ML, OCR, integrasi, multi-tenant, atau write-back**.
- **Commercial MVP:** shared core (ingestion, entity resolution, rules, baseline ML, graph basic, document AI basic, alert triage, case management, reporting, model governance, admin, security, audit); P&C auto/properti; 1 integrasi claims; OCR basic; dashboard; audit; mobile reviewer/supervisor read-only + push.
- Modul jiwa dan kesehatan pada Phase 2.
- Integrasi dengan claims/PAS, dokumen, dan data eksternal pada Phase 1b/2.
- Dukungan Bahasa Indonesia, IDR, format NIK/NPWP, dan data residency Indonesia untuk target-state.
- Kepatuhan UU PDP, OJK, dan praktik perlindungan konsumen untuk produksi; thin slice menggunakan data sintetis.

### 6.2 Out of scope

- Menjadi core policy administration atau claims system.
- Pemrosesan pembayaran klaim.
- Document management system penuh.
- Transaction monitoring AML umum (kecuali integrasi untuk asuransi jiwa yang relevan).
- Penuntutan pidana atau representasi hukum.
- Core BPJS Kesehatan.
- Penolakan klaim otomatis.
- Underwriting pricing/rating; fokus hanya fraud risk.
- Fraud di luar klaim asuransi (mis. investasi, aset perusahaan).
- **Thin slice (BRD v0.5):** ML scoring, OCR/forensik dokumen, API/SFTP/real-time, write-back ke claims system, multi-tenant/SSO/MFA, data eksternal/konsorsium, DSR/DPIA, sertifikasi WCAG, dan SLA komersial. Semua ini tetap target-state PRD, bukan Phase 0/1.

### 6.3 MVP vs Phase 2 vs Phase 3

**Layer scope:**

| Layer | Scope | Timeline |
|---|---|---|
| **Thin slice (BRD v0.5)** | Rules A–G, CSV/JSON, queue, evidence, coverage, audit, usage metering; P&C; data sintetis; lokal; tanpa ML/OCR/integrasi/write-back | Hackathon / Phase 0/1 |
| **Commercial MVP** | P&C auto/properti; rules + baseline ML; OCR basic; 1 integrasi; case management; dashboard; audit; mobile reviewer/supervisor read-only + push | M2–M9 (build M2–M6; pilot M6–M9) |
| **Phase 2** | Jiwa/kesehatan; graph lanjutan; document forensics; external connectors | M9–M15 |
| **Phase 3** | GenAI assistant; agentic evidence; consortium; ASEAN | M15–M24 |

> Kolom **MVP** pada tabel di bawah adalah **commercial MVP**, bukan thin slice. Thin slice adalah subset yang divalidasi lebih dulu dan tidak boleh dijadikan alasan untuk menunda kontrol produksi.

| Area | MVP (M2–M6 build; M6–M9 pilot) | Phase 2 (M9–M15) | Phase 3 (M15–M24) |
|---|---|---|---|
| Lini | P&C auto & properti | Jiwa & kesehatan | Multi-negara ASEAN |
| Deteksi | Rules + baseline ML | Model per lini, provider analytics | Ensemble, adaptive models |
| Graph | Basic entity graph & link alerts | Ring detection, community detection, graph explorer lanjutan | Cross-tenant consortium graph (opt-in) |
| Dokumen | OCR basic, duplicate detection | Forensik dokumen, tamper detection, cross-document consistency | Agentic evidence gathering |
| Case management | End-to-end untuk P&C | Template jiwa/kesehatan, provider cases | Orkestrasi multi-pihak |
| Integrasi | 1 claims system + CSV/SFTP | Konektor eksternal (identitas, kendaraan, provider, cuaca) | Ecosystem marketplace |
| AI | Scoring + reason codes | Model tuning UI, champion/challenger | GenAI assistant, autonomous investigation support |
| Pelanggan | 2 design partner P&C | 4–6 pelanggan (termasuk jiwa/kesehatan) | 10+ pelanggan, ekspansi regional |
| Mobile | Reviewer/supervisor: alert inbox, case read-only, push; aksi tulis di web | Notifikasi lebih kaya | App pemegang polis; mobile write |

### 6.3.1 Mobile MVP (M-01..M-10)

Mobile MVP adalah channel tambahan reviewer/supervisor, **bukan bagian thin slice**. Thin slice tetap web + Docker Compose. Mobile MVP bersifat **read-only + push**; semua aksi tulis (assign, escalate, catatan, keputusan) dilakukan di web.

| ID | Requirement | Acceptance | Trace |
|---|---|---|---|
| M-01 | iOS native Swift 6+/SwiftUI | TestFlight build; Keychain; Face ID/Touch ID | TRD §8.10, §23.5 |
| M-02 | Android native Kotlin/Compose | Play internal; Keystore/BiometricPrompt | TRD §8.11, §23.5 |
| M-03 | Shared business logic KMP/KMM | Contract test iOS/Android; domain/use case sama | TRD §8.8, §8.9 |
| M-04 | Clean Architecture/Clean Code mobile | Dependency rule; import lint; review | TRD §8.8, §19.1 |
| M-05 | Clean Code backend (lintas platform) | Typed error; dependency rule; coverage gate | TRD §7.12, §16.2 |
| M-06 | Unit test semua platform (lintas platform) | CI coverage gate Go, web, KMP, iOS, Android | TRD §17.9 |
| M-07 | Offline read-only | Cache alert/case; tidak ada antrean write | TRD §8.12 |
| M-08 | Auth aman + biometrik native | OIDC PKCE; MFA; re-auth biometrik; MASVS L1 | TRD §8.13 |
| M-09 | Push tanpa PII + deep link | Payload test; deep link ke alert/case | TRD §8.12, §8.14 |
| M-10 | Distribusi + staged rollout | TestFlight/Play internal; kill switch | TRD §8.15, §16.9 |

> Mobile MVP tidak memuat upload dokumen, aksi tulis, atau keputusan akhir. Jika product memutuskan mobile write, requirement baru harus ditambahkan ke PRD dan traceability TRD §23.5.

### 6.4 Design partner

- **Thin slice:** tidak memerlukan design partner; menggunakan data sintetis dan dataset berlabel. Ini adalah validasi teknis, bukan pilot komersial.
- **Commercial MVP:** 2 perusahaan asuransi P&C (idealnya 1 Tier 1 dan 1 mid-market) yang bersedia memberikan data historis, akses shadow mode, dan feedback mingguan. Minimal 2 paid pilot direkomendasikan; pilot fee dikreditkan ke kontrak tahunan (BRD §3A).
- **Phase 2:** 1 perusahaan jiwa dan 1 perusahaan kesehatan.
- **Kriteria design partner:** memiliki volume klaim yang cukup, tim SIU aktif, sponsor eksekutif, dan komitmen untuk menyediakan label/outcome.

---

## 7. Kebutuhan Fungsional

### 7.1 Ringkasan epic

| Epic | Outcome | Prioritas | Fase |
|---|---|---|---|
| E1. Data ingestion & integration | Data klaim/polis/pihak/dokumen masuk secara akurat dan tepat waktu | P0 | MVP |
| E2. Entity resolution & identity | Satu Golden Record per pihak; relasi terlihat | P0 | MVP |
| E3. Rules engine | Aturan dapat dibuat, diuji, dan dipublikasikan oleh bisnis | P0 | MVP |
| E4. ML scoring | Skor risiko yang terkalibrasi dan dapat dijelaskan | P0 | MVP |
| E5. Graph & network analytics | Fraud ring dan relasi tersembunyi terdeteksi | P1 | MVP basic / Phase 2 advanced |
| E6. Document & media forensics | Dokumen palsu/dimodifikasi terdeteksi | P1 | MVP basic / Phase 2 advanced |
| E7. Alert triage | Alert diprioritaskan berdasarkan risiko, nilai, dan kapasitas | P0 | MVP |
| E8. Case management | Investigasi end-to-end dengan bukti dan audit trail | P0 | MVP |
| E9. Decisioning & workflow | Keputusan konsisten, human-in-the-loop; MVP read-only, write-back Phase 2 opt-in | P0 | MVP (read-only) / Phase 2 (write-back) |
| E10. Reporting & dashboards | Visibilitas operasional, model, dan eksekutif | P0 | MVP |
| E11. Model monitoring & feedback | Model belajar dari outcome; drift terkelola | P0 | MVP |
| E12. Admin & multi-tenancy | Tenant, peran, konfigurasi, usage metering | P0 | MVP |
| E13. Security, privacy & audit | Kepatuhan UU PDP dan audit-ready | P0 | MVP |
| E14. External data & consortium | Data eksternal dan berbagi data opt-in | P1 | Phase 2/3 |
| E15. GenAI investigation assistant | Ringkasan, Q&A, draft report dengan sitasi | P2 | Phase 3 |
| E16. Notifications & collaboration | Tim SIU bekerja bersama secara aman | P1 | MVP basic |
| E17. Thin slice (BRD v0.5) | Rules-only P&C, evidence, coverage, metering; tanpa ML/OCR/integrasi | P0 | Thin slice |

### 7.2 E1 — Data Ingestion & Integration

**Outcome:** data dari claims/PAS, polis, pihak, agen, provider, dan dokumen dapat masuk secara real-time maupun batch, dengan kualitas yang terukur.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-101 | Sebagai Data Engineer, saya dapat menghubungkan sumber klaim, polis, pihak, agen, dan provider melalui API, SFTP, atau batch. | Koneksi teruji; mapping field; idempotensi; error queue; data lineage; log akses. | P0 | MVP |
| FR-102 | Sebagai Claims Ops, saya dapat mengirim klaim real-time saat FNOL dan batch harian. | API p95 <2 detik; batch hingga 1 juta klaim/hari (platform); retry otomatis; status ingestion. | P0 | MVP |
| FR-103 | Sebagai Admin, saya dapat memetakan field khusus lini (P&C, jiwa, kesehatan) tanpa deploy ulang. | Template mapping per lini; validasi; versi; rollback. | P0 | MVP |
| FR-104 | Sebagai Data Steward, saya dapat melihat kualitas data per sumber dan per lini. | Dashboard completeness, duplicate, match rate, timeliness; alert jika di bawah threshold. | P0 | MVP |
| FR-105 | Sebagai Auditor, saya dapat melacak asal setiap field yang memengaruhi skor. | Data lineage; timestamp; source; transformasi; immutable log. | P0 | MVP |
| FR-106 | Sebagai Admin, saya dapat mengimpor data manual via CSV/Excel untuk sumber yang belum terintegrasi. | Template; validasi; preview; error report; audit. | P1 | MVP |

### 7.3 E2 — Entity Resolution & Identity

**Outcome:** pihak yang sama tidak terpecah di banyak ID; relasi antar pihak, kendaraan, alamat, rekening, dan provider terlihat.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-201 | Sebagai Analis, saya dapat mencocokkan pihak menggunakan NIK, NPWP, telepon, email, alamat, plat, dan rekening. | Deterministic + probabilistic matching; confidence score; manual merge/split; audit. | P0 | MVP |
| FR-202 | Sebagai Investigator, saya dapat melihat Golden Record dan riwayat relasi pihak. | 360 view; tipe relasi; riwayat perubahan; sumber data. | P0 | MVP |
| FR-203 | Sebagai Data Engineer, saya dapat menormalisasi identifier Indonesia. | Validasi NIK 16 digit, NPWP, telepon +62, plat, alamat; match rate terukur. | P0 | MVP |
| FR-204 | Sebagai Compliance, saya dapat membatasi akses PII berdasarkan peran. | Field-level masking; RBAC/ABAC; audit akses. | P0 | MVP |
| FR-205 | Sebagai Analis, saya dapat meninjau kandidat false merge. | Review queue; approve/reject; metrik false merge; audit. | P1 | MVP |
| FR-206 | Sebagai Investigator, saya dapat menautkan entitas baru (provider, bengkel, agen) ke kasus. | One-click link; evidence; audit; graph update. | P1 | Phase 2 |

### 7.4 E3 — Rules Engine

**Outcome:** aturan fraud dapat dikelola oleh tim bisnis, diuji pada data historis, dan dipublikasikan dengan aman.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-301 | Sebagai Analis Fraud, saya dapat membuat aturan tanpa coding. | Rule builder; kondisi; bobot; versioning; sandbox; approval; rollback. | P0 | MVP |
| FR-302 | Sebagai Analis, saya dapat menggunakan rule pack per lini dan tipologi. | P&C auto/properti pada MVP; jiwa/kesehatan pada Phase 2; rule pack dapat dikustomisasi. | P0 | MVP |
| FR-303 | Sebagai Analis, saya dapat menguji aturan pada klaim historis. | Backtesting; impact analysis; precision/recall; volume alert; segment. | P0 | MVP |
| FR-304 | Sebagai Kepala SIU, saya dapat melihat performa setiap aturan. | Hit rate; confirmation rate; false positive; override; trend. | P0 | MVP |
| FR-305 | Sebagai Compliance, saya dapat melihat perubahan aturan dan approval. | Change log; approver; effective date; audit. | P0 | MVP |
| FR-306 | Sebagai Analis, saya dapat mengatur threshold berdasarkan kapasitas SIU. | Simulasi volume; rekomendasi threshold; approval. | P1 | Phase 2 |

### 7.5 E4 — ML Scoring

**Outcome:** skor risiko yang lebih akurat daripada rules-only, terkalibrasi, dan dapat dijelaskan.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-401 | Sebagai Kepala SIU, saya dapat menggunakan model baseline per lini. | AUC/PR-AUC; lift vs rules; kalibrasi; evaluasi per segmen. | P0 | MVP |
| FR-402 | Sebagai Data Scientist, saya dapat membangun feature dari klaim, polis, pihak, jaringan, dokumen, provider, geo, dan waktu. | Feature store; point-in-time correctness; lineage. | P0 | MVP |
| FR-403 | Sebagai Claims Ops, saya dapat melakukan scoring real-time dan batch. | API p95 <2 detik; batch; idempotensi; fallback jika model gagal. | P0 | MVP |
| FR-404 | Sebagai Investigator, saya dapat melihat alasan skor. | Top drivers; SHAP/feature contribution; bahasa manusia; evidence. | P0 | MVP |
| FR-405 | Sebagai Data Scientist, saya dapat menjalankan champion/challenger dan shadow mode. | Traffic split; comparison; promotion approval; rollback. | P1 | Phase 2 |
| FR-406 | Sebagai Kepala SIU, saya dapat merekomendasikan threshold berdasarkan kapasitas dan expected loss. | Simulasi; target volume; ROI; approval. | P1 | Phase 2 |
| FR-407 | Sebagai Compliance, saya dapat memastikan model tidak memakai atribut terlarang dan diuji fairness. | Protected attributes excluded; bias test; segment report; approval. | P0 | MVP |
| FR-408 | Sebagai Data Scientist, saya dapat memantau drift dan melakukan retraining. | PSI/KS; alert; retraining workflow; model card; audit. | P0 | MVP |

### 7.6 E5 — Graph & Network Analytics

**Outcome:** fraud ring, kolusi, dan relasi tersembunyi terdeteksi dan dapat dieksplorasi.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-501 | Sebagai Analis, saya dapat membangun graph dari entitas dan event. | Node/edge; incremental update; query <3 detik; provenance. | P1 | MVP basic |
| FR-502 | Sebagai Analis, saya dapat mendeteksi pola ring. | Shared phone/address/bank/vehicle; community detection; risk score. | P1 | Phase 2 |
| FR-503 | Sebagai Investigator, saya dapat mengeksplorasi graph. | Search; expand; filter; path finding; timeline; export. | P1 | Phase 2 |
| FR-504 | Sebagai Investigator, saya dapat menautkan alert graph ke kasus. | One-click; evidence; audit; graph snapshot. | P1 | Phase 2 |
| FR-505 | Sebagai Kepala SIU, saya dapat melihat network risk score per klaim. | Centrality/community features; explainability; segment. | P2 | Phase 2 |

### 7.7 E6 — Document & Media Forensics

**Outcome:** dokumen dan media palsu/dimodifikasi terdeteksi, dan temuan dapat dipertahankan sebagai bukti.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-601 | Sebagai Investigator, saya dapat mengekstrak data dari dokumen Indonesia. | KTP, SIM, STNK, BPKB, invoice, kuitansi, rekam medis, surat kematian, laporan polisi; akurasi field utama ≥95% pada benchmark 500 dokumen sintetis + 200 dokumen real teranonimisasi (legal basis); per-field accuracy; confidence; fallback review manual jika confidence <60%. | P1 | MVP basic |
| FR-602 | Sebagai Investigator, saya dapat mendeteksi manipulasi dokumen. | Metadata; ELA; duplicate hash; font/format anomaly; confidence; evidence. | P1 | Phase 2 |
| FR-603 | Sebagai Investigator, saya dapat memeriksa konsistensi antar dokumen. | Nama, tanggal, jumlah, plat, provider; mismatch alert. | P1 | Phase 2 |
| FR-604 | Sebagai Investigator, saya dapat melihat dan menganotasi dokumen. | Zoom; highlight; redaction; audit. | P1 | Phase 2 |
| FR-605 | Sebagai Analis, saya dapat memberi feedback atas hasil ekstraksi. | Accept/reject; label; feedback loop. | P1 | Phase 2 |
| FR-606 | Sebagai Compliance, saya dapat memastikan dokumen sensitif dimasking. | PII/PHI redaction; role-based access; audit. | P0 | MVP |

### 7.8 E7 — Alert Triage & Prioritization

**Outcome:** tim SIU bekerja pada alert yang paling berdampak, bukan yang paling berisik.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-701 | Sebagai Investigator, saya dapat melihat semua alert dalam satu queue. | Rules/ML/graph/doc alerts; dedupe; grouping; filter. | P0 | MVP |
| FR-702 | Sebagai Investigator, saya dapat melihat skor, alasan, expected loss, dan confidence. | Reason codes; evidence links; severity; SLA. | P0 | MVP |
| FR-703 | Sebagai Kepala SIU, saya dapat memprioritaskan berdasarkan kapasitas dan nilai. | Capacity-aware ranking; SLA; claim value; expected loss. | P0 | MVP |
| FR-704 | Sebagai Kepala SIU, saya dapat melakukan assignment dan routing. | Skill-based; workload; escalation; audit. | P0 | MVP |
| FR-705 | Sebagai Analis, saya dapat melakukan suppression/whitelist dengan approval. | Reason; expiry; approver; audit. | P1 | MVP |
| FR-706 | Sebagai Investigator, saya dapat memberi feedback true/false positive. | Structured reason; label; model feedback; audit. | P0 | MVP |
| FR-707 | Sebagai Investigator, saya dapat melihat alert terkait dalam satu grup. | Group by entity/network/claim; merge/split; audit. | P1 | Phase 2 |

### 7.9 E8 — Investigation Case Management

**Outcome:** investigasi end-to-end dengan bukti, kolaborasi, dan keputusan yang dapat diaudit.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-801 | Sebagai Investigator, saya dapat membuat kasus dari alert. | Auto-populate claim, policy, parties, timeline, docs, score; editable. | P0 | MVP |
| FR-802 | Sebagai Investigator, saya dapat menggunakan template investigasi per tipologi. | P&C auto/properti pada MVP; jiwa/kesehatan pada Phase 2; tasks; SLA; checklist. | P0 | MVP |
| FR-803 | Sebagai Investigator, saya dapat mengelola bukti. | Upload; link; chain of custody; notes; interviews; field visits; audit. | P0 | MVP |
| FR-804 | Sebagai Investigator, saya dapat berkolaborasi. | Comments; mentions; watchers; role-based access; audit. | P1 | MVP |
| FR-805 | Sebagai Investigator, saya dapat membuat keputusan. | Pay/investigate/deny/refer/recover; approval matrix; reason; no auto denial. | P0 | MVP |
| FR-806 | Sebagai Kepala SIU, saya dapat menutup kasus dan mencatat outcome. | Disposition; recovered amount; root cause; feedback; audit. | P0 | MVP |
| FR-807 | Sebagai Auditor, saya dapat mengekspor case file. | PDF/zip; evidence index; redaction; audit. | P1 | Phase 2 |
| FR-808 | Sebagai Investigator, saya dapat melihat ringkasan kasus otomatis. | Summary with citations; editable; confidence; human review. | P2 | Phase 3 |

### 7.10 E9 — Decisioning & Workflow

**Outcome:** keputusan konsisten, sesuai kebijakan, dan terhubung kembali ke claims system.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-901 | Sebagai Admin, saya dapat mengonfigurasi workflow per tenant/lini. | States; transitions; approvals; SLA; versioning; audit. | P0 | MVP |
| FR-902 | Sebagai Investigator, saya dapat menjalankan aksi otomatis. | Request docs; external checks; assign; hold payment (dengan rules); notify. | P1 | Phase 2 |
| FR-903 | Sebagai Compliance, saya dapat memastikan tidak ada penolakan otomatis. | Denial requires human; reason; second approval untuk nilai tinggi; audit. | P0 | MVP |
| FR-904 | Sebagai Claims Ops, saya dapat mengirim keputusan ke claims system. | Status/decision/notes; API; idempotensi; error handling. **Default MVP: read-only; write-back Phase 2, opt-in dengan sign-off tenant dan review compliance.** | P1 | Phase 2 |
| FR-905 | Sebagai Kepala SIU, saya dapat mengelola SLA dan eskalasi. | Timers; notifications; breach report; audit. | P1 | Phase 2 |
| FR-906 | Sebagai Admin, saya dapat mengatur approval matrix. | Role/value/risk-based; delegation; audit. | P0 | MVP |

### 7.11 E10 — Reporting & Dashboards

**Outcome:** eksekutif, operasional, dan compliance melihat gambaran yang sama dan dapat mengekspor bukti.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1001 | Sebagai Eksekutif, saya dapat melihat fraud prevented/recovered, leakage, ROI, dan productivity. | Dashboard; filter lini/periode; drill-down; export. | P0 | MVP |
| FR-1002 | Sebagai Kepala SIU, saya dapat melihat queue aging, cycle time, workload, dan SLA. | Real-time/near-real-time; alert; drill-down. | P0 | MVP |
| FR-1003 | Sebagai Data Scientist, saya dapat melihat precision/recall, drift, calibration, dan segment. | Dashboard; alert; export. | P0 | MVP |
| FR-1004 | Sebagai Kepala Klaim, saya dapat melihat dampak per lini/portofolio. | Loss ratio impact; typology; provider/agent network; export. | P1 | Phase 2 |
| FR-1005 | Sebagai Auditor, saya dapat mengunduh laporan kepatuhan. | Case file; decisions; evidence; access logs; PDF/Excel. | P0 | MVP |
| FR-1006 | Sebagai Admin, saya dapat menjadwalkan laporan. | Email; PDF/Excel; role-based; audit. | P1 | Phase 2 |

### 7.12 E11 — Model Monitoring & Feedback

**Outcome:** model dan aturan belajar dari outcome, dengan kontrol risiko model.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1101 | Sebagai Investigator, saya dapat mencatat disposition dan outcome. | Structured reason codes; recovered amount; time; audit. | P0 | MVP |
| FR-1102 | Sebagai Data Scientist, saya dapat mengelola label store dan active learning. | Queue; sampling; labeling; quality checks. | P1 | Phase 2 |
| FR-1103 | Sebagai Data Scientist, saya dapat memantau performa model. | PR-AUC; calibration; drift; fairness; segment. | P0 | MVP |
| FR-1104 | Sebagai Data Scientist, saya dapat melakukan retraining dan approval. | Pipeline; validation; model card; approval; rollback. | P1 | Phase 2 |
| FR-1105 | Sebagai Data Scientist, saya dapat menjalankan eksperimen. | Champion/challenger; A/B; holdout; statistical significance. | P2 | Phase 3 |

### 7.13 E12 — Admin & Multi-Tenancy

**Outcome:** tenant dapat onboard cepat, aman, dan terisolasi.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1201 | Sebagai Admin, saya dapat melakukan onboarding tenant. | <1 hari; konfigurasi lini; branding; SSO; data residency. | P0 | MVP |
| FR-1202 | Sebagai Admin, saya dapat mengelola peran dan izin. | RBAC/ABAC; least privilege; access review; audit. | P0 | MVP |
| FR-1203 | Sebagai Admin, saya dapat mengelola versi konfigurasi. | Draft/publish/rollback; audit; approval. | P0 | MVP |
| FR-1204 | Sebagai Finance, saya dapat melihat usage metering. | Claims scored; documents processed; flags; case groups; users; API calls; billing export; rekonsiliasi 100% dengan event sumber; tanpa PII. | P0 | MVP |
| FR-1205 | Sebagai Data Engineer, saya dapat menggunakan sandbox. | Synthetic data; safe testing; isolated. | P0 | MVP |
| FR-1206 | Sebagai Admin, saya dapat mengelola feature flags. | Per tenant; audit; rollback. | P1 | Phase 2 |
| FR-1207 | Sebagai Kepala SIU/Finance, saya dapat melihat value report per batch/tenant. | Confirmed cases; value protected/recovered; time saved; formula version; confidence; tanpa double counting; ekspor tanpa PII. | P0 | MVP |
| FR-1208 | Sebagai Admin, saya dapat mengonfigurasi plan/entitlement. | Volume band; feature flags; peringatan overage 80%/100%; audit; tanpa silent blocking. | P1 | MVP |
| FR-1209 | Sebagai Finance, saya dapat mengekspor data billing. | Periode; plan; usage; overage; unit price; checksum; rekonsiliasi 100% dengan metering; tanpa PII. | P0 | MVP |
| FR-1210 | Sebagai Kepala SIU, saya dapat melihat pilot scorecard. | Baseline vs pilot; konversi; sign-off Supervisor/Admin; rekomendasi. | P1 | MVP |

### 7.14 E13 — Security, Privacy & Audit

**Outcome:** platform memenuhi kebutuhan keamanan dan privasi data asuransi Indonesia.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1301 | Sebagai User, saya dapat login dengan SSO/MFA. | SAML/OIDC; MFA; session timeout; audit. | P0 | MVP |
| FR-1302 | Sebagai Security Officer, saya dapat memastikan data terenkripsi. | AES-256 at rest; TLS 1.2+; KMS; key rotation. | P0 | MVP |
| FR-1303 | Sebagai Compliance, saya dapat membatasi PII/PHI. | Field-level masking; role-based; audit. | P0 | MVP |
| FR-1304 | Sebagai Data Subject, saya dapat mengajukan akses/koreksi/penghapusan. | DSR workflow; SLA; audit; verifikasi identitas. | P0 | MVP |
| FR-1305 | Sebagai Auditor, saya dapat melihat audit log. | Immutable; who/what/when; export; retention. | P0 | MVP |
| FR-1306 | Sebagai Admin, saya dapat memilih data residency Indonesia. | Region Indonesia; tenant config; verifikasi. | P0 | MVP |
| FR-1307 | Sebagai Security Officer, saya dapat menjalankan incident response. | Runbook; notification; post-mortem; audit. | P0 | MVP |
| FR-1308 | Sebagai Compliance, saya dapat mengatur retensi dan legal hold. | Configurable; deletion; hold; audit. | P0 | MVP |

### 7.15 E14 — External Data & Consortium

**Outcome:** data eksternal dan berbagi data lintas insurer meningkatkan deteksi tanpa melanggar privasi/kompetisi.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1401 | Sebagai Data Engineer, saya dapat menambahkan konektor data eksternal. | Auth; rate limit; caching; cost; provenance; error handling. | P1 | Phase 2 |
| FR-1402 | Sebagai Analis, saya dapat memakai data identitas, kendaraan, perusahaan, pengadilan, sanksi, geo, dan cuaca. | Configurable; legal basis; consent; audit. | P1 | Phase 2 |
| FR-1403 | Sebagai Analis Kesehatan, saya dapat memakai referensi ICD-10, prosedur, dan tarif. | Versioning; mapping; audit. | P1 | Phase 2 |
| FR-1404 | Sebagai Kepala SIU, saya dapat ikut data konsorsium secara opt-in. | Governance; anonymization; audit; opt-out. | P2 | Phase 3 |
| FR-1405 | Sebagai Finance, saya dapat melihat biaya data eksternal. | Per-query cost; budget alert; export. | P2 | Phase 3 |

### 7.16 E15 — GenAI Investigation Assistant

**Outcome:** investigator mendapat ringkasan, Q&A, dan draft laporan dengan sitasi, tanpa menggantikan keputusan manusia.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1501 | Sebagai Investigator, saya dapat membuat ringkasan kasus dengan sitasi. | Fakta klaim; dokumen; timeline; citations; confidence. | P2 | Phase 3 |
| FR-1502 | Sebagai Investigator, saya dapat bertanya tentang isi kasus. | RAG; citations; refusal jika bukti tidak cukup. | P2 | Phase 3 |
| FR-1503 | Sebagai Investigator, saya dapat meminta saran langkah investigasi. | Typology-based; human approval; audit. | P2 | Phase 3 |
| FR-1504 | Sebagai Investigator, saya dapat membuat draft laporan. | Structured; citations; editable; audit. | P2 | Phase 3 |
| FR-1505 | Sebagai Compliance, saya dapat memastikan guardrails LLM. | PII redaction; no training on tenant data; prompt injection defense; hallucination eval; human review. | P2 | Phase 3 |

### 7.17 E16 — Notifications & Collaboration

**Outcome:** tim SIU dan pemangku kepentingan mendapat informasi tepat waktu tanpa membocorkan data.

| ID | User story | Acceptance criteria | Prioritas | Fase |
|---|---|---|---|---|
| FR-1601 | Sebagai User, saya dapat mengatur notifikasi in-app/email. | Preferences; digest; escalation; audit. | P1 | MVP |
| FR-1602 | Sebagai Investigator, saya dapat berkolaborasi dalam kasus. | Comments; mentions; tasks; role-based; audit. | P1 | MVP |
| FR-1603 | Sebagai Admin, saya dapat mengirim notifikasi eksternal yang aman. | No PII; opt-in; template; audit. | P2 | Phase 2 |

### 7.18 Kebutuhan khusus per lini

| Aspek | P&C | Jiwa | Kesehatan |
|---|---|---|---|
| **Data inti** | Polis, klaim, kendaraan, properti, bengkel, vendor, estimasi kerusakan, foto | Polis, peserta, ahli waris, agen, riwayat medis, dokumen kematian | Peserta, provider, diagnosis, prosedur, tarif, resep, hasil lab, klaim cashless/reimbursement |
| **Tipologi utama** | Staged accident, inflated repair, duplikasi klaim, arson, contractor/vendor fraud, total loss | Non-disclosure, klaim dini, beneficiary fraud, fake death, misrepresentasi agen, klaim luar negeri | Upcoding, phantom billing, duplikasi, layanan tidak perlu, provider collusion, member eligibility fraud |
| **Fitur/model** | Repair estimate anomaly, vehicle history, geo/weather, bengkel/vendor network, severity | Policy timeline, medical history, beneficiary graph, agent/broker pattern, waiting period | Provider peer benchmark, diagnosis–procedure mismatch, frequency/severity, network provider |
| **Dokumen** | STNK, BPKB, SIM, invoice bengkel, foto, laporan polisi, estimasi | KTP, surat kematian, rekam medis, polis, dokumen ahli waris | Kartu peserta, resep, invoice rumah sakit, rekam medis, hasil lab |
| **Regulasi/kepatuhan** | OJK, AAUI, perlindungan konsumen | OJK, AAJI, APU-PPT, perlindungan konsumen | OJK, UU Kesehatan, kerahasiaan PHI, perlindungan konsumen |
| **KPI lini** | Loss ratio, severity, repair cycle time, recovery | Early claim ratio, mortality claim, agent persistency, recovery | Loss ratio, provider cost, claim frequency, recovery |
| **Integrasi** | Bengkel, Samsat, telematik, cuaca, reasuransi | Rumah sakit/lab, agen, reasuransi | Provider, TPA, asosiasi, sistem klaim |

> Akses data kesehatan, rekam medis, dan data pribadi sensitif hanya boleh dilakukan dengan legal basis yang sah, persetujuan yang diperlukan, dan kontrol akses yang ketat.

### 7.19 E17 — Thin slice (BRD v0.5) requirements delta

Thin slice adalah implementasi Phase 0/1 yang **rules-only, P&C auto/properti, data sintetis, lokal/offline**. Requirement di bawah mengikat thin slice dan tidak menggantikan epic target-state.

| ID | Requirement | Acceptance criteria | Prioritas |
|---|---|---|---|
| FR-1701 | Katalog 43 rule (A–H) dengan ID, kondisi, parameter, severity, owner, dan mapping tipologi | A–G = 33 rule aktif; H = 10 rule bergerbang off; katalog dapat diaudit | P0 |
| FR-1702 | Flag bundle + penjelasan wajib 4 elemen | 100% flag punya rule+versi, nilai pemicu, referensi record/dokumen, penjelasan bahasa manusia | P0 |
| FR-1703 | Triage-only/read-only | Status klaim identik dengan sumber; tidak ada auto-denial; tidak ada write-back | P0 |
| FR-1704 | Reason code taxonomy | `CLEARED` wajib reason code; kategori FP/NI/CF tersedia; library final `TBD` | P0 |
| FR-1705 | Prioritas deterministik | Severity → jumlah flag → nilai klaim; SLA High 24 jam, Medium 3 hari, Low 7 hari kerja; tanpa skor tersembunyi | P0 |
| FR-1706 | Suppression berbasis fingerprint | Masa berlaku; approval Supervisor; flag tetap tercatat; audit | P1 |
| FR-1707 | Case group + evidence package | Klaim terkait dibungkus; evidence package memuat hash dan referensi | P1 |
| FR-1708 | Coverage matrix + residual risk | Denominator 14 tipologi in-scope; 10/14 Phase 0/1; residual risk register terisi | P0 |
| FR-1709 | Rule lifecycle + kill switch | Registrasi, aktivasi, backtest, perubahan, pensiun <30% setelah 3 bulan; kill switch per rule/global | P0 |
| FR-1710 | Deterministic backtest + label leakage test | Input & versi sama → output identik; label ground truth tidak dibaca engine; test otomatis | P0 |
| FR-1711 | Control-claim false-positive gate | 0 false flag untuk rule Must; <5% untuk Should/Could; <8% untuk H9–H10 eksperimental | P0 |
| FR-1712 | Data contract Phase 0/1 | Skema klaim/invoice/polis/provider/identity/dokumen; validasi fail-fast; identitas hash; hanya MOTOR/PROPERTY | P0 |
| FR-1713 | Retention matrix | Data demo 12 bulan; produksi per kategori `TBD` Legal; audit log immutable | P1 |
| FR-1714 | Usage metering, value report, billing export | Rekonsiliasi 100%; formula version; checksum; tanpa PII | P0 |
| FR-1715 | UI language guardrail | Dilarang "terbukti/palsu/fraud" sebagai kesimpulan; gunakan "perlu verifikasi/indikasi" | P0 |

> **Acceptance thin slice:** BRD §13 AC-1..AC-16. **AT target-state** di §18.3 tidak boleh dipakai sebagai acceptance thin slice tanpa pemetaan.

---

## 8. Kebutuhan Non-Fungsional

| ID | Kategori | Requirement | Acceptance criteria |
|---|---|---|---|
| NFR-1 | Performance | Scoring real-time dan batch | API p95 <2 detik; batch platform 1 juta klaim <4 jam; UI <3 detik |
| NFR-2 | Availability | Ketersediaan layanan | 99.9% MVP; 99.95% GA; RTO 4 jam; RPO 15 menit |
| NFR-3 | Scalability | Skala multi-tenant | 10 juta klaim/tahun per tenant (target Enterprise); 1.000 concurrent users; horizontal scaling; batas Core/Pro ditegakkan oleh entitlement |
| NFR-4 | Security | Keamanan aplikasi dan data | Kontrol selaras ISO 27001 (sertifikasi target Phase 2); OWASP ASVS L2; enkripsi; MFA; pen test tahunan. Tier 1 enterprise memerlukan ISO 27001/SOC 2 sebelum produksi atau compensating controls yang disetujui |
| NFR-5 | Privacy | UU PDP | Consent/legal basis; DSR; retensi; deletion; pseudonymization; data residency |
| NFR-6 | Explainability | Alasan alert dan keputusan | 100% alert memiliki reason code; 100% keputusan memiliki audit trail |
| NFR-7 | Fairness | Keadilan model | Atribut terlarang dikecualikan; bias test; monitoring per segmen |
| NFR-8 | Localization | Bahasa dan format Indonesia | Bahasa Indonesia utama; English toggle; IDR; WIB/WITA/WIT; NIK/NPWP; alamat |
| NFR-9 | Accessibility | Aksesibilitas | Thin slice: baseline (keyboard, kontras, label); MVP/GA: WCAG 2.1 AA |
| NFR-10 | Observability | Monitoring dan troubleshooting | Metrics; logs; traces; model monitoring; alerting |
| NFR-11 | Maintainability | Kualitas engineering | API-first; CI/CD; IaC; test coverage; dokumentasi |
| NFR-12 | Portability | Exit dan portabilitas | Export data, model, konfigurasi; tidak ada lock-in |
| NFR-13 | Compliance | Retensi dan audit | Retensi per kategori (lihat §12.6); produksi 5–10 tahun configurable `TBD` Legal; thin slice data sintetis 12 bulan; legal hold; audit log immutable |
| NFR-14 | Usability | Efisiensi investigator | Dari alert ke kasus ≤3 klik; training ≤1 hari |
| NFR-15 | Determinism | Thin slice rules-only | Input & versi rule sama → output identik; backtest 2× diff = 0 |
| NFR-16 | Reproducibility | Semua temuan dapat direproduksi | Versi rule/engine/parameter tercatat; re-analisis berversi |
| NFR-17 | Metering accuracy | Usage dapat ditagih | 100% rekonsiliasi dengan event sumber; selisih 0 |
| NFR-18 | Billing export integrity | Export siap-finance | Versioned; checksum; tanpa PII; dapat direproduksi |
| NFR-19 | Value report reproducibility | ROI dapat diaudit | Input & formula version sama → output identik; formula terdokumentasi |
| NFR-20 | Read-only claim status | Tidak ada penolakan otomatis | 100% status klaim identik dengan sumber setelah batch; tidak ada jalur write-back di thin slice |

---

## 9. Data, AI/ML, dan Tata Kelola Model

### 9.1 Sumber data

| Kategori | Contoh data | Sumber | Frekuensi | Sensitivitas | Catatan |
|---|---|---|---|---|---|
| Klaim & polis | Nomor klaim, tanggal, nilai, status, coverage | PAS/claims system | Real-time/batch | Tinggi | Data inti |
| Pihak | Tertanggung, pengemudi, ahli waris, agen, broker | PAS/CRM | Batch/real-time | Tinggi | Perlu entity resolution |
| Kendaraan/properti | Plat, STNK, BPKB, alamat, nilai properti | Claims/polis | Batch | Tinggi | Normalisasi |
| Dokumen | KTP, invoice, kuitansi, foto, rekam medis, surat kematian | DMS/upload | Event | Sangat tinggi | PII/PHI |
| Provider | Rumah sakit, klinik, bengkel, vendor | Claims/network | Batch | Tinggi | Benchmark |
| Pembayaran | Rekening, metode bayar, status | Finance | Batch | Tinggi | Fraud link |
| Telematik | GPS, perilaku mengemudi | Partner | Streaming/batch | Tinggi | Auto |
| Identitas eksternal | Dukcapil, alamat, telepon | Partner resmi | API | Sangat tinggi | Legal basis |
| Kendaraan eksternal | Samsat, riwayat kendaraan | Partner resmi | API | Tinggi | Optional |
| Perusahaan | AHU, kepemilikan | Partner resmi | API | Sedang | Commercial |
| Pengadilan/sanksi | Perkara, watchlist | Partner | API | Sedang | Due diligence |
| Geo/cuaca | Lokasi, bencana, cuaca | Partner | API | Rendah | Property |
| Referensi medis | ICD-10, prosedur, tarif | Asosiasi/provider | Batch | Sedang | Health |
| OSINT | Publik, media sosial | Terbatas | Manual/API | Tinggi | Hanya dengan legal basis |

### 9.2 Model data konseptual

| Entitas | Deskripsi | Field kunci |
|---|---|---|
| Tenant | Perusahaan asuransi | tenant_id, nama, lini, region |
| Policy | Polis | policy_id, product, line, coverage, inception, status |
| Claim | Klaim | claim_id, policy_id, date, amount, status, channel |
| ClaimItem | Item klaim | item_id, type, amount, diagnosis/procedure, provider |
| Party | Pihak | party_id, role, NIK/NPWP, name, phone, address |
| Provider | Provider/bengkel/vendor | provider_id, type, name, license, address |
| Document | Dokumen | doc_id, type, hash, extracted_fields, status |
| Alert | Alert | alert_id, claim_id, score, reasons, priority, status |
| Case | Kasus investigasi | case_id, alert_ids, owner, status, decision |
| Evidence | Bukti | evidence_id, case_id, type, source, chain_of_custody |
| Decision | Keputusan | decision_id, case_id, outcome, reason, approver, timestamp |
| Model | Model | model_id, version, line, metrics, status |
| Rule | Aturan | rule_id, version, condition, score, status |
| Label | Label/outcome | label_id, case_id, disposition, confidence |
| AuditEvent | Audit | event_id, actor, action, object, timestamp, hash |
| UsageRecord | Usage metering | period, tenant_id, claims_scored, documents_processed, flags_created, case_groups, active_users, confirmed_cases, value_protected_estimate, value_recovered |
| ReasonCode | Taksonomi alasan | code, category, label, description, active |
| Asset | Aset kendaraan/properti | asset_id, type, identifier, policy_id, attributes |
| Payment | Pembayaran klaim | payment_id, claim_id, bank_account_hash, amount, status |
| Consent | Consent/legal basis | consent_id, subject_id, purpose, source, expiry |

### 9.3 Labeling dan feedback

- **Label utama:** confirmed fraud, not fraud, inconclusive.
- **Label tambahan:** fraud typology, reason code, recovered amount, time to resolve, investigator confidence.
- **Sumber label:** keputusan investigator, hasil recovery, laporan provider, audit, dan (jika ada) data penegak hukum.
- **Kualitas label:** review sampling, adjudication untuk kasus ambigu, dan audit label.
- **Active learning:** prioritaskan kasus dengan uncertainty tinggi dan nilai tinggi.
- **Feedback loop:** label masuk ke model dan rule tuning; perubahan harus melalui approval.

### 9.4 Evaluasi model

- **Metrik utama:** PR-AUC, recall@k, precision@k, calibration, lift over rules-only.
- **Metrik operasional:** alert volume, confirmation rate, false positive rate, time-to-flag.
- **Metrik fairness:** performa per segmen (lini, wilayah, ukuran klaim, kanal, tipe nasabah), tanpa atribut terlarang.
- **Metrik stabilitas:** PSI/KS, drift, performance decay.
- **Validasi:** backtest pada data historis, shadow mode, live pilot, dan holdout.
- **Threshold:** dipilih berdasarkan kapasitas SIU dan expected loss, bukan hanya F1.

### 9.5 Tata kelola model

- **Model owner:** Head of Data Science.
- **Validasi independen:** risk/compliance atau fungsi validasi model.
- **Model card:** tujuan, data, fitur, metrik, limitasi, fairness, approval.
- **Approval:** model baru/threshold baru harus disetujui sebelum produksi.
- **Monitoring:** performa, drift, fairness, insiden.
- **Retraining:** terjadwal dan/atau trigger drift; validasi ulang.
- **Audit:** versi model, data latih, parameter, approver, dan perubahan.
- **Kill switch:** model dapat dinonaktifkan dan fallback ke rules.

### 9.6 Guardrails GenAI

- Gunakan model privat/tenant-isolated; data tenant tidak digunakan untuk training model publik.
- Redaksi PII/PHI sebelum prompt; retrieval-augmented generation dengan sitasi.
- Tolak menjawab jika bukti tidak cukup; tandai confidence.
- Pertahanan prompt injection; evaluasi hallucination; human review wajib.
- Audit setiap prompt/output yang memengaruhi kasus.
- GenAI tidak boleh membuat keputusan penolakan; hanya ringkasan, saran, dan draft.

### 9.7 Data contract Phase 0/1 (BRD v0.5)

- **Source of truth:** BRD §10.7 untuk entitas Alert, Case, Evidence, Decision, Rule, Label, AuditEvent, UsageRecord, Tenant, Consent, Asset, Payment.
- **Skema input:** klaim, invoice, polis, provider, identity map (hashed), metadata dokumen (Fase 2).
- **Validasi:** fail-fast per batch; laporan per baris; tolak re-impor identik; hanya `MOTOR`/`PROPERTY` di Phase 0/1.
- **Identitas:** hash (`sha256:`) di thin slice; produksi tokenization/vault; raw NIK dilarang di aplikasi/log.
- **Label isolation:** `is_fraud_ground_truth` hanya untuk evaluasi; rule engine tidak boleh membacanya (test otomatis).
- **Retention:** §12.6; data demo 12 bulan; produksi `TBD` Legal.

### 9.8 Cold start

- Thin slice dan commercial MVP dimulai dengan **rules + expert priors**, bukan ML.
- Dataset awal: sintetis berlabel + planted cases + kontrol; label design partner menyusul.
- ML hanya diaktifkan setelah label cukup, ada baseline rules, dan model card disetujui (Phase 2/3).
- Jangan mencampur skor ML ke prioritas thin slice; keputusan tetap deterministik.

### 9.9 Metrik: planted vs produksi

- **Planted-test recall** mengukur kemampuan rule pada dataset berlabel; bukan business lift.
- **Production lift** diukur dengan control group/holdout dan baseline tenant.
- **Value report** memakai formula version; tidak boleh double count; nilai yang tidak terverifikasi ditandai `TBD`.
- Laporkan pasangan metrik: recall@k + false positive; precision proxy + volume; coverage + denominator.

---

## 10. Integrasi dan Arsitektur

### 10.1 Prinsip arsitektur

- **API-first** dan **event-driven**.
- **Multi-tenant** dengan isolasi logis; opsi dedicated tenant untuk enterprise.
- **Cloud-native** di region Indonesia; opsi private cloud/on-prem.
- **Modular**: rules, ML, graph, document AI, case management dapat di-deploy/scale terpisah.
- **Terbuka**: export data/model/konfigurasi; tidak lock-in.
- **Aman**: zero-trust, enkripsi, audit, least privilege.

### 10.2 Komponen utama

| Komponen | Fungsi |
|---|---|
| Ingestion service | API, SFTP, batch, event; validasi; mapping |
| Data lakehouse | Penyimpanan data mentah dan terkurasi; lineage |
| Feature store | Fitur model; point-in-time correctness |
| Entity resolution service | Matching, Golden Record, relasi |
| Graph service | Graph storage, query, ring detection |
| Rules engine | Rule authoring, testing, versioning, execution |
| ML scoring service | Model serving, explainability, fallback |
| Document AI service | OCR, extraction, tamper detection, consistency |
| Alert & triage service | Queue, prioritisasi, assignment, feedback |
| Case service | Case, evidence, tasks, decisions, audit |
| Reporting service | Dashboard, export, scheduled reports |
| Admin service | Tenant, user, role, config, usage |
| Audit service | Immutable audit log, legal hold |
| LLM gateway | RAG, redaction, citations, guardrails |
| Integration service | Connectors, webhooks, write-back |

### 10.3 Alur data

```
[Sumber: PAS/Claims, Polis, Pihak, Dokumen, Provider, Data Eksternal]
        |
        v
[Ingestion: API / SFTP / Batch / Event]
        |
        v
[Data Lakehouse + Feature Store + Entity Resolution + Graph]
        |
        +--> [Rules Engine] --------+
        +--> [ML Scoring] ----------+--> [Alert & Triage] --> [Case Management] --> [Decision & Write-back]
        +--> [Document AI] ---------+                              |
        +--> [Graph Analytics] -----+                              v
                                                        [Reporting & Audit]
Cross-cutting: Security, Privacy, Observability, Model Governance, LLM Gateway
```

### 10.4 Deployment dan data residency

- **Region utama:** Indonesia (mis. Jakarta) untuk data residency.
- **Opsi:** multi-tenant SaaS; dedicated VPC; private cloud; on-prem untuk Tier 1 dengan kebutuhan khusus.
- **Isolasi:** logical tenant isolation; encryption per tenant; akses terbatas.
- **DR:** replikasi lintas availability zone; backup terenkripsi; uji restore.
- **Observability:** metrics, logs, traces, alerting, status page.

### 10.5 API dan integrasi

- REST/JSON untuk sinkron; webhook untuk event; SFTP/CSV untuk batch.
- Dukungan ACORD XML/JSON opsional untuk insurer yang menggunakannya.
- Autentikasi OAuth2/OIDC; API key untuk server-to-server; rate limit.
- Idempotency key; retry dengan backoff; dead-letter queue.
- Dokumentasi OpenAPI; sandbox; contoh kode.
- Write-back ke claims system: status, decision, notes, attachment metadata. **Phase 2, opt-in; default MVP read-only; bukan thin slice; memerlukan sign-off tenant dan review compliance.**

### 10.6 Non-goals arsitektur

- Bukan core PAS/claims system.
- Tidak memproses pembayaran.
- Tidak menyimpan seluruh arsip dokumen selamanya; retensi sesuai kebijakan.
- Tidak menggantikan DMS.

### 10.7 Profil integrasi thin slice (BRD v0.5)

| Aspek | Thin slice | Target-state PRD |
|---|---|---|
| Input | CSV/JSON upload; dokumen tidak diproses (H off) | API, SFTP, event, OCR |
| Deployment | Docker Compose lokal/offline | Cloud region Indonesia, multi-tenant, private cloud/on-prem |
| Identitas | Hash sintetis | SSO/MFA, RBAC/ABAC, tokenization/vault |
| Integrasi | Tidak ada | Claims/PAS, DMS, external data |
| Write-back | Tidak ada; status klaim read-only | Phase 2, opt-in |
| Multi-tenant | Tidak ada | Isolasi tenant |
| DSR/DPIA | N/A data sintetis | Wajib produksi |
| Observability | Log lokal + audit | Metrics, traces, alerting, status page |

> Profil ini adalah **N/A register** yang mengikat: thin slice tidak boleh mengklaim kapabilitas kolom target-state. Detail ada di §19.3.

---

## 11. UX/UI dan Lokalisasi

### 11.1 Prinsip desain

1. **Evidence-first:** setiap alert menampilkan "mengapa" sebelum "apa."
2. **Capacity-aware:** queue diurutkan berdasarkan dampak, SLA, dan kapasitas.
3. **Minimal clicks:** dari alert ke kasus ≤3 klik.
4. **Bahasa Indonesia first:** istilah bisnis lokal; English toggle.
5. **Role-based:** investigator, analis, kepala SIU, compliance melihat hal yang relevan.
6. **Accessible:** baseline thin slice; WCAG 2.1 AA untuk MVP/GA; keyboard navigation; kontras.
7. **Responsive:** desktop untuk investigasi dan semua aksi tulis; mobile MVP untuk alert inbox + case read-only + push (M-01..M-10); upload bukti dan approval via web, mobile write menyusul Phase 2/3.
8. **Flag bukan bukti:** dilarang menampilkan "terbukti", "palsu", atau "fraud" sebagai kesimpulan. Gunakan "perlu verifikasi", "indikasi", atau "periksa dokumen". Ini berlaku untuk semua layar dan ekspor.

### 11.2 Layar utama

| Layar | Tujuan | Fitur kunci |
|---|---|---|
| Login/SSO | Akses aman | SSO, MFA, tenant selection |
| Dashboard eksekutif | ROI dan risiko | Fraud protected, leakage, coverage, trend |
| Dashboard operasional | Kinerja SIU | Queue aging, cycle time, workload, SLA |
| Alert Queue | Triage | Skor, alasan, prioritas, filter, assignment |
| Case Workspace | Investigasi | Timeline, parties, docs, graph, tasks, notes, decision |
| Graph Explorer | Analisis jaringan | Search, expand, path, filter, export |
| Document Viewer | Verifikasi dokumen | Zoom, annotate, redact, extracted fields |
| Reports | Pelaporan | Regulatory, audit, portfolio, export |
| Model Monitoring | Tata kelola model | Performance, drift, fairness, versions |
| Admin | Konfigurasi | Tenant, roles, rules, thresholds, usage |
| Audit | Kepatuhan | Search, filter, export, legal hold |

### 11.3 Lokalisasi dan aksesibilitas

- Bahasa Indonesia utama; English untuk tim regional/global.
- Mata uang IDR; format angka Indonesia; zona waktu WIB/WITA/WIT.
- Format NIK 16 digit, NPWP, telepon +62, plat nomor, alamat.
- Nama Indonesia dengan variasi (nama tunggal, gelar, ejaan).
- Dokumen Indonesia: KTP, SIM, STNK, BPKB, invoice, kuitansi, rekam medis, surat kematian.
- WCAG 2.1 AA untuk MVP/GA; ukuran font; kontras; label; keyboard.

### 11.4 Inventaris layar per fase

| Layar | Tujuan | Thin slice | Target-state |
|---|---|:---:|:---:|
| Import & validation | Unggah CSV/JSON, validasi, error report | Ya | Ya |
| Review queue | Antrean flag, filter, SLA | Ya | Ya |
| Flag bundle/explanation | 4 elemen penjelasan + evidence | Ya | Ya |
| Case group | Klaim terkait, evidence package | Ya | Ya |
| Document viewer | Side-by-side, OCR/forensik | Tidak (H off) | Ya |
| Coverage matrix | Tipologi → rule → residual risk | Ya | Ya |
| Batch dashboard | Volume, flag rate, precision proxy | Ya | Ya |
| Admin | Rule, threshold, user, feature flag | Ya | Ya |
| Audit | Audit log, ekspor | Ya | Ya |
| Usage & value report | Metering, ROI, value protected/recovered | Ya | Ya |
| Billing export | Usage, plan, overage, checksum | Ya | Ya |
| Graph explorer | Analitik jaringan | Tidak (basic link only) | Ya |
| Model monitoring | Drift, fairness, model card | Tidak | Ya |

> Layar thin slice harus dapat dioperasikan tanpa API, SSO, atau integrasi. Layar target-state tidak boleh ditampilkan sebagai "sudah tersedia" pada demo thin slice.

---

## 12. Kepatuhan, Keamanan, dan Privasi

### 12.1 Peta regulasi

| Regulasi/standar | Relevansi | Implikasi produk |
|---|---|---|
| **UU No. 40 Tahun 2014 tentang Perasuransian** | Kerangka usaha, tata kelola, perlindungan pemegang polis | Keputusan klaim harus dapat dipertanggungjawabkan; audit trail |
| **UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi (UU PDP)** | Data pribadi, data sensitif, consent, hak subjek data | Privacy-by-design; DSR; breach notification; DPO; retensi |
| **UU No. 17 Tahun 2023 tentang Kesehatan** | Kerahasiaan data kesehatan | PHI masking; akses terbatas; legal basis |
| **POJK No. 22 Tahun 2023 tentang Pelindungan Konsumen dan Masyarakat di Sektor Jasa Keuangan** | Perlakuan adil, transparansi, penanganan keluhan | Tidak ada penolakan otomatis; komunikasi jelas; jalur sengketa |
| **POJK No. 8 Tahun 2024 tentang Produk Asuransi dan Saluran Pemasaran Produk Asuransi** | Produk, pemasaran, dan perlindungan pemegang polis | Transparansi produk; komunikasi; tata kelola pemasaran. Mencabut POJK 23/POJK.05/2015 |
| **POJK No. 44/POJK.05/2020 tentang Penerapan Manajemen Risiko bagi LJK Non-Bank** | Manajemen risiko, termasuk risiko model dan outsourcing | Model risk management; vendor risk; kontinuitas |
| **POJK No. 73/POJK.05/2016 tentang Tata Kelola Perusahaan yang Baik** | Akuntabilitas, transparansi, pengawasan | Approval matrix; audit; pelaporan |
| **UU No. 8 Tahun 2010 tentang APU-PPT** | Kewajiban anti pencucian uang (relevan untuk jiwa) | Integrasi watchlist/sanksi; audit |
| **UU No. 11 Tahun 2008 jo. UU No. 19 Tahun 2016 jo. UU No. 1 Tahun 2024 tentang ITE; PP No. 71 Tahun 2019** | Bukti elektronik, penyelenggaraan sistem elektronik | Audit log; integritas data; data residency |
| **Panduan/kajian OJK terkait AI dan manajemen risiko model** | Tata kelola AI | Explainability; fairness; validasi; dokumentasi |
| **AAUI/AAJI dan asosiasi industri terkait** | Praktik industri, berbagi data fraud | Kepatuhan asosiasi; data sharing opt-in |
| **ISO 27001 / SOC 2** | Keamanan dan kontrol | Sertifikasi/roadmap; audit |
| **NIST AI RMF (opsional)** | Tata kelola risiko AI | Framework validasi dan monitoring |

> **Koreksi v0.2:** POJK 23/POJK.05/2015 **bukan** regulasi perlindungan konsumen; itu regulasi produk asuransi dan pemasaran yang telah dicabut oleh POJK 8/2024. Perlindungan konsumen mengacu **POJK 22/2023**. UU ITE telah diubah kedua kali oleh **UU 1/2024**. Nomor dan interpretasi regulasi dapat berubah; Legal/Compliance wajib memvalidasi versi terbaru dan relevansinya per lini.

### 12.2 Kontrol wajib

- **Human-in-the-loop:** AI tidak boleh menolak klaim secara otomatis; keputusan akhir ada pada manusia yang berwenang.
- **Explainability:** reason code, feature contribution, evidence link untuk setiap alert dan keputusan.
- **Audit trail:** immutable; siapa, apa, kapan, mengapa; dapat diekspor.
- **Fairness:** atribut terlarang dikecualikan; bias test; monitoring per segmen; review manusia.
- **Consent/legal basis:** setiap sumber data memiliki dasar pemrosesan; consent registry.
- **Data minimization:** hanya data yang relevan; retensi terbatas.
- **DSR:** akses, koreksi, penghapusan; SLA; verifikasi identitas.
- **Breach notification:** sesuai UU PDP dan kebijakan tenant.
- **Data residency:** region Indonesia; tenant config.
- **Vendor risk:** DPA dengan tenant dan subprocessors; audit; exit plan.
- **Consumer protection:** template komunikasi; jalur keluhan; tidak ada tekanan tidak wajar.
- **Record retention:** per kategori (§12.6); produksi 5–10 tahun configurable `TBD` Legal; thin slice 12 bulan; legal hold; deletion yang dapat diaudit.

### 12.3 Perlindungan konsumen dan fairness

- Jangan gunakan atribut terlarang (mis. agama, suku, gender, disabilitas) dalam model.
- Uji dampak per segmen dan dokumentasikan.
- Berikan alasan yang dapat dipahami nasabah melalui jalur resmi.
- Pastikan klaim sah tidak tertunda tanpa alasan yang jelas.
- Sediakan mekanisme koreksi jika data yang digunakan tidak akurat.
- Libatkan Compliance dalam desain komunikasi dan escalation.

### 12.4 Keamanan

- Enkripsi at rest dan in transit; KMS; key rotation.
- RBAC/ABAC; least privilege; MFA; SSO; session management.
- Segmentasi jaringan; WAF; DDoS protection; secrets management.
- SAST/DAST; dependency scanning; pen test tahunan; bug bounty (opsional).
- Incident response runbook; tabletop; notifikasi.
- Backup terenkripsi; uji restore; DR.
- Audit akses dan perubahan konfigurasi.

### 12.5 Retensi dan hak subjek data

- Retensi default produksi: 5 tahun setelah kasus ditutup; configurable hingga 10 tahun sesuai kebijakan tenant dan kewajiban hukum. Angka final `TBD` Legal.
- **Thin slice:** data sintetis; retensi demo 12 bulan; tidak ada DSR/DPIA karena tidak ada PII nyata.
- Legal hold menghentikan penghapusan.
- DSR: verifikasi identitas; akses/koreksi/penghapusan; catat di audit log. DSR workflow adalah target-state produksi, bukan thin slice.
- Data yang dihapus harus dihapus dari sistem produksi dan backup sesuai kebijakan.
- Fairness testing memerlukan akses terbatas ke atribut terlindungi untuk pengujian; atribut tidak dipakai sebagai fitur keputusan.

### 12.6 Matriks retensi & applicability per fase

| Data | Thin slice (sintetis) | Produksi (target) | Catatan |
|---|---|---|---|
| Klaim/invoice/polis/provider | 12 bulan setelah batch ditutup | 5–10 tahun `TBD` | Legal |
| Dokumen & turunan forensik | 12 bulan | 5–10 tahun `TBD` | BRD DR-18 |
| Audit log | 12 bulan | ≥5 tahun `TBD` | Immutable |
| Usage/billing record | 24 bulan | 7 tahun `TBD` | Finance/audit |
| Label/keputusan | 12 bulan | `TBD` | Feedback model |
| PII | Hanya hash | Tokenization/vault | Raw NIK dilarang |
| Backup | 30 hari | `TBD` | Uji restore |

**Applicability:**

| Kewajiban | Thin slice | Produksi |
|---|---|---|
| UU PDP (consent, DSR, breach, DPO) | N/A — data sintetis | Wajib |
| DPIA | N/A | Wajib sebelum pilot produksi |
| Data residency | N/A — lokal | Region Indonesia |
| OJK consumer protection | N/A — demo | Wajib; bahasa "flag bukan bukti" |
| APU-PPT/watchlist | N/A | Phase 2/3 sesuai lini |
| ISO 27001/SOC 2 | N/A | Enterprise/Tier 1 |

> **Consumer protection:** keputusan klaim tetap di manusia; tidak ada penolakan otomatis; komunikasi ke nasabah tidak boleh menyebut "fraud" tanpa verifikasi. Ini adalah guardrail, bukan fitur opsional.

---

## 13. Analitik, Pelaporan, dan Eksperimen

### 13.1 Event taxonomy

| Event | Kapan | Properti kunci |
|---|---|---|
| claim_ingested | Klaim masuk | tenant, line, claim_id, source, latency |
| score_generated | Skor dibuat | model_version, score, reasons, latency |
| alert_created | Alert dibuat | alert_id, source, priority, expected_loss |
| alert_viewed | Alert dibuka | user, alert_id, duration |
| case_opened | Kasus dibuat | case_id, alert_id, owner |
| evidence_added | Bukti ditambahkan | case_id, type, source |
| decision_made | Keputusan dibuat | case_id, outcome, reason, approver |
| case_closed | Kasus ditutup | case_id, cycle_time, recovered_amount |
| feedback_submitted | Feedback investigator | case_id, label, reason |
| model_retrained | Model di-retrain | model_id, version, approver |
| report_exported | Laporan diekspor | user, report_type, format |
| claim_scored | Klaim di-score (thin slice) | batch_id, claim_id, rule_count, flag_count |
| document_processed | Dokumen diproses (Fase 2) | doc_id, engine_version, status, confidence |
| flag_created | Flag dibuat | flag_id, rule_id, severity, claim_id |
| case_group_created | Case group dibuat | group_id, claim_ids, rule_ids |
| usage_recorded | Usage ditutup | period, claims_scored, documents_processed, flags |
| value_report_generated | Value report dibuat | period, formula_version, value_protected, value_recovered |
| billing_exported | Billing export dibuat | period, plan, usage, checksum |

### 13.2 Dashboard

- **Eksekutif:** fraud protected, recovered, leakage, ROI, coverage, trend per lini.
- **Operasional:** queue aging, cycle time, workload, SLA, alert volume.
- **Model:** precision/recall, calibration, drift, fairness, override rate.
- **Portofolio:** loss ratio impact, typology, provider/agent network, region.
- **Kepatuhan:** case audit, access log, DSR, retention, incident.

### 13.3 Desain pilot

> **Thin slice vs pilot:** thin slice = validasi teknis dengan data sintetis. Pilot komersial = 2 design partner P&C, minimal 2 paid pilot; pilot fee dikreditkan ke kontrak tahunan (BRD §3A). Jangan menyebut thin slice sebagai pilot produksi.

**Fase A — Historical backtest (4–6 minggu):**
- Gunakan 12–24 bulan klaim tertutup dengan outcome.
- Ukur recall pada fixed alert volume, precision, lift, dan performa per segmen.
- Tentukan threshold awal berdasarkan kapasitas SIU.

**Fase B — Shadow mode (4–6 minggu):**
- Platform berjalan paralel tanpa memengaruhi operasi.
- Bandingkan alert platform vs alert existing.
- Ukur false positive, coverage, dan kualitas reason code.

**Fase C — Live pilot (8–12 minggu):**
- Gunakan A/B atau stepped-wedge.
- Ukur cycle time, coverage, recovery, false positive, dan kepuasan investigator.
- Review mingguan dengan design partner.

**Kriteria sukses pilot:**
- +25% deteksi fraud terkonfirmasi pada volume alert yang sama, atau target yang disepakati.
- -30% triage time.
- ≥95% coverage alert risiko tinggi.
- ≥4/5 kepuasan investigator.
- Tidak ada temuan kritis compliance/security.
- Pilot fee, conversion target, dan churn assumption disepakati sebelum pilot.
- ROI positif atau jalur jelas menuju ROI.

### 13.4 Formula ROI

```
ROI = (V_time + V_fraud + V_friction + V_audit − Total biaya) / Total biaya

V_time     = jam investigator yang dihemat × fully loaded cost
V_fraud    = fraud dicegah + fraud dipulihkan (diverifikasi dengan control group)
V_friction = false positive yang dihindari × cost per false positive
V_audit    = temuan audit yang dihindari × cost per finding
Total biaya = subscription + implementasi + biaya data + effort internal

Customer ROI = (V_time + V_fraud + V_friction + V_audit − Price) / Price
```

**Aturan:**
- Gunakan holdout/control group untuk mengukur fraud dicegah; jangan mengklaim seluruh penurunan klaim sebagai dampak platform.
- Jangan double count: satu klaim hanya boleh masuk satu komponen nilai.
- Value report memakai formula version; nilai yang belum terverifikasi ditandai `TBD`.
- **Planted-test recall bukan ROI.** Laporkan secara terpisah dari production lift.

### 13.5 Metering, value report, dan billing

- **Usage record:** klaim di-score, dokumen diproses, flag, case group, reviewer aktif, per periode; immutable.
- **Value report:** confirmed cases, value protected/recovered, time saved, formula version, confidence; tanpa PII; tanpa double counting.
- **Billing export:** periode, plan, usage, overage, unit price, checksum; rekonsiliasi 100% dengan usage record.
- **Pilot scorecard:** baseline vs pilot, conversion recommendation, sign-off Supervisor/Admin.
- Formula revenue dan contoh placeholder ada di BRD §3A; angka final `TBD` sampai validasi design partner.

---

## 14. Model Komersial dan GTM

### 14.1 Packaging

| Paket | Target | Cakupan | Batas indikatif |
|---|---|---|---|
| **Core** | Mid-market P&C | P&C auto/properti; rules + baseline ML; alert & case; dashboard; 10 users | ≤100K klaim/tahun |
| **Pro** | Multi-line insurer | Semua lini; graph; document forensics; external connectors; model tuning; 50 users | ≤1 juta klaim/tahun |
| **Enterprise** | Tier 1 / group | Unlimited; private cloud/on-prem; custom model; consortium; SLA 99.95%; dedicated support | Custom |

> **Thin slice bukan paket komersial.** Thin slice adalah validasi teknis; pilot komersial memakai fixed fee yang dikreditkan ke kontrak tahunan. Jangan menjual thin slice sebagai Core/Pro.

### 14.2 Prinsip pricing

- **Platform fee + per-claim scoring + document add-on + implementation fee.**
- **Pilot fixed fee**, dikreditkan ke kontrak tahunan jika lanjut.
- **Hindari per-alert pricing** yang menghukum deteksi; per-claim scoring dapat diterima.
- **Outcome-based upside/success fee** opsional (shared savings) untuk tenant matang; wajib capped, diverifikasi dengan control group, dan disepakati sebelum pilot.
- **Transparansi overage** dan biaya data eksternal; entitlement memberi peringatan 80%/100%, bukan silent block.
- **Harga = min(value_ceiling, max(cost_floor, competitive_anchor))**; `value_ceiling` indikatif 10–20% dari quantified annual value, `cost_floor` = COGS + target gross margin. Semua angka `TBD`.
- **Usage metering adalah product requirement** (FR-1204, FR-1207–FR-1209), bukan spreadsheet manual.
- **Harga indikatif** harus divalidasi dengan riset willingness-to-pay; jangan mengutip angka final sebelum pilot.

### 14.3 GTM

- **Design partner**: 2 P&C untuk MVP; 1 jiwa + 1 kesehatan untuk Phase 2.
- **Channel:** AAUI/AAJI, reasuransi, broker, system integrator, cloud partner.
- **Sales motion:** pilot → production; ROI case study; reference selling.
- **Customer success:** onboarding, training, quarterly model review, executive business review.
- **Kepatuhan:** DPA, security questionnaire, audit support, model documentation.

### 14.4 Kriteria sukses komersial

- 2 design partner MVP bertransisi ke kontrak berbayar.
- 4–6 pelanggan pada akhir Phase 2.
- 10+ pelanggan pada akhir Phase 3.
- Net revenue retention ≥110%.
- Gross margin SaaS ≥70% (setelah biaya cloud/data).
- CAC payback ≤18 bulan.
- LTV/CAC ≥3.
- Pilot conversion ≥50% (target `TBD`, divalidasi pilot).
- Rekonsiliasi usage/billing 100%; tidak ada double count.

---

## 15. Rencana Rilis dan Milestone

### 15.0 Phase map dengan BRD v0.5

| BRD v0.5 | PRD | Scope | Catatan |
|---|---|---|---|
| Fase 1 — Hackathon MVP | Phase 0/1 thin slice | Rules A–G, CSV/JSON, queue, evidence, coverage, audit, metering | Tanpa ML/OCR/integrasi/write-back |
| Fase 2 — 1–3 bulan | Phase 1b/2 | OCR & forensik H1–H7, identity/graph minimal, pilot | Chain of custody |
| Fase 3 — 3–6 bulan | Phase 2/3 | ML scoring, CV forensics, SIU case management, external data | Model governance |

> Pemetaan ini bukan 1:1; lihat §0A.2.

| Fase | Timeline | Scope | Exit criteria |
|---|---|---|---|
| **Phase 0 — Discovery** | M0–M2 | Design partner, data profiling, workflow mapping, compliance assessment, arsitektur, metrik | 2 pilot agreement; data contract; backlog prioritas; arsitektur disetujui |
| **Phase 1 — MVP** | M2–M6 | Shared core + P&C auto/properti; 1 integrasi claims; rules + baseline ML; case management; dashboard; audit | MVP acceptance; security review; pilot readiness |
| **Phase 1b — Pilot** | M6–M9 | Live pilot 2 insurer; shadow mode; feedback | Kriteria pilot tercapai; GA readiness |
| **GA gate** | M9 | DR, 99.95%, security, mobile MASVS L2 | 2 paid pilot; HA/DR drill; GA sign-off |
| **Phase 2 — Multi-line** | M9–M15 | Jiwa & kesehatan; provider analytics; graph lanjutan; document forensics; external connectors | 2 pelanggan jiwa/kesehatan live; target model; compliance review |
| **Phase 3 — Intelligence** | M15–M24 | GenAI assistant; agentic evidence; consortium; ASEAN; outcome-based pricing | 10+ pelanggan; ARR target; retention; SOC 2 Type II |

### 15.1 Tim inti

| Peran | Jumlah | Fokus |
|---|---:|---|
| Product Manager | 1 | Prioritas, roadmap, stakeholder |
| Product Designer | 1 | UX investigasi, design system |
| Frontend Engineer | 2 | Case workspace, dashboard, graph |
| Mobile Engineer | 2 | iOS/Android native, KMP shared, store release |
| Backend Engineer | 3 | Ingestion, rules, case, API |
| Data Engineer | 2 | Pipeline, lakehouse, feature store |
| ML Engineer | 2 | Model, explainability, monitoring |
| DevOps/SRE | 1 | Infra, security, reliability |
| QA Engineer | 1 | Test automation, acceptance |
| Compliance/Legal | 0.5 | Regulasi, DPA, audit |
| Customer Success | 1 | Pilot, training, feedback |
| Sales | 1 | Design partner, pipeline |

### 15.2 Dependensi

- Akses data historis dan label dari design partner.
- API/batch dari claims/PAS.
- Perjanjian data eksternal (identitas, kendaraan, provider).
- Cloud region Indonesia dan persetujuan data residency.
- Review Legal/Compliance untuk UU PDP, OJK, dan kontrak.
- Kapasitas labeling dari tim SIU.
- Sertifikasi keamanan (ISO 27001/SOC 2) untuk enterprise.

### 15.3 Kesiapan eksekusi & RACI

**Keputusan v0.2 (rev. 3):** commercial MVP memiliki ~53 FR P0 dan 16 epic; target **13 engineer (11 core + 2 mobile)**, MVP build **M2–M6**, pilot **M6–M9**, dan **GA gate M9**. Risiko scope tetap dipantau; fallback jika headcount tidak tersedia adalah GA M15 atau pemangkasan P0. **Thin slice** (BRD v0.5) tetap subset yang divalidasi lebih dulu.

**Tim thin slice (indikatif):**

| Peran | Jumlah | Fokus |
|---|---:|---|
| Product Manager | 1 | Scope, prioritas, stakeholder |
| Tech Lead | 1 | Arsitektur, review, NFR |
| Backend Engineer | 2 | Ingestion, rules, queue, audit |
| Frontend Engineer | 1 | Queue, explanation, dashboard |
| Data Engineer | 1 | Dataset, validasi, metering |
| QA Engineer | 1 | Acceptance, determinism, pentest |
| Compliance/Legal | 0.5 | Regulasi, retensi, bahasa UI |
| Reviewer SME | 1 | Validasi rule & reason code |

**RACI ringkas:** Product (A/R untuk roadmap), Engineering (R untuk NFR), Data Science (R untuk rule/dataset), Compliance (C/A untuk legal), QA (R untuk acceptance), Finance (C untuk pricing/metering). Budget thin slice `TBD`.

---

## 16. Risiko dan Mitigasi

### 16.1 Risiko proyek

| Risiko | Likelihood | Impact | Mitigasi | Owner |
|---|---|---|---|---|
| Data tidak lengkap/kotor | Tinggi | Tinggi | Data profiling; minimum data contract; data quality dashboard; phased integration | Data Lead |
| Integrasi claims system kompleks | Tinggi | Tinggi | Konektor standar; middleware; professional services; pilot terbatas | CTO |
| False positive/alert fatigue | Tinggi | Tinggi | Capacity-aware threshold; feedback loop; explainability; tuning | Head of Data Science |
| Model bias/fairness | Sedang | Tinggi | Atribut terlarang; bias test; monitoring; human review | Compliance + Data Science |
| Ketidakpastian regulasi | Sedang | Tinggi | Legal review; compliance-by-design; auditability; regulatory watch | Legal/Compliance |
| Hallucination/privasi LLM | Sedang | Tinggi | RAG + citations; PII redaction; no training; human review; eval | CTO + Compliance |
| Adopsi SIU/change management | Tinggi | Tinggi | Design partner; training; champion; workflow alignment; quick wins | Customer Success |
| Governance data konsorsium | Tinggi | Sedang | Opt-in; anonymization; legal agreement; anti-competitive review | Legal + Product |
| Security breach | Sedang | Tinggi | ISO 27001; pen test; IR runbook; encryption; cyber insurance | CISO/SRE |
| Scope creep multi-line | Tinggi | Sedang | Shared core; line-by-line rollout; stage gates | Product |
| Vendor lock-in/build vs buy | Sedang | Sedang | Open API; export; modular; build-partner decision | CTO |
| Key person dependency | Sedang | Sedang | Dokumentasi; cross-training; code review | Engineering Lead |

### 16.2 Residual risk register (BRD §16.2)

| Fraud yang tidak terjangkau penuh | Alasan | Kontrol | Rencana |
|---|---|---|---|
| Forgery berkualitas tinggi | Butuh forensik tingkat lanjut | H1/H2/H4 menangkap sebagian | Fase 3: model CV + spesialis |
| Foto ulang layar | Metadata/kompresi baru | Cross-check OCR | Catat batas; edukasi reviewer |
| Copy-move/splicing halus | ELA eksperimental | Tidak jadi dasar prioritas | Fase 3: model CV |
| Staged accident/arson | Butuh investigasi lapangan | Tidak ada | SIU + data eksternal |
| Fraud internal | Butuh audit sistem internal | Tidak ada | Fase 3 |
| Non-disclosure underwriting | Di luar siklus klaim | Tidak ada | Underwriting analytics terpisah |
| Modus baru | Rule retrospektif | Review coverage bulanan | ML dari label reviewer |
| Kolusi tanpa atribut bersama | Graph buta | D-family parsial | Perkaya graf |
| Cross-insurer fraud | Butuh data sharing | Tidak ada | Konsorsium Fase 3 |

### 16.3 Risiko revenue (BRD §3A.9)

| Risiko | Mitigasi | Owner |
|---|---|---|
| Atribusi fraud dicegah diperdebatkan | Baseline + control group + formula disepakati | Product + Finance |
| Budget owner tidak jelas | Multi-stakeholder mapping; sponsor eksekutif | Sales |
| Sensitivitas harga | Tiered packaging; pilot credit; value proof | Product |
| Biaya OCR/dokumen membengkak | Kuota; confidence routing; local OCR | Engineering |
| Pilot tidak konversi | Scope pilot tanpa integrasi; success criteria | Product |
| Churn setelah pilot | Onboarding; training; quarterly value review | Customer Success |

### 16.4 Kill switch

- Rule: global dan per-rule, dapat dinonaktifkan Admin tanpa deploy; audit.
- Document pipeline: `doc_forensics_enabled`; default off di thin slice.
- Model: fallback ke rules; model card dan approval.
- Tidak ada kill switch yang menghapus audit trail.

---

## 17. Asumsi, Ketergantungan, dan Open Questions

### 17.1 Asumsi

1. 2 design partner P&C bersedia memberikan data historis, label, dan akses shadow mode.
2. Minimal satu claims system memiliki API atau batch extract yang dapat diintegrasikan dalam MVP.
3. Insurer menerima prinsip human-in-the-loop dan tidak menuntut penolakan otomatis.
4. Cloud region Indonesia tersedia dan memenuhi kebutuhan data residency.
5. Akses data eksternal (Dukcapil, Samsat, AAUI, provider) mungkin terbatas; platform harus connector-agnostic dan mendukung CSV/manual.
6. Bahasa Indonesia adalah bahasa utama UI; English sebagai opsi.
7. Multi-line dirollout bertahap; P&C terlebih dahulu.
8. Pricing willingness-to-pay akan divalidasi pada pilot, bukan diasumsikan.
9. Thin slice menggunakan data sintetis; tidak ada PII nyata.
10. BRD v0.5 adalah Phase 0/1 thin slice dari PRD ini; jika stakeholder memutuskan produk terpisah, decision log §19.4 diperbarui.
11. Retensi dan skema tokenization final ditetapkan Legal sebelum produksi.
12. Pilot economics (fee, konversi, churn, cost to serve) divalidasi dengan design partner.

### 17.2 Ketergantungan

- Design partner dan data contract.
- Integrasi claims/PAS.
- Perjanjian data eksternal.
- Legal/Compliance review.
- Cloud/infra dan keamanan.
- Tim labeling/SIU.

### 17.3 Open questions

1. Siapa 2–3 design partner pertama, dan apa komitmen mereka?
2. Claims/PAS apa yang harus didukung pertama (vendor lokal, Guidewire, Duck Creek, custom)?
3. Dasar hukum/consent untuk data eksternal dan data kesehatan?
4. Apakah Tier 1 membutuhkan on-prem/private cloud?
5. Lini mana setelah P&C: jiwa atau kesehatan?
6. Build vs partner untuk document AI dan LLM?
7. Apakah insurer bersedia berbagi data konsorsium? Bagaimana governance-nya?
8. Berapa periode retensi yang disetujui Legal?
9. Bagaimana interpretasi OJK atas AI dalam keputusan klaim?
10. Apa threshold sukses pilot yang disepakati?
11. Model pricing dan ACV yang dapat diterima pasar?
12. Apakah perlu dukungan bahasa Inggris untuk group insurer regional?
13. Apakah BRD v0.5 adalah thin slice PRD ini atau produk terpisah?
14. Siapa target tenant pertama dan berapa volume klaim/tahun?
15. Berapa range willingness-to-pay dan harga pilot?
16. Berapa cost to serve per klaim dan per dokumen?
17. Berapa durasi pilot, conversion target, dan churn assumption?
18. Apakah success fee dipakai? Bagaimana verifikasinya?
19. Skema tokenization NIK dan retensi final yang disetujui Legal?
20. Kapan health masuk roadmap, dan apakah TRAVEL dihapus dari enum Phase 0/1?

---

## 18. Lampiran

### 18.1 Glosarium

| Istilah | Definisi |
|---|---|
| **FNOL** | First Notice of Loss; pemberitahuan pertama klaim |
| **SIU** | Unit investigasi khusus/fraud di insurer |
| **PAS** | Policy Administration System |
| **P&C** | Property and Casualty |
| **TPA** | Third-Party Administrator |
| **Leakage** | Kebocoran nilai klaim akibat fraud, error, atau inefisiensi |
| **Alert** | Sinyal risiko yang dihasilkan rules/model/graph/document |
| **Case** | Kasus investigasi yang menautkan alert, bukti, dan keputusan |
| **Golden Record** | Representasi tunggal entitas setelah entity resolution |
| **Entity resolution** | Proses mencocokkan dan menggabungkan entitas yang sama |
| **Graph analytics** | Analisis relasi antar entitas (pihak, klaim, provider, kendaraan) |
| **Upcoding** | Penggunaan kode diagnosis/prosedur yang lebih mahal dari seharusnya |
| **Phantom billing** | Penagihan layanan yang tidak diberikan |
| **Non-disclosure** | Tidak mengungkapkan informasi material saat underwriting |
| **DPA** | Data Processing Agreement |
| **DSR** | Data Subject Request |
| **PHI** | Protected Health Information |
| **PII** | Personally Identifiable Information |
| **PSI** | Population Stability Index; metrik drift |
| **SHAP** | Metode explainability untuk kontribusi fitur |
| **RAG** | Retrieval-Augmented Generation |
| **OJK** | Otoritas Jasa Keuangan |
| **AAUI** | Asosiasi Asuransi Umum Indonesia |
| **AAJI** | Asosiasi Asuransi Jiwa Indonesia |
| **Flag** | Penanda klaim cocok dengan rule; bukan tuduhan, bukan bukti |
| **Flag bundle** | Semua flag pada satu klaim; severity = tertinggi |
| **Case group** | Kumpulan klaim terkait di thin slice; bukan investigation case PRD |
| **Reason code** | Kode terstruktur untuk keputusan/review outcome |
| **Coverage matrix** | Pemetaan tipologi fraud → rule → status cakupan |
| **Residual risk** | Fraud yang tidak terjangkau sistem beserta alasannya |
| **Precision proxy** | `CONFIRMED / (CONFIRMED + CLEARED)` per rule |
| **Thin slice** | Implementasi Phase 0/1 rules-only, data sintetis, P&C, tanpa ML/OCR/integrasi |
| **Source of truth** | Dokumen pemilik definitif untuk suatu domain |
| **Traceability matrix** | Peta item BRD ↔ PRD ↔ status ↔ fase |
| **N/A register** | Daftar kapabilitas yang sengaja tidak berlaku di thin slice |
| **ARR** | Annual Recurring Revenue |
| **CAC** | Customer Acquisition Cost |
| **LTV** | Lifetime Value |
| **WTP** | Willingness to Pay |
| **Value protected** | Nilai klaim berisiko yang tidak dibayar/dikurangi |
| **Value recovered** | Recovery dari investigasi |
| **Pilot fee** | Biaya pilot tetap; dikreditkan bila lanjut |
| **Success fee** | Persentase nilai fraud terverifikasi; opsional dan capped |

### 18.2 Skema data contoh

| Entitas | Field wajib | Field opsional | Catatan |
|---|---|---|---|
| Claim | claim_id, policy_id, line, date, amount, status | channel, cause, location | Line-specific extension |
| Policy | policy_id, product, line, inception, expiry, coverage | agent_id, branch | — |
| Party | party_id, role, name, NIK/NPWP | phone, email, address | Masking |
| Provider | provider_id, type, name, license | address, specialty | Health/auto |
| Document | doc_id, type, hash, status | extracted_fields, confidence | PII/PHI |
| Alert | alert_id, claim_id, score, reasons, priority | expected_loss, model_version | Explainability |
| Case | case_id, alert_ids, owner, status | decision, recovered_amount | Audit |
| Evidence | evidence_id, case_id, type, source | chain_of_custody | Audit |
| Decision | decision_id, case_id, outcome, reason | approver, timestamp | No auto-denial |
| Label | label_id, case_id, disposition | confidence, reviewer | Feedback |
| UsageRecord | period, tenant_id, claims_scored, documents_processed, flags_created, case_groups, active_users | confirmed_cases, value_protected_estimate, value_recovered | Tanpa PII; billing |
| ReasonCode | code, category, label | description, active | FP/NI/CF |
| Asset | asset_id, type, identifier, policy_id | attributes | D7/B5 |
| Payment | payment_id, claim_id, bank_account_hash, amount | status | N/A thin slice |

> **Data contract Phase 0/1:** BRD §10.7 adalah source of truth untuk entitas Alert, Case, Evidence, Decision, Rule, Label, AuditEvent, UsageRecord, Tenant, Consent, Asset, Payment. Skema di atas adalah target-state; jangan mencampur field thin slice dengan field target-state tanpa mapping.

### 18.3 Skenario uji penerimaan

| ID | Skenario | Expected result |
|---|---|---|
| AT-01 | Klaim duplikat lintas polis dalam 30 hari | Alert dibuat; kedua klaim tertaut; reason code; bukti |
| AT-02 | Ring staged accident dengan telepon/alamat/rekening sama | Graph alert; klaster; kasus jaringan |
| AT-03 | Klaim jiwa dengan indikasi non-disclosure | Alert; timeline polis; rekomendasi investigasi |
| AT-04 | Provider health dengan pola upcoding | Provider alert; peer benchmark; kasus provider |
| AT-05 | Klaim sah yang salah ditandai | Investigator override; label; feedback tersimpan |
| AT-06 | Invoice properti yang dimodifikasi | Document forensics alert; bukti; confidence |
| AT-07 | Investigator mencoba melihat tenant lain | Akses ditolak; audit log |
| AT-08 | Data residency | Data tersimpan di region Indonesia; verifikasi |
| AT-09 | Drift model | Alert drift; workflow retraining; approval |
| AT-10 | Upaya penolakan otomatis | Sistem memblokir; meminta approval manusia |
| AT-11 | Audit rekonstruksi keputusan | Semua langkah, bukti, dan approver terlihat |
| AT-12 | Load test 1 juta klaim/hari | Batch selesai <4 jam; tidak ada data loss |
| AT-13 | Metering satu batch | Usage record rekonsiliasi 100% dengan event sumber; selisih 0 |
| AT-14 | Value report | Setiap CONFIRMED case punya value fields/formula version; ekspor tanpa PII |
| AT-15 | Entitlement overage | Peringatan 80%/100%; tidak ada silent block; audit |
| AT-16 | Billing export | Checksum; rekonsiliasi dengan metering; siap finance |
| AT-17 | Chain of custody dokumen (Fase 2) | Hash asli = hash di evidence package; versi engine tercatat |
| AT-18 | OCR gagal (Fase 2) | Confidence <60% dirutekan manual; tidak pernah auto-clear |

**Mapping acceptance BRD → PRD:**

| BRD | PRD | Status |
|---|---|---|
| AC-1..AC-8 | AT-01..AT-12 (sebagian) | Partial — perlu mapping rinci |
| AC-9..AC-12 | AT-06, AT-11, AT-17, AT-18 | Partial — forensik target-state |
| AC-13..AC-16 | AT-13..AT-16 | Match |
| AC-1..AC-16 | Thin slice acceptance | **BRD §13 adalah acuan thin slice**, bukan AT di atas |

### 18.4 Referensi

**Internal:**

- [Market & Vendor Landscape 2026](./insurance-fraud-detection-market-landscape-2026.md)
- [BRD Insurance Claim Fraud Detection v0.5](./BRD-Insurance-Claim-Fraud-Detection-v0.5.md)
- [TRD Platform Deteksi Fraud Asuransi v0.1](./TRD-Platform-Deteksi-Fraud-Asuransi-2026.md)

**Regulasi dan standar:**

- Otoritas Jasa Keuangan (OJK): https://www.ojk.go.id
- Asosiasi Asuransi Umum Indonesia (AAUI): https://www.aaui.org
- Asosiasi Asuransi Jiwa Indonesia (AAJI): https://www.aaji.or.id
- UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi (akses via JDIH/BPK)
- UU No. 40 Tahun 2014 tentang Perasuransian (akses via JDIH/BPK)
- UU No. 17 Tahun 2023 tentang Kesehatan (akses via JDIH/BPK)
- POJK No. 22 Tahun 2023 tentang Pelindungan Konsumen dan Masyarakat di Sektor Jasa Keuangan
- POJK No. 8 Tahun 2024 tentang Produk Asuransi dan Saluran Pemasaran Produk Asuransi (mencabut POJK 23/POJK.05/2015)
- POJK No. 44/POJK.05/2020 tentang Penerapan Manajemen Risiko bagi LJK Non-Bank
- POJK No. 73/POJK.05/2016 tentang Tata Kelola Perusahaan yang Baik bagi Perusahaan Perasuransian
- UU No. 8 Tahun 2010 tentang Pencegahan dan Pemberantasan Tindak Pidana Pencucian Uang
- UU No. 11 Tahun 2008 jo. UU No. 19 Tahun 2016 jo. UU No. 1 Tahun 2024 tentang Informasi dan Transaksi Elektronik
- PP No. 71 Tahun 2019 tentang Penyelenggaraan Sistem dan Transaksi Elektronik

**Standar teknis:**

- ISO/IEC 27001: https://www.iso.org/standard/27001
- OWASP ASVS: https://owasp.org/www-project-application-security-verification-standard/
- WCAG 2.1: https://www.w3.org/WAI/standards-guidelines/wcag/
- NIST AI Risk Management Framework: https://www.nist.gov/itl/ai-risk-management-framework

---

## 19. Sync Register & Traceability (BRD v0.5)

### 19.1 Hubungan dokumen

- **Default v0.2:** PRD ini adalah target-state; BRD v0.5 adalah Phase 0/1 thin slice.
- **Thin slice** = rules-only, data sintetis, P&C auto/properti, CSV/JSON, tanpa ML/OCR/integrasi/write-back.
- **Commercial MVP** = P&C, rules + baseline ML, OCR basic, 1 integrasi, case management, dashboard, audit, mobile read-only + push (M2–M9).
- Jika stakeholder memutuskan BRD adalah produk terpisah, status register menjadi *reference only* dan decision log diperbarui. Status: `TBD`.

### 19.2 Traceability matrix

| BRD v0.5 | PRD | Status | Aksi |
|---|---|---|---|
| §3A business case | §13.4, §13.5, §14 | Partial | PRD menjadi acuan pricing; BRD acuan pilot economics |
| §7 rule catalog | E3, E6, E17 | Partial | Katalog 43 rule diadopsi di thin slice |
| §8 FR-01..30 | E1, E3, E7, E8, E10, E12, E17 | Partial | Map FR ke epic; tandai N/A |
| §8.2 DR-01..18 | E6 | Partial | Detail forensik tetap di BRD/Fase 2 |
| §8.3 FR-31..35 | E12 FR-1204/1207–1210 | Match | Usage, value report, entitlement, billing |
| §9 NFR-01..24 | §8 NFR-1..20 | Partial | Determinism, reproducibility, metering |
| §10 data contract | §9.2, §9.7, §18.2 | Partial | Field-level mapping `TBD` |
| §11 RBAC | FR-1202/1301, §4 | Partial | 3 peran thin slice vs 10 persona target |
| §12 metrics/coverage | §3.3, §13.5 | Partial | Denominator & planted vs production |
| §13 AC | §18.3 AT + mapping | Partial | AC-1..8 mapping rinci `TBD` |
| §15 governance | §12 | Partial | Retensi/tokenization `TBD` Legal |
| §16.2 residual risk | §16.2 | Match | Register diadopsi |
| §16.3 roadmap | §15.0 | Partial | Phase map |
| §17 milestones | §15 | Partial | Milestone map |
| §19 open questions | §17.3 | Partial | Decision log |
| TRD v0.1 | §8 NFR-1..20, §18.3 AT, §23 | Match | Implementasi teknis; traceability NFR/acceptance di TRD §23; mobile M-01..M-10 di §6.3.1 |
| §20 glossary | §18.1 | Partial | Glossary digabung |

### 19.3 N/A register thin slice

| Kapabilitas PRD | Status thin slice | Referensi PRD |
|---|---|---|
| Multi-tenant, SSO/MFA | N/A | E12, E13 |
| API/SFTP/real-time | N/A — CSV/JSON upload | E1 |
| Entity resolution/Golden Record | N/A — hash matching saja | E2 |
| Graph analytics/explorer | N/A — case grouping saja | E5 |
| Case management workspace | N/A — evidence package saja | E8 |
| ML scoring/model governance | N/A sampai Phase 2/3 | E4, §9 |
| GenAI assistant | N/A | E15 |
| External data/consortium | N/A | E14 |
| Notifications/collaboration | N/A | E16 |
| Mobile MVP (M-01..M-10) | N/A thin slice — MVP read-only + push | §6.3.1 |
| Underwriting fraud | N/A | J5 |
| DSR/DPIA/data residency | N/A — data sintetis | §12 |
| Availability/DR/observability | N/A — lokal | §8 |
| Localization/WCAG | Lokalisasi N/A (IDR only); sertifikasi WCAG N/A — baseline aksesibilitas tetap wajib | §11 |
| Commercial GTM | N/A — pilot economics saja | §14 |
| ISO/SOC 2/vendor risk | N/A | §12 |

### 19.4 Decision log

| ID | Keputusan | Owner | Status | Review |
|---|---|---|---|---|
| D-01 | ML Phase 2/3; thin slice rules-only | Product + Data Science | Default v0.2 | `TBD` |
| D-02 | Thin slice read-only; write-back Phase 2 opt-in | Product + Legal | Default v0.2 | `TBD` |
| D-03 | Retensi per kategori; produksi `TBD` Legal | Compliance | Default v0.2 | `TBD` |
| D-04 | Tokenization/vault produksi; raw NIK dilarang | CTO + Compliance | Default v0.2 | `TBD` |
| D-05 | P&C auto/properti MVP; health/life Phase 2 | Product | Default v0.2 | `TBD` |
| D-06 | Platform + per-claim + implementation + pilot credit; no per-alert | Product + Finance | Default v0.2 | `TBD` |
| D-07 | Prioritas deterministik thin slice; expected loss Phase 2 | Product | Default v0.2 | `TBD` |
| D-08 | Pisahkan case group/review outcome dari investigation case/claim decision | Product | Default v0.2 | `TBD` |
| D-09 | Mobile MVP read-only + push; aksi tulis hanya di web | Product + Eng | Keputusan v0.2 (rev. 3) | M9 |
| D-10 | GA gate M9; tim 13 engineer (11 core + 2 mobile) | Product + Finance | Keputusan v0.2 (rev. 3) | M9 |

### 19.5 BRD items yang diadopsi ke PRD

1. Katalog rule + mapping tipologi (E3/E17).
2. Coverage denominator & residual risk (§3.3, §16.2).
3. Control-claim false-positive gates + rule retirement (E3/E17).
4. Label leakage test + deterministic backtest (NFR-15/NFR-16, §9.7).
5. UI language guardrail "flag bukan bukti" (§11.1).
6. Data contract + data quality rules (§9.7, §18.2).
7. Document forensics thresholds + chain of custody (E6, Fase 2).
8. Adversarial/red-team scenarios (BRD §12.3; QA).
9. Usage metering/value report/billing export (E12).
10. Revenue formula + unit economics placeholders (§13.4, §14; BRD §3A).

### 19.6 Open sync issues

1. Apakah BRD thin slice bagian dari PRD atau produk terpisah? `TBD`
2. Retensi produksi final per kategori. `TBD` Legal
3. Skema tokenization NIK. `TBD` Legal/CTO
4. Harga, WTP, pilot fee, cost to serve. `TBD` pilot
5. Konversi pilot, churn, CAC. `TBD` pilot
6. Mapping rinci AC-1..AC-8 → AT-01..AT-12. `TBD` QA
7. Mapping field-level data contract thin slice ↔ target-state. `TBD` Data
8. Timeline health & enum TRAVEL. `TBD` Product
9. On-prem vs gross margin ≥70%. `TBD` Finance/CTO
10. Success fee: dipakai atau tidak, formula verifikasi. `TBD` Product/Finance

### 19.7 Change control

- Setiap perubahan yang memengaruhi dokumen lain harus dicatat di changelog kedua dokumen.
- Satu change set memperbarui PRD, BRD, dan decision log.
- Review sync dilakukan setiap akhir fase dan sebelum pilot.
- Dokumen dianggap sinkron hanya jika §0A.5 terpenuhi.

---

**Akhir dokumen.**  
Dokumen ini akan diperbarui setelah review design partner, validasi Legal/Compliance, dan hasil pilot. Angka revenue, retensi, tokenization, dan keputusan konflik bertanda `TBD`/default dan belum final.

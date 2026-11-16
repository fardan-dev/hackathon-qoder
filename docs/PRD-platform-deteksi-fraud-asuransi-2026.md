# PRD — Platform Deteksi Fraud Asuransi Multi-Lini

**Nama kerja:** Sentinel Fraud Platform (placeholder)  
**Versi:** 0.1 — Draft untuk review  
**Tanggal:** 16 November 2026  
**Pemilik:** Product Management  
**Status:** Draft  
**Pasar awal:** Indonesia (OJK)  
**Lini produk:** P&C, jiwa, dan kesehatan  
**Model bisnis:** B2B SaaS multi-tenant, dengan opsi private cloud/on-prem untuk enterprise

**Reviewer wajib:** Head of Claims/SIU, Chief Compliance Officer, Legal, CTO, Head of Data Science, UX Lead, Head of Customer Success.

> Dokumen ini adalah product requirements document, bukan nasihat hukum. Seluruh interpretasi regulasi wajib divalidasi oleh Legal/Compliance sebelum implementasi.

---

## Daftar Isi

0. Ringkasan Eksekutif
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
| Deteksi | Recall pada fixed alert volume | Baseline rules | +25% |
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

> Catatan: baseline “benchmark vendor” bersifat indikatif dan tidak menggantikan pengukuran baseline tenant. Semua target harus dikunci bersama design partner sebelum pilot.

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
| **Provider Network Manager (kesehatan)** | Mengelola risiko provider | Tidak ada benchmark billing; audit manual | Provider peer analytics, case management | Provider anomali turun |
| **Auditor/Regulator (read-only)** | Memeriksa keputusan dan bukti | Tidak bisa merekonstruksi keputusan | Read-only export, audit trail, evidence index | Pemeriksaan lebih cepat |

---

## 5. Alur Kerja Utama

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

- Shared core: ingestion, entity resolution, rules, ML scoring, graph, document AI, alert triage, case management, reporting, model governance, admin, security, audit.
- P&C auto dan properti pada MVP.
- Modul jiwa dan kesehatan pada Phase 2.
- Integrasi dengan claims/PAS, dokumen, dan data eksternal.
- Dukungan Bahasa Indonesia, IDR, format NIK/NPWP, dan data residency Indonesia.
- Kepatuhan UU PDP, OJK, dan praktik perlindungan konsumen.

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

### 6.3 MVP vs Phase 2 vs Phase 3

| Area | MVP (M0–M6 build; M6–M9 pilot) | Phase 2 (M9–M15) | Phase 3 (M15–M24) |
|---|---|---|---|
| Lini | P&C auto & properti | Jiwa & kesehatan | Multi-negara ASEAN |
| Deteksi | Rules + baseline ML | Model per lini, provider analytics | Ensemble, adaptive models |
| Graph | Basic entity graph & link alerts | Ring detection, community detection, graph explorer lanjutan | Cross-tenant consortium graph (opt-in) |
| Dokumen | OCR basic, duplicate detection | Forensik dokumen, tamper detection, cross-document consistency | Agentic evidence gathering |
| Case management | End-to-end untuk P&C | Template jiwa/kesehatan, provider cases | Orkestrasi multi-pihak |
| Integrasi | 1 claims system + CSV/SFTP | Konektor eksternal (identitas, kendaraan, provider, cuaca) | Ecosystem marketplace |
| AI | Scoring + reason codes | Model tuning UI, champion/challenger | GenAI assistant, autonomous investigation support |
| Pelanggan | 2 design partner P&C | 4–6 pelanggan (termasuk jiwa/kesehatan) | 10+ pelanggan, ekspansi regional |

### 6.4 Design partner

- **MVP:** 2 perusahaan asuransi P&C (idealnya 1 Tier 1 dan 1 mid-market) yang bersedia memberikan data historis, akses shadow mode, dan feedback mingguan.
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
| E9. Decisioning & workflow | Keputusan konsisten, human-in-the-loop, write-back | P0 | MVP |
| E10. Reporting & dashboards | Visibilitas operasional, model, dan eksekutif | P0 | MVP |
| E11. Model monitoring & feedback | Model belajar dari outcome; drift terkelola | P0 | MVP |
| E12. Admin & multi-tenancy | Tenant, peran, konfigurasi, usage metering | P0 | MVP |
| E13. Security, privacy & audit | Kepatuhan UU PDP dan audit-ready | P0 | MVP |
| E14. External data & consortium | Data eksternal dan berbagi data opt-in | P1 | Phase 2/3 |
| E15. GenAI investigation assistant | Ringkasan, Q&A, draft report dengan sitasi | P2 | Phase 3 |
| E16. Notifications & collaboration | Tim SIU bekerja bersama secara aman | P1 | MVP basic |

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
| FR-302 | Sebagai Analis, saya dapat menggunakan rule pack per lini dan tipologi. | P&C auto/properti; jiwa; kesehatan; rule pack dapat dikustomisasi. | P0 | MVP |
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
| FR-601 | Sebagai Investigator, saya dapat mengekstrak data dari dokumen Indonesia. | KTP, SIM, STNK, BPKB, invoice, kuitansi, rekam medis, surat kematian, laporan polisi; akurasi field utama ≥95% pada test set. | P1 | MVP basic |
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
| FR-802 | Sebagai Investigator, saya dapat menggunakan template investigasi per tipologi. | P&C auto/properti; jiwa; kesehatan; tasks; SLA; checklist. | P0 | MVP |
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
| FR-904 | Sebagai Claims Ops, saya dapat mengirim keputusan ke claims system. | Status/decision/notes; API; idempotensi; error handling. | P0 | MVP |
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
| FR-1204 | Sebagai Finance, saya dapat melihat usage metering. | Claims scored; users; API calls; billing export. | P1 | Phase 2 |
| FR-1205 | Sebagai Data Engineer, saya dapat menggunakan sandbox. | Synthetic data; safe testing; isolated. | P0 | MVP |
| FR-1206 | Sebagai Admin, saya dapat mengelola feature flags. | Per tenant; audit; rollback. | P1 | Phase 2 |

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

---

## 8. Kebutuhan Non-Fungsional

| ID | Kategori | Requirement | Acceptance criteria |
|---|---|---|---|
| NFR-1 | Performance | Scoring real-time dan batch | API p95 <2 detik; batch platform 1 juta klaim <4 jam; UI <3 detik |
| NFR-2 | Availability | Ketersediaan layanan | 99.9% MVP; 99.95% GA; RTO 4 jam; RPO 15 menit |
| NFR-3 | Scalability | Skala multi-tenant | 10 juta klaim/tahun per tenant; 1.000 concurrent users; horizontal scaling |
| NFR-4 | Security | Keamanan aplikasi dan data | Kontrol selaras ISO 27001 (sertifikasi target Phase 2); OWASP ASVS L2; enkripsi; MFA; pen test tahunan |
| NFR-5 | Privacy | UU PDP | Consent/legal basis; DSR; retensi; deletion; pseudonymization; data residency |
| NFR-6 | Explainability | Alasan alert dan keputusan | 100% alert memiliki reason code; 100% keputusan memiliki audit trail |
| NFR-7 | Fairness | Keadilan model | Atribut terlarang dikecualikan; bias test; monitoring per segmen |
| NFR-8 | Localization | Bahasa dan format Indonesia | Bahasa Indonesia utama; English toggle; IDR; WIB/WITA/WIT; NIK/NPWP; alamat |
| NFR-9 | Accessibility | Aksesibilitas | WCAG 2.1 AA |
| NFR-10 | Observability | Monitoring dan troubleshooting | Metrics; logs; traces; model monitoring; alerting |
| NFR-11 | Maintainability | Kualitas engineering | API-first; CI/CD; IaC; test coverage; dokumentasi |
| NFR-12 | Portability | Exit dan portabilitas | Export data, model, konfigurasi; tidak ada lock-in |
| NFR-13 | Compliance | Retensi dan audit | Retensi 5–10 tahun configurable; legal hold; audit log immutable |
| NFR-14 | Usability | Efisiensi investigator | Dari alert ke kasus ≤3 klik; training ≤1 hari |

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
- Write-back ke claims system: status, decision, notes, attachment metadata.

### 10.6 Non-goals arsitektur

- Bukan core PAS/claims system.
- Tidak memproses pembayaran.
- Tidak menyimpan seluruh arsip dokumen selamanya; retensi sesuai kebijakan.
- Tidak menggantikan DMS.

---

## 11. UX/UI dan Lokalisasi

### 11.1 Prinsip desain

1. **Evidence-first:** setiap alert menampilkan "mengapa" sebelum "apa."
2. **Capacity-aware:** queue diurutkan berdasarkan dampak, SLA, dan kapasitas.
3. **Minimal clicks:** dari alert ke kasus ≤3 klik.
4. **Bahasa Indonesia first:** istilah bisnis lokal; English toggle.
5. **Role-based:** investigator, analis, kepala SIU, compliance melihat hal yang relevan.
6. **Accessible:** WCAG 2.1 AA; keyboard navigation; kontras.
7. **Responsive:** desktop untuk investigasi; mobile untuk upload bukti dan approval.

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
- WCAG 2.1 AA; ukuran font; kontras; label; keyboard.

---

## 12. Kepatuhan, Keamanan, dan Privasi

### 12.1 Peta regulasi

| Regulasi/standar | Relevansi | Implikasi produk |
|---|---|---|
| **UU No. 40 Tahun 2014 tentang Perasuransian** | Kerangka usaha, tata kelola, perlindungan pemegang polis | Keputusan klaim harus dapat dipertanggungjawabkan; audit trail |
| **UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi (UU PDP)** | Data pribadi, data sensitif, consent, hak subjek data | Privacy-by-design; DSR; breach notification; DPO; retensi |
| **UU No. 17 Tahun 2023 tentang Kesehatan** | Kerahasiaan data kesehatan | PHI masking; akses terbatas; legal basis |
| **POJK No. 23/POJK.05/2015 tentang Perlindungan Konsumen Sektor Jasa Keuangan** | Perlakuan adil, transparansi, penanganan keluhan | Tidak ada penolakan otomatis; komunikasi jelas; jalur sengketa |
| **POJK No. 44/POJK.05/2020 tentang Penerapan Manajemen Risiko bagi LJK Non-Bank** | Manajemen risiko, termasuk risiko model dan outsourcing | Model risk management; vendor risk; kontinuitas |
| **POJK No. 73/POJK.05/2016 tentang Tata Kelola Perusahaan yang Baik** | Akuntabilitas, transparansi, pengawasan | Approval matrix; audit; pelaporan |
| **UU No. 8 Tahun 2010 tentang APU-PPT** | Kewajiban anti pencucian uang (relevan untuk jiwa) | Integrasi watchlist/sanksi; audit |
| **UU No. 11 Tahun 2008 jo. UU No. 19 Tahun 2016 tentang ITE; PP No. 71 Tahun 2019** | Bukti elektronik, penyelenggaraan sistem elektronik | Audit log; integritas data; data residency |
| **Panduan/kajian OJK terkait AI dan manajemen risiko model** | Tata kelola AI | Explainability; fairness; validasi; dokumentasi |
| **AAUI/AAJI dan asosiasi industri terkait** | Praktik industri, berbagi data fraud | Kepatuhan asosiasi; data sharing opt-in |
| **ISO 27001 / SOC 2** | Keamanan dan kontrol | Sertifikasi/roadmap; audit |
| **NIST AI RMF (opsional)** | Tata kelola risiko AI | Framework validasi dan monitoring |

> Nomor dan interpretasi regulasi dapat berubah. Legal/Compliance wajib memvalidasi versi terbaru dan relevansinya per lini.

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
- **Record retention:** 5–10 tahun configurable; legal hold; deletion yang dapat diaudit.

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

- Retensi default: 5 tahun setelah kasus ditutup; configurable hingga 10 tahun sesuai kebijakan tenant dan kewajiban hukum.
- Legal hold menghentikan penghapusan.
- DSR: verifikasi identitas; akses/koreksi/penghapusan; catat di audit log.
- Data yang dihapus harus dihapus dari sistem produksi dan backup sesuai kebijakan.

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

### 13.2 Dashboard

- **Eksekutif:** fraud protected, recovered, leakage, ROI, coverage, trend per lini.
- **Operasional:** queue aging, cycle time, workload, SLA, alert volume.
- **Model:** precision/recall, calibration, drift, fairness, override rate.
- **Portofolio:** loss ratio impact, typology, provider/agent network, region.
- **Kepatuhan:** case audit, access log, DSR, retention, incident.

### 13.3 Desain pilot

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
- ROI positif atau jalur jelas menuju ROI.

### 13.4 Formula ROI

```
ROI = (Fraud dicegah + Fraud dipulihkan + Penghematan operasional - Total biaya) / Total biaya

Fraud dicegah   = nilai klaim berisiko tinggi yang tidak dibayar atau dikurangi
Fraud dipulihkan = recovery dari investigasi
Penghematan     = jam investigator yang dihemat × fully loaded cost
Total biaya     = subscription + implementasi + biaya data + effort internal
```

Gunakan holdout/control group untuk mengukur fraud dicegah; jangan mengklaim seluruh penurunan klaim sebagai dampak platform.

---

## 14. Model Komersial dan GTM

### 14.1 Packaging

| Paket | Target | Cakupan | Batas indikatif |
|---|---|---|---|
| **Core** | Mid-market P&C | P&C auto/properti; rules + baseline ML; alert & case; dashboard; 10 users | ≤100K klaim/tahun |
| **Pro** | Multi-line insurer | Semua lini; graph; document forensics; external connectors; model tuning; 50 users | ≤1 juta klaim/tahun |
| **Enterprise** | Tier 1 / group | Unlimited; private cloud/on-prem; custom model; consortium; SLA 99.95%; dedicated support | Custom |

### 14.2 Prinsip pricing

- **Annual subscription + volume tier + implementation fee.**
- **Pilot fixed fee**, dikreditkan ke kontrak tahunan jika lanjut.
- **Hindari per-alert pricing** yang menghukum deteksi; per-claim scoring dapat diterima.
- **Outcome-based upside** opsional (shared savings) untuk tenant matang.
- **Transparansi overage** dan biaya data eksternal.
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

---

## 15. Rencana Rilis dan Milestone

| Fase | Timeline | Scope | Exit criteria |
|---|---|---|---|
| **Phase 0 — Discovery** | M0–M2 | Design partner, data profiling, workflow mapping, compliance assessment, arsitektur, metrik | 2 pilot agreement; data contract; backlog prioritas; arsitektur disetujui |
| **Phase 1 — MVP** | M2–M6 | Shared core + P&C auto/properti; 1 integrasi claims; rules + baseline ML; case management; dashboard; audit | MVP acceptance; security review; pilot readiness |
| **Phase 1b — Pilot** | M6–M9 | Live pilot 2 insurer; shadow mode; feedback | Kriteria pilot tercapai; GA readiness |
| **Phase 2 — Multi-line** | M9–M15 | Jiwa & kesehatan; provider analytics; graph lanjutan; document forensics; external connectors | 2 pelanggan jiwa/kesehatan live; target model; compliance review |
| **Phase 3 — Intelligence** | M15–M24 | GenAI assistant; agentic evidence; consortium; ASEAN; outcome-based pricing | 10+ pelanggan; ARR target; retention; SOC 2 Type II |

### 15.1 Tim inti

| Peran | Jumlah | Fokus |
|---|---:|---|
| Product Manager | 1 | Prioritas, roadmap, stakeholder |
| Product Designer | 1 | UX investigasi, design system |
| Frontend Engineer | 2 | Case workspace, dashboard, graph |
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

---

## 16. Risiko dan Mitigasi

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

### 18.4 Referensi

**Internal:**

- [Market & Vendor Landscape 2026](./insurance-fraud-detection-market-landscape-2026.md)

**Regulasi dan standar:**

- Otoritas Jasa Keuangan (OJK): https://www.ojk.go.id
- Asosiasi Asuransi Umum Indonesia (AAUI): https://www.aaui.org
- Asosiasi Asuransi Jiwa Indonesia (AAJI): https://www.aaji.or.id
- UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi (akses via JDIH/BPK)
- UU No. 40 Tahun 2014 tentang Perasuransian (akses via JDIH/BPK)
- UU No. 17 Tahun 2023 tentang Kesehatan (akses via JDIH/BPK)
- POJK No. 23/POJK.05/2015 tentang Perlindungan Konsumen Sektor Jasa Keuangan
- POJK No. 44/POJK.05/2020 tentang Penerapan Manajemen Risiko bagi LJK Non-Bank
- POJK No. 73/POJK.05/2016 tentang Tata Kelola Perusahaan yang Baik bagi Perusahaan Perasuransian
- UU No. 8 Tahun 2010 tentang Pencegahan dan Pemberantasan Tindak Pidana Pencucian Uang
- UU No. 11 Tahun 2008 jo. UU No. 19 Tahun 2016 tentang Informasi dan Transaksi Elektronik
- PP No. 71 Tahun 2019 tentang Penyelenggaraan Sistem dan Transaksi Elektronik

**Standar teknis:**

- ISO/IEC 27001: https://www.iso.org/standard/27001
- OWASP ASVS: https://owasp.org/www-project-application-security-verification-standard/
- WCAG 2.1: https://www.w3.org/WAI/standards-guidelines/wcag/
- NIST AI Risk Management Framework: https://www.nist.gov/itl/ai-risk-management-framework

---

**Akhir dokumen.**  
Dokumen ini akan diperbarui setelah review design partner, validasi Legal/Compliance, dan hasil pilot.

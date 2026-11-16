# TRD — Platform Deteksi Fraud Asuransi Multi-Lini

**Nama kerja:** Sentinel Fraud Platform (placeholder)  
**Versi:** 0.1 — Draft untuk review  
**Tanggal:** 17 November 2026  
**Pemilik:** Engineering  
**Status:** Draft  
**Pasar awal:** Indonesia (OJK)  
**Dokumen terkait:** BRD Insurance Claim Fraud Detection v0.5, PRD Platform Deteksi Fraud Asuransi v0.2, QA Report  
**Cakupan teknis:** Phase 0/1 thin slice, commercial MVP, dan target-state; profil HA untuk MVP/GA  
**Stack baseline:** Next.js, Go, PostgreSQL, Redis, Kafka, Kubernetes  
**Cloud utama:** Alibaba Cloud Indonesia (Jakarta, `ap-southeast-5`), 3 availability zone

**Reviewer wajib:** CTO, Lead Backend, Lead Frontend, SRE/Platform Lead, Security Architect, Data Architect, QA Lead, Compliance/Legal.

> Dokumen ini adalah technical requirements document, bukan nasihat hukum. Seluruh interpretasi regulasi, retensi, dan residensi data wajib divalidasi oleh Legal/Compliance. Semua angka kapasitas adalah **design envelope** dan harus divalidasi dengan benchmark, pilot, serta data design partner.

---

## Changelog

| Versi | Tanggal | Perubahan |
|---|---|---|
| 0.1 | 16 November 2026 | Draft awal: arsitektur logis, microservices Go, Next.js, PostgreSQL, Redis, Kafka, topologi HA multi-AZ, sizing, DR, observability, keamanan, test HA, runbook, ADR, dan traceability ke PRD/BRD |
| 0.1 (rev. 2) | 17 November 2026 | Strategi deploy eksplisit ke Alibaba Cloud `ap-southeast-5`: landing zone, ACK Pro, ApsaraDB RDS, Tair, ApsaraMQ, OSS WORM/CRR, ALB/WAF/Anti-DDoS, ASM, SLS/ARMS, referensi R36–R50 |

---

## Daftar Isi

0. Document Control & Scope
1. Ringkasan Eksekutif & Keputusan Arsitektur
2. Tujuan, Non-Goals, dan Requirement Traceability
3. Prinsip Arsitektur & Standar Teknologi
4. Konteks Sistem & Batas Layanan
5. Arsitektur Logis
6. Arsitektur Deployment & High Availability
7. Desain Layanan Backend (Go)
8. Desain Frontend (Next.js)
9. Desain Data
10. Alur Data & Konsistensi
11. API & Kontrak Integrasi
12. Keamanan & Kepatuhan
13. Observability & SLO
14. Kapasitas, Concurrency, dan Sizing
15. Backup, Disaster Recovery, dan Business Continuity
16. CI/CD, Release, dan Environment
17. Test Strategy & Validasi HA
18. Runbook Operasional
19. Alternatif Teknologi & ADR
20. Risiko Teknis & Mitigasi
21. Rencana Implementasi & Milestone
22. Open Questions & Assumptions
23. Traceability Matrix
24. Referensi
25. Lampiran

---

## 0. Document Control & Scope

### 0.1 Status dan hubungan dokumen

| Item | Nilai |
|---|---|
| Dokumen | TRD — Platform Deteksi Fraud Asuransi Multi-Lini |
| Versi | 0.1 |
| Tanggal | 16 November 2026 |
| Status | Draft untuk review |
| Pemilik | Engineering |
| Target-state | 24 bulan (M0–M24) |
| Thin slice | BRD v0.5 — Phase 0/1, rules-only, data sintetis, P&C auto/properti |
| Dokumen acuan | BRD v0.5, PRD v0.2, QA Report |
| Perubahan | TRD baru; belum mengubah scope PRD/BRD |

**Hierarki dokumen:**

1. **BRD v0.5** mengikat problem, scope Phase 0/1, rule, data contract, acceptance, dan business case.
2. **PRD v0.2** mengikat target-state 24 bulan, requirement fungsional/non-fungsional, compliance, metrik, dan model komersial.
3. **TRD ini** mengikat realisasi teknis: arsitektur, deployment, HA, data, API, keamanan, observability, sizing, dan operasi.
4. **QA Report** mengikat bukti pengujian dan hasil acceptance.

Jika terjadi konflik: **scope/requirement mengikuti PRD/BRD; cara implementasi mengikuti TRD**. Konflik yang tidak dapat diselesaikan harus masuk decision log PRD §19.5/BRD §18.5.

### 0.2 Phase map dan stance HA

| BRD v0.5 | PRD v0.2 | Profil deployment | HA |
|---|---|---|---|
| Fase 1 — Hackathon MVP | Phase 0/1 thin slice | Docker Compose lokal/offline; single node | **Tidak HA** — validasi teknis, data sintetis |
| Fase 2 — 1–3 bulan | Phase 1b/2 | Kubernetes 3 AZ; managed data services | **HA** — target 99.9% MVP |
| Fase 3 — 3–6 bulan | Phase 2/3 | Kubernetes 3 AZ; HA + DR | **HA** — target 99.95% GA + DR |
| — | Commercial MVP (M0–M9) | Kubernetes 3 AZ; managed data services | **HA** — 99.9% |

> **Thin slice ≠ commercial MVP.** Thin slice berjalan non-HA untuk alasan biaya dan kecepatan, tetapi **kode, kontrak, dan state harus HA-ready**: stateless, state eksternal, health probes, idempotency, dan tidak menyimpan state di filesystem lokal. TRD ini mendefinisikan HA untuk commercial MVP/GA; thin slice tidak boleh diklaim sebagai produksi HA.

### 0.3 Source of truth teknis

| Domain | Source of truth | Catatan |
|---|---|---|
| Scope, rule, data contract Phase 0/1 | BRD v0.5 §7, §10, §13 | TRD tidak boleh menambah rule/scope |
| Requirement produk, NFR, compliance | PRD v0.2 §7–§12 | TRD memetakan ke komponen |
| Arsitektur, service boundary, deployment | TRD §5–§6 | Perubahan arsitektur melalui ADR |
| API contract | OpenAPI di repo; TRD §11 | Contract testing wajib |
| Event contract | Schema registry; TRD §9.7 | Versioned, backward-compatible |
| Data model fisik | Migration repo; TRD §9 | ERD target ada di PRD §9.2/§18.2 |
| SLO, alert, runbook | TRD §13, §18 | Di-review SRE |
| Security control | TRD §12; PRD §12 | Di-review Security/Compliance |
| DR plan | TRD §15 | Diuji quarterly |

### 0.4 Asumsi dan constraint

- Cloud utama: **Alibaba Cloud Indonesia (Jakarta, `ap-southeast-5`)** dengan **3 availability zone** (Jakarta Zone A/B/C) [R36][R37]. Provider alternatif hanya untuk contingency dan harus melalui ADR; seluruh layanan managed dipetakan ke ekuivalen Alibaba Cloud.
- Jika tenant enterprise memilih private cloud/on-prem, topologi HA direplikasi ke 3 rack/DC dengan komponen self-managed (ACK/kubernetes, Patroni, Redis Sentinel, Kafka, MinIO) dan tetap mengikuti pola failure domain yang sama.
- Thin slice: Docker Compose, data sintetis, tanpa PII nyata, tanpa internet/API eksternal, tanpa ML/OCR/integrasi/write-back.
- Commercial MVP: P&C auto/properti, rules + baseline ML adapter, OCR basic, 1 integrasi claims.
- Tidak ada active-active multi-region pada MVP/GA. DR memakai **warm standby** in-country.
- Tidak ada write-back status klaim sampai Phase 2 dan opt-in tenant.
- Semua angka kapasitas adalah `design envelope`, bukan forecast. Harga/biaya cloud `TBD`.

### 0.5 Definisi "TRD siap"

TRD dianggap siap untuk implementasi jika:

1. Setiap NFR PRD/BRD punya komponen, kontrol, dan cara verifikasi.
2. Setiap service punya boundary, data ownership, API/event, dan scaling policy.
3. Topologi HA, failure domain, RTO/RPO, dan runbook jelas.
4. Sizing punya formula dan benchmark gate, bukan angka tanpa asal.
5. Security, privacy, audit, dan compliance terpetakan.
6. Backup/restore dan DR diuji, bukan hanya didokumentasikan.
7. Tidak ada keputusan arsitektur besar yang belum punya ADR.
8. Dokumen saling mereferensikan dan versi konsisten.

### 0.6 Ringkasan keputusan

| ID | Keputusan | Status |
|---|---|---|
| ADR-001 | Alibaba Cloud ACK managed cluster multi-AZ sebagai compute platform | Accepted |
| ADR-002 | ApsaraDB RDS for PostgreSQL Cluster Edition/HA Edition multi-zone sebagai default; Patroni untuk self-managed | Accepted |
| ADR-003 | Tair (Redis OSS-compatible) multi-AZ untuk cache/session/rate limit; Sentinel hanya self-managed | Accepted |
| ADR-004 | ApsaraMQ for Kafka multi-zone sebagai durable event backbone; self-managed Kafka/NATS sebagai alternatif | Accepted |
| ADR-005 | Next.js App Router sebagai UI + BFF tipis; tanpa business logic | Accepted |
| ADR-006 | Shared schema + Row-Level Security (RLS) per tenant; dedicated DB untuk enterprise | Accepted |
| ADR-007 | OSS dengan versioning, retention policy (WORM), dan CRR | Accepted |
| ADR-008 | ASM (Istio-based) untuk mTLS/traffic management pada GA; NetworkPolicy + app resilience pada MVP | Accepted |
| ADR-009 | OpenTelemetry + SLS + ARMS Managed Prometheus/Grafana | Accepted |
| ADR-010 | DR warm standby in-country; RTO 4 jam / RPO 15 menit | Accepted |
| ADR-011 | Alibaba Cloud sebagai cloud utama; provider lain hanya contingency | Accepted |

---

## 1. Ringkasan Eksekutif & Keputusan Arsitektur

Sentinel adalah platform multi-tenant untuk mendeteksi fraud klaim asuransi. Phase 0/1 thin slice membuktikan pipeline rules-only dengan data sintetis. Commercial MVP menambahkan baseline ML, OCR basic, integrasi claims, case management, dan multi-tenant. TRD ini merancang platform yang **stateless di compute, stateful di layanan data terkelola, dan HA secara default** pada commercial MVP/GA.

**Bentuk solusi:**

- **Frontend:** Next.js App Router (TypeScript) sebagai UI dan BFF tipis. Tidak ada business logic atau penyimpanan state lokal. Semua instance stateless dan berbagai cache melalui Redis.
- **Backend:** microservices Go dengan boundary domain yang jelas. Sinkron melalui REST/JSON (eksternal) dan gRPC opsional (internal); asinkron melalui Kafka. Semua service bisnis ditulis Go; Phase 2 ML serving memakai Go/ONNX, sedangkan training offline boleh Python dan tidak berjalan di production.
- **Data:** PostgreSQL sebagai system of record; Redis untuk cache/session/rate limit; object storage untuk dokumen dan evidence; Kafka untuk event dan outbox.
- **HA:** 3 AZ, minimal 3 replica per service, PodDisruptionBudget, topology spread, managed multi-AZ untuk PostgreSQL/Redis/Kafka, backup + PITR, dan DR warm standby in-country.
- **Security:** OIDC/MFA, RBAC/ABAC, mTLS, RLS, KMS, WAF, secrets management, immutable audit, dan supply-chain controls.
- **Operability:** OpenTelemetry, SLO/error budget, dashboards, runbook, chaos test, load test, dan DR drill.

**Yang tidak diklaim:**

- Thin slice tidak HA dan tidak multi-tenant.
- MVP/GA tidak active-active multi-region.
- Tidak ada auto-denial klaim; write-back hanya Phase 2 opt-in.
- Angka kapasitas bukan jaminan performa sampai benchmark dan pilot selesai.

### 1.1 Stack decision

| Layer | Pilihan | Status | Alasan |
|---|---|---|---|
| Frontend | Next.js App Router | Wajib | SSR/RSC, BFF tipis, ekosistem, self-host |
| Backend | Go microservices | Wajib | Performa, concurrency, binary kecil, operasi sederhana |
| Database | PostgreSQL (ApsaraDB RDS for PostgreSQL) | Wajib | ACID, RLS, partitioning, ekosistem, portabilitas |
| Cache | Redis (Tair Redis OSS-compatible) | Wajib | Session, cache, rate limit, queue index; multi-AZ HA |
| Integrasi | REST/JSON + Kafka (ApsaraMQ for Kafka) | Wajib | API-first + event-driven; replay dan audit |
| Orchestration | Alibaba Cloud ACK multi-AZ | Tambahan | HA, autoscaling, PDB, topology spread |
| Object storage | Alibaba Cloud OSS | Tambahan | Dokumen/evidence/export; versioning, WORM, CRR |
| Edge | Alibaba Cloud ALB + WAF + Anti-DDoS | Tambahan | TLS, routing, JWT, rate limit, WAF |
| Observability | OTel + SLS + ARMS Prometheus/Grafana | Tambahan | SLO, trace, log, alert |
| Delivery | Argo CD + Terraform provider `alicloud` | Tambahan | GitOps, IaC, rollback |
| ML serving | Go/ONNX (Phase 2) | Tambahan | Model portable; training offline Python |

**Prinsip stack:** managed service dipakai jika tersedia dan SLA-nya memenuhi target; self-managed hanya untuk private cloud/on-prem atau kebutuhan kontrol. Setiap penggantian komponen harus melalui ADR dan review HA.

---

## 2. Tujuan, Non-Goals, dan Requirement Traceability

### 2.1 Tujuan teknis

| ID | Tujuan | Ukuran |
|---|---|---|
| TG-01 | Ketersediaan layanan | 99.9% MVP; 99.95% GA; error budget terkelola |
| TG-02 | Latensi | API p95 <2 detik; UI <3 detik; scoring real-time p95 <2 detik |
| TG-03 | Throughput | 1.000 concurrent users; 500 rps peak API; 1 juta klaim <4 jam |
| TG-04 | Skala data | 10 juta klaim/tahun per tenant; design envelope 50 juta/tahun aggregate |
| TG-05 | Durabilitas | RPO 15 menit; RTO 4 jam; backup restore teruji |
| TG-06 | Determinisme | Input + versi rule sama → output identik; backtest 2× diff = 0 |
| TG-07 | Keamanan | OWASP ASVS L2; enkripsi; tenant isolation; audit immutable |
| TG-08 | Observability | Metrics, logs, traces, SLO, alert, runbook |
| TG-09 | Portabilitas | Tidak lock-in; export data/model/config; opsi private cloud |
| TG-10 | Maintainability | API-first, CI/CD, IaC, contract test, dokumentasi |

### 2.2 Non-goals

- Active-active multi-region; multi-cloud aktif; global load balancing.
- Sharding horizontal di luar PostgreSQL pada MVP.
- ML/OCR/forensik dokumen di thin slice.
- Write-back status klaim di MVP.
- Penggantian PAS/claims system atau DMS.
- Pemrosesan pembayaran.
- Auto-denial atau keputusan klaim tanpa manusia.
- Data lakehouse penuh di MVP (Phase 2).
- SLA komersial sebelum pilot selesai.

### 2.3 Requirement sources

| Sumber | Fokus | Bagian TRD |
|---|---|---|
| PRD §8 NFR-1..20 | Performance, availability, security, privacy, metering | §2.4, §6, §12, §13, §14 |
| PRD §10 arsitektur | API-first, event-driven, multi-tenant, portability | §5, §6, §11 |
| PRD §12 compliance | UU PDP, OJK, ISO/SOC, audit | §12 |
| BRD §9 NFR-01..24 | Thin slice performance, determinism, privacy, billing | §7, §9, §14, §17 |
| BRD §10 data | Schema, data quality, retention, metering | §9 |
| BRD §13 acceptance | AC-1..16, AT mapping | §17, §23 |
| PRD §15 phase map | Thin slice → MVP → Phase 2/3 | §0.2, §21 |

### 2.4 NFR target

| ID | Kategori | Target | Verifikasi |
|---|---|---|---|
| TRD-NFR-01 | Availability MVP | 99.9%/bulan; error budget 43,2 menit | SLO dashboard + synthetic |
| TRD-NFR-02 | Availability GA | 99.95%/bulan; error budget 21,6 menit | SLO dashboard + audit |
| TRD-NFR-03 | API latency | p95 <2 detik; p99 <5 detik | Load test + APM |
| TRD-NFR-04 | UI latency | p95 <3 detik pada 1.000 users | k6 browser + RUM |
| TRD-NFR-05 | Throughput API | 500 rps peak; burst 1.000 rps 60 detik | Load test |
| TRD-NFR-06 | Batch | 1 juta klaim <4 jam; 50 ribu <90 detik | Benchmark |
| TRD-NFR-07 | Concurrency | 1.000 concurrent users | Load test |
| TRD-NFR-08 | Durability | RPO 15 menit; RTO 4 jam | Backup + DR drill |
| TRD-NFR-09 | AZ failover | RTO <5 menit; RPO 0 untuk PostgreSQL sync | Chaos test |
| TRD-NFR-10 | Determinism | Backtest 2× diff = 0 | Automated test |
| TRD-NFR-11 | Security | OWASP ASVS L2; 0 critical/high terbuka | Pentest + SAST/DAST |
| TRD-NFR-12 | Privacy | 0 PII di log; raw NIK ditolak | Log scan + test |
| TRD-NFR-13 | Audit | 100% decision punya audit trail | Test + audit query |
| TRD-NFR-14 | Metering | 100% rekonsiliasi; selisih 0 | Reconciliation test |
| TRD-NFR-15 | Observability | 100% service punya metrics/logs/traces | Review |
| TRD-NFR-16 | Recovery | Restore <4 jam; DR <4 jam | Drill |
| TRD-NFR-17 | Portability | Export data/model/config <30 hari | Tabletop + test |
| TRD-NFR-18 | Accessibility | WCAG 2.1 AA | Audit |

### 2.5 SLO dan error budget

| SLO | Target | Budget/bulan (30 hari) | Catatan |
|---|---:|---:|---|
| Edge/API availability | 99.95% | 21,6 menit | 5xx / total request valid; 4xx tidak dihitung |
| API latency p95 | <2 detik | 5% request | Diukur di edge |
| UI availability | 99.95% | 21,6 menit | Halaman utama + API BFF |
| Batch success | 99.9% | 43,2 menit | Batch 1 juta <4 jam |
| Metering reconciliation | 100% | 0 | Defect, bukan error budget |
| Data durability | 99.999999% | — | Mengikuti object storage + backup |

**Aturan dependency:** SLO 99.95% hanya dapat dipertahankan jika dependency kritis (IdP, KMS, object storage, managed DB/Redis/Kafka) memiliki SLA ≥99.95% dan arsitektur redundan. Jika SLA dependency 99.9%, target platform harus diturunkan atau dependency diganti. Review SLA dilakukan sebelum GA.

**Error budget policy:**

- Burn rate >2× dalam 1 jam → alert; >10× → page on-call.
- Budget habis 50% dalam 7 hari → freeze perubahan non-perbaikan.
- Budget habis 100% → freeze release fitur; fokus reliability.
- Planned maintenance diumumkan ≥5 hari; emergency maintenance dihitung sebagai downtime.

---

## 3. Prinsip Arsitektur & Standar Teknologi

### 3.1 Prinsip

1. **Stateless compute, stateful data services.** Pod/container tidak menyimpan state yang harus bertahan.
2. **Externalized session dan cache.** Session di Redis atau cookie terenkripsi; cache di Redis; object di object storage.
3. **Idempotency by default.** Semua operasi tulis punya idempotency key; consumer punya dedupe.
4. **Loose coupling.** Service berkomunikasi lewat API/event; tidak ada shared table lintas service.
5. **Deterministic core.** Rules engine tidak bergantung pada waktu lokal, random, atau urutan map.
6. **Multi-AZ default.** Setiap komponen stateful dan stateless tersebar di 3 AZ.
7. **Graceful degradation.** Cache down → fallback ke DB; LLM/OCR down → queue; reporting down → tidak memblokir review.
8. **Security by design.** Least privilege, tenant isolation, enkripsi, audit, no PII di log.
9. **Observable by default.** Setiap request punya correlation_id dan trace; setiap service punya SLO.
10. **Boring technology.** Pilih teknologi yang dipahami tim; managed service jika tersedia.
11. **No distributed transaction.** Gunakan outbox/saga; terima eventual consistency antar-service.
12. **Reversibility.** Feature flag, canary, expand/contract migration, rollback <10 menit.

### 3.2 Standar versi

| Komponen | Baseline | Catatan |
|---|---|---|
| Go | 1.24+ (pin di `go.mod`) | Gunakan versi stable yang didukung security patch |
| Next.js | App Router, versi stable di-pin | TypeScript strict; Node LTS |
| PostgreSQL | 17+ (18 setelah tersedia di ApsaraDB RDS) | Extension minimal; `pg_stat_statements` |
| Redis | 7.4+ (Tair Redis OSS-compatible) | TLS; ACL; multi-AZ failover |
| Kafka | ApsaraMQ for Kafka; 4.x KRaft untuk self-managed | Managed multi-zone; RF 3, minISR 2 |
| Kubernetes | ACK managed (versi didukung); 1.32+ untuk self-managed | ACK control plane 3 AZ; kubeadm HA untuk private cloud |
| Load balancer | Alibaba Cloud ALB/NLB | Multi-AZ; health check; TLS |
| Service mesh | Alibaba Cloud ASM (Istio-based) | Opsional di MVP; wajib untuk mTLS skala penuh |
| Envoy Gateway | Gateway API v1 (opsional) | Bila butuh routing/JWT di dalam ACK |
| OpenTelemetry | SDK stable | OTLP |
| Observability | SLS + ARMS Managed Prometheus/Grafana | Logs, metrics, traces, alert |
| Argo CD/Rollouts | Stable | GitOps + canary |
| IaC | Terraform provider `alicloud` | Resource Directory, ACK, RDS, Tair, ApsaraMQ, OSS |

### 3.3 Pilihan library dan tooling

| Area | Pilihan | Alasan |
|---|---|---|
| HTTP router Go | `chi` atau `net/http` | Ringan, kompatibel standard library |
| DB access | `pgx` + `sqlc` | Typed query, performa, no ORM berat |
| Migration | `golang-migrate` atau Atlas | Expand/contract, versioned |
| Kafka client | `franz-go` | Performa, idempotent producer, consumer group |
| Redis client | `go-redis/v9` | TLS; Tair/Sentinel-compatible |
| Object storage | `aliyun-oss-go-sdk` atau S3-compatible SDK | OSS; MinIO untuk lokal |
| Auth | OIDC + JWT (`go-jose`) | Standar, SSO/MFA |
| RBAC | Casbin atau policy engine internal | Simple, auditable |
| Observability | OTel SDK, ARMS Prometheus, SLS, Grafana | Vendor-neutral; managed Alibaba |
| Frontend data | TanStack Query + Zod | Cache, validation, typed |
| UI | Tailwind + headless components | Konsisten, accessible |
| E2E | Playwright | Browser test |
| Load | k6 | Load/soak/spike |
| Chaos | Chaos Mesh atau LitmusChaos | Kubernetes-native |
| ML serving | ONNX Runtime (Go binding) | Model portable; training offline Python tidak berjalan di production |

### 3.4 Struktur repository

```
/
  apps/
    web/                       # Next.js App Router
  services/
    identity/
    admin/                     # MVP: tenant config, entitlement, feature flag
    ingestion/
    rules/
    alert/
    case/
    audit/
    metering/
    reporting/
    notification/
    document/                  # Phase 2
    entity-resolution/         # Phase 2
    feature-store/             # Phase 2
    ml/                        # Phase 2: training offline (Python), serving Go/ONNX
    graph/                     # Phase 2
    integration/               # Phase 2
    llm-gateway/               # Phase 3
  packages/
    contracts/                 # OpenAPI, protobuf, JSON Schema
    go-common/                 # config, logging, auth, errors, otel
    rulesengine/               # shared deterministic engine
  infra/
    terraform/alicloud/        # landing zone, ACK, RDS, Tair, ApsaraMQ, OSS
    helm/
    kustomize/
    argo/
  db/
    migrations/
  tests/
    load/
    chaos/
    e2e/
  docs/
```

**Aturan:** `packages/contracts` adalah source of truth kontrak; perubahan harus melalui PR dan contract test. `services/*` tidak boleh mengimpor kode service lain; hanya `packages/*`.

### 3.5 Konvensi

- **API:** REST/JSON, OpenAPI 3.1, versioning `/v1`, error model RFC 7807.
- **Event:** nama `domain.entity.action.vN`; payload versioned; schema registry.
- **ID:** UUIDv7 untuk waktu terurut; alert ID deterministik dari `claim_id + rule_id + rule_version`.
- **Waktu:** UTC di storage/API; tampilan WIB/WITA/WIT.
- **Uang:** integer minor unit (IDR tanpa desimal); tidak pernah float.
- **Log:** JSON terstruktur; `correlation_id`, `tenant_id`, `service`, `route`, `status`, `latency_ms`; tanpa PII.
- **Error:** kode stabil, pesan aman, retryable flag, detail terredaksi.
- **Config:** environment variable + secret manager; tidak ada secret di repo/image.
- **Feature flag:** default off; audit perubahan; kill switch.

---

## 4. Konteks Sistem & Batas Layanan

### 4.1 Aktor

| Aktor | Peran | Phase |
|---|---|---|
| Reviewer | Meninjau alert, mengumpulkan bukti, memberi outcome | Thin slice, MVP |
| Supervisor | Assign, review kualitas, approve keputusan | Thin slice, MVP |
| Admin | Tenant, user, role, config, usage | Thin slice, MVP |
| Finance/Management | Read-only usage/value report/billing export | MVP |
| Auditor/Regulator | Read-only audit/export | MVP |
| Data Scientist | Rule/model governance | MVP/Phase 2 |
| Claims/PAS | Sumber klaim, penerima write-back opt-in | Phase 2 |
| DMS/Object storage | Dokumen | Phase 2 |
| IdP | SSO/MFA | MVP |
| KMS/Secrets | Enkripsi/key management (Alibaba Cloud KMS + Secrets Manager) | MVP |
| External data | Dukcapil, Samsat, watchlist | Phase 2/3 |

### 4.2 Batas sistem

**In scope:** ingestion, validasi, rules, scoring, alert/triage, case/evidence, audit, metering, reporting, notification, admin, document AI (Phase 2), ML (Phase 2), integration (Phase 2).

**Out of scope:** core PAS/claims system, pembayaran, DMS, underwriting pricing, auto-denial, fraud di luar klaim, replacement DMS.

### 4.3 Trust boundary

| Boundary | Di luar | Di dalam | Kontrol |
|---|---|---|---|
| Internet → edge | User, browser | Alibaba Cloud WAF, CDN, ALB | TLS, WAF, Anti-DDoS, rate limit |
| Edge → app | ALB/Envoy | Next.js, Go services | JWT, mTLS, NetworkPolicy |
| Service → service | Service lain | Namespace | mTLS, authorization, timeout |
| App → data | Compute | ApsaraDB RDS/PostgreSQL, Tair/Redis, ApsaraMQ/Kafka, OSS | TLS, RAM, RLS, encryption |
| Tenant → tenant | Tenant A | Tenant B | RLS, token claim, audit |
| Admin → prod | Operator | Cluster | SSO, MFA, RAM least privilege, break-glass |

### 4.4 Region dan residensi

- Region utama: **Alibaba Cloud Indonesia (Jakarta, `ap-southeast-5`)**, 3 AZ: `ap-southeast-5a/b/c` [R36][R37].
- DR: lokasi kedua **di Indonesia** (co-location/private cloud) atau region Alibaba Cloud terdekat **hanya dengan persetujuan Legal dan tenant**. Alibaba Cloud Indonesia saat ini satu region; region failure tanpa DR in-country adalah residual risk yang harus disetujui.
- Data tenant tidak boleh keluar Indonesia kecuali diizinkan eksplisit.
- Object storage (OSS), backup, dan log (SLS) harus mengikuti residensi.

### 4.5 Kapabilitas per phase

| Kapabilitas | Thin slice | Commercial MVP | Phase 2 | Phase 3 |
|---|---|---|---|---|
| Rules A–G | Ya | Ya | Ya | Ya |
| Rules H/OCR | Tidak | Basic | Advanced | Advanced |
| ML scoring | Tidak | Baseline adapter | Ya | Ya |
| Ingestion | CSV/JSON | API/SFTP/batch | Event/connectors | Event/connectors |
| Multi-tenant | Tidak | Ya | Ya | Ya |
| Case management | Basic | Ya | Ya | Ya |
| Graph | Tidak | Basic | Advanced | Advanced |
| Write-back | Tidak | Tidak | Opt-in | Opt-in |
| External data | Tidak | Tidak | Ya | Ya |
| GenAI | Tidak | Tidak | Tidak | Ya |

---

## 5. Arsitektur Logis

### 5.1 Diagram komponen

```
┌────────────────────────────────────────────────────────────────────┐
│                        Client / Browser                            │
└───────────────────────────────┬────────────────────────────────────┘
                                │ HTTPS
┌───────────────────────────────▼────────────────────────────────────┐
│ CDN + WAF + Anti-DDoS + Alibaba Cloud DNS (multi-AZ)               │
└───────────────────────────────┬────────────────────────────────────┘
                                │
┌───────────────────────────────▼────────────────────────────────────┐
│ ALB multi-AZ (TLS, routing) + Envoy Gateway opsional (JWT/rate)    │
└───────────────┬───────────────────────────────────┬────────────────┘
                │                                   │
┌───────────────▼──────────────┐      ┌─────────────▼────────────────┐
│ Next.js web/BFF (3+ replica) │      │ External API (v1)            │
│ RSC, UI, session, no logic   │      │ REST/JSON, OpenAPI           │
└───────────────┬──────────────┘      └─────────────┬────────────────┘
                │                                   │
                └───────────────┬───────────────────┘
                                │
┌───────────────────────────────▼────────────────────────────────────┐
│ Go microservices on ACK (3+ replica per service, 3 AZ)              │
│ identity | admin | ingestion | rules | alert | case | audit         │
│ metering | reporting | notification | document* | ml* | integration*│
└───────┬───────────────────┬───────────────────┬────────────────────┘
        │                   │                   │
┌───────▼────────┐  ┌───────▼────────┐  ┌───────▼────────────────────┐
│ ApsaraDB RDS   │  │ Tair (Redis)   │  │ ApsaraMQ for Kafka         │
│ PostgreSQL     │  │ primary + 2    │  │ multi-zone, RF=3, minISR=2 │
│ Cluster/HA     │  │ replica        │  │ topics + DLQ + schema      │
│ Edition        │  │ multi-AZ       │  │ registry                   │
└───────┬────────┘  └────────────────┘  └───────┬────────────────────┘
        │                                        │
┌───────▼────────────────────────────────────────▼───────────────────┐
│ Alibaba Cloud OSS: documents, evidence, export, audit              │
│ versioning, SSE-KMS, retention policy (WORM), CRR                  │
└────────────────────────────────────────────────────────────────────┘
        │
┌───────▼────────────────────────────────────────────────────────────┐
│ Observability: OTel Collector → SLS, ARMS Managed Prometheus,      │
│ Managed Grafana, ARMS tracing; ActionTrail audit terpisah          │
└────────────────────────────────────────────────────────────────────┘
* Phase 2
```

### 5.2 Katalog service

| Service | Tanggung jawab | Data yang dimiliki | API/event | Phase | Replica HA |
|---|---|---|---|---|---|
| `identity` | Tenant, user, role, session mapping, policy | tenant, user, role, consent metadata | `/v1/auth`, `/v1/tenants`, `identity.*` | Thin slice/MVP | 3 |
| `admin` | Konfigurasi tenant, entitlement, feature flag, view usage | tenant_config, entitlement, feature_flag | `/v1/admin`, `admin.*` | MVP | 3 |
| `ingestion` | Upload CSV/JSON, validasi, mapping, batch status | claim_raw, import_batch, idempotency | `/v1/imports`, `claim.ingested.v1` | Thin slice/MVP | 3 |
| `rules` | Rule catalog, versioning, evaluasi deterministik, backtest | rule, rule_pack, evaluation | `/v1/rules`, `rules.evaluated.v1` | Thin slice/MVP | 3 |
| `alert` | Alert creation, bundle, priority, queue | alert, alert_bundle | `/v1/alerts`, `alert.created.v1` | Thin slice/MVP | 3 |
| `case` | Case group, evidence, task, decision, audit metadata | case, evidence, decision | `/v1/cases`, `case.updated.v1` | Thin slice/MVP | 3 |
| `audit` | Append-only audit, legal hold, export | audit_event | `/v1/audit`, `audit.event.v1` | Thin slice/MVP | 3 |
| `metering` | Usage event, period close, value report, billing export | usage_record, value_report, billing_export | `/v1/usage`, `usage.recorded.v1` | Thin slice/MVP | 3 |
| `reporting` | Dashboard, export, scheduled report | materialized view/read model | `/v1/reports`, `report.*` | Thin slice/MVP | 3 |
| `notification` | Email/webhook/in-app notification | notification, delivery | `/v1/notifications`, `notification.*` | MVP | 3 |
| `document` | OCR, forensics, metadata, chain of custody | document, extracted_field, forensic | `/v1/documents`, `document.processed.v1` | Phase 2 | 3 |
| `entity-resolution` | Matching, Golden Record, relasi pihak/aset | entity, entity_edge, match_decision | `/v1/entities`, `entity.*` | Phase 2 | 3 |
| `feature-store` | Definisi fitur, online/offline, point-in-time correctness | feature_def, feature_value, feature_snapshot | `/v1/features`, `feature.*` | Phase 2 | 3 |
| `ml` | Model serving, explainability, drift | model, prediction, model_metric | `/v1/scores`, `model.*` | Phase 2 | 3 |
| `integration` | Connectors, webhook, write-back opt-in | connector, sync_state | `/v1/connectors`, `integration.*` | Phase 2 | 3 |
| `graph` | Graph query, ring detection | graph projection | `/v1/graph`, `graph.*` | Phase 2 | 3 |
| `llm-gateway` | RAG, redaction, citation, guardrail | prompt_audit (redacted) | `/v1/assistant` | Phase 3 | 3 |

**Aturan data ownership:** hanya service pemilik yang menulis tabelnya. Service lain membaca melalui API/event/read model. Untuk MVP, semua service dapat memakai satu cluster PostgreSQL dengan schema terpisah dan role terpisah; enterprise dapat memakai dedicated cluster.

> **Data lakehouse (Phase 2):** komponen data, bukan service transaksional. Object storage + Iceberg/Trino untuk data mentah/terkurasi, lineage, training, dan analitik; tidak menjadi source of truth transaksi. **Feature store (Phase 2):** dibangun di atas PostgreSQL/Redis + Iceberg, dengan point-in-time join; bukan source of truth keputusan.

### 5.3 Thin slice subset

Thin slice menjalankan subset berikut:

- `web` (queue, explanation, coverage, dashboard).
- `identity` (local admin/reviewer/supervisor; tanpa SSO).
- `ingestion` (CSV/JSON upload; fail-fast; hash identity).
- `rules` (A–G; H off; deterministic; version pinned).
- `alert` (bundle per claim; priority deterministic).
- `case` (case group, evidence, decision CONFIRMED/CLEARED/NEEDS_INFO).
- `audit` (append-only).
- `metering` (usage record, value report, billing export).
- `reporting` (coverage matrix, queue, basic dashboard).
- PostgreSQL, Redis, Kafka (single-node KRaft), MinIO.

Tidak ada: ML, OCR, document service, integration, notification, graph, multi-tenant, SSO/MFA, write-back, external API.

> Thin slice memakai role admin lokal di `identity`; service `admin` terpisah baru ada di MVP saat entitlement/feature flag multi-tenant dibutuhkan.

### 5.4 Cross-cutting

| Komponen | Fungsi |
|---|---|
| ALB + Envoy opsional | TLS, routing, JWT validation, rate limit, request size limit |
| ASM (GA) | mTLS, retry, timeout, traffic split, telemetry |
| OTel Collector | Trace/metric/log pipeline |
| ARMS Managed Prometheus | Metrics + long-term |
| SLS | Logs + query |
| ARMS tracing | Traces |
| Managed Grafana + ARMS alert | Dashboard + alert |
| KMS Secrets Manager | Sinkron secret dari KMS |
| Cert Manager | Rotasi certificate |
| Kyverno/Gatekeeper | Policy admission |
| Trivy/Cosign | Scan + sign image |
| ACR | Image registry + vulnerability scan |
| Velero | Backup resource Kubernetes |
| Argo CD/Rollouts | GitOps + canary |
| ActionTrail | Audit API cloud |

### 5.5 Aturan dependensi

1. Tidak ada shared database table lintas service.
2. Tidak ada synchronous call cycle; gunakan event.
3. Sync call hanya untuk query yang membutuhkan respons langsung; timeout wajib.
4. Semua event punya schema dan version; consumer harus backward-compatible.
5. Service tidak boleh mengakses Redis/DB service lain.
6. `audit` dan `metering` menerima event dari service lain; tidak boleh menjadi dependency blocking.
7. Reporting membaca read replica/read model; tidak boleh membebani primary.

### 5.6 Sequence: ingest → rules → alert

```
Reviewer/System
  │ POST /v1/imports (CSV/JSON + Idempotency-Key)
  ▼
Ingestion
  │ validasi, normalisasi, simpan claim_raw + import_batch
  │ tulis outbox claim.ingested.v1 (transaksi sama)
  ▼
Kafka
  │ claim.ingested.v1 (key: tenant_id:claim_id)
  ▼
Rules worker
  │ load rule pack version, evaluasi deterministik
  │ simpan evaluation + outbox rules.evaluated.v1
  ▼
Alert
  │ buat alert/bundle, priority deterministik
  │ simpan alert + outbox alert.created.v1
  ▼
Queue/UI
  │ reviewer melihat alert + evidence + reason code
```

### 5.7 Sequence: review → decision → audit/metering

```
Reviewer
  │ PATCH /v1/cases/{id}/decision (CONFIRMED/CLEARED/NEEDS_INFO)
  ▼
Case
  │ validasi RBAC, optimistic lock, simpan decision
  │ tulis outbox case.updated.v1 + decision.recorded.v1
  ▼
Kafka
  ├─► Audit: append audit_event (immutable)
  ├─► Metering: decision_made, confirmed_cases, value fields
  └─► Reporting: update read model
```

### 5.8 Rule engine

- **Input:** claim, policy, provider, invoice, identity hash, historical aggregates, `as_of`.
- **Rule pack:** immutable version; setiap rule punya `rule_id`, `version`, `family`, `severity`, `params`, `mapping_typology`, `owner`.
- **Eksekusi:** urutan stabil (family, rule_id); tidak ada map iteration tanpa sort; tidak ada random/waktu lokal.
- **Output:** `triggered`, `severity`, `trigger_value`, `evidence_ref`, `reason_code`, `explanation` 4 elemen.
- **Determinism:** canonical JSON; alert ID = `sha256(tenant_id|claim_id|rule_id|rule_version)`; backtest 2× diff = 0.
- **Isolation:** `is_fraud_ground_truth` tidak dapat dibaca rule engine; test otomatis.
- **Performance:** evaluasi per rule target ≤0,5 ms; rule yang butuh data historis memakai cache/aggregate, bukan query N+1.
- **Kill switch:** rule pack dapat dinonaktifkan per tenant/family; fallback ke rules-only.

### 5.9 Metering dan reporting

- **Usage event:** `claim_scored`, `document_processed`, `flag_created`, `case_group_created`, `decision_made`, `active_user`.
- **Usage record:** minimal `period`, `tenant_id`, `claims_scored`, `documents_processed`, `flags_created`, `case_groups`, `active_users`, `confirmed_cases`, `value_protected_estimate`, `value_recovered`, `formula_version`; immutable setelah close; rekonsiliasi 100% dengan event sumber.
- **Value report:** confirmed cases, value protected/recovered, time saved, formula version, confidence; input/version sama → output identik; tanpa PII; no double counting; **planted-test recall dilaporkan terpisah dari production lift**.
- **Billing export:** versioned, checksum, tanpa PII; dapat direproduksi; rekonsiliasi 100% dengan usage record.
- **Reporting:** materialized view di read replica; refresh async; tidak blocking transaksi.

---

## 6. Arsitektur Deployment & High Availability

### 6.1 Profil environment

| Profil | Tujuan | Topologi | HA | Data |
|---|---|---|---|---|
| Local dev | Developer | Docker Compose | Tidak | Sintetis |
| Thin slice demo | Hackathon | Docker Compose single node | Tidak | Sintetis |
| Staging | Pre-prod | Kubernetes 3 AZ kecil | Ya (HA-lite) | Sintetis/anonymized |
| Production MVP | Pilot | Kubernetes 3 AZ + managed data | Ya, 99.9% | Tenant pilot |
| Production GA | Komersial | Kubernetes 3 AZ + managed data + DR | Ya, 99.95% | Tenant produksi |
| Private cloud | Enterprise | Kubernetes 3 rack/DC + self-managed data | Ya | Tenant |

> **Observability per profil:** thin slice hanya memakai log/audit lokal dan tidak memiliki SLO/alert produksi; staging dan production memakai stack OTel §13. Jangan mengklaim thin slice memenuhi NFR availability/DR/observability PRD.

### 6.2 Topologi referensi 3 AZ

```
                         ┌──────────────────┐
                         │ Alibaba Cloud DNS│
                         │ health + failover│
                         └────────┬─────────┘
                                  │
                 ┌────────────────▼────────────────┐
                 │ ALB multi-AZ + WAF + Anti-DDoS  │
                 └────────────────┬────────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │ AZ-A                    │ AZ-B                    │ AZ-C
        │                         │                         │
  ┌─────▼─────┐            ┌─────▼─────┐            ┌─────▼─────┐
  │ ACK node  │            │ ACK node  │            │ ACK node  │
  │ app 3x    │            │ app 3x    │            │ app 3x    │
  │ batch 1x  │            │ batch 1x  │            │ batch 1x  │
  └─────┬─────┘            └─────┬─────┘            └─────┬─────┘
        │                        │                        │
  ┌─────▼─────┐            ┌─────▼─────┐            ┌─────▼─────┐
  │ Envoy     │            │ Envoy     │            │ Envoy     │
  │ opsional  │            │ opsional  │            │ opsional  │
  └─────┬─────┘            └─────┬─────┘            └─────┬─────┘
        │                        │                        │
        └───────────────┬────────┴───────────────┬────────┘
                        │                        │
             ┌──────────▼──────────┐  ┌──────────▼─────────┐
             │ ApsaraDB RDS        │  │ Tair (Redis)       │
             │ PostgreSQL          │  │ primary AZ-A       │
             │ primary AZ-A        │  │ replica AZ-B       │
             │ secondary AZ-B      │  │ replica AZ-C       │
             │ readable AZ-C       │  │ multi-AZ failover  │
             └──────────┬──────────┘  └──────────┬─────────┘
                        │                        │
             ┌──────────▼────────────────────────▼─────────┐
             │ ApsaraMQ for Kafka multi-zone, RF=3, minISR=2│
             └──────────┬──────────────────────────────────┘
                        │
             ┌──────────▼──────────────────────────────────┐
             │ OSS multi-AZ + versioning + WORM + CRR      │
             └─────────────────────────────────────────────┘
```

### 6.3 Kubernetes HA baseline

| Kontrol | Baseline | Alasan |
|---|---|---|
| Control plane | ACK managed/ACK Pro control plane across 3 AZ; self-managed 3 control plane + 3 etcd | K8s menyarankan ≥3 failure zone untuk control plane/etcd [R1][R2]; ACK mengelola control plane 3 AZ [R38] |
| Node pool | Minimal 3 node app, satu per AZ; system pool terpisah; Terway CNI; ESSD | Hindari correlated failure |
| Replica | Minimal 3 per service; critical 6 | Tahan 1 AZ down |
| Topology spread | `maxSkew: 1`, `topologyKey: topology.kubernetes.io/zone`, `whenUnsatisfiable: DoNotSchedule` | Sebar pod [R33] |
| Pod anti-affinity | Prefer different nodes/AZ | Hindari satu node |
| PDB | `minAvailable: 2` untuk replica 3; `maxUnavailable: 1` | Cegah drain habis [R32] |
| HPA | CPU 60% + custom RPS/queue; min 3 max 12 | Skala horizontal [R34] |
| Probes | startup, readiness, liveness terpisah | Cegah restart loop |
| Graceful shutdown | `terminationGracePeriodSeconds: 60`, preStop 10s | Selesaikan request |
| Resources | requests/limits; QoS Burstable/Guaranteed untuk critical | Cegah noisy neighbor |
| Priority | system > gateway > core > batch | Preemption saat kapasitas ketat |
| NetworkPolicy | default deny; allow per dependency | Segmentasi |
| SecurityContext | non-root, read-only root FS, drop capabilities | Hardening |
| Disruption | node drain dengan PDB; cluster autoscaler buffer | Maintenance aman |

**Catatan:** Kubernetes tidak menyediakan cross-zone resilience untuk API server endpoint; gunakan load balancer dengan health check [R1]. ACK managed/ACK Pro mengelola control plane across 3 AZ [R38].

### 6.4 Edge dan network HA

- DNS: Alibaba Cloud DNS health check + failover ke DR; TTL rendah saat DR.
- Load balancer: ALB multi-AZ (minimal 2 AZ; target 3); NLB untuk TCP; health check per target group [R43].
- WAF: Alibaba Cloud WAF managed rules + rate limit + bot control [R44].
- DDoS: Anti-DDoS Pro/Premium sesuai kebutuhan [R45].
- NAT: satu NAT Gateway per AZ; jangan satu NAT untuk semua AZ.
- Egress: VPC endpoint/PrivateLink untuk OSS/KMS/Secrets; tidak lewat internet.
- TLS: Alibaba Cloud Certificate Management Service/ACME; rotasi otomatis; HSTS.
- CDN: Alibaba Cloud CDN untuk static assets; cache invalidation via deployment ID.
- Gateway: ALB 3 AZ; Envoy Gateway opsional 3 replica di ACK; JWT, request size, rate limit.
- Service mesh: ASM (Istio-based) untuk mTLS/retry/traffic split pada GA [R42]; MVP dapat memakai NetworkPolicy + TLS + app-level resilience.

### 6.5 Stateful HA summary

| Komponen | HA mechanism | AZ failure | Region failure | Catatan |
|---|---|---|---|---|
| ApsaraDB RDS PostgreSQL (Cluster/HA Edition) | Multi-zone; primary + secondary/readable secondary; automatic failover | RTO <30 detik (HA Edition); <2 menit (Cluster Edition); RPO ≤15 menit* | RTO ≤2 jam; RPO ≤15 menit | RDS HA/DR [R40]; PITR |
| PostgreSQL self-managed | Patroni + etcd 3 node; synchronous_mode; HAProxy/PgBouncer | RTO <2 menit; RPO 0 | Standby cluster manual promote | Patroni multi-DC [R6] |
| Tair (Redis) | Multi-AZ primary + replica; automatic failover | RTO <60 detik; cache/session loss window kecil | Rebuild cache; RTO ≤2 jam | Tair DR [R41]; async replication |
| ApsaraMQ for Kafka | Multi-zone; RF=3; minISR=2; acks=all | RTO <5 menit; no loss untuk acked | Replay/outbox; RPO ≤15 menit | ApsaraMQ multi-zone [R46] |
| Object storage | OSS multi-AZ; versioning; CRR | RTO ~0 | RTO ≤1 jam; RPO ≤15 menit | OSS durability + CRR [R47] |
| Kubernetes | ACK managed control plane 3 AZ; node 3 AZ | RTO <5 menit | Redeploy DR | ACK HA [R38] |
| Next.js/Go | 3+ replica; stateless | RTO <1 menit | Redeploy DR | PDB, HPA |

> **Catatan RPO PostgreSQL:** RPO 0 hanya berlaku jika mode replikasi sinkron dan secondary sehat. ApsaraDB RDS HA Edition/Cluster Edition dapat melakukan failover otomatis, tetapi mode replikasi aktual harus diverifikasi di console sebelum menjanjikan RPO 0. Target platform tetap RPO ≤15 menit [R40].

### 6.6 Failure mode matrix

| Failure | Dampak | Deteksi | Respons otomatis | Respons manual | Target |
|---|---|---|---|---|---|
| Pod crash | 1 replica hilang | Readiness/liveness | Restart, route ke replica lain | Lihat log/trace | <1 menit |
| Node failure | Pod di node hilang | Node NotReady | Reschedule | Drain/cordon | <5 menit |
| AZ failure | 1/3 kapasitas hilang | AZ health, error rate | Route ke 2 AZ; RDS/Tair failover | Verifikasi kapasitas | <5 menit |
| DB primary failure | Write terhenti | DB health, connection error | ApsaraDB RDS automatic failover / Patroni | Verifikasi lag | <2 menit |
| Redis primary failure | Cache/session terputus | Tair health event | Tair multi-AZ failover | Verifikasi client | <60 detik |
| Kafka broker failure | Broker/zone hilang | ApsaraMQ health, URP | Multi-zone failover, producer retry | Rebalance/repair | <5 menit |
| Object storage error | Upload/download gagal | OSS 5xx/latency | Retry/backoff | Failover endpoint | <15 menit |
| Bad deploy | Error rate naik | Canary analysis | Rollback otomatis | Postmortem | <10 menit |
| Queue backlog | Alert terlambat | Consumer lag | Scale consumer | Drop/priority | <30 menit |
| Disk full | DB/Kafka gagal | Disk alert | Autoscale storage | Cleanup/expand | <30 menit |
| Certificate expiry | TLS gagal | Cert alert | Auto-renew | Manual renew | <15 menit |
| KMS outage | Enkripsi gagal | Alibaba KMS error | Retry/cache data key | Failover key | <30 menit |
| Region failure | Semua down | Synthetic + DNS | DNS failover | DR runbook | ≤4 jam |

### 6.7 Kapasitas headroom

- Setiap AZ harus mampu menampung **≥67% beban** saat satu AZ down; target **N+1 per AZ**.
- CPU/memory cluster target ≤60% steady; burst ≤80%.
- ApsaraDB RDS: ≤60% CPU, ≤70% storage, replication lag <5 detik.
- Tair: ≤60% memory, eviction policy `allkeys-lru` untuk cache; session keys tanpa eviction.
- ApsaraMQ for Kafka: ≤60% disk, URP 0, consumer lag <5 menit.
- OSS: lifecycle; tidak ada kapasitas fisik.
- Headroom minimum 40% untuk peak; autoscaling buffer 20%.

### 6.8 Anti-pattern yang dilarang

- Satu NAT gateway untuk semua AZ.
- ApsaraDB RDS single-zone atau read replica sebagai pengganti HA.
- Tair single node untuk session/rate limit.
- ApsaraMQ Kafka single-zone, RF=1, atau `min.insync.replicas=1` untuk topik transaksional.
- Session di memory pod atau filesystem lokal.
- Upload dokumen ke filesystem lokal.
- Cron/job tanpa leader election atau idempotency.
- Long transaction saat batch besar.
- Retry tanpa backoff/jitter.
- Tanpa PDB saat drain.
- Menyimpan secret di image/env repo.
- Mengakses DB service lain langsung.
- Menjadikan cache sebagai source of truth.

### 6.9 Maintenance dan upgrade

| Komponen | Strategi | Downtime |
|---|---|---|
| ACK | Rolling node pool; surge; drain dengan PDB | 0 |
| App | Rolling/canary; maxUnavailable 0 | 0 |
| ApsaraDB RDS minor | Maintenance window / secondary upgrade | <1 menit |
| ApsaraDB RDS major | Blue-green replica + switchover | Terjadwal |
| Tair | Failover ke replica; upgrade replica dulu | <1 menit |
| ApsaraMQ for Kafka | Multi-zone rolling; controlled shutdown | 0 untuk client resilient |
| ASM | Canary control plane | 0 |
| ALB/Gateway | Rolling; config canary | 0 |
| OSS | Tidak ada downtime | 0 |
| ACR | Immutable tag; replication | 0 |

### 6.10 Service mesh dan mTLS

- **GA:** ASM (Alibaba Cloud Service Mesh, Istio-based), mTLS strict antar-service, authorization policy per service account, retry/timeout, traffic split canary [R42].
- **MVP:** jika tim belum siap, gunakan NetworkPolicy default-deny + TLS ingress + JWT + app-level retry/timeout. mTLS service-to-service dapat ditunda, tetapi harus masuk risk register.
- **Tanpa mesh:** pastikan setiap service memvalidasi JWT, memakai HTTP client dengan timeout, dan tidak mempercayai header tenant dari client.

### 6.11 Multi-region stance

- **Active-passive** warm standby; bukan active-active.
- DNS failover Alibaba Cloud DNS; ApsaraDB RDS cross-region backup/restore atau read replica/promote; OSS CRR; event replay/outbox untuk ApsaraMQ.
- RTO 4 jam, RPO 15 menit.
- Data residency: DR in-country; cross-border (mis. Singapore) hanya dengan persetujuan Legal/tenant.
- Active-active ditinjau ulang setelah GA jika ada kebutuhan tenant dan biaya terjustifikasi.

### 6.12 Strategi deploy Alibaba Cloud

**Prinsip landing zone**

- **Region:** Alibaba Cloud Indonesia (Jakarta, `ap-southeast-5`), 3 AZ `ap-southeast-5a/b/c` [R36][R37].
- **Akun:** Resource Directory dengan akun terpisah untuk `prod`, `staging`, `dev`, `shared`, `security`, dan `dr`; tidak ada data produksi di akun non-prod [R39].
- **Identitas cloud:** RAM + CloudSSO; least privilege; role per workload; tidak ada AccessKey permanen di pod/CI.
- **Jaringan:** satu VPC per environment; vSwitch per AZ; NAT Gateway per AZ; tidak ada public IP di node; ALB satu-satunya ingress publik.
- **IaC:** Terraform provider `alicloud`; remote state di OSS + locking; plan review; drift detection; tidak ada perubahan manual di prod kecuali break-glass.
- **Tagging wajib:** `env`, `tenant`, `service`, `owner`, `cost-center`, `data-classification`.
- **Data residency:** OSS, SLS, backup, dan DR in-country; cross-border hanya dengan persetujuan Legal/tenant.

**Topologi deployment**

| Layer | Layanan Alibaba Cloud | Konfigurasi HA |
|---|---|---|
| DNS/edge | Alibaba Cloud DNS, CDN, WAF, Anti-DDoS | DNS health check; WAF managed rules; Anti-DDoS Pro/Premium |
| Load balancer | ALB (L7), NLB (TCP) | Multi-AZ minimal 2 AZ, target 3 AZ; health check |
| Compute | ACK Pro managed cluster | Control plane 3 AZ; node pool per AZ; Terway; ESSD; PDB; HPA |
| Registry | ACR Enterprise | Immutable tag; vulnerability scan; geo-replication jika DR |
| Database | ApsaraDB RDS for PostgreSQL Cluster Edition (GA) / HA Edition (MVP) | Multi-zone; automatic failover; readable secondary; backup + PITR |
| Cache | Tair (Redis OSS-compatible) | Multi-AZ primary/replica; automatic failover; cluster edition jika perlu |
| Messaging | ApsaraMQ for Kafka Professional | Multi-zone; RF=3; minISR=2; acks=all |
| Object storage | OSS | Versioning; retention policy (WORM); CRR; lifecycle |
| Secrets/keys | KMS + Secrets Manager | CMK per env; rotation; envelope encryption; RRSA |
| Observability | SLS, ARMS Managed Prometheus, Managed Grafana, ARMS tracing | Multi-AZ managed; alert; SLO dashboard |
| Audit | ActionTrail | Organization trail; immutable delivery ke OSS/SLS |
| Security | Cloud Firewall, Security Center, WAF, RAM | Default deny; vulnerability/threat; least privilege |
| Service mesh | ASM (Istio-based) | mTLS; traffic split; retry/timeout; multi-AZ |
| GitOps | Argo CD + Argo Rollouts | Canary; rollback; DR cluster sync |
| Backup/DR | DBS, OSS CRR, RDS backup, ACK DR cluster | RTO 4 jam; RPO 15 menit; drill quarterly |

**Urutan deployment**

1. Landing zone: Resource Directory, akun, RAM/CloudSSO, budget, tagging, logging pusat.
2. Network: VPC, vSwitch 3 AZ, NAT per AZ, ALB/NLB, WAF, DNS, PrivateLink.
3. Platform: ACK Pro, node pool 3 AZ, ACR, ingress, cert-manager, policy admission, observability agent.
4. Data: RDS PostgreSQL multi-zone, Tair multi-AZ, ApsaraMQ Kafka multi-zone, OSS bucket + WORM + CRR, KMS.
5. Platform add-ons: Argo CD, secrets, backup, autoscaling, PDB, NetworkPolicy/ASM.
6. Aplikasi: identity → ingestion → rules → alert → case/audit → metering/reporting; canary per service.
7. Validasi: load 500 rps/1.000 users, batch 1M/4h, chaos AZ/DB/Tair/ApsaraMQ, restore test.
8. Pilot: tenant pilot, SLO 99.9%, runbook, on-call.
9. DR: cluster DR, DNS failover, RDS restore, OSS CRR, game day; GA 99.95%.

**Environment matrix**

| Environment | Akun | Region/AZ | ACK | RDS | Tair | ApsaraMQ | OSS |
|---|---|---|---|---|---|---|---|
| dev | dev | `ap-southeast-5`, 1 AZ | ACK kecil | HA Edition single-zone | single-AZ | single-zone | lokal/OSS |
| staging | staging | 3 AZ | ACK Pro 3 AZ | HA Edition multi-zone | multi-AZ | multi-zone | versioning |
| prod MVP | prod | 3 AZ | ACK Pro 3 AZ | HA Edition/Cluster multi-zone | multi-AZ | multi-zone | versioning + CRR |
| prod GA | prod | 3 AZ + DR | ACK Pro 3 AZ + DR | Cluster Edition multi-zone + DBS | multi-AZ + DR | multi-zone | WORM + CRR + RTC |
| private cloud | tenant | 3 rack/DC | ACK/kubernetes | Patroni | Sentinel/Cluster | Kafka | MinIO |

**Kontrol HA Alibaba Cloud**

- **ACK:** gunakan ACK Pro; control plane dikelola across 3 AZ [R38]; node pool minimal 3, satu per AZ; topology spread `DoNotSchedule`; PDB; HPA; cluster autoscaler dengan buffer 20%.
- **RDS:** pilih multi-zone; HA Edition untuk MVP (failover <30 detik) atau Cluster Edition untuk readable secondary dan skala GA; aktifkan backup otomatis + PITR; uji failover [R40].
- **Tair:** multi-AZ primary/replica; client harus reconnect; cache loss = fallback DB; session loss = re-login; rate limit reset window [R41].
- **ApsaraMQ:** Professional Edition multi-zone; RF=3; minISR=2; acks=all; consumer idempotent; DLQ; monitor lag [R46].
- **OSS:** versioning + retention policy WORM untuk audit/evidence; CRR untuk DR; RTC untuk RPO ketat [R47][R48].
- **ALB:** deploy minimal 2 AZ (target 3); health check; WAF/anti-DDoS di depan [R43].
- **ASM:** mTLS strict; authorization policy; retry/timeout; canary; multi-AZ [R42].
- **Observability:** SLS untuk log, ARMS Managed Prometheus untuk metrics, Managed Grafana untuk dashboard, ARMS tracing [R49][R50].
- **Audit:** ActionTrail organization trail ke OSS/SLS; akses read-only untuk auditor.

**Catatan dan guardrail Alibaba Cloud**

- **Ketersediaan layanan:** verifikasi ketersediaan ACK Pro, RDS Cluster Edition PostgreSQL, Tair multi-AZ, ApsaraMQ Professional, ALB, ASM, dan kuota di `ap-southeast-5` sebelum komitmen. Jika tidak tersedia, gunakan fallback yang terdaftar di ADR.
- **ApsaraMQ 2 AZ vs 3 AZ:** jika multi-zone Jakarta hanya mendukung 2 AZ untuk edisi tertentu, dokumentasikan; untuk simetri 3 AZ, gunakan Kafka self-managed di ACK 3 AZ atau terima risiko 2 AZ dengan RF=3 dan minISR=2.
- **Region failure:** Alibaba Cloud Indonesia satu region. DR in-country harus disiapkan via co-location/private cloud; jika tidak, cross-border (mis. Singapore) hanya dengan persetujuan Legal/tenant. Tanpa DR site, region failure adalah residual risk yang harus disetujui.
- **Quota:** ajukan kuota ECS/ESSD/EIP/ALB/RDS sebelum peak; quota keterlambatan adalah risiko proyek.
- **Biaya:** subscription/commitment discount jika tersedia; jangan mengklaim margin sebelum data billing pilot. Harga `TBD`.
- **Kepatuhan:** OSS/SLS/backup harus di region Indonesia; ActionTrail dan log akses disimpan sesuai retensi.

---

## 7. Desain Layanan Backend (Go)

### 7.1 Template service

Setiap service Go memiliki:

```
cmd/server/main.go
internal/
  config/        # env + validation
  http/          # handler, middleware
  service/       # business logic
  repo/          # pgx/sqlc
  event/         # producer/consumer
  auth/          # JWT, RBAC
  observability/ # otel, metrics, logging
  health/        # live/ready
migrations/
openapi.yaml
```

**Baseline main:**

- Load config; fail fast jika invalid.
- Init OTel tracer/meter/logger.
- Init DB pool (pgx) + PgBouncer.
- Init Kafka producer/consumer.
- Init Tair/Redis client (multi-AZ endpoint; Sentinel hanya self-managed).
- Register health endpoints.
- Start HTTP server dengan timeouts.
- Handle SIGTERM: stop accepting, drain, close consumers, flush OTel.
- Panic recovery middleware; tidak membocorkan stack ke client.

### 7.2 Concurrency dan resilience

| Kontrol | Baseline |
|---|---|
| Context timeout | HTTP 2s default; scoring 5s; report 30s; batch per item 1s |
| Retry | Hanya idempotent; exponential backoff + jitter; max 3 |
| Circuit breaker | Per dependency; open setelah error threshold; half-open |
| Bulkhead | Pool/worker terpisah untuk report, batch, upload |
| Rate limit | Per tenant/user/IP di gateway + app |
| Singleflight | Cegah cache stampede |
| Worker pool | Bounded; backpressure; tidak unbounded goroutine |
| Graceful shutdown | 30–60s; stop consume; finish in-flight |
| Panic | Recover; log; metric; 500 aman |
| Queue | Kafka consumer; max.poll.records; pause/resume |

**Aturan:** jangan retry operasi non-idempotent tanpa idempotency key. Jangan retry 4xx kecuali 408/429. Jangan memakai `time.Now()` di rules engine; gunakan `as_of`.

### 7.3 Data access

- `pgxpool` per service; max conn dihitung dari pool budget.
- Semua query via `sqlc` atau query konstan; tidak ada string concat.
- Transaksi: `BEGIN; SET LOCAL app.tenant_id = $1; ...; COMMIT;` untuk RLS.
- Outbox ditulis di transaksi yang sama dengan aggregate.
- Optimistic locking via `version`; conflict → 409.
- Pagination cursor (keyset), bukan offset untuk data besar.
- Statement timeout; idle-in-transaction timeout; lock timeout.
- Read replica hanya untuk reporting; jangan untuk read-after-write.
- Tidak ada `SELECT *`; pilih kolom; index review.

### 7.4 Kafka producer/consumer

**Producer:**

- `acks=all`, `enable.idempotence=true`, `max.in.flight=5`.
- Key = `tenant_id:aggregate_id` untuk ordering.
- Header: `event_id`, `event_type`, `schema_version`, `correlation_id`, `trace_id`.
- Outbox relay: `SELECT ... FOR UPDATE SKIP LOCKED`, publish, mark; retry dengan backoff; DLQ setelah N.
- Tidak publish di luar transaksi DB.

**Consumer:**

- Manual commit setelah sukses.
- Idempotent: unique `(consumer, event_id)` di tabel `processed_event`.
- Retry topic dengan delay; DLQ; alert.
- `max.poll.interval.ms` sesuai worst-case processing.
- Rebalance listener: commit, pause, drain.
- Poison message: log redacted, DLQ, continue.

### 7.5 Rules engine determinism

- Rule pack immutable; hash disimpan di evaluation.
- Evaluasi tanpa network call; data historis di-preload.
- Sorting eksplisit; canonical JSON.
- `as_of` dari input, bukan clock.
- Alert ID deterministik.
- Backtest: dataset + rule pack + engine version → hash output; 2× harus sama.
- Label isolation test: rule engine tidak bisa membaca `is_fraud_ground_truth`.
- Kill switch per tenant/family.

### 7.6 Batch worker

- Input: import batch; chunk 1.000–5.000 claim.
- Parallel worker pool; bounded; backpressure.
- Idempotent per claim; skip jika sudah diproses dengan versi sama.
- Progress di `import_batch`; checkpoint.
- Gagal per baris → laporan; fail-fast untuk batch sesuai BRD.
- Limit baseline: 5.000–50.000 klaim/batch; 10 MB/dokumen; maks 3 dokumen/klaim; maks 1.000 dokumen/batch; konfigurasi per tenant; melebihi limit → 413/422 dengan pesan jelas.
- Target 50 ribu <90 detik; 1 juta <4 jam.
- Pisahkan pool batch dari API agar tidak mengganggu latency.

### 7.7 Auth middleware

- Validasi JWT signature, issuer, audience, expiry, `tenant_id`, `role`.
- JWKS cache dengan TTL; rotasi key.
- Jangan percaya `X-Tenant-ID` dari client; tenant dari token.
- Authorization: RBAC per route; ABAC untuk case/tenant jika diperlukan.
- Audit akses sensitif.
- Untuk service-to-service: mTLS + service account token; bukan user token.

### 7.8 Catatan per service

| Service | Catatan desain |
|---|---|
| Ingestion | Idempotency key; checksum file; validasi MIME/magic bytes; CSV injection escape; fail-fast; laporan per baris |
| Rules | Cache rule pack; evaluasi paralel per claim; deterministik; backtest API |
| Alert | Bundle per claim; priority deterministik; dedupe; queue index |
| Case | Optimistic lock; evidence hash; chain of custody; decision immutable |
| Audit | Append-only; hash chain; export; legal hold; no update/delete role |
| Metering | Event sourcing ringan; period close; reconciliation; billing checksum |
| Reporting | Read replica; materialized view; async export; no PII |
| Notification | Outbox; retry; provider fallback; no sensitive data in email |
| Document | Phase 2; sandbox OCR; engine version; object storage; queue; OCR 1 halaman ≤15 detik; 200 dokumen background <20 menit; antrean tetap responsif |
| ML | Phase 2; model registry; shadow; fallback rules; drift alert + retraining approval; explainability |
| Integration | Phase 2; connector secret; idempotent sync; write-back opt-in |

### 7.9 Error dan retry policy

| Kode | Arti | Retry | Catatan |
|---|---|---|---|
| 400 | Validasi | Tidak | Pesan aman |
| 401 | Auth | Tidak | Refresh token |
| 403 | Authorization | Tidak | Audit |
| 404 | Tidak ditemukan | Tidak | — |
| 409 | Conflict/duplicate | Ya dengan idempotency | Optimistic lock |
| 422 | Semantik | Tidak | Detail field |
| 429 | Rate limit | Ya dengan backoff | `Retry-After` |
| 500 | Internal | Ya terbatas | Alert |
| 503 | Dependency | Ya dengan backoff | Circuit breaker |
| 504 | Timeout | Ya jika idempotent | Budget |

### 7.10 Performance budget

| Operasi | Budget p95 |
|---|---|
| Auth/me | 200 ms |
| List alert | 500 ms |
| Case detail | 800 ms |
| Decision write | 1.000 ms |
| Scoring single claim | 2.000 ms |
| Upload 10 MB | 5.000 ms |
| Report export | 30.000 ms async |
| Batch 50k | 90.000 ms |

### 7.11 Guardrail domain (wajib)

| Guardrail | Implementasi | Verifikasi |
|---|---|---|
| Determinisme | Rule pack immutable; `as_of`; sorting eksplisit; canonical JSON | Backtest 2× diff = 0 |
| No write-back MVP | Tidak ada credential/endpoint write ke PAS/claims; integration Phase 2 opt-in | Code review + network policy |
| No auto-denial | Keputusan hanya manusia; sistem tidak mengubah status klaim | Test + audit |
| No auto-clear | OCR/ML confidence rendah (<60%) dirutekan manual; tidak pernah auto-clear | Test + review |
| Label isolation | `is_fraud_ground_truth` tidak dapat dibaca rule engine/ML training path tanpa izin | Automated test + IAM |
| Tenant isolation | `tenant_id` + RLS + object prefix + cache key | AT-07/pentest |
| Fairness | Atribut terlindungi dikecualikan; bias test per segmen; rule berkorelasi demografi dicatat | Review + monitoring |
| Audit immutability | Append-only + hash chain; role tanpa update/delete | Test + audit |
| Metering immutability | Usage record immutable setelah close; rekonsiliasi 100% | Reconciliation test |
| UI language | "perlu verifikasi/indikasi", bukan "terbukti/palsu" | Review + automated lint |

---

## 8. Desain Frontend (Next.js)

### 8.1 Arsitektur

- **App Router**, React Server Components untuk initial data; client components untuk interaksi.
- **BFF tipis:** Next.js route handlers memanggil API Go; tidak ada business logic, tidak ada akses DB langsung.
- **Session:** HttpOnly, Secure, SameSite=Lax; token disimpan server-side/Redis; refresh rotation.
- **API client:** typed dari OpenAPI; Zod validation; error mapping.
- **No local state persistence:** tidak ada upload ke filesystem; semua ke object storage via presigned URL.
- **Feature flags:** per tenant; kill switch.

### 8.2 Rendering dan caching

| Konten | Strategi | Cache |
|---|---|---|
| Static assets | Immutable, hash | CDN 1 tahun |
| Marketing/login | Static/ISR | CDN |
| Queue/dashboard | Dynamic, tenant-specific | No CDN; short Redis cache |
| Case detail | Dynamic | No cache; ETag |
| Report export | Async | Object storage |
| User session | Dynamic | Redis |

**Aturan:** data tenant tidak boleh di-cache di CDN publik. Gunakan `Cache-Control: private, no-store` untuk halaman dinamis. Untuk multi-instance, gunakan custom cache handler Redis; jangan andalkan cache lokal [R12].

### 8.3 Multi-instance HA

| Kontrol | Baseline |
|---|---|
| Replica | 3+ across AZ; HPA 3–12 |
| Reverse proxy | Envoy Gateway/nginx di depan; jangan expose Next.js langsung [R12] |
| Shared cache | Custom Redis cache handler; `cacheMaxMemorySize: 0`; TTL + bound [R12] |
| Server Actions key | `NEXT_SERVER_ACTIONS_ENCRYPTION_KEY` sama di semua instance [R12] |
| Deployment ID | `deploymentId` untuk version skew [R12] |
| Build ID | `generateBuildId` dari git hash [R12] |
| Health | `/healthz` dan `/readyz`; tidak render halaman |
| Graceful shutdown | SIGTERM; drain 10–30 detik [R12] |
| Image optimization | Loader/CDN; limit memory; atau disable |
| Streaming | Proxy tidak buffer; `X-Accel-Buffering: no` [R12] |
| Static assets | CDN + assetPrefix; immutable |

### 8.4 Keamanan frontend

- CSP strict; nonce; no inline script.
- Cookie HttpOnly/Secure/SameSite; CSRF token untuk mutation.
- XSS: escape; no `dangerouslySetInnerHTML` tanpa sanitasi.
- Upload: validasi tipe/ukuran di server; presigned URL; scan malware.
- IDOR: authorization di API, bukan UI.
- No secret di `NEXT_PUBLIC_*`.
- Dependency scan; SRI untuk third-party.
- Accessibility WCAG 2.1 AA; keyboard; contrast; screen reader.

### 8.5 UX saat degraded

- Tampilkan status "data mungkin tertunda" jika queue lag.
- Retry idempotent; jangan double-submit.
- Error boundary per panel.
- Skeleton/streaming; jangan blank.
- Decision tetap bisa dilakukan jika reporting down; audit/metering async.
- Jika Redis down, login existing session mungkin gagal; arahkan re-login; jangan kehilangan draft.

### 8.6 Performance budget

| Metrik | Target |
|---|---|
| LCP | <2,5 detik |
| INP | <200 ms |
| CLS | <0,1 |
| JS initial | <250 KB gzip |
| API p95 | <2 detik |
| Queue page | <3 detik |
| Alert → case | ≤3 klik |
| Training | ≤1 hari |

---

## 9. Desain Data

### 9.1 Gambaran

| Store | Peran | Source of truth | HA |
|---|---|---|---|
| ApsaraDB RDS PostgreSQL | Transaksi, claim, alert, case, rule, audit, metering | Ya | Multi-zone |
| Tair (Redis) | Cache, session, rate limit, ephemeral | Tidak | Multi-AZ |
| ApsaraMQ for Kafka | Event, outbox, integration, replay | Tidak (event log) | Multi-zone |
| OSS | Dokumen, evidence, export, backup | Ya untuk file | Multi-AZ + CRR |
| Lakehouse (Phase 2) | Data mentah/terkurasi, lineage, training, analitik | Ya untuk analitik | OSS + Iceberg/Trino; multi-AZ |
| Feature store (Phase 2) | Fitur online/offline, point-in-time correctness | Ya untuk fitur | RDS/Tair + Iceberg |
| Observability storage | Metrics/logs/traces | Tidak | SLS/ARMS multi-AZ |

### 9.2 PostgreSQL topology

**Default managed (Alibaba Cloud):**

- **ApsaraDB RDS for PostgreSQL Cluster Edition** untuk GA: primary + multiple secondary, readable secondary, automatic failover [R40].
- **ApsaraDB RDS for PostgreSQL HA Edition** untuk MVP: primary/standby multi-zone; failover <30 detik [R40].
- Writer endpoint untuk write; read-only endpoint/readable secondary untuk reporting.
- RDS database proxy (jika tersedia untuk PostgreSQL) atau PgBouncer untuk pooling [R17].
- Automated backup + PITR; backup cross-region via DBS; retention baseline 35 hari (final `TBD`).
- Parameter: `max_connections`, `statement_timeout`, `idle_in_transaction_session_timeout`, `lock_timeout`, `work_mem`, `maintenance_work_mem`.
- Monitoring: CPU, IOPS, replication lag, connections, locks, bloat, slow query; integrasi ARMS/CloudMonitor.

**Self-managed/private cloud:**

- Patroni + etcd 3 node across 3 AZ; `synchronous_mode: true`; satu sync standby; HAProxy/PgBouncer [R6].
- Minimal 3 etcd; odd number; quorum 2.
- `synchronous_mode_strict: false` untuk availability; RPO 0 selama sync standby sehat; jika standby hilang, tradeoff availability vs RPO harus disetujui.
- Fencing/STONITH untuk cegah split-brain.
- pgBackRest untuk backup/PITR [R29].
- Read replica async untuk reporting.

**Aturan:**

- Jangan akses primary untuk reporting berat.
- Read replica hanya jika lag <5 detik; otherwise fallback.
- Tidak ada long transaction >30 detik; batch chunk.
- Connection budget: total app connections ≤80% `max_connections`; PgBouncer transaction mode.
- RLS: `SET LOCAL app.tenant_id` per transaksi; PgBouncer transaction pooling kompatibel karena setting lokal transaksi [R15][R17].

### 9.3 Multi-tenancy dan RLS

| Model | Isolasi | Kapan |
|---|---|---|
| Shared schema + RLS | Logis, biaya rendah | Default MVP |
| Schema per tenant | Lebih kuat, migrasi lebih berat | Tenant besar |
| Dedicated database/cluster | Paling kuat | Enterprise/private cloud |

**RLS baseline:**

```sql
ALTER TABLE claim ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON claim
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
```

- Semua tabel tenant-owned punya `tenant_id NOT NULL`.
- Role aplikasi bukan superuser; `BYPASSRLS` tidak diberikan.
- Test otomatis: tenant A tidak bisa membaca/menulis tenant B (AT-07).
- Audit akses lintas tenant (admin break-glass) dengan alasan.
- Cache key wajib menyertakan `tenant_id`.
- Object storage prefix per tenant; presigned URL terikat tenant.

### 9.4 Table design dan partition

| Tabel | Service | Partition | Index kunci | Retensi |
|---|---|---|---|---|
| `claim` | ingestion | Range bulanan `claim_date` | `(tenant_id, claim_id)`, `(tenant_id, policy_id)`, `(tenant_id, status)` | 5–10 tahun `TBD` |
| `claim_item` | ingestion | Ikut claim | `(tenant_id, claim_id)` | 5–10 tahun |
| `import_batch` | ingestion | Tidak | `(tenant_id, created_at)` | 24 bulan |
| `rule` | rules | Tidak | `(tenant_id, family, status)` | Permanen/versioned |
| `evaluation` | rules | Range bulanan | `(tenant_id, claim_id, rule_version)` | 5–10 tahun |
| `alert` | alert | Range bulanan | `(tenant_id, status, priority, created_at)` | 5–10 tahun |
| `case` | case | Range bulanan | `(tenant_id, status, owner)` | 5–10 tahun |
| `evidence` | case | Ikut case | `(tenant_id, case_id)` | 5–10 tahun |
| `decision` | case | Ikut case | `(tenant_id, case_id, decided_at)` | 5–10 tahun |
| `audit_event` | audit | Range bulanan | `(tenant_id, object_type, object_id, timestamp)` | ≥5 tahun |
| `usage_record` | metering | Range bulanan | `(tenant_id, period)` | 7 tahun |
| `outbox_event` | semua | Tidak | `(status, next_attempt_at)` | 7 hari setelah terkirim |
| `processed_event` | consumer | Range bulanan | `(consumer, event_id)` | 30 hari |
| `idempotency_key` | semua | Tidak | `(tenant_id, key)` | 24 jam–7 hari |

**Aturan:**

- UUIDv7 untuk PK; waktu UTC.
- Uang integer minor unit; currency eksplisit.
- JSONB untuk payload/evidence; index GIN hanya jika perlu.
- Partition drop untuk retensi; archive ke object storage.
- `created_at`/`updated_at`; `version` untuk optimistic lock.
- Soft delete hanya untuk DSR; hard delete via job teraudit.
- Tidak ada `SELECT *`; query review.

### 9.5 Transaksi dan konsistensi

- **Strong consistency** dalam aggregate/service.
- **Eventual consistency** antar-service.
- **Outbox pattern:** aggregate + outbox dalam satu transaksi; relay publish.
- **Saga:** case closure/decision lintas service; kompensasi idempoten.
- **Idempotency:** header `Idempotency-Key`; simpan request hash + response; unique.
- **Read-your-writes:** response mengembalikan `version`; UI memakai versi; read replica tidak dipakai.
- **No distributed transaction/XA.**
- **Exactly-once:** tidak diklaim; at-least-once + dedupe.

### 9.6 Redis design

| Penggunaan | Key pattern | TTL | Eviction |
|---|---|---|---|
| Session | `sess:{tenant}:{user}:{sid}` | 30 menit idle | No eviction |
| Cache | `cache:{tenant}:{domain}:{id}` | 5–60 menit + jitter | allkeys-lru |
| Rate limit | `rl:{tenant}:{route}:{window}` | window | volatile |
| Queue index | `queue:{tenant}:{priority}` | — | no eviction |
| Next.js cache | `next:{deployment}:{key}` | TTL | allkeys-lru |
| Lock | `lock:{resource}` | short | no eviction |

**HA:**

- **Tair (Redis OSS-compatible) multi-AZ**: primary + replica di AZ berbeda; automatic failover [R41].
- Client harus reconnect ke endpoint baru; gunakan retry/backoff; jangan cache koneksi permanen tanpa health check.
- Jika self-managed: Redis Sentinel 3 node di 3 AZ, quorum 2; `min-replicas-to-write 1`, `min-replicas-max-lag 10` [R7].
- TLS + ACL; no `FLUSHALL` di prod; `SCAN` bukan `KEYS`.
- Cache loss = fallback DB; session loss = re-login; rate limit loss = reset window.
- Tair cluster edition jika dataset/throughput > single primary; 3 shard × 2 replica; hash tags [R8].
- Tair global distributed cache/geo-replication untuk DR jika tersedia dan disetujui [R41].

**Larangan:** Redlock untuk correctness; gunakan fencing token/DB unique. Cache sebagai source of truth. Menyimpan PII mentah di cache.

### 9.7 Kafka design

| Topic | Key | Partitions | Retention | Producer |
|---|---|---:|---:|---|
| `claim.ingested.v1` | `tenant:claim` | 24 | 7 hari | Ingestion |
| `rules.evaluated.v1` | `tenant:claim` | 24 | 7 hari | Rules |
| `alert.created.v1` | `tenant:claim` | 24 | 7 hari | Alert |
| `case.updated.v1` | `tenant:case` | 12 | 30 hari | Case |
| `decision.recorded.v1` | `tenant:case` | 12 | 30 hari | Case |
| `audit.event.v1` | `tenant:object` | 24 | 30 hari | Semua |
| `usage.recorded.v1` | `tenant:period` | 12 | 30 hari | Semua |
| `document.processed.v1` | `tenant:doc` | 12 | 30 hari | Document |
| `dlq.*` | original | 6 | 30 hari | Consumer |

**Konfigurasi:**

- RF=3; `min.insync.replicas=2`; `acks=all`; idempotent producer [R10].
- `unclean.leader.election.enable=false`.
- ApsaraMQ Professional multi-zone; rack awareness per AZ untuk self-managed; verify zone coverage [R46].
- Schema registry; Protobuf/JSON Schema; backward-compatible.
- Consumer group per service; `max.poll.records` bounded.
- Retry topic dengan backoff; DLQ; alert.
- ACL per service; TLS/SASL.
- Quota per tenant; monitor lag/URP/disk.
- ApsaraMQ Professional multi-zone; broker CPU <60% [R46].

**Kafka vs NATS:** Kafka dipilih untuk replay, ekosistem, dan audit; NATS JetStream lebih ringan dan HA dengan R3/odd server [R11]. NATS dapat menjadi alternatif thin slice/private cloud jika operasi Kafka terlalu berat.

### 9.8 Object storage (OSS)

- Bucket terpisah: `documents`, `evidence`, `exports`, `backups`, `logs`.
- Prefix: `tenant/{tenant_id}/claim/{claim_id}/doc/{doc_id}/{version}`.
- SSE-KMS; bucket policy RAM least privilege; block public access.
- Versioning; **retention policy (WORM)** untuk evidence/audit [R47].
- Presigned URL short TTL (≤5 menit); upload langsung; scan malware async.
- Lifecycle: Standard → IA → Archive; retention per kategori.
- CRR + Replication Time Control (RTC) untuk DR [R48].
- Access logging; checksum; hash verification.
- Immutable original; turunan berversi.

### 9.9 Volume dan growth

| Data | Asumsi | Volume/tahun | 5 tahun |
|---|---|---:|---:|
| Claim + item | 10 juta klaim; 10 KB/klaim | 100 GB | 500 GB |
| Evaluation/alert/case | 5× claim; 2 KB | 100 GB | 500 GB |
| Audit/event | 20 event/klaim; 1 KB | 200 GB | 1 TB |
| Index/overhead | 2× | 400 GB | 2 TB |
| **PostgreSQL total** | — | **~800 GB** | **~4 TB** |
| Dokumen | 25% klaim × 2 × 1 MB | 5 TB | 25 TB |
| Versi/CRR | 1,5× + 2× | 15 TB | 75 TB |
| **Object storage** | — | **~20 TB** | **~100 TB** |
| Kafka | 20 event/klaim; 1 KB; 7 hari | ~4 GB | — |

Angka ini **design envelope** untuk satu tenant Enterprise 10 juta klaim/tahun; aggregate 50 juta/tahun = 5×. Validasi dengan data design partner.

### 9.10 Retensi

| Data | Thin slice | Produksi | Catatan |
|---|---|---|---|
| Claim/invoice/policy/provider | 12 bulan | 5–10 tahun `TBD` | Legal |
| Dokumen/forensik | 12 bulan | 5–10 tahun `TBD` | BRD DR-18 |
| Audit | 12 bulan | ≥5 tahun `TBD` | Immutable |
| Usage/billing | 24 bulan | 7 tahun `TBD` | Finance |
| Label/decision | 12 bulan | `TBD` | Feedback |
| PII | Hash/token | Tokenization/vault | Raw NIK dilarang |
| Backup | 30 hari | 35 hari + archive (baseline; final `TBD`) | Uji restore |

### 9.11 Reporting read path

- Read replica untuk dashboard/export.
- Materialized view per tenant/periode; refresh async.
- Export job ke object storage; signed URL; expiry.
- Tidak ada query berat di primary.
- Jika lag >5 detik, tampilkan "data tertunda"; jangan block queue.
- Phase 2: data lakehouse (OSS + Iceberg/Trino) untuk analitik lanjutan; bukan MVP.

### 9.12 Data quality dan determinisme

- Validasi fail-fast per batch; laporan per baris.
- Tolak re-impor identik; checksum.
- `loss_type` hanya MOTOR/PROPERTY di Phase 0/1; currency IDR only; multi-currency Phase 2+.
- Tanggal ISO-8601; `incident_date ≤ claim_date`.
- Identitas hash `sha256:`; raw ditolak.
- Label ground truth terisolasi.
- Backtest deterministic; hash output.
- Metering reconciliation 100%.
- Data quality metrics: reject rate, duplicate rate, missing rate.

### 9.14 Lakehouse, feature store, dan entity resolution (Phase 2)

- **Lakehouse:** OSS + Iceberg/Trino; ingest dari CDC/event; Bronze/Silver/Gold; lineage; akses read-only untuk analitik/training; bukan source of truth transaksi.
- **Feature store:** definisi fitur versioned; offline (Iceberg) dan online (Redis/PostgreSQL); point-in-time join untuk mencegah leakage; freshness SLI; akses per tenant.
- **Entity resolution:** deterministic keys + probabilistic matching; threshold versioned; Golden Record + edge; audit keputusan matching; tidak menimpa data sumber.
- **Retensi:** mengikuti §9.10; PII dimasking; akses per peran.
- **HA:** lakehouse memakai object storage multi-AZ + query engine stateless; feature store online memakai Redis/PostgreSQL HA; entity resolution 3 replica.
- **Sizing:** `TBD` setelah pilot; mulai dari serverless query + OSS; jangan mengklaim kapasitas sebelum benchmark.

### 9.13 Backup matrix

| Komponen | Metode | Frekuensi | Retention | RPO | RTO |
|---|---|---|---|---|---|
| PostgreSQL | Automated snapshot + WAL/PITR | Harian + kontinu | 35 hari | ≤15 menit | ≤2 jam |
| PostgreSQL archive | DBS/pgBackRest + OSS | Harian | 1 tahun | — | ≤4 jam |
| Tair (Redis) | Snapshot (opsional) | Harian | 7 hari | Cache loss OK | ≤30 menit |
| Kafka | RF=3 + tiered/MirrorMaker | Kontinu | 7–30 hari | ≤15 menit | ≤2 jam |
| Object storage | Versioning + CRR | Kontinu | Sesuai retensi | ≤15 menit | ≤1 jam |
| Kubernetes | Velero + IaC | Harian | 30 hari | — | ≤2 jam |
| Secrets | KMS/Vault backup | Harian | 30 hari | — | ≤1 jam |

> Retensi backup final mengikuti keputusan Legal/Compliance (`TBD`); angka di atas adalah baseline teknis, bukan komitmen retensi.

---

## 10. Alur Data & Konsistensi

### 10.1 Ingest

1. Client upload file + `Idempotency-Key`.
2. Gateway validasi ukuran/rate limit; auth.
3. Ingestion cek idempotency; simpan batch; validasi per baris.
4. Dalam transaksi: simpan claim_raw + outbox `claim.ingested.v1`.
5. Relay publish ke Kafka.
6. Rules consume; evaluasi; simpan evaluation + outbox.
7. Alert consume; buat alert/bundle; simpan + outbox.
8. UI poll/SSE; queue muncul.

### 10.2 Review

1. Reviewer buka case; API query read model.
2. Decision write dengan `If-Match`/version.
3. Case simpan decision + outbox.
4. Audit append; metering event; reporting read model.
5. UI optimistic update; rollback jika 409.

### 10.3 Failure semantics

| Titik gagal | Perilaku | Jaminan |
|---|---|---|
| Upload putus | Retry idempotent | Tidak duplikat |
| DB commit gagal | 5xx; tidak ada event | Atomic |
| Outbox publish gagal | Retry; event tetap di DB | At-least-once |
| Kafka duplicate | Consumer dedupe | Efek sekali |
| Rules crash | Kafka offset belum commit; retry | At-least-once |
| Alert crash | Idempotent alert ID | Tidak duplikat |
| Audit down | Event tetap di Kafka/outbox | Eventually |
| Metering down | Event tetap; rekonsiliasi | Eventually |
| Reporting down | Queue tetap jalan | Tidak blocking |

### 10.4 Exactly-once stance

Tidak ada exactly-once end-to-end. Gunakan **at-least-once + idempotent consumer + unique constraint**. Untuk billing, rekonsiliasi harian dan period close. Untuk audit, append-only + hash chain; duplikat dicegah `event_id` unik.

### 10.5 Ordering

- Kafka key `tenant_id:aggregate_id` menjaga urutan per aggregate.
- Tidak ada global ordering.
- Consumer per key parallelism; jangan asumsikan urutan lintas key.
- Outbox relay menjaga urutan per aggregate dengan `sequence`.
- Case decision menggunakan optimistic lock; bukan urutan event.

### 10.6 Read-your-writes

- Write response mengembalikan `version`/`etag`.
- UI memakai response; tidak langsung query read replica.
- Jika perlu query, gunakan primary atau cache dengan versi.
- Read replica hanya untuk reporting.

---

## 11. API & Kontrak Integrasi

### 11.1 Prinsip

- REST/JSON; OpenAPI 3.1 [R30].
- Versioning URL `/v1`; breaking change → `/v2`.
- Auth OIDC/JWT [R31]; API key untuk server-to-server; mTLS opsional.
- Idempotency-Key untuk POST/PATCH tertentu.
- Rate limit per tenant/user/IP; header `X-RateLimit-*`, `Retry-After`.
- Pagination cursor; limit default 50, max 200.
- Error RFC 7807: `type`, `title`, `status`, `detail`, `instance`, `code`, `retryable`.
- ETag/If-Match untuk update.
- Correlation ID `X-Request-ID`.
- Webhook signed HMAC; retry + DLQ.
- OpenAPI contract test di CI.

### 11.2 Endpoint catalog (thin slice/MVP)

| Method | Path | Fungsi | Phase |
|---|---|---|---|
| POST | `/v1/auth/login` | Login lokal/SSO | Thin slice/MVP |
| GET | `/v1/me` | Profil/role | Thin slice/MVP |
| POST | `/v1/imports` | Upload CSV/JSON | Thin slice |
| GET | `/v1/imports/{id}` | Status batch | Thin slice |
| GET | `/v1/alerts` | Queue + filter | Thin slice |
| GET | `/v1/alerts/{id}` | Detail + evidence | Thin slice |
| POST | `/v1/cases` | Buat case group | Thin slice |
| PATCH | `/v1/cases/{id}` | Assign/update | Thin slice |
| POST | `/v1/cases/{id}/decision` | Outcome | Thin slice |
| GET | `/v1/cases/{id}/audit` | Audit trail | Thin slice |
| GET | `/v1/coverage` | Coverage matrix | Thin slice |
| GET | `/v1/usage` | Usage record | Thin slice |
| GET | `/v1/reports/value` | Value report | Thin slice |
| GET | `/v1/billing/export` | Billing export | Thin slice |
| POST | `/v1/documents` | Upload dokumen | Phase 2 |
| GET | `/v1/scores/{claim}` | ML score | Phase 2 |
| POST | `/v1/connectors` | Integrasi | Phase 2 |

### 11.3 Event catalog

| Event | Payload minimum | Producer |
|---|---|---|
| `claim.ingested.v1` | tenant_id, claim_id, batch_id, checksum, occurred_at | Ingestion |
| `rules.evaluated.v1` | tenant_id, claim_id, rule_pack_version, alerts | Rules |
| `alert.created.v1` | tenant_id, alert_id, claim_id, severity, reasons | Alert |
| `case.updated.v1` | tenant_id, case_id, status, owner, version | Case |
| `decision.recorded.v1` | tenant_id, case_id, outcome, reason_code, reviewer | Case |
| `audit.event.v1` | tenant_id, actor, action, object, hash | Semua |
| `usage.recorded.v1` | tenant_id, period, metric, quantity | Semua |
| `document.processed.v1` | tenant_id, doc_id, status, engine_version | Document |

### 11.4 API gateway/BFF

- Gateway: TLS, WAF, JWT, rate limit, routing, request size, timeout.
- BFF: agregasi read-only untuk UI; tidak menyimpan business state.
- Timeout budget: gateway 3s; BFF 2,5s; service 2s.
- Circuit breaker per upstream.
- Retry hanya GET/idempotent.
- Response cache hanya untuk public/static.

### 11.5 Integrasi Phase 2

- Connectors: REST, SFTP, webhook, ACORD opsional.
- Sync state; checkpoint; idempotent.
- Write-back opt-in; approval; audit.
- External data: legal basis, contract, rate limit, fallback.
- DMS: presigned upload/download; metadata.
- Tidak ada write-back di MVP.

---

## 12. Keamanan & Kepatuhan

### 12.1 Threat model (STRIDE)

| Ancaman | Contoh | Kontrol |
|---|---|---|
| Spoofing | Token palsu, tenant palsu | OIDC, JWT validation, mTLS, tenant claim |
| Tampering | Modifikasi audit/evidence | Append-only, hash chain, OSS WORM |
| Repudiation | Deny decision | Audit immutable, digital timestamp |
| Information disclosure | Lintas tenant, PII log | RLS, masking, log redaction, encryption |
| Denial of service | Flood upload/API | WAF, rate limit, quota, autoscale |
| Elevation of privilege | Role escalation | RBAC, least privilege, approval |
| Supply chain | Image/dependency compromised | Scan, sign, SBOM, admission |

### 12.2 Identity dan authorization

- OIDC Authorization Code + PKCE; MFA untuk admin/supervisor [R31].
- JWT short-lived (≤15 menit); refresh rotation; revoke.
- RBAC: reviewer, supervisor, admin, finance read-only, auditor.
- ABAC: tenant, line, region, case assignment.
- Service-to-service: mTLS + service account; bukan user token.
- Break-glass admin: time-bound, approval, audit.
- Session idle ≤30 menit (BRD NFR-08).

### 12.3 Tenant isolation

- `tenant_id` di token dan semua tabel.
- RLS + role non-superuser.
- Cache key/object prefix per tenant.
- Rate limit/quota per tenant.
- Quota: batch, dokumen, API, user aktif; peringatan 80%/100%; tidak ada silent block; penolakan dengan pesan jelas; audit (AT-15).
- Dedicated DB/cluster untuk enterprise.
- Test IDOR/tenant escape di CI dan pentest.
- Audit akses lintas tenant.

### 12.4 Enkripsi dan key management

- TLS 1.2+; mTLS internal.
- At rest: PostgreSQL/Kafka/Redis/object storage SSE; KMS.
- Envelope encryption; key rotation tahunan/90 hari untuk data key.
- Tenant-specific data key opsional untuk enterprise.
- Secrets: External Secrets/Vault; tidak di repo/image.
- Backup terenkripsi; restore key tersedia.
- Crypto-shredding untuk DSR jika diperlukan.

### 12.5 Network security

- VPC terpisah per environment; private subnet.
- Security group least privilege; no public DB/Redis/Kafka.
- NetworkPolicy default deny.
- WAF managed rules; rate limit; bot control [R44].
- DDoS protection.
- Egress via NAT per AZ/VPC endpoint.
- Bastion tidak ada; SSM/Session Manager.
- mTLS service-to-service (GA).

### 12.6 Application security

- OWASP ASVS L2 [R26].
- Input validation; output encoding; parameterized query.
- File upload: MIME + magic bytes + size + extension; antivirus; sandbox.
- CSV injection escape.
- SSRF: allowlist URL; no metadata endpoint.
- XSS/CSP; CSRF token.
- IDOR: authorization server-side.
- Rate limit; account lockout.
- Dependency scan; SAST/DAST; secret scan.
- Pen test tahunan; bug bounty opsional.
- Security headers: HSTS, CSP, X-Content-Type-Options, Referrer-Policy.

### 12.7 Supply chain

- Build reproducible; pinned dependencies.
- SBOM (CycloneDX).
- Trivy scan; fail on critical/high.
- Cosign sign; verify admission.
- Image digest, bukan tag.
- Base image minimal/distroless.
- Renovate/Dependabot; patch SLA.
- Artifact registry private; retention.

### 12.8 Audit dan immutability

- Audit event append-only; role tanpa UPDATE/DELETE.
- Hash chain per tenant/periode; anchor ke OSS dengan retention policy WORM.
- Audit berisi actor, action, object, before/after (redacted), timestamp, correlation_id.
- Export signed; legal hold.
- Retensi ≥5 tahun `TBD`.
- Audit tidak boleh menjadi bottleneck; async event.

### 12.9 Privacy

- UU PDP: consent/legal basis, DSR, breach notification, DPO.
- Data minimization; purpose limitation.
- PII/PHI masking; raw NIK ditolak.
- Tokenization/vault produksi; hash thin slice.
- Retensi per kategori; deletion job teraudit.
- Data residency Indonesia.
- DPIA sebelum pilot produksi.
- Vendor DPA; subprocessor list.

### 12.10 Compliance mapping

| Requirement | Kontrol teknis | Bukti |
|---|---|---|
| UU PDP | Encryption, DSR, retention, breach | DPIA, audit, test |
| OJK consumer protection | No auto-denial, audit, wording | Test, review |
| UU ITE | Audit integrity, e-signature | Hash chain, timestamp |
| ISO 27001 | ISMS controls, access, incident | Audit, policy |
| SOC 2 | Availability, security, confidentiality | Report |
| OWASP ASVS L2 | App security controls | Pentest, SAST/DAST |
| NIST AI RMF | Model governance (Phase 2) | Model card, monitoring |

### 12.11 Incident response

- Severity: SEV1 data breach/outage; SEV2 partial; SEV3 degraded.
- On-call page; incident commander; comms lead.
- Containment, eradication, recovery, postmortem.
- Breach notification sesuai UU PDP dan kontrak.
- Evidence preservation; legal hold.
- Tabletop quarterly.

---

## 13. Observability & SLO

### 13.1 SLI/SLO

| SLI | Definisi | SLO |
|---|---|---|
| Availability | 2xx+3xx / total valid | 99.95% GA; 99.9% MVP |
| Latency | p95 request duration | <2 detik API; <3 detik UI |
| Error rate | 5xx / total | <0.1% |
| Saturation | CPU/mem/conn | <60% steady |
| Batch duration | 1 juta klaim | <4 jam |
| Queue age | alert oldest | <5 menit |
| Replication lag | DB | <5 detik |
| Kafka lag | consumer | <5 menit |
| URP | under-replicated partitions | 0 |
| Metering recon | usage vs event | 0 selisih |

### 13.2 Metrics/logs/traces

- OTel SDK; OTLP ke collector [R21].
- Metrics: RED (rate, error, duration) + USE (utilization, saturation, errors).
- Logs: JSON; correlation_id; tenant_id; no PII; sampling.
- Traces: 10% head + tail sampling error/slow; span per dependency.
- Exemplars ke trace.
- Audit log terpisah dari app log.

### 13.3 Dashboard dan alert

| Dashboard | Isi |
|---|---|
| Platform | Availability, latency, error, saturation, SLO burn |
| Service | RED, dependency, queue, DB |
| Data | Replication lag, connections, locks, disk |
| Kafka | URP, lag, throughput, disk |
| Tair (Redis) | Memory, hit rate, eviction, failover |
| Batch | Progress, duration, rejects, backlog |
| Business | Claims, alerts, decisions, queue age |
| Tenant | Usage, quota, value report |
| Security | Auth failures, WAF, audit anomalies |

**Alerting:** multi-window burn rate (1h/6h/3d); page only for SLO/user impact; ticket for capacity. Runbook link di setiap alert.

### 13.4 HA observability

- OTel collector 2–3 per AZ; queue + retry.
- Prometheus HA pair/Thanos/Mimir; dedupe.
- Loki/Tempo object storage; replicated.
- Grafana 2+ replica; Alertmanager 3 replica.
- Observability di namespace/node terpisah; jangan berebut dengan app.
- Synthetic checks dari luar region.
- Status page.

### 13.5 Error budget policy

- Burn >2×/1h → warning.
- Burn >10×/1h → page.
- Budget 50% dalam 7 hari → freeze non-fix.
- Budget 100% → freeze fitur.
- Postmortem untuk SEV1/SEV2; action item.

### 13.6 Business observability

- Claims ingested/scored; alert rate; confirmation rate; queue age.
- Batch duration vs NFR.
- Metering reconciliation; billing export checksum.
- Data quality: rejects/duplicates.
- Rule pack version/determinism.
- Tenant usage/quota.

---

## 14. Kapasitas, Concurrency, dan Sizing

### 14.1 Design envelope

| Parameter | Nilai | Sumber/asumsi |
|---|---:|---|
| Klaim/tahun per tenant Enterprise | 10.000.000 | PRD NFR-3 |
| Aggregate design envelope | 50.000.000/tahun | Asumsi 5 tenant Enterprise |
| Concurrent users | 1.000 | PRD NFR-3 |
| Peak API | 500 rps; burst 1.000 rps 60s | Asumsi 0,5 req/s/user + burst |
| Batch stress | 50.000 klaim/90s | BRD NFR-01 |
| Batch target | 1.000.000 klaim/4h | PRD NFR-1 |
| Dokumen | 25% klaim × 2 × 1 MB | Asumsi |
| Peak factor | 10× average | CAT surge/seasonal |

### 14.2 Formula

**API concurrency (Little's Law):**

```
concurrency = rps × latency_seconds
500 × 0,3 = 150 in-flight
× 2 headroom = 300
6 pod × 50 concurrency = 300
```

**Batch CPU:**

```
cores = claims_per_sec × rules_per_claim × ms_per_rule / (1000 × target_utilization)
50.000/90 = 556 claims/s
556 × 33 × 0,5 / (1000 × 0,5) = 18,3 cores
```

Jika benchmark rule >0,5 ms, cores naik proporsional. Batch pool 16–24 vCPU; benchmark gate wajib.

**DB writes:**

```
writes/s = claims/s × rows_per_claim
556 × 20 = 11.120 writes/s burst 90s
69 × 20 = 1.380 writes/s untuk 1 juta/4h
```

Gunakan batch insert, 16 vCPU, NVMe/ESSD PL1 12k IOPS; monitor.

**Storage:**

```
structured = claims × KB_per_claim × years × index_factor
10M × 10KB × 5 × 2 = 1 TB
documents = claims × doc_ratio × docs_per_claim × MB × years × version_factor
10M × 0,25 × 2 × 1MB × 5 × 1,5 = 37,5 TB
```

**Kafka:**

```
events/s = claims/year / 31.536.000 × events_per_claim
10M / 31,5M × 20 = 6,3 events/s average; peak 63/s
```

### 14.3 Sizing table

| Komponen | Thin slice | Commercial MVP | GA |
|---|---|---|---|
| Kubernetes | Docker Compose 1 node 8 vCPU/32 GB | ACK Pro 3 AZ; 3× app node 8 vCPU/32 GB; 3× batch 8 vCPU/32 GB | ACK Pro 3 AZ; 6× app 16 vCPU/64 GB; 3× batch 16 vCPU/64 GB |
| Next.js | 1 container 1 vCPU/1 GB | 3–6 pod 1 vCPU/1 GB | 6–12 pod 2 vCPU/2 GB |
| Go services | 6 container 0,5 vCPU/512 MB | 3 pod/service 0,5 vCPU/512 MB | 3–6 pod/service 1 vCPU/1 GB |
| PostgreSQL | 1 node 4 vCPU/16 GB | ApsaraDB RDS 8 vCPU/64 GB; readable secondary; 1 TB ESSD | ApsaraDB RDS 16 vCPU/128 GB; readable secondary; 4 TB; 12k IOPS |
| Redis | 1 node 2 GB | Tair 4 GB multi-AZ (primary+replica) | Tair 8–16 GB multi-AZ |
| Kafka | 1 broker KRaft 2 vCPU/4 GB | ApsaraMQ 3 broker 4 vCPU/16 GB/1 TB | ApsaraMQ 3 broker 8 vCPU/32 GB/2 TB |
| Object storage | MinIO 1 node 100 GB | OSS 20 TB + CRR | OSS 100 TB + CRR |
| Observability | Local logs | 3× OTel 1 vCPU/2 GB; managed metrics | 3× OTel 2 vCPU/4 GB; HA Prometheus/Mimir, Loki, Tempo |
| Estimated vCPU | 8 | ~50 | ~150 |

Angka ini bukan biaya; harga `TBD`. **Lakehouse/feature store Phase 2:** sizing `TBD`; mulai dari OSS + query engine serverless, skala setelah pilot.

### 14.4 Autoscaling

| Workload | Metric | Min | Max | Behavior |
|---|---|---:|---:|---|
| Next.js | CPU 60% + RPS | 3 | 12 | scale up 60s; down 300s |
| Gateway | RPS + connections | 3 | 12 | scale up 30s |
| Go API | CPU 60% + RPS | 3 | 12 | scale up 60s |
| Batch worker | queue lag | 3 | 24 | scale up 120s; down 600s |
| Consumer | lag | 3 | 18 | scale up 60s |
| Reporting | CPU | 2 | 6 | scale up 120s |

Cluster autoscaler/Karpenter; node pool buffer 20%; PDB membatasi disruption.

### 14.5 DB connection budget

| Service | Pods | Conn/pod | Total |
|---|---:|---:|---:|
| identity | 3 | 10 | 30 |
| ingestion | 3 | 20 | 60 |
| rules | 3 | 20 | 60 |
| alert | 3 | 10 | 30 |
| case | 3 | 10 | 30 |
| audit | 3 | 10 | 30 |
| metering | 3 | 10 | 30 |
| reporting | 3 | 10 | 30 |
| **Total app** | — | — | **300** |
| PgBouncer | — | — | 400 |
| DB max | — | — | 500 |

Sisakan 20% untuk admin/maintenance. Jangan `max_connections` besar tanpa pool.

### 14.6 Performance budget dan benchmark

| Test | Target | Gate |
|---|---|---|
| 50k batch | <90 detik | Wajib thin slice |
| 1M batch | <4 jam | Wajib MVP |
| API 500 rps | p95 <2s; error <0,1% | Wajib MVP |
| 1.000 users | UI p95 <3s | Wajib MVP |
| Soak 24h | no leak; stable | Wajib GA |
| DB failover under load | <2 menit; no data loss | Wajib GA |
| AZ failure under load | <5 menit; error <1% | Wajib GA |
| Tair failover | <60 detik | Wajib GA |
| ApsaraMQ broker failure | no loss acked; lag recover | Wajib GA |
| DR drill | RTO <4h; RPO <15m | Wajib GA |

### 14.7 Cost guardrails

- Budget per environment; alert 80%/100%.
- Right-sizing review bulanan.
- Object storage lifecycle; log sampling.
- Kafka retention; tiered storage.
- Read replica hanya yang diperlukan.
- `TBD` untuk harga; jangan mengklaim gross margin tanpa data.

---

## 15. Backup, Disaster Recovery, dan Business Continuity

### 15.1 Objective

| Skenario | RTO | RPO | Cakupan |
|---|---:|---:|---|
| Pod failure | <1 menit | 0 | Service |
| Node failure | <5 menit | 0 | Service |
| AZ failure | <5 menit | 0 (DB sync); cache loss | Region |
| DB corruption | <2 jam | ≤15 menit | Data |
| Region failure | ≤4 jam | ≤15 menit | Platform |
| Ransomware | ≤4 jam | ≤15 menit | Data + config |
| Bad deploy | <10 menit | 0 | Service |

### 15.2 Backup

- PostgreSQL: automated backup + WAL/PITR 35 hari; pgBackRest untuk self-managed [R29]; cross-region copy.
- Object storage: versioning + CRR; retention policy WORM untuk audit/evidence [R47][R48].
- Kafka: ApsaraMQ RF=3; MirrorMaker 2/replication tool jika tersedia; event replay dari Postgres outbox jika perlu.
- Kubernetes: Velero + IaC; secrets backup.
- Restore test: bulanan sample; quarterly full; catat waktu.
- Backup encryption; key terpisah; akses terbatas.
- Backup tidak boleh di region yang sama dengan primary jika residensi mengizinkan.

### 15.3 DR topology

**Warm standby in-country:**

- DR site in-country via co-location/private cloud, atau region Alibaba terdekat (mis. Singapore) hanya dengan persetujuan Legal/tenant; Kubernetes, database read replica, Tair, ApsaraMQ, object storage CRR.
- DNS failover; health check.
- Secrets/KMS replicated.
- IaC siap deploy; image registry replicated.
- Runbook dan tim on-call.

**Urutan failover:**

| Langkah | Durasi target | Aksi |
|---|---:|---|
| 1. Deklarasi | 15 menit | Incident commander; freeze write; comms |
| 2. Isolasi | 10 menit | Pastikan primary tidak reachable; fencing |
| 3. Promote DB | 30 menit | Promote read replica; verifikasi lag |
| 4. Deploy app | 60 menit | Argo CD sync DR; config/secret |
| 5. Verifikasi data | 30 menit | Checksum; sample; audit |
| 6. DNS cutover | 15 menit | Alibaba Cloud DNS failover; TTL rendah |
| 7. Smoke test | 30 menit | Login, queue, decision, export |
| 8. Monitor | 60 menit | SLO; error; rollback plan |
| **Total** | **≤4 jam** | — |

### 15.4 Split-brain dan fencing

- Jangan promote jika primary masih reachable.
- Gunakan STONITH/fencing untuk self-managed; managed failover mengikuti provider.
- Satu writer; DNS/connection string mengarah ke writer.
- Setelah failover, old primary di-rebuild sebagai replica; jangan dual-write.
- Untuk Patroni standby cluster, manual promote dan pastikan source down [R6].
- Untuk Kafka, `unclean.leader.election=false`.
- Untuk self-managed Redis, quorum 2; jangan write di minority partition. Untuk Tair multi-AZ, ikuti automatic failover dan jangan intervensi manual.

### 15.5 Restore test

- Restore DB ke environment terisolasi; verifikasi checksum, row count, audit chain.
- Restore object sample; verifikasi hash.
- Ukur RTO aktual; catat gap.
- Uji restore dari backup terenkripsi.
- Uji DSR deletion + legal hold.
- Dokumentasikan hasil di QA Report.

### 15.6 Data residency

- DR in-country; jika tidak ada region kedua, gunakan co-location/private cloud Indonesia.
- Cross-border DR hanya dengan persetujuan Legal dan tenant; encrypt + contract.
- Backup/log/object storage mengikuti residensi.
- Data residency test (AT-08).

### 15.7 Business continuity

- Kritikal: review queue, decision, audit, metering.
- Non-kritikal: reporting, notification, export.
- Manual fallback: reviewer dapat bekerja dari export queue jika UI down? Tidak untuk MVP; dokumentasikan.
- Comms: status page, email, account manager.
- Vendor contact; escalation.

---

## 16. CI/CD, Release, dan Environment

### 16.1 Environment

| Environment | Tujuan | Data | Akses |
|---|---|---|---|
| dev | Developer | Sintetis | Tim |
| sandbox | Integrasi tenant/API | Sintetis | Tenant + tim |
| staging | Pre-prod/HA test | Sintetis/anonymized | Tim + QA |
| prod | Pilot/GA | Tenant | Restricted |
| dr | Failover | Replikasi | SRE |

Akun cloud terpisah; VPC terpisah; IAM least privilege; tidak ada data prod di dev.

### 16.2 Pipeline

1. Lint/format/typecheck.
2. Unit test + coverage gate.
3. SAST + secret scan + dependency scan.
4. Build image; SBOM; Trivy; Cosign sign.
5. Contract test (OpenAPI/event schema).
6. Integration test (Testcontainers).
7. Push image + chart.
8. Argo CD sync staging.
9. E2E/load/chaos staging.
10. Manual approval prod.
11. Argo Rollouts canary prod.
12. Smoke + SLO analysis; promote/rollback.

### 16.3 GitOps

- Argo CD; semua manifest di Git [R23].
- Environment overlay (Kustomize/Helm).
- Drift detection; auto-sync staging; manual prod.
- Sealed secrets/External Secrets; no plaintext.
- Policy as code (Kyverno).

### 16.4 Progressive delivery

- Canary 5% → 25% → 50% → 100%.
- Analysis: error rate, p95, saturation, business KPI.
- Auto rollback jika error > baseline + threshold.
- Feature flag untuk fitur berisiko.
- Blue-green untuk perubahan besar/DB.

### 16.5 DB migration

- Expand/contract; backward-compatible.
- Migration job terpisah; lock timeout; no long lock.
- Backfill async; idempotent.
- Rollback plan; tested di staging.
- Tidak drop column dalam release yang sama.
- Version tabel; migration checksum.

### 16.6 Rollback

- App: Argo Rollouts rollback <10 menit.
- DB: restore/PITR atau reverse migration; butuh approval.
- Config: Git revert.
- Feature flag kill switch.
- Postmortem.

### 16.7 IaC

- Terraform; state remote + lock.
- Modul per environment.
- Plan review; policy check.
- Drift detection.
- Tidak ada manual console untuk prod (kecuali break-glass).

### 16.8 Release checklist

- [ ] Tests lulus; coverage.
- [ ] Security scan; no critical/high.
- [ ] Contract test.
- [ ] Migration reviewed.
- [ ] Runbook updated.
- [ ] Dashboard/alert updated.
- [ ] Capacity checked.
- [ ] Rollback tested.
- [ ] Change log/comms.
- [ ] Approval.

---

## 17. Test Strategy & Validasi HA

### 17.1 Test pyramid

| Level | Cakupan | Tool |
|---|---|---|
| Unit | Logic, rules, validation | `go test`, Vitest |
| Integration | DB/Redis/Kafka/object storage | Testcontainers |
| Contract | OpenAPI/event schema | Schemathesis, Pact |
| E2E | Queue → decision → audit | Playwright |
| Performance | Load/soak/spike | k6 [R25] |
| Chaos | Pod/node/AZ/dependency | Chaos Mesh/Litmus [R24] |
| Security | SAST/DAST/pentest | Semgrep, ZAP, external |
| DR | Restore/failover | Runbook drill |

### 17.2 Functional/acceptance

- Semua AC BRD §13 dan AT PRD §18.3.
- Metering reconciliation, billing checksum, value report reproducibility.
- Determinism/backtest 2×.
- Tenant isolation.
- No auto-denial; no write-back.
- UI wording guardrail ("flag bukan bukti").

### 17.3 Performance test

| Test | Beban | Pass |
|---|---|---|
| Thin slice batch | 50k/90s | Selesai; deterministik |
| MVP batch | 1M/4h | Selesai; no loss |
| API load | 500 rps 30 menit | p95 <2s; error <0,1% |
| UI load | 1.000 users | p95 <3s |
| Spike | 0→1.000 rps 60s | no crash; recover |
| Soak | 24 jam | no leak; stable |
| DB failover | under load | <2 menit; no loss |
| AZ failure | under load | <5 menit; error <1% |

### 17.4 Chaos test

| Eksperimen | Harapan | Pass |
|---|---|---|
| Kill pod | Restart; no user impact | Error <0,1% |
| Drain node | Reschedule; PDB | No outage |
| Kill AZ | 2 AZ handle; DB/Tair failover | <5 menit |
| DB primary kill | Failover; retry | <2 menit; RPO 0 |
| Tair primary kill | Multi-AZ auto-failover | <60 detik |
| Kafka broker kill | Leader election | No loss acked |
| Network latency | Timeout/retry | No cascade |
| Network partition | Quorum; no split-brain | No dual write |
| Disk full | Alert; recover | No corruption |
| Dependency 5xx | Circuit breaker | Graceful |
| Bad deploy | Canary rollback | <10 menit |

### 17.5 DR test

- Quarterly full DR drill; monthly restore sample.
- Ukur RTO/RPO aktual.
- Verifikasi DNS, KMS, secrets, object storage, Kafka.
- Verifikasi audit chain dan metering.
- Catat gap; action item.

### 17.6 Security test

- SAST/DAST setiap PR/release.
- Dependency scan; secret scan.
- Pentest tahunan; retest critical/high.
- Tenant isolation/IDOR.
- Upload malware/CSV injection/SSRF.
- Auth/session/MFA.
- Audit immutability.

### 17.7 Data test

- Data quality rules.
- Determinism/backtest.
- Label leakage.
- Metering reconciliation.
- Billing export checksum.
- Retention/deletion.
- Backup restore.

### 17.8 Game day

- Skenario: AZ down, DB failover, Kafka broker loss, bad deploy, dependency outage.
- Peserta: engineering, SRE, QA, product, comms.
- Output: timeline, gap, action, runbook update.
- Frekuensi: sebelum GA, lalu quarterly.

---

## 18. Runbook Operasional

### 18.1 On-call

- Rotasi 24/7; primary/secondary.
- Escalation: SRE → Lead → CTO.
- Severity: SEV1 (data breach/outage), SEV2 (partial), SEV3 (degraded).
- Page hanya untuk user impact/SLO burn.
- Status page update.

### 18.2 Runbook index

| Runbook | Trigger | RTO |
|---|---|---|
| DB failover | Primary down/lag | <2 menit |
| Tair failover | Multi-AZ failover event | <60 detik |
| Kafka broker loss | URP/offline | <5 menit |
| AZ outage | AZ health | <5 menit |
| Region DR | Region down | ≤4 jam |
| Bad deploy | Canary fail | <10 menit |
| Queue backlog | Lag >5 menit | <30 menit |
| Disk full | Disk >85% | <30 menit |
| Cert expiry | Cert <14 hari | <15 menit |
| KMS outage | Encrypt/decrypt fail | <30 menit |
| Tenant isolation | Suspected breach | Immediate |
| Data breach | PII leak | Immediate |
| Restore | Corruption | <2 jam |

### 18.3 DB failover runbook

1. Cek alert: primary unreachable/lag.
2. Cek managed console/Patroni state.
3. Pastikan failover otomatis berjalan; jangan intervensi jika sehat.
4. Jika manual: promote standby; update endpoint; verifikasi.
5. Cek aplikasi reconnect; error rate; latency.
6. Cek data: sync lag; last transaction; audit.
7. Rebuild old primary sebagai replica.
8. Postmortem.

### 18.4 Tair (Redis) failover runbook

1. Cek event failover Tair (console/event log) dan `READONLY`/connection error pada client.
2. Verifikasi master baru; client reconnect.
3. Cek session/login; cache hit; rate limit.
4. Jika failover tidak selesai dalam SLO, eskalasi ke support Alibaba Cloud; jangan force failover manual.
5. Cek data loss window; re-login jika perlu.
6. Postmortem.

### 18.5 Kafka broker loss

1. Cek URP/offline partitions.
2. Pastikan leader election; producer retry.
3. Jangan restart semua broker.
4. Cek disk/CPU; rebalance.
5. Cek consumer lag; DLQ.
6. Postmortem.

### 18.6 AZ outage

1. Cek cloud status; AZ health.
2. Pastikan traffic ke 2 AZ; HPA scale.
3. Cek DB/Tair/ApsaraMQ failover.
4. Cek capacity 67%.
5. Jangan failback sebelum stabil.
6. Postmortem.

### 18.7 Region DR

Ikuti §15.3. Incident commander; freeze write; fencing; promote; deploy; verify; DNS; smoke; monitor.

### 18.8 Bad deploy

1. Hentikan rollout.
2. Rollback Argo.
3. Cek DB migration compatibility.
4. Cek error/latency.
5. Feature flag off.
6. Postmortem.

### 18.9 Queue backlog

1. Cek consumer lag; error; DLQ.
2. Scale consumer; cek DB/Kafka.
3. Prioritaskan high severity.
4. Jangan drop tanpa audit.
5. Postmortem.

### 18.10 Disk full

1. Cek volume; log/temp.
2. Expand storage; cleanup aman.
3. Cek DB WAL/Kafka retention.
4. Jangan hapus audit/backup.
5. Postmortem.

### 18.11 Certificate

1. Cek cert manager; expiry.
2. Renew; reload; verify.
3. Cek client.
4. Postmortem.

### 18.12 KMS

1. Cek KMS status; retry.
2. Gunakan data key cache jika aman.
3. Failover key/region.
4. Cek backup restore.
5. Postmortem.

### 18.13 Tenant isolation/breach

1. Isolasi akses; revoke token.
2. Preserve evidence; legal.
3. Assess scope; notify.
4. Patch; test.
5. Postmortem; regulator notification.

### 18.14 Restore

1. Pilih backup/PITR.
2. Restore ke isolated environment.
3. Verifikasi checksum/row count/audit.
4. Switch; smoke.
5. Postmortem.

---

## 19. Alternatif Teknologi & ADR

### 19.1 ADR ringkas

| ADR | Keputusan | Alternatif | Alasan | Konsekuensi |
|---|---|---|---|---|
| ADR-001 | Kubernetes multi-AZ di ACK Pro | ACK Serverless, ECS, self-managed k8s | Ekosistem, portabilitas, HA terkelola | Kompleksitas operasi |
| ADR-002 | ApsaraDB RDS PostgreSQL Cluster/HA Edition multi-zone | PolarDB, self-managed Patroni | RTO/RPO, operasi | Biaya, lock-in |
| ADR-003 | Tair multi-AZ | Tair cluster, Redis Cluster, Memcached | Cukup untuk cache/session; HA terkelola | Scale limit |
| ADR-004 | ApsaraMQ for Kafka | NATS JetStream, RabbitMQ | Replay, ekosistem, audit | Operasi berat |
| ADR-005 | Next.js self-host di ACK | Vercel | Data residency, kontrol | Ops frontend |
| ADR-006 | RLS shared schema | Schema/DB per tenant | Biaya, skala | Risiko isolation |
| ADR-007 | OSS + WORM | MinIO | Durabilitas, CRR | Biaya |
| ADR-008 | ASM (Istio-based) | Linkerd, Cilium, no mesh | mTLS, traffic | Kompleksitas |
| ADR-009 | OTel stack | Vendor APM | Portabilitas | Ops |
| ADR-010 | Warm standby DR | Active-active | Biaya, residensi | RTO 4 jam |

### 19.2 Perbandingan detail

**Compute**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| ACK Pro (Kubernetes) | HA, autoscale, ekosistem, managed control plane | Kompleks | Default |
| ACK Serverless (ECI) | Tanpa node, burst | Kontrol terbatas, cold start | Burst/edge |
| ECS (VM-based) | Familiar, kontrol | Tidak cloud-native | Legacy |
| Self-managed k8s | Portabilitas | Operasi berat | Private cloud |

**Database**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| ApsaraDB RDS PostgreSQL HA/Cluster Edition | Managed failover, PITR, readable secondary | Biaya, fitur terbatas | Default |
| PolarDB for PostgreSQL | Performa, storage terpisah, HA | Lock-in, biaya | Skala besar |
| Patroni | Kontrol, on-prem | Operasi | Private cloud |
| CockroachDB | Multi-region | Kompleks, biaya | Active-active |

**Cache**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| Tair multi-AZ | HA terkelola, Redis-compatible | Single primary | Default |
| Tair cluster/sharded | Sharding | Kompleks | > single primary |
| Memcached | Sederhana | No persistence/HA | Cache murni |
| Self-managed Redis Sentinel | Kontrol, on-prem | Operasi | Private cloud |

**Event backbone**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| ApsaraMQ for Kafka | Replay, ekosistem, audit, multi-zone | Operasi | Default |
| NATS JetStream | Ringan, HA R3 | Ekosistem lebih kecil | Thin slice/private |
| RabbitMQ | Routing, mature | Replay terbatas | Task queue |
| Apache Pulsar | Multi-tenant, tiered | Kompleks | Skala besar |

**Frontend**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| Next.js self-host | Kontrol, residensi | Ops | Default |
| Vercel | DX, edge | Residensi/biaya | Non-regulated |
| SPA + API | Sederhana | SEO/streaming | Internal |

**Tenancy**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| RLS shared | Murah, skala | Risiko isolation | Default |
| Schema per tenant | Isolation | Migrasi | Tenant besar |
| DB per tenant | Isolation max | Biaya | Enterprise |

**Service mesh**

| Opsi | Pro | Kontra | Kapan |
|---|---|---|---|
| ASM (Istio-based) | mTLS, traffic, telemetry, managed | Kompleks | GA |
| Istio self-managed | Kontrol penuh | Operasi berat | Private cloud |
| Linkerd | Ringan | Fitur | MVP |
| Cilium | eBPF, network | mTLS terbatas | Network-centric |

### 19.3 Revisit trigger

- Tenant >10 atau aggregate >50M klaim/tahun → evaluasi sharding/partition.
- Redis memory >60% → Cluster.
- Kafka throughput >60% → broker/partition.
- DB CPU >60% steady → scale/read replica.
- DR RTO tidak tercapai → active-active/in-region DR.
- Team tidak mampu operasi Kafka → NATS/managed.
- Compliance cross-border → DR topology.

### 19.4 Portability dan exit

- Export PostgreSQL (dump/CSV) + object storage (asli + manifest hash) + Kafka/outbox event.
- Export konfigurasi: rule pack, tenant config, integration mapping; tanpa secret.
- Export model (Phase 2): ONNX/artifact + model card + feature definition.
- Format terbuka; dokumentasi skema; checksum; prosedur restore di lingkungan baru.
- Target exit ≤30 hari setelah kontrak berakhir; data dihapus setelah retensi.
- Uji exit di staging setiap akhir phase; hasil masuk QA Report.

---

## 20. Risiko Teknis & Mitigasi

| ID | Risiko | Dampak | Mitigasi | Owner |
|---|---|---|---|---|
| TR-01 | Kompleksitas microservices/Kafka | Operasi berat, bug | Managed service, SRE, observability | CTO |
| TR-02 | DB hotspot/tenant noisy | Latency | Partition, pool, rate limit, dedicated | Data |
| TR-03 | Tair/Redis loss | Session/cache down | Tair multi-AZ, fallback DB, re-login | SRE |
| TR-04 | Kafka rebalance/lag | Delay alert | Tuning, partition, monitor | SRE |
| TR-05 | Next.js cache inconsistency | Data stale | Shared Redis cache, no-store | FE |
| TR-06 | RLS bypass | Data leak | Test, role, pentest | Security |
| TR-07 | DR tidak siap | RTO gagal | Drill, IaC, runbook | SRE |
| TR-08 | Dependency SLA < target | SLO gagal | Review SLA, redundancy | CTO |
| TR-09 | Cost overrun | Margin turun | Budget, right-sizing | Finance |
| TR-10 | Skill gap | Delivery lambat | Training, hire, managed | Eng |
| TR-11 | Data residency | Legal | In-country DR, DPA | Compliance |
| TR-12 | Vendor lock-in | Exit sulit | Open standard, export | CTO |
| TR-13 | Supply chain | Compromise | SBOM, sign, scan | Security |
| TR-14 | Model/OCR drift | False positive | Monitoring, fallback | Data |
| TR-15 | Batch SLA | Demo gagal | Benchmark, pool, optimize | Eng |

**Residual risk:** risiko yang tersisa setelah mitigasi harus dicatat di PRD §16.2/BRD §16.2. Tidak ada risiko yang "diterima" tanpa owner dan review.

**Kill switch:** rule pack, ML model, document OCR, integration, LLM gateway; fallback ke rules-only/manual.

---

## 21. Rencana Implementasi & Milestone

### 21.1 Phase plan

| Phase | Timeline | Deliverable | Exit |
|---|---|---|---|
| 0/1 Thin slice | Hackathon | Docker Compose; 6 service; rules A–G; queue; case; audit; metering; 50k/90s | AC BRD; demo |
| Commercial MVP | M0–M9 | K8s 3 AZ; managed DB/Redis/Kafka; multi-tenant; ML adapter; OCR basic; 1 integrasi; HA | 99.9%; pilot |
| GA | M9–M15 | DR; 99.95%; SOC 2 readiness; scale | 2 paid; HA/DR drill |
| Phase 2 | M9–M15 | Jiwa/kesehatan; graph; document advanced; external | Multi-line |
| Phase 3 | M15–M24 | GenAI; consortium; ASEAN; active-active review | 10+ tenant |

### 21.2 Workstream

| Workstream | Fokus | Lead |
|---|---|---|
| Platform | K8s, IaC, CI/CD, observability | SRE |
| Frontend | Next.js, UX, accessibility | FE Lead |
| Backend | Go services, rules, case, audit | BE Lead |
| Data | PostgreSQL, Kafka, Redis, object storage | Data |
| Security | Auth, RLS, KMS, pentest | Security |
| QA | Test, load, chaos, DR | QA |
| Compliance | UU PDP/OJK, DPA, retention | Compliance |

### 21.3 Milestone

| Milestone | Kriteria |
|---|---|
| M1 Architecture approved | TRD sign-off; ADR |
| M2 Thin slice | 50k/90s; determinism; queue; audit; metering |
| M3 MVP HA | 3 AZ; failover; load; multi-tenant |
| M4 Pilot | 2 design partner; SLO; security |
| M5 GA | 99.95%; DR; SOC 2 readiness |
| M6 Scale | Multi-line; 10+ tenant |

### 21.4 Definition of done

- Kode, test, docs, runbook.
- SLO/dashboard/alert.
- Security scan; no critical/high.
- Load/chaos pass.
- Backup/restore tested.
- Migration/rollback tested.
- Capacity headroom.
- Sign-off.

### 21.5 Dependencies

- Cloud region/account; network; KMS.
- Managed DB/Redis/Kafka quota.
- IdP/SSO.
- Design partner data/labels.
- Legal/DPA/DPIA.
- Team hiring/training.

---

## 22. Open Questions & Assumptions

| ID | Pertanyaan | Dampak | Owner | Default |
|---|---|---|---|---|
| OQ-01 | Cloud provider final? | Topologi/managed service | CTO | Alibaba Cloud Jakarta (`ap-southeast-5`) |
| OQ-02 | Managed vs self-managed? | Operasi/biaya | CTO | Managed |
| OQ-03 | DR region/lokasi? | RTO/residensi | Compliance | In-country |
| OQ-04 | Tenant isolation model? | RLS/dedicated | Security | RLS |
| OQ-05 | Kafka vs NATS? | Operasi | Eng | Kafka |
| OQ-06 | Service mesh? | mTLS/kompleksitas | Eng | GA |
| OQ-07 | IdP? | SSO/MFA | Security | OIDC |
| OQ-08 | Retensi final? | Storage/compliance | Legal | `TBD` |
| OQ-09 | KMS/HSM? | Enkripsi | Security | Cloud KMS |
| OQ-10 | Budget? | Sizing | Finance | `TBD` |
| OQ-11 | On-prem hardware? | Private cloud | CTO | `TBD` |
| OQ-12 | SLA tenant? | SLO | Product | 99.9/99.95 |

**Asumsi:** 3 AZ; managed services; data residency Indonesia; no active-active; thin slice non-HA; team 9 engineer; pilot 2 insurer.

---

## 23. Traceability Matrix

### 23.1 PRD NFR → TRD

| PRD | Requirement | TRD | Verifikasi |
|---|---|---|---|
| NFR-1 | p95 <2s; batch 1M <4h; UI <3s | §7.10, §8.6, §14.6 | Load/benchmark |
| NFR-2 | 99.9/99.95; RTO 4h; RPO 15m | §2.5, §6, §15 | SLO/DR drill |
| NFR-3 | 10M/tahun; 1.000 users; scale | §14 | Load test |
| NFR-4 | ISO/ASVS/encryption/MFA | §12 | Pentest |
| NFR-5 | UU PDP; DSR; residency | §12.9, §15.6 | Audit |
| NFR-6 | Explainability (reason code + audit) | §5.8, §12.8 | Test |
| NFR-7 | Fairness | §7.11, §12 | Review + monitoring |
| NFR-8 | Bahasa/IDR/WIB | §3.5, §8 | UAT |
| NFR-9 | WCAG AA | §8.4 | Audit |
| NFR-10 | Observability | §13 | Review |
| NFR-11 | API-first/CI/CD/IaC | §3, §16 | Review |
| NFR-12 | Portability | §19.4 | Export/exit test |
| NFR-13 | Retensi/audit | §9.10, §12.8 | Audit |
| NFR-14 | 3 klik/training | §8.6 | UAT |
| NFR-15 | Determinism | §7.5 | Backtest |
| NFR-16 | Reproducibility | §7.5 | Test |
| NFR-17 | Metering | §9.5, §13.6 | Reconciliation |
| NFR-18 | Billing export | §9.5 | Checksum |
| NFR-19 | Value report | §9.5 | Test |
| NFR-20 | Read-only | §5, §7.11 | Code review + network policy |

### 23.2 BRD NFR → TRD

| BRD | Requirement | TRD | Verifikasi |
|---|---|---|---|
| NFR-01 | 50k/90s | §14.2, §14.6 | Benchmark |
| NFR-02 | Queue/dashboard <3s | §8.6 | Load |
| NFR-03 | Determinism | §7.5 | Test |
| NFR-04 | Explainability | §5.8 | Test |
| NFR-05 | No write-back | §5, §7.11 | Code review + network policy |
| NFR-06 | Fairness | §7.11 | Review + monitoring |
| NFR-07 | Privacy | §12.9 | Pentest |
| NFR-08 | Password/session | §12.2 | Pentest |
| NFR-09 | Upload validation | §12.6 | Pentest |
| NFR-10 | CSV injection | §12.6 | Pentest |
| NFR-11 | OWASP | §12.6 | Pentest |
| NFR-12 | Audit append-only | §12.8 | Test |
| NFR-13 | Atomic/idempotent | §10 | Test |
| NFR-14 | Usability | §8 | UAT |
| NFR-15 | Docker Compose | §6.1 | Demo |
| NFR-16 | OCR performance | §7.8 | Phase 2 |
| NFR-17 | Storage integrity | §9.8 | Test |
| NFR-18 | Encryption | §12.4 | Review |
| NFR-19 | Offline | §6.1 | Demo |
| NFR-20 | Reproducibility | §7.5 | Test |
| NFR-21 | Quota | §11, §12 | Test |
| NFR-22 | Metering | §9.5 | Reconciliation |
| NFR-23 | Billing export | §9.5 | Checksum |
| NFR-24 | Value report | §9.5 | Test |

### 23.3 Acceptance → TRD

| Acceptance | TRD |
|---|---|
| AC-1..AC-8 (partial → AT-01..AT-12) | §5.6, §7.5, §9; mapping rinci tetap `TBD` QA |
| AC-9..AC-12 (→ AT-06, AT-11, AT-17, AT-18) | §9.8, §12.8, Phase 2 |
| AC-13..AC-16 (→ AT-13..AT-16) | §9.5, §13.6 |
| AT-01..AT-18 | §17.2–§17.7; AT-01..AT-05/07/08/10–16 relevan MVP; AT-06/17/18 Phase 2 |
| BRD §13 AC-1..AC-16 | Acuan thin slice; §17.2 dan §23.2 |

### 23.4 PRD component → TRD service

| PRD §10.2 | TRD | Phase |
|---|---|---|
| Ingestion service | `ingestion` §5.2 | Thin slice/MVP |
| Data lakehouse | §5.2 note, §9.14 | Phase 2 |
| Feature store | `feature-store`, §9.14 | Phase 2 |
| Entity resolution service | `entity-resolution` | Phase 2 |
| Graph service | `graph` | Phase 2 |
| Rules engine | `rules` | Thin slice/MVP |
| ML scoring service | `ml` | Phase 2 |
| Document AI service | `document` | Phase 2 |
| Alert & triage service | `alert` | Thin slice/MVP |
| Case service | `case` | Thin slice/MVP |
| Reporting service | `reporting` | Thin slice/MVP |
| Admin service | `identity` (thin slice) + `admin` (MVP) | Thin slice/MVP |
| Audit service | `audit` | Thin slice/MVP |
| LLM gateway | `llm-gateway` | Phase 3 |
| Integration service | `integration` | Phase 2 |

---

## 24. Referensi

Sumber diakses 16–17 November 2026. Referensi eksternal adalah bahan best practice, bukan pengganti validasi Legal/Compliance.

| ID | Sumber | URL |
|---|---|---|
| R1 | Kubernetes — Running in multiple zones | https://kubernetes.io/docs/setup/best-practices/multiple-zones/ |
| R2 | Kubernetes — Creating Highly Available Clusters with kubeadm | https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/high-availability/ |
| R3 | Amazon EKS — Best Practices for Reliability (pembanding provider alternatif) | https://docs.aws.amazon.com/eks/latest/best-practices/reliability.html |
| R4 | Amazon RDS — Multi-AZ DB cluster deployments (pembanding) | https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/multi-az-db-clusters-concepts.html |
| R5 | PostgreSQL — Log-Shipping Standby Servers | https://www.postgresql.org/docs/current/warm-standby.html |
| R6 | Patroni — HA multi datacenter | https://patroni.readthedocs.io/en/latest/ha_multi_dc.html |
| R7 | Redis — High availability with Redis Sentinel | https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/ |
| R8 | Redis — Scale with Redis Cluster | https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/ |
| R9 | Amazon ElastiCache — Multi-AZ automatic failover (pembanding) | https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/AutoFailover.html |
| R10 | Apache Kafka — Documentation/Design | https://kafka.apache.org/documentation/ |
| R11 | NATS — JetStream in a cluster | https://docs.nats.io/learn/topologies/jetstream-in-a-cluster |
| R12 | Next.js — How to self-host | https://nextjs.org/docs/app/guides/self-hosting |
| R13 | Kubernetes Gateway API | https://gateway-api.sigs.k8s.io/ |
| R14 | Istio — Security/mTLS | https://istio.io/latest/docs/concepts/security/ |
| R15 | PostgreSQL — Row Security Policies | https://www.postgresql.org/docs/current/ddl-rowsecurity.html |
| R16 | PostgreSQL — Table Partitioning | https://www.postgresql.org/docs/current/ddl-partitioning.html |
| R17 | PgBouncer — Configuration | https://www.pgbouncer.org/config.html |
| R18 | Amazon S3 — Object Lock (pembanding WORM) | https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html |
| R19 | Amazon S3 — Replication (pembanding CRR) | https://docs.aws.amazon.com/AmazonS3/latest/userguide/replication.html |
| R20 | AWS WAF — Developer Guide (pembanding) | https://docs.aws.amazon.com/waf/latest/developerguide/what-is-aws-waf.html |
| R21 | OpenTelemetry — Documentation | https://opentelemetry.io/docs/ |
| R22 | Google SRE — Service Level Objectives | https://sre.google/sre-book/service-level-objectives/ |
| R23 | Argo CD — Documentation | https://argo-cd.readthedocs.io/en/stable/ |
| R24 | Chaos Mesh — Overview | https://chaos-mesh.org/docs/ |
| R25 | Grafana k6 — Documentation | https://grafana.com/docs/k6/latest/ |
| R26 | OWASP ASVS | https://owasp.org/projects/asvs |
| R27 | AWS — Asia Pacific (Jakarta) Region, 3 AZ (pembanding region) | https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html |
| R28 | Google Cloud — Jakarta region, 3 zones; Azure — Indonesia Central, 3 AZ | https://cloud.google.com/about/locations ; https://learn.microsoft.com/en-us/azure/reliability/availability-zones-overview |
| R29 | pgBackRest — User Guide | https://pgbackrest.org/user-guide.html |
| R30 | OpenAPI Specification | https://spec.openapis.org/oas/v3.0.3.html |
| R31 | OpenID Connect — How it works | https://openid.net/developers/how-connect-works/ |
| R32 | Kubernetes — Pod Disruption Budget | https://kubernetes.io/docs/tasks/run-application/configure-pdb/ |
| R33 | Kubernetes — Pod Topology Spread Constraints | https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/ |
| R34 | Kubernetes — Horizontal Pod Autoscaling | https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/ |
| R35 | Amazon MSK — Best practices (pembanding) | https://docs.aws.amazon.com/msk/latest/developerguide/bestpractices.html |
| R36 | Alibaba Cloud — Regions and Zones (ECS) | https://www.alibabacloud.com/help/en/ecs/regions-and-zones |
| R37 | Alibaba Cloud — Global Locations | https://www.alibabacloud.com/global-locations |
| R38 | Alibaba Cloud ACK — High availability | https://www.alibabacloud.com/help/en/ack/product-overview/high-availability |
| R39 | Alibaba Cloud — Resource Directory | https://www.alibabacloud.com/help/en/resource-management/product-overview/what-is-resource-directory |
| R40 | ApsaraDB RDS for PostgreSQL — High availability and disaster recovery | https://www.alibabacloud.com/help/en/rds/apsaradb-rds-for-postgresql/user-guide/high-availability |
| R41 | Tair — Disaster recovery and multi-AZ deployment | https://www.alibabacloud.com/help/en/redis/product-overview/disaster-recovery |
| R42 | Alibaba Cloud Service Mesh (ASM) — Overview | https://www.alibabacloud.com/help/en/asm/product-overview/what-is-asm |
| R43 | Application Load Balancer (ALB) — Overview | https://www.alibabacloud.com/help/en/alb/product-overview/what-is-alb |
| R44 | Alibaba Cloud WAF — Overview | https://www.alibabacloud.com/help/en/waf/product-overview/what-is-waf |
| R45 | Anti-DDoS Pro/Premium — Overview | https://www.alibabacloud.com/help/en/anti-ddos-pro/product-overview/what-is-anti-ddos-pro |
| R46 | ApsaraMQ for Kafka — Overview and multi-zone deployment | https://www.alibabacloud.com/help/en/apsaramq-for-kafka/product-overview/what-is-apsaramq-for-kafka |
| R47 | OSS — Retention policy (WORM) | https://www.alibabacloud.com/help/en/oss/user-guide/retention-policy |
| R48 | OSS — Cross-region replication (CRR/RTC) | https://www.alibabacloud.com/help/en/oss/user-guide/cross-region-replication |
| R49 | Alibaba Cloud ARMS — Overview | https://www.alibabacloud.com/help/en/arms/product-overview/what-is-arms |
| R50 | Simple Log Service (SLS) — Overview | https://www.alibabacloud.com/help/en/sls/product-overview/what-is-log-service |

---

## 25. Lampiran

### 25.1 HA checklist pra-produksi

- [ ] 3 AZ; node pool tersebar; control plane HA.
- [ ] Minimal 3 replica service; PDB; topology spread.
- [ ] HPA + cluster autoscaler; headroom 40%.
- [ ] ApsaraDB RDS PostgreSQL multi-zone; sync standby; PITR; restore tested.
- [ ] Tair multi-AZ; client reconnect/backoff; session fallback teruji.
- [ ] ApsaraMQ RF=3; minISR=2; acks=all; URP 0.
- [ ] OSS versioning; retention policy WORM; CRR.
- [ ] Backup + restore drill; RTO/RPO terukur.
- [ ] DR runbook; DNS failover; KMS/secrets.
- [ ] SLO dashboard; burn-rate alert; on-call.
- [ ] Load test 500 rps/1.000 users; batch 1M/4h.
- [ ] Chaos test pod/node/AZ/DB/Tair/ApsaraMQ.
- [ ] Security scan/pentest; no critical/high.
- [ ] Tenant isolation test; RLS.
- [ ] Metering reconciliation; billing checksum.
- [ ] Runbook + game day.
- [ ] Capacity/cost review.

### 25.2 Kubernetes baseline (contoh ringkas)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rules
spec:
  replicas: 3
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    spec:
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: rules
      terminationGracePeriodSeconds: 60
      containers:
        - name: rules
          image: registry/rules@sha256:...
          resources:
            requests: {cpu: "500m", memory: "512Mi"}
            limits: {cpu: "1", memory: "1Gi"}
          readinessProbe:
            httpGet: {path: /readyz, port: 8080}
          livenessProbe:
            httpGet: {path: /healthz, port: 8080}
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: rules
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: rules
```

### 25.3 Kafka topic baseline

```
replication.factor=3
min.insync.replicas=2
unclean.leader.election.enable=false
compression.type=zstd
retention.ms=604800000
message.max.bytes=1048576
```

### 25.4 PostgreSQL baseline

```
max_connections=500
shared_buffers=25% RAM
effective_cache_size=50% RAM
work_mem=16MB
maintenance_work_mem=1GB
statement_timeout=30s
idle_in_transaction_session_timeout=30s
lock_timeout=5s
log_min_duration_statement=500ms
pg_stat_statements.track=all
```

### 25.5 SLO alert example

```
- alert: APIErrorBudgetBurn
  expr: |
    (sum(rate(http_requests_total{status=~"5.."}[1h]))
     / sum(rate(http_requests_total[1h]))) > (14.4 * 0.0005)
  for: 5m
  labels: {severity: page}
  annotations: {runbook: "https://.../runbooks/api"}
```

### 25.6 Glosarium teknis

| Istilah | Arti |
|---|---|
| AZ | Availability Zone |
| RTO | Recovery Time Objective |
| RPO | Recovery Point Objective |
| RLS | Row-Level Security |
| PITR | Point-in-Time Recovery |
| PDB | Pod Disruption Budget |
| HPA | Horizontal Pod Autoscaler |
| URP | Under-Replicated Partitions |
| SLI/SLO | Service Level Indicator/Objective |
| Outbox | Tabel event ditulis dalam transaksi yang sama |
| Idempotency | Operasi berulang menghasilkan efek sama |
| Warm standby | DR siap tapi tidak menerima traffic penuh |
| Fencing | Mencegah dua writer aktif |
| Tiered storage | Pindahkan data lama ke storage murah |

---

**Akhir dokumen.**

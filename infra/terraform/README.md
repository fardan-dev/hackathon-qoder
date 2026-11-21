# Terraform — Google Cloud landing zone

Landing zone Phase 1 untuk platform fraud detection. Region terkunci ke **`asia-southeast2` (Jakarta)** dengan tiga zone: `asia-southeast2-a/b/c`.

## Cakupan saat ini

| Stack | Path | Isi |
| --- | --- | --- |
| Org | `org/` | Project (subset purpose), baseline API, guardrail, budget per project; folder + policy folder-level hanya jika `org_id` diisi |
| Bootstrap | `bootstrap/` | GCS bucket state (versioning, CMEK, UBLA, PAP), Cloud KMS keyring/key state |
| Env | `envs/dev`, `envs/staging`, `envs/prod` | Cloud KMS key aplikasi + enablement API baseline |

Stack network, GKE, data, observability, dan DR dikerjakan di INFRA-2..INFRA-6.

## Status 20 Nov 2026

Akun ini **tidak punya Google Cloud Organization** (`gcloud organizations list` kosong; project tanpa parent), jadi stack `org` berjalan dalam mode tanpa organization:

- Tidak ada folder. Organization policy juga tidak bisa dipasang: role `roles/orgpolicy.policyAdmin` tidak dapat diberikan pada project, dan `setOrgPolicy` ditolak untuk owner. Resource policy otomatis dilewati di mode ini.
- Guardrail yang tetap berlaku: setting bucket (`PAP`/`UBLA`) dari `bootstrap`, `auto_create_network = false`, label wajib, dan budget alert per project.
- Project aktif: `miku-project-487309` (staging; bootstrap + env sudah apply, state di GCS) dan `revas-ifd-dev` (number `158798799196`, dibuat untuk demo/UAT).
- Project `staging` hasil factory, `prod`, `shared`, `security`, `dr` belum dibuat.

## Prasyarat

- Terraform >= 1.12 dan provider `hashicorp/google` ~> 7.0.
- `gcloud auth application-default login` dengan akun yang punya akses:
  - `roles/resourcemanager.projectCreator` dan `roles/billing.user` (mode tanpa organization)
  - `roles/resourcemanager.organizationAdmin` dan `roles/resourcemanager.folderAdmin` (hanya mode dengan `org_id`)
  - `roles/billing.admin` (budget dan binding billing account)
  - `roles/cloudkms.admin`, `roles/storage.admin` untuk bootstrap
- Billing account ID dan project ID unik global. Organization ID hanya diperlukan bila akun sudah punya Organization.

Jangan menaruh service account key JSON di repo, CI, atau pod. Gunakan ADC lokal dan Workload Identity Federation untuk CI.

## Urutan eksekusi

### 1. Org (project, policy, budget)

```bash
cd infra/terraform/org
cp terraform.tfvars.example terraform.tfvars   # isi billing_account + project_ids (+ org_id bila ada)
terraform init
terraform plan
terraform apply
```

`project_ids` menerima subset purpose apa pun. Minimal untuk demo/UAT:

```hcl
project_ids = { dev = "revas-ifd-dev" }
```

Project dibuat tanpa network default (`auto_create_network = false`).

Dua mode:

- `org_id = ""` — tanpa organization. Tidak ada folder dan tidak ada org policy (role `roles/orgpolicy.policyAdmin` hanya ada di level organization). Guardrail mengandalkan setting resource dan proses review.
- `org_id = "<id>"` — folder platform dibuat (opsional di bawah `parent_folder_id`), lalu guardrail dipasang di level folder:
  - `iam.disableServiceAccountKeyCreation`
  - `storage.uniformBucketLevelAccess`
  - `storage.publicAccessPrevention`

Budget per project memakai mata uang billing account; `budget_currency` harus sama dengan mata uang akun (akun ini `IDR`).

Provider memakai `user_project_override` dengan `quota_project` karena `billingbudgets.googleapis.com` menolak kredensial ADC tanpa quota project eksplisit.

State `org` disimpan di GCS: `cp backend.tf.example backend.tf` lalu isi `bucket`/`prefix` sebelum `terraform init`.

### 2. Bootstrap per project environment

Jalankan sekali per project `dev`, `staging`, dan `prod`:

```bash
cd infra/terraform/bootstrap
cp terraform.tfvars.example terraform.tfvars   # isi project_id + environment
terraform init
terraform plan
terraform apply
```

Output `backend_config` memberi nama bucket dan prefix. Setelah bucket siap:

```bash
cp backend.tf.example backend.tf               # ganti bucket
terraform init -migrate-state
```

State bootstrap pindah dari lokal ke GCS. Ulangi untuk project environment lain.

### 3. Environment stack

```bash
cd infra/terraform/envs/staging
cp terraform.tfvars.example terraform.tfvars   # isi project_id
cp backend.hcl.example backend.hcl             # bucket + prefix envs/staging
terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

Bucket state per environment mengikuti pola `ifd-tfstate-<env>-<project_id>`.

## State dan locking

- State di GCS dengan versioning, default encryption CMEK, Uniform Bucket-Level Access, dan Public Access Prevention `enforced`.
- Locking memakai native GCS backend; tidak ada layanan lock terpisah.
- Noncurrent version dihapus setelah 90 hari.
- File berikut tidak boleh di-commit: `*.tfstate*`, `terraform.tfvars`, `backend.hcl`, `backend.tf`, `.terraform/`, plan file.

## Guardrail

- Service account key creation diblokir lewat org policy `iam.disableServiceAccountKeyCreation` (hanya mode dengan `org_id`); mode tanpa organization mengandalkan proses review dan WIF-only.
- PAP/UBLA ditegakkan lewat setting bucket di `bootstrap`, bukan org policy.
- Tidak ada kredensial statis di repo, CI, atau pod; CI memakai OIDC/WIF.
- Label wajib: `env`, `tenant`, `service`, `owner`, `cost-center`, `data-classification`.
- Perubahan produksi hanya lewat Terraform; akses manual dicatat di runbook break-glass.
- Hancurkan resource lewat `terraform destroy`; jangan hapus manual.

## Pemetaan layanan dari desain awal

| Kebutuhan | Google Cloud |
| --- | --- |
| Resource Directory / akun terpisah | Organization, folder, project |
| RAM / CloudSSO | Cloud IAM, Workforce Identity Federation |
| KMS + Secrets Manager | Cloud KMS + Secret Manager |
| OSS + WORM + CRR | Cloud Storage versioning + retention lock + dual-region/turbo replication |
| TableStore state lock | Native GCS backend locking |
| ACK Pro | GKE regional (3 zone) |
| ACR Enterprise | Artifact Registry |
| ApsaraDB RDS PostgreSQL | Cloud SQL for PostgreSQL (regional HA) |
| Tair | Memorystore for Redis (Standard tier) |
| ApsaraMQ for Kafka | Pub/Sub (MVP) atau Managed Service for Apache Kafka (jika semantics Kafka wajib) |
| ALB + WAF + Anti-DDoS | Global external Application Load Balancer + Cloud Armor |
| Alibaba Cloud DNS | Cloud DNS routing policy + health check |
| SLS / ARMS | Cloud Logging, Cloud Monitoring, Managed Service for Prometheus, Cloud Trace |
| ActionTrail | Cloud Audit Logs + aggregated sink |
| DBS | Backup and DR Service + Cloud SQL backups |
| RRSA | Workload Identity Federation for GKE |

## Verifikasi kuota sebelum komitmen

Isi tabel ini saat akses project tersedia. Hasilnya dilampirkan ke HAC-71.

| Layanan | Tersedia di `asia-southeast2` | Kuota | Catatan |
| --- | --- | --- | --- |
| GKE regional (3 zone) | | | control plane regional |
| Cloud SQL for PostgreSQL HA | | | regional HA, PITR |
| Memorystore for Redis Standard | | | multi-zone |
| Pub/Sub / Managed Kafka | | | pilih sesuai semantics |
| Global external ALB + Cloud Armor | | | |
| Cloud KMS | | | keyring Jakarta |
| Cloud Storage (CMEK, retention lock) | | | |
| Backup and DR Service | | | |
| Cloud DNS | | | failover routing |

## Biaya

Ukuran awal adalah yang terkecil yang layak: satu bucket state kecil + CMK per environment, enablement API, dan budget alert per project. GKE, Cloud SQL, Memorystore, dan ALB baru dibuat di ticket berikutnya setelah kuota dan biaya diverifikasi.

## Known gaps

- Akun belum punya Organization; folder dan org policy tidak bisa dipasang sampai Cloud Identity/Workspace tersedia. `iam.disableServiceAccountKeyCreation` belum ditegakkan.
- Provider harus memakai `quota_project` untuk org policy dan billing budget; tanpa itu API menolak ADC (`x-goog-user-project` tidak dikirim).
- Budget harus memakai mata uang billing account (`IDR`); nominal USD ditolak API.
- Org policy baru mencakup pemblokiran SA key, UBLA, dan PAP; constraint lokasi resource belum dipasang.
- Budget memakai `budget_currency` yang harus sama dengan mata uang billing account.
- Google Cloud hanya punya satu region di Indonesia (`asia-southeast2`). DR in-country belum tersedia; opsi DR adalah region Singapore dengan persetujuan Legal/tenant atau co-location, dan residual risk harus disetujui.
- TRD/BRD/PRD masih memetakan layanan Alibaba Cloud; sinkronisasi dokumen ke GCP belum dilakukan.
- Resource Alibaba yang sempat terbuat (KMS key + OTS instance) tidak terkelola lagi dan tidak bisa dihapus karena akun dinonaktifkan.
- CI hanya menjalankan `fmt` + `validate` tanpa kredensial; plan/apply otomatis menunggu keputusan WIF audience dan IAM role.

data "google_project" "current" {
  project_id = var.project_id
}

locals {
  bucket_name = lower("${var.name_prefix}-tfstate-${var.environment}-${data.google_project.current.project_id}")

  labels = {
    env                 = var.environment
    tenant              = var.tenant
    service             = "terraform-state"
    owner               = var.owner
    cost-center         = var.cost_center
    data-classification = "internal"
  }

  services = [
    "cloudkms.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "serviceusage.googleapis.com",
    "storage.googleapis.com",
  ]
}

resource "google_project_service" "required" {
  for_each = toset(local.services)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

resource "google_kms_key_ring" "state" {
  name     = "${var.name_prefix}-${var.environment}-state"
  location = var.region
  project  = var.project_id

  depends_on = [google_project_service.required]
}

resource "google_kms_crypto_key" "state" {
  name            = "${var.name_prefix}-${var.environment}-state"
  key_ring        = google_kms_key_ring.state.id
  purpose         = "ENCRYPT_DECRYPT"
  rotation_period = "7776000s"
  labels          = local.labels
}

resource "google_kms_crypto_key_iam_member" "gcs" {
  crypto_key_id = google_kms_crypto_key.state.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.current.number}@gs-project-accounts.iam.gserviceaccount.com"

  depends_on = [google_kms_crypto_key.state]
}

resource "google_storage_bucket" "tfstate" {
  name                        = local.bucket_name
  project                     = var.project_id
  location                    = "ASIA-SOUTHEAST2"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = var.state_bucket_force_destroy
  labels                      = local.labels

  versioning {
    enabled = true
  }

  encryption {
    default_kms_key_name = google_kms_crypto_key.state.id
  }

  lifecycle_rule {
    condition {
      days_since_noncurrent_time = 90
    }
    action {
      type = "Delete"
    }
  }

  depends_on = [
    google_kms_crypto_key_iam_member.gcs,
    google_project_service.required,
  ]
}

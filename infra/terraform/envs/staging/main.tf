data "google_project" "current" {
  project_id = var.project_id
}

locals {
  labels = {
    env                 = var.environment
    tenant              = var.tenant
    service             = "platform"
    owner               = var.owner
    cost-center         = var.cost_center
    data-classification = var.data_classification
  }

  services = [
    "artifactregistry.googleapis.com",
    "backupdr.googleapis.com",
    "certificatemanager.googleapis.com",
    "cloudkms.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "cloudtrace.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "dns.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "networksecurity.googleapis.com",
    "pubsub.googleapis.com",
    "redis.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "sqladmin.googleapis.com",
    "storage.googleapis.com",
  ]
}

resource "google_project_service" "required" {
  for_each = toset(local.services)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

resource "google_kms_key_ring" "app" {
  name     = "${var.name_prefix}-${var.environment}-app"
  location = var.region
  project  = var.project_id

  depends_on = [google_project_service.required]
}

resource "google_kms_crypto_key" "app" {
  name            = "${var.name_prefix}-${var.environment}-app"
  key_ring        = google_kms_key_ring.app.id
  purpose         = "ENCRYPT_DECRYPT"
  rotation_period = "7776000s"
  labels          = local.labels
}

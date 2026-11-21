output "project_id" {
  value = data.google_project.current.project_id
}

output "project_number" {
  value = data.google_project.current.number
}

output "state_bucket" {
  description = "GCS bucket that stores Terraform state."
  value       = google_storage_bucket.tfstate.name
}

output "state_kms_key_id" {
  description = "Cloud KMS key used for the bucket default encryption."
  value       = google_kms_crypto_key.state.id
}

output "backend_config" {
  description = "Values for the GCS backend after migrating bootstrap state."
  value = {
    bucket = google_storage_bucket.tfstate.name
    prefix = "bootstrap"
  }
}

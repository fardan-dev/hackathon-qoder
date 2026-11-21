output "project_id" {
  value = data.google_project.current.project_id
}

output "project_number" {
  value = data.google_project.current.number
}

output "kms_key_ring_id" {
  value = google_kms_key_ring.app.id
}

output "kms_key_id" {
  value = google_kms_crypto_key.app.id
}

output "enabled_services" {
  value = sort(local.services)
}

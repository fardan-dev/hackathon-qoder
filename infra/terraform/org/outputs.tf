output "folder_id" {
  description = "Numeric ID of the platform folder; null when the stack runs without an organization."
  value       = try(google_folder.platform[0].folder_id, null)
}

output "folder_name" {
  description = "Resource name of the platform folder; null when the stack runs without an organization."
  value       = try(google_folder.platform[0].name, null)
}

output "project_ids" {
  description = "Project IDs keyed by purpose."
  value       = { for purpose, project in google_project.member : purpose => project.project_id }
}

output "project_numbers" {
  description = "Project numbers keyed by purpose."
  value       = { for purpose, project in google_project.member : purpose => project.number }
}

output "budget_names" {
  description = "Billing budget resource names keyed by purpose."
  value       = { for purpose, budget in google_billing_budget.project : purpose => budget.name }
}

provider "google" {
  region = var.region

  # orgpolicy.googleapis.com and billingbudgets.googleapis.com reject user ADC
  # without an explicit quota project; the org stack runs from a separate
  # project than the ones it manages.
  user_project_override = var.quota_project != ""
  billing_project       = var.quota_project != "" ? var.quota_project : null
}

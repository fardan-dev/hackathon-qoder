locals {
  has_org = var.org_id != ""
  parent  = local.has_org ? (var.parent_folder_id != "" ? "folders/${var.parent_folder_id}" : "organizations/${var.org_id}") : null

  labels = {
    tenant      = var.tenant
    service     = "platform"
    owner       = var.owner
    cost-center = var.cost_center
  }

  base_services = [
    "cloudresourcemanager.googleapis.com",
    "serviceusage.googleapis.com",
  ]

  platform_services = [
    "billingbudgets.googleapis.com",
    "cloudkms.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "storage.googleapis.com",
  ]

  base_matrix = {
    for pair in setproduct(keys(var.project_ids), local.base_services) :
    "${pair[0]}/${pair[1]}" => { purpose = pair[0], service = pair[1] }
  }

  platform_matrix = {
    for pair in setproduct(keys(var.project_ids), local.platform_services) :
    "${pair[0]}/${pair[1]}" => { purpose = pair[0], service = pair[1] }
  }

  # Org policy roles only exist at organization level, so without an
  # organization the policies are skipped instead of failing the whole apply.
  policy_targets = local.has_org ? { platform = google_folder.platform[0].name } : {}
}

resource "google_folder" "platform" {
  count = local.has_org ? 1 : 0

  display_name = var.folder_name
  parent       = local.parent
}

resource "google_project" "member" {
  for_each = var.project_ids

  name                = "${var.name_prefix}-${each.key}"
  project_id          = each.value
  folder_id           = local.has_org ? google_folder.platform[0].folder_id : null
  billing_account     = var.billing_account
  auto_create_network = false
  labels              = merge(local.labels, { env = each.key })
}

resource "google_project_service" "base" {
  for_each = local.base_matrix

  project            = google_project.member[each.value.purpose].project_id
  service            = each.value.service
  disable_on_destroy = false
}

resource "google_project_service" "platform" {
  for_each = local.platform_matrix

  project            = google_project.member[each.value.purpose].project_id
  service            = each.value.service
  disable_on_destroy = false

  depends_on = [google_project_service.base]
}

resource "google_org_policy_policy" "disable_service_account_keys" {
  for_each = local.policy_targets

  name   = "${each.value}/policies/iam.disableServiceAccountKeyCreation"
  parent = each.value

  spec {
    rules {
      enforce = "TRUE"
    }
  }
}

resource "google_org_policy_policy" "uniform_bucket_level_access" {
  for_each = local.policy_targets

  name   = "${each.value}/policies/storage.uniformBucketLevelAccess"
  parent = each.value

  spec {
    rules {
      enforce = "TRUE"
    }
  }
}

resource "google_org_policy_policy" "public_access_prevention" {
  for_each = local.policy_targets

  name   = "${each.value}/policies/storage.publicAccessPrevention"
  parent = each.value

  spec {
    rules {
      enforce = "TRUE"
    }
  }
}

resource "google_billing_budget" "project" {
  for_each = var.project_ids

  billing_account = var.billing_account
  display_name    = "${each.value}-monthly"

  budget_filter {
    projects = ["projects/${google_project.member[each.key].number}"]
  }

  amount {
    specified_amount {
      currency_code = var.budget_currency
      units         = tostring(var.budget_amount)
    }
  }

  threshold_rules {
    threshold_percent = 0.5
  }

  threshold_rules {
    threshold_percent = 0.8
  }

  threshold_rules {
    threshold_percent = 1.0
  }

  depends_on = [google_project_service.platform]
}

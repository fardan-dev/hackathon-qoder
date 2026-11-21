variable "org_id" {
  type        = string
  default     = ""
  description = "Optional GCP organization ID. Empty runs the stack without an organization: projects have no folder and guardrail policies are applied per project."
}

variable "quota_project" {
  type        = string
  default     = ""
  description = "Project billed for API quota (x-goog-user-project). Required for org policy and billing budget calls; use a project where orgpolicy.googleapis.com and billingbudgets.googleapis.com are enabled."
}

variable "billing_account" {
  type        = string
  description = "Billing account ID in XXXXXX-XXXXXX-XXXXXX format."
}

variable "parent_folder_id" {
  type        = string
  default     = ""
  description = "Optional numeric parent folder ID. Empty creates the platform folder at the organization root."
}

variable "folder_name" {
  type        = string
  default     = "platform"
  description = "Folder that holds the platform projects."
}

variable "name_prefix" {
  type        = string
  default     = "ifd"
  description = "Short lowercase prefix for platform resource names."

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,18}[a-z0-9]$", var.name_prefix))
    error_message = "name_prefix must be 3-20 lowercase characters, letters/digits/hyphen, starting with a letter."
  }
}

variable "region" {
  type        = string
  default     = "asia-southeast2"
  description = "GCP region for the platform."

  validation {
    condition     = var.region == "asia-southeast2"
    error_message = "Phase 1 landing zone is pinned to asia-southeast2 (Jakarta)."
  }
}

variable "project_ids" {
  type        = map(string)
  description = "Project IDs keyed by purpose (dev, staging, prod, shared, security, dr). Any subset is allowed. Projects must not already exist."

  validation {
    condition = alltrue([
      for id in values(var.project_ids) :
      can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", id))
    ])
    error_message = "Each project ID must be 6-30 lowercase characters, letters/digits/hyphen, starting with a letter."
  }
}

variable "budget_amount" {
  type        = number
  default     = 100
  description = "Monthly budget per project."
}

variable "budget_currency" {
  type        = string
  default     = "USD"
  description = "Currency of the billing account; must match the billing account currency."
}

variable "tenant" {
  type        = string
  default     = "platform"
  description = "Tenant recorded in labels."
}

variable "owner" {
  type        = string
  default     = "platform-team"
  description = "Owning team recorded in labels."
}

variable "cost_center" {
  type        = string
  default     = "fraud-platform"
  description = "Cost center recorded in labels."
}

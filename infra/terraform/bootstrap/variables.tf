variable "project_id" {
  type        = string
  description = "GCP project that hosts the Terraform state bucket and state CMK."
}

variable "environment" {
  type        = string
  description = "Environment this state backend belongs to."

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of dev, staging, prod."
  }
}

variable "region" {
  type        = string
  default     = "asia-southeast2"
  description = "GCP region for the state bucket and state CMK."

  validation {
    condition     = var.region == "asia-southeast2"
    error_message = "Phase 1 landing zone is pinned to asia-southeast2 (Jakarta)."
  }
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

variable "state_bucket_force_destroy" {
  type        = bool
  default     = false
  description = "Allow Terraform to delete the state bucket with its objects. Keep false outside throwaway projects."
}

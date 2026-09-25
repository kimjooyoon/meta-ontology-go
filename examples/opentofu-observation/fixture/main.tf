terraform {
  required_version = ">= 1.12.0"
}

locals {
  observed_value = "opentofu-observation"
}

output "observed_value" {
  value = local.observed_value
}

terraform {
  required_providers {
    backstage = {
      source = "tdabasinskas/backstage"
    }
  }
}

provider "backstage" {
  base_url = "https://demo.backstage.io"
}

data "backstage_entities" "test" {
  filters = {
    "kind"               = "User",
    "metadata.namespace" = "default",
  }
}

output "streetlights" {
  value = data.backstage_entities.test
}

terraform {
  required_providers {
    backstage = {
      source  = "tdabasinskas/backstage"
    }
  }
}

provider "backstage" {
  base_url = "https://demo.backstage.io"
}

data "backstage_api" "test" {
  name = "streetlights"
}

output "streetlights" {
  value = data.backstage_api.test
}

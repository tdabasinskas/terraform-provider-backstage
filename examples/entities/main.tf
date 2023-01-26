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

data "backstage_component" "test" {
  name = "shuffle-api"
}

output "streetlights" {
  value = data.backstage_component.test
}

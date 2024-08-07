# Retrieves data about multiple entities that match the given filters:
data "backstage_entities" "example" {
  // The filters to apply to the query:
  filters = [
    "kind=User,metadata.namespace=default",
    "kind=Group,metadata.namespace=default",
  ]
}

# Outputs data from `spec` from an entity:
output "example" {
  value = jsondecode(data.backstage_entities.example.entities[0].spec)["profile"]["email"]
}

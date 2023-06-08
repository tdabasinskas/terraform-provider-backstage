# Retrieves data about multiple entities that match the given filters:
data "backstage_entities" "example" {
  // The filters to apply to the query:
  filters = [
    "kind=User",
    "metadata.namespace=default",
  ]
}

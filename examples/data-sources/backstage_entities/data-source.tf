data "backstage_entities" "example" {
  filters = {
    "kind"               = "User",
    "metadata.namespace" = "default",
  }
}

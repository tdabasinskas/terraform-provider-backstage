# Retrieves specific group data:
data "backstage_group" "example" {
  # Required name of the group:
  name = "example-group"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

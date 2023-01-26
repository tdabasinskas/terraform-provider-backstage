# Retrieves specific component data:
data "backstage_component" "example" {
  # Required name of the component:
  name = "example-component"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

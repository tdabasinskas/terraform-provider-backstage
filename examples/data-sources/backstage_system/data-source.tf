# Retrieves specific system data:
data "backstage_system" "example" {
  # Required name of the system:
  name = "example-system"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

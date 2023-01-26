# Retrieves specific resource data:
data "backstage_resource" "example" {
  # Required name of the system:
  name = "example-resource"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

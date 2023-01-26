# Retrieves specific api data:
data "backstage_api" "example" {
  # Required name of the api:
  name = "example-api"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

# Retrieves specific API data:
data "backstage_api" "example" {
  # Required name of the API:
  name = "example-api"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

# Retrieves specific location data:
data "backstage_location" "example" {
  # Required name of the location:
  name = "example-location"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

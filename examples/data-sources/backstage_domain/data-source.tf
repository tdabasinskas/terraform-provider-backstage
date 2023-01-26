# Retrieves specific domain data:
data "backstage_domain" "example" {
  # Required name of the domain:
  name = "example-domain"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

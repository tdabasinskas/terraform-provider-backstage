# Retrieves specific user data:
data "backstage_user" "example" {
  # Required name of the user:
  name = "example-user"
  # If not provided, namespace defaults to "default" or the the one set in the provider:
  namespace = "example-namespace"
}

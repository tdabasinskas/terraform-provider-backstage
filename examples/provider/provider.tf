# Setup Backstage provider:
provider "backstage" {
  # Provide the URL to Backstage instance:
  base_url = "https://demo.backstage.io"
  # Override the name of default namespace:
  default_namespace = "custom-default"
  # Set custom headers (might be useful for authentication):
  headers = {
    "Custom-Header" = "header_value"
  }
}

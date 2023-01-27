# Ensures the location exists in Backstage.
resource "backstage_location" "example" {
  # URL to the location target:
  target = "http://example-target"
}

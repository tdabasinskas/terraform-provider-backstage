package backstage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLocation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceLocationConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_location.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "kind", "Location"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "metadata.annotations.backstage.io/view-url",
						"https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/all-components.yaml"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "metadata.description", "A collection of all Backstage example components"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "spec.targets.0", "./components/artist-lookup-component.yaml"),
				),
			},
		},
	})
}

const testAccDataSourceLocationConfig = `
data "backstage_location" "test" {
  name = "example-components"
}
`

func TestAccDataSourceLocation_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_location" "test" {
						name = "non_existent_location_a9ab8"
						namespace = "default"
						fallback = {
							name = "fallback_location"
							namespace = "default"
							metadata = {
								description = "Fallback Location"
							}
							spec = {
								targets = ["./components/artist-lookup-component.yaml"]
							}
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_location.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "kind", "Location"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "metadata.description", "Fallback Location"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "name", "fallback_location"),
					resource.TestCheckResourceAttr("data.backstage_location.test", "spec.targets.0", "./components/artist-lookup-component.yaml"),
				)},
		},
	})
}

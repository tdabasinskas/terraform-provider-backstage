package backstage

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_resource.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "kind", "Resource"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "metadata.annotations.backstage.io/managed-by-origin-location",
						"url:https://github.com/backstage/backstage/blob/master/packages/catalog-model/examples/all-resources.yaml"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "metadata.description", "Stores artist details"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "relations.0.type", "dependencyOf"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "spec.system", "artist-engagement-portal"),
				),
			},
		},
	})
}

const testAccDataSourceResourceConfig = `
data "backstage_resource" "test" {
  name = "artists-db"
}
`

func TestAccDataSourceResource_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_resource" "test" {
						name = "artist_not_found_resource_a9ab8"
						fallback = {
							name = "fallback_artists"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_resource.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "kind", "Resource"),
					resource.TestCheckResourceAttr("data.backstage_resource.test", "name", "fallback_artists"),
				),
			},
		},
	})
}

func TestAccDataSourceResource_WithoutFallback_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_resource" "test" {
						name = "artist_not_found_resource_a9ab8"
					}
				`,
				ExpectError: regexp.MustCompile(`default/artist_not_found_resource_a9ab8: 404 Not Found`),
			},
		},
	})
}

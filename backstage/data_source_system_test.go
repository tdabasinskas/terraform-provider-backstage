package backstage

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSystem(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceSystemConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_system.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "kind", "System"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "metadata.annotations.backstage.io/edit-url",
						"https://github.com/backstage/backstage/edit/master/packages/catalog-model/examples/systems/artist-engagement-portal-system.yaml"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "metadata.description", "Everything related to artists"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "relations.0.type", "hasPart"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "spec.domain", "artists"),
				),
			},
		},
	})
}

const testAccDataSourceSystemConfig = `
data "backstage_system" "test" {
  name = "artist-engagement-portal"
}
`

func TestAccDataSourceSystem_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_system" "test" {
						name = "system_not_found_humungous_a9ab8"
						fallback = {
							name = "fallback_system"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_system.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "kind", "System"),
					resource.TestCheckResourceAttr("data.backstage_system.test", "name", "fallback_system"),
				),
			},
		},
	})
}

func TestAccDataSourceSystem_WithoutFallback_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_system" "test" {
						name = "system_not_found_huppla_a9ab8"
					}
				`,
				ExpectError: regexp.MustCompile(`404 Not Found`),
			},
		},
	})
}

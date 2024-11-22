package backstage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceComponent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceComponentConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_component.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "kind", "Component"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "metadata.annotations.backstage.io/managed-by-location",
						"url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/components/shuffle-api-component.yaml"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "metadata.description", "Shuffle API"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "metadata.tags.0", "go"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "relations.0.target_ref", "user:default/guest"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "spec.system", "audio-playback"),
				),
			},
		},
	})
}

const testAccDataSourceComponentConfig = `
data "backstage_component" "test" {
  name = "shuffle-api"
}
`

func TestAccDataSourceComponent_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_component" "test" {
						name = "non_existent_component_a9ab8"
						namespace = "default"
						fallback = {
							id = "123456"
							kind = "Component"
							name = "fallback_component"
							namespace = "default"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_component.test", "kind", "Component"),
					resource.TestCheckResourceAttr("data.backstage_component.test", "name", "fallback_component"),
					resource.TestCheckNoResourceAttr("data.backstage_component.test", "api_version"),
					resource.TestCheckNoResourceAttr("data.backstage_component.test", "api_version"),
					resource.TestCheckNoResourceAttr("data.backstage_component.test", "metadata"),
				),
			},
		},
	})
}

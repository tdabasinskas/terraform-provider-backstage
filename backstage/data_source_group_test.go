package backstage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceGroupConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_group.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "kind", "Group"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "metadata.description", "Team A"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "metadata.annotations.backstage.io/source-location",
						"url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/acme/"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "relations.0.type", "childOf"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "spec.parent", "backstage"),
				),
			},
		},
	})
}

const testAccDataSourceGroupConfig = `
data "backstage_group" "test" {
  name = "team-a"
}
`

func TestAccDataSourceGroup_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_group" "test" {
						name = "team_does_not_compute_a9ab8"
						fallback = {
							api_version = "backstage.io/v1alpha1"
							name = "fallback_team"
							kind = "Group"
							metadata = {
								description = "Fallback Team A"
								annotations = {
									"backstage.io/source-location" = "url:fallback.com"
								}
							}
							spec = {
								parent = "backstage"
							}
							relations = [
								{
									type = "childOf"
									target_ref = "backstage"
								}
							]
						}
					}	
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_group.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "kind", "Group"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "metadata.description", "Fallback Team A"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "metadata.annotations.backstage.io/source-location",
						"url:fallback.com"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "relations.0.type", "childOf"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "spec.parent", "backstage"),
					resource.TestCheckResourceAttr("data.backstage_group.test", "name", "fallback_team"),
				),
			},
		},
	})
}

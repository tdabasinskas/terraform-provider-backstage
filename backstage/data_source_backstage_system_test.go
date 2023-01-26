package backstage

import (
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

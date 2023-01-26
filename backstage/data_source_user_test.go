package backstage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceUserConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_user.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "kind", "User"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "metadata.annotations.backstage.io/managed-by-location",
						"url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/acme/team-a-group.yaml"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "metadata.name", "janelle.dawe"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "relations.0.target_ref", "group:default/team-a"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "spec.member_of.0", "team-a"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "spec.profile.display_name", "Janelle Dawe"),
				),
			},
		},
	})
}

const testAccDataSourceUserConfig = `
data "backstage_user" "test" {
  name = "janelle.dawe"
}
`

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
					resource.TestCheckResourceAttr("data.backstage_group.test", "spec.profile.picture",
						"https://avatars.dicebear.com/api/identicon/team-a@example.com.svg?background=%23fff&margin=25"),
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

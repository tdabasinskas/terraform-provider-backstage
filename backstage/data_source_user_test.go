package backstage

import (
	"regexp"
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

func TestAccDataSourceUser_WithFallback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_user" "test" {
						name = "user_not_found_blasted_a9ab8"
						fallback = {
							name = "fallback_user"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_user.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "kind", "System"),
					resource.TestCheckResourceAttr("data.backstage_user.test", "name", "fallback_user"),
				),
			},
		},
	})
}

func TestAccDataSourceUser_WithoutFallback_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "backstage_user" "test" {
						name = "user_not_found_emoji_a9ab8"
					}
				`,
				ExpectError: regexp.MustCompile(`default/user_not_found_emoji_a9ab8`),
			},
		},
	})
}

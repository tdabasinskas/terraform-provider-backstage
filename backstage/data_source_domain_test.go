package backstage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceDomainConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_domain.test", "api_version", "backstage.io/v1alpha1"),
					resource.TestCheckResourceAttr("data.backstage_domain.test", "kind", "Domain"),
					resource.TestCheckResourceAttr("data.backstage_domain.test", "metadata.annotations.backstage.io/source-location",
						"url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/domains/"),
					resource.TestCheckResourceAttr("data.backstage_domain.test", "metadata.description", "Everything related to artists"),
					resource.TestCheckResourceAttr("data.backstage_domain.test", "relations.0.target.name", "artist-engagement-portal"),
					resource.TestCheckResourceAttr("data.backstage_domain.test", "spec.owner", "team-a"),
				),
			},
		},
	})
}

const testAccDataSourceDomainConfig = `
data "backstage_domain" "test" {
  name = "artists"
}
`

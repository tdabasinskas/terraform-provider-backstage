package backstage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEntities(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + testAccDataSourceEntitiesConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.backstage_entities.test", "id", fmt.Sprint([]string{
						"kind=user,metadata.name=janelle.dawe",
						"metadata.uid=b70bed0a-e955-4f93-b03f-e03f7954da99",
					})),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.#", "2"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.0.kind", "User"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.0.metadata.annotations.backstage.io/managed-by-location",
						"url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/acme/team-a-group.yaml"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.0.metadata.name", "janelle.dawe"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.1.metadata.uid", "b70bed0a-e955-4f93-b03f-e03f7954da99"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.0.relations.0.target_ref", "group:default/team-a"),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.1.relations.0.target_ref", "group:default/team-a"),
				),
			},
		},
	})
}

const testAccDataSourceEntitiesConfig = `
data "backstage_entities" "test" {
  filters = [
    "kind=user,metadata.name=janelle.dawe",
    "metadata.uid=b70bed0a-e955-4f93-b03f-e03f7954da99",
  ]
}
`

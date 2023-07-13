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
						"kind=component,metadata.description=Searcher",
					})),
					resource.TestCheckResourceAttr("data.backstage_entities.test", "entities.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.backstage_entities.test", "entities.*", map[string]string{
						"kind":                   "Component",
						"metadata.name":          "searcher",
						"relations.0.target_ref": "user:default/guest",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.backstage_entities.test", "entities.*", map[string]string{
						"kind": "User",
						"metadata.annotations.backstage.io/managed-by-location": "url:https://github.com/backstage/backstage/tree/master/packages/catalog-model/examples/acme/team-a-group.yaml",
						"relations.0.target_ref":                                "group:default/team-a",
					}),
				),
			},
		},
	})
}

const testAccDataSourceEntitiesConfig = `
data "backstage_entities" "test" {
  filters = [
    "kind=user,metadata.name=janelle.dawe",
    "kind=component,metadata.description=Searcher",
  ]
}
`

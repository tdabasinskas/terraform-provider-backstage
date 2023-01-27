//go:build !resources

package backstage

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceLocation(t *testing.T) {
	if os.Getenv("ACCTEST_SKIP_RESOURCE_TEST") != "" {
		t.Skip("Skipping as ACCTEST_SKIP_RESOURCE_LOCATION is set")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create testing
			{
				Config: testAccProviderConfig + testAccResourceLocationConfig1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("backstage_location.test", "type", "url"),
					resource.TestCheckResourceAttr("backstage_location.test", "target", "http://test1"),
					resource.TestCheckResourceAttrSet("backstage_location.test", "id"),
					resource.TestCheckResourceAttrSet("backstage_location.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "backstage_location.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: testAccProviderConfig + testAccResourceLocationConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("backstage_location.test", "type", "url"),
					resource.TestCheckResourceAttr("backstage_location.test", "target", "http://test2"),
				),
			},
		},
	})

}

const testAccResourceLocationConfig1 = `
resource "backstage_location" "test" {
  target = "http://test1"
}
`
const testAccResourceLocationConfig2 = `
resource "backstage_location" "test" {
  target = "http://test2"
}
`

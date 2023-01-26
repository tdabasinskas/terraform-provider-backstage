package backstage

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const testAccProviderConfig = `
provider "backstage" {
	base_url = "https://demo.backstage.io"
}
`

// testAccProtoV6ProviderFactories are used to instantiate a provider during acceptance testing. The factory function will be invoked for  every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"backstage": providerserver.NewProtocol6WithError(New("test")()),
}

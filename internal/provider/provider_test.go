package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	providerFactory = map[string]func() (tfprotov6.ProviderServer, error){
		"httpclient": providerserver.NewProtocol6WithError(
			New("test")(),
		),
	}
)

package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-httpclient/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		provider.New(version),
		providerserver.ServeOpts{
			ProtocolVersion: 6,
			Address:         "registry.terraform.io/kvaidas/httpclient",
			Debug:           debug,
		},
	)

	if err != nil {
		log.Fatal(err.Error())
	}
}

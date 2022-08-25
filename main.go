package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/alphagov/terraform-provider-concourse/pkg/provider"
)

func main() {
	log.SetPrefix("[DEBUG] ")

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider.Provider()
		},
	})
}

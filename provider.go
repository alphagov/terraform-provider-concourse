package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"target": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("FLY_TARGET", nil),
				Description: "Target as in 'fly --target'",
			},
		},
		ConfigureFunc: ProviderConfigurationBuilder,

		DataSourcesMap: map[string]*schema.Resource{
			"concourse_pipeline": dataPipeline(),
			"concourse_team":     dataTeam(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"concourse_pipeline": resourcePipeline(),
			"concourse_team":     resourceTeam(),
		},
	}
}

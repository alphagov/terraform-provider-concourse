package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"target": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("FLY_TARGET", nil),
				Description: "Target as in 'fly --target', do not use if using team/username/password",
				Optional:    true,
			},
			"url": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("FLY_URL", nil),
				Description: "URL, do not use if using target ",
				Optional:    true,
			},
			"team": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("FLY_TEAM", nil),
				Description: "Team name, do not use if using target ",
				Optional:    true,
			},
			"username": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("FLY_USERNAME", nil),
				Description: "Username, do not use if using target",
				Optional:    true,
			},
			"password": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("FLY_PASSWORD", nil),
				Description: "Password, do not use if using target ",
				Optional:    true,
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

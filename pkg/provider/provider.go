package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"ca_cert_file": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("REQUESTS_CA_BUNDLE", nil),
				Description: "Client CA certificates that are used while connecting to concourse",
				Optional:    true,
			},
			"insecure_skip_verify": {
				Type:        schema.TypeBool,
				Description: "Set this to true if concourse uses a custom certificate",
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
			"concourse_teams":    dataTeams(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"concourse_pipeline": resourcePipeline(),
			"concourse_team":     resourceTeam(),
		},
	}
}

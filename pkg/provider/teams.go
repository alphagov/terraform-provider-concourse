package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataTeamsReads,
		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataTeamsReads(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teams, err := client.ListTeams()
	if err != nil {
		return diag.FromErr(err)
	}

	var names []string

	for _, team := range teams {
		names = append(names, team.Name)
	}

	d.SetId("concourse_teams")
	if err := d.Set("names", names); err != nil {
		return diag.Errorf("error setting team names: %s", err)
	}

	return nil
}

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataTeams() *schema.Resource {
	return &schema.Resource{
		Read: dataTeamsReads,
		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataTeamsReads(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teams, err := client.ListTeams()
	if err != nil {
		return err
	}

	var names []string

	for _, team := range teams {
		names = append(names, team.Name)
	}

	d.SetId("concourse_teams")
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting team names: %s", err)
	}

	return nil
}

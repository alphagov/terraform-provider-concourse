package provider

import (
	"context"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var roleNames = []string{
	"owner",
	"member",
	"pipeline-operator",
	"viewer",
}

var roleTypes = []string{
	"users",
	"groups",
}

func dataTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataTeamRead,

		Schema: map[string]*schema.Schema{
			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"owners": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"members": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"pipeline_operators": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"viewers": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,

		Schema: map[string]*schema.Schema{

			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"owners": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Set:      schema.HashString,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"members": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"pipeline_operators": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"viewers": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: resourceTeamResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTeamStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

type teamHelper struct {
	TeamName          string
	Owners            []interface{}
	Members           []interface{}
	PipelineOperators []interface{}
	Viewers           []interface{}
}

func (t *teamHelper) appendElem(field string, elem string) {
	switch field {
	case "owner":
		t.Owners = append(t.Owners, elem)
	case "member":
		t.Members = append(t.Members, elem)
	case "pipeline-operator":
		t.PipelineOperators = append(t.PipelineOperators, elem)
	case "viewer":
		t.Viewers = append(t.Viewers, elem)
	}
}

func readTeam(
	ctx context.Context,
	client concourse.Client,
	teamName string,
) (teamHelper, diag.Diagnostics) {

	retVal := teamHelper{
		TeamName: teamName,
	}

	teams, err := client.ListTeams()

	if err != nil {
		return retVal, diag.FromErr(err)
	}

	var foundTeam *atc.Team

	for _, team := range teams {
		if team.Name == teamName {
			foundTeam = &team
			break
		}
	}

	if foundTeam == nil {
		return retVal, diag.Errorf("Could not find team %s", teamName)
	}

	var (
		ok   bool
		role map[string][]string
	)

	for _, roleName := range roleNames {
		if role, ok = foundTeam.Auth[roleName]; !ok {
			continue
		}

		users, user_ok := role["users"]
		groups, group_ok := role["groups"]

		if user_ok {
			for _, user := range users {
				retVal.appendElem(roleName, "user:"+user)
			}
		}

		if group_ok {
			for _, group := range groups {
				retVal.appendElem(roleName, "group:"+group)
			}
		}
	}

	return retVal, nil
}

func dataTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	team, err := readTeam(ctx, client, teamName)

	if err != nil {
		return err
	}

	d.SetId(teamName)
	d.Set("owners", schema.NewSet(schema.HashString, team.Owners))
	d.Set("members", schema.NewSet(schema.HashString, team.Members))
	d.Set("pipeline_operators", schema.NewSet(schema.HashString, team.PipelineOperators))
	d.Set("viewers", schema.NewSet(schema.HashString, team.Viewers))
	return nil
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTeamCreateUpdate(ctx, d, m, true)
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTeamCreateUpdate(ctx, d, m, false)
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	team, err := readTeam(ctx, client, teamName)

	if err != nil {
		return err
	}

	d.SetId(teamName)
	d.Set("owners", schema.NewSet(schema.HashString, team.Owners))
	d.Set("members", schema.NewSet(schema.HashString, team.Members))
	d.Set("pipeline_operators", schema.NewSet(schema.HashString, team.PipelineOperators))
	d.Set("viewers", schema.NewSet(schema.HashString, team.Viewers))
	return nil
}

func resourceTeamCreateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, create bool) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)
	auths := make(map[string][]string)

	var authline []string
	roleEnabled := make(map[string]bool)

	// fetches input from terraform and breaks out user/groups that prepend the string
	for _, role := range roleNames {

		// concourse calls things: "pipeline-operator", terraform calls them: "pipeline_operators"
		terraformRoleName := strings.ReplaceAll(role, "-", "_") + "s"

		for _, terraformInput := range d.Get(terraformRoleName).(*schema.Set).List() {
			roleEnabled[role] = true
			authline = strings.Split(terraformInput.(string), ":")
			switch authline[0] {
			case "user":
				auths[role+"_users"] = append(auths[role+"_users"], strings.Join(authline[1:], ":"))
			case "group":
				auths[role+"_groups"] = append(auths[role+"_groups"], strings.Join(authline[1:], ":"))
			}
		}
	}

	teamDetails := atc.Team{
		Name: teamName,
		Auth: atc.TeamAuth{},
	}

	// we cant set a role into the TeamAuth struct if it doesnt exist
	// otherwise sending the atc.Team to concourse creates "role": null entries
	for _, role := range roleNames {
		if roleEnabled[role] == true {
			teamDetails.Auth[role] = map[string][]string{}
			for _, roleType := range roleTypes {
				roleValues := auths[role+"_"+roleType]
				if len(roleValues) > 0 {
					teamDetails.Auth[role][roleType] = roleValues
				}
			}
		}
	}

	team := client.Team(teamName)

	if d.HasChange("team_name") && !create {
		_, warnings, err := team.RenameTeam(d.Id(), d.Get("team_name").(string))

		if err != nil {
			return diag.Errorf("Could not rename team %s %s", teamName, SerializeWarnings(warnings))
		}
	}

	_, created, updated, warnings, err := team.CreateOrUpdate(teamDetails)

	if err != nil {
		return diag.Errorf("Error creating/updating team %s: %s %s", teamName, err, SerializeWarnings(warnings))
	}

	if !created && !updated {
		return diag.Errorf("Could not create or update team %s", teamName)
	}

	return resourceTeamRead(ctx, d, m)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	if teamName == "main" {
		return diag.Errorf("Cannot delete main team")
	}

	team := client.Team(teamName)

	err := team.DestroyTeam(teamName)

	if err != nil {
		return diag.Errorf("Could not delete team %s: %s", teamName, err)
	}

	d.SetId("")
	return nil
}

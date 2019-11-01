package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform/helper/schema"
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
		Read: dataTeamRead,

		Schema: map[string]*schema.Schema{
			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"owners": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"members": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"pipeline_operators": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"viewers": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTeamCreate,
		Read:   resourceTeamRead,
		Update: resourceTeamUpdate,
		Delete: resourceTeamDelete,

		Schema: map[string]*schema.Schema{

			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"owners": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"members": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"pipeline_operators": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"viewers": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type teamHelper struct {
	TeamName          string
	Owners            []string
	Members           []string
	PipelineOperators []string
	Viewers           []string
}

func (t *teamHelper) sort() {
	sort.Strings(t.Owners)
	sort.Strings(t.Members)
	sort.Strings(t.PipelineOperators)
	sort.Strings(t.Viewers)
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
	client concourse.Client,
	teamName string,
) (teamHelper, error) {

	retVal := teamHelper{
		TeamName: teamName,
	}

	teams, err := client.ListTeams()

	if err != nil {
		return retVal, err
	}

	var foundTeam *atc.Team

	for _, team := range teams {
		if team.Name == teamName {
			foundTeam = &team
			break
		}
	}

	if foundTeam == nil {
		return retVal, fmt.Errorf("Could not find team %s", teamName)
	}

	var role map[string][]string
	var ok bool
	for _, roleName := range roleNames {
		if role, ok = foundTeam.Auth[roleName]; !ok {

			// owner must exist
			if roleName == "owner" {
				return retVal, fmt.Errorf(
					"Could not find any details for role %s in team %s",
					roleName,
					teamName,
				)
			}

			continue
		}

		users, user_ok := role["users"]
		groups, group_ok := role["groups"]

		if !user_ok && !group_ok {
			return retVal, fmt.Errorf(
				"Could not find users or group field for role %s in team %s",
				roleName,
				teamName,
			)
		}

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

	retVal.sort()
	return retVal, nil
}

func dataTeamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	team, err := readTeam(client, teamName)

	if err != nil {
		return err
	}

	d.SetId(teamName)
	d.Set("owners", team.Owners)
	d.Set("members", team.Members)
	d.Set("pipeline_operators", team.PipelineOperators)
	d.Set("viewers", team.Viewers)
	return nil
}

func resourceTeamCreate(d *schema.ResourceData, m interface{}) error {
	return resourceTeamUpdate(d, m)
}

func resourceTeamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	team, err := readTeam(client, teamName)

	if err != nil {
		return err
	}

	d.SetId(teamName)
	d.Set("owners", team.Owners)
	d.Set("members", team.Members)
	d.Set("pipeline_operators", team.PipelineOperators)
	d.Set("viewers", team.Viewers)
	return nil
}

func resourceTeamUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)
	auths := make(map[string][]string)

	var authline []string
	roleEnabled := make(map[string]bool)

	// fetches input from terraform and breaks out user/groups that prepend the string
	for _, role := range roleNames {

		// concourse calls things: "pipeline-operator", terraform calls them: "pipeline_operators"
		terraformRoleName := strings.ReplaceAll(role, "-", "_") + "s"

		for _, terraformInput := range d.Get(terraformRoleName).([]interface{}) {
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

	if d.HasChange("team_name") {
		_, err := team.RenameTeam(d.Id(), d.Get("team_name").(string))

		if err != nil {
			return fmt.Errorf("Could not rename team %s", teamName)
		}
	}

	_, created, updated, err := team.CreateOrUpdate(teamDetails)

	if err != nil {
		return fmt.Errorf("Error creating/updating team %s: %s", teamName, err)
	}

	if !created && !updated {
		return fmt.Errorf("Could not create or update team %s", teamName)
	}

	return resourceTeamRead(d, m)
}

func resourceTeamDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)

	if teamName == "main" {
		return fmt.Errorf("Cannot delete main team")
	}

	team := client.Team(teamName)

	err := team.DestroyTeam(teamName)

	if err != nil {
		return fmt.Errorf("Could not delete team %s: %s", teamName, err)
	}

	d.SetId("")
	return nil
}

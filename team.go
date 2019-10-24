package main

import (
	"fmt"
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

func (t *teamHelper) appendElem(field string, elem string) {
	switch field {
	case "owner":
		t.Owners = append(t.Owners, elem)
	case "member":
		t.Members = append(t.Members, elem)
	case "pipeline-operator":
		t.PipelineOperators = append(t.PipelineOperators, elem)
	case "viewer":
		t.PipelineOperators = append(t.Viewers, elem)
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

	var ok, user_ok, group_ok bool
	var role map[string][]string
	var groups []string
	var users []string

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

		users, user_ok = role["users"]
		groups, group_ok = role["groups"]

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

	for _, owners := range d.Get("owners").([]interface{}) {
		authline = strings.Split(owners.(string), ":")
		switch authline[0] {
		case "user":
			auths["owner_users"] = append(auths["owner_users"], strings.Join(authline[1:], ":"))
		case "group":
			auths["owner_groups"] = append(auths["owner_groups"], strings.Join(authline[1:], ":"))
		}
	}

	for _, members := range d.Get("members").([]interface{}) {
		authline = strings.Split(members.(string), ":")
		switch authline[0] {
		case "user":
			auths["member_users"] = append(auths["member_users"], strings.Join(authline[1:], ":"))
		case "group":
			auths["member_groups"] = append(auths["member_groups"], strings.Join(authline[1:], ":"))
		}
	}

	for _, pipelineOperators := range d.Get("pipeline_operators").([]interface{}) {
		authline = strings.Split(pipelineOperators.(string), ":")
		switch authline[0] {
		case "user":
			auths["pipeline_operator_users"] = append(auths["pipeline_operator_users"], strings.Join(authline[1:], ":"))
		case "group":
			auths["pipeline_operator_groups"] = append(auths["pipeline_operator_groups"], strings.Join(authline[1:], ":"))
		}
	}

	for _, viewers := range d.Get("viewers").([]interface{}) {
		authline = strings.Split(viewers.(string), ":")
		switch authline[0] {
		case "user":
			auths["viewer_users"] = append(auths["viewer_users"], strings.Join(authline[1:], ":"))
		case "group":
			auths["viewer_groups"] = append(auths["viewer_groups"], strings.Join(authline[1:], ":"))
		}
	}

	teamDetails := atc.Team{
		Name: teamName,
		Auth: atc.TeamAuth{
			"owner": map[string][]string{
				"users":  auths["owner_users"],
				"groups": auths["owner_groups"],
			},
			"member": map[string][]string{
				"users":  auths["owner_users"],
				"groups": auths["owner_groups"],
			},
			"pipeline-operator": map[string][]string{
				"users":  auths["pipeline_operator_users"],
				"groups": auths["pipeline_operator_groups"],
			},
			"viewer": map[string][]string{
				"users":  auths["viewer_users"],
				"groups": auths["viewer_groups"],
			},
		},
	}

	team := client.Team(teamName)

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

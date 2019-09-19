package main

import (
	"fmt"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform/helper/schema"
)

var roleName = "owner"

func dataTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataTeamRead,

		Schema: map[string]*schema.Schema{
			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"groups": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"users": &schema.Schema{
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

			"groups": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"users": &schema.Schema{
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
	TeamName string
	Groups   []string
	Users    []string
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

	var ok bool
	var role map[string][]string
	if role, ok = foundTeam.Auth[roleName]; !ok {
		return retVal, fmt.Errorf(
			"Could not find any details for role %s in team %s",
			roleName,
			teamName,
		)
	}

	var groups []string
	if groups, ok = role["groups"]; !ok {
		return retVal, fmt.Errorf(
			"Could not find groups field for role %s in team %s",
			roleName,
			teamName,
		)
	}
	retVal.Groups = groups

	var users []string
	if users, ok = role["users"]; !ok {
		return retVal, fmt.Errorf(
			"Could not find users field in team %s",
			teamName,
		)
	}
	retVal.Users = users

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
	d.Set("groups", team.Groups)
	d.Set("users", team.Users)

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
	d.Set("groups", team.Groups)
	d.Set("users", team.Users)

	return nil
}

func resourceTeamUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	teamName := d.Get("team_name").(string)
	groups := make([]string, 0)
	users := make([]string, 0)

	for _, group := range d.Get("groups").([]interface{}) {
		groups = append(groups, group.(string))
	}

	for _, user := range d.Get("users").([]interface{}) {
		users = append(users, user.(string))
	}

	teamDetails := atc.Team{
		Name: teamName,
		Auth: atc.TeamAuth{
			roleName: map[string][]string{
				"groups": groups,
				"users":  users,
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

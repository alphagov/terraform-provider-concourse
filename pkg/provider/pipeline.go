package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataPipeline() *schema.Resource {
	return &schema.Resource{
		Read: dataPipelineRead,

		Schema: map[string]*schema.Schema{
			"pipeline_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"is_exposed": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Computed: true,
			},

			"is_paused": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Computed: true,
			},

			"yaml": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},

			"json": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},
		},
	}
}

func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourcePipelineCreate,
		Read:   resourcePipelineRead,
		Update: resourcePipelineUpdate,
		Delete: resourcePipelineDelete,

		Schema: map[string]*schema.Schema{
			"pipeline_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"is_exposed": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},

			"is_paused": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},

			"pipeline_config_format": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"pipeline_config": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"json": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"yaml": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type pipelineHelper struct {
	TeamName      string
	PipelineName  string
	IsExposed     bool
	IsPaused      bool
	JSON          string
	YAML          string
	ConfigVersion string
}

func pipelineID(teamName string, pipelineName string) string {
	return fmt.Sprintf("%s:%s", teamName, pipelineName)
}

func readPipeline(
	client concourse.Client,
	teamName string,
	pipelineName string,
) (pipelineHelper, bool, error) {

	retVal := pipelineHelper{
		TeamName:      teamName,
		PipelineName:  pipelineName,
		ConfigVersion: "0",
	}

	team := client.Team(teamName)

	pipeline, pipelineFound, err := team.Pipeline(pipelineName)

	if err != nil {
		return retVal, false, err
	}

	if !pipelineFound {
		return retVal, false, nil
	}

	atcConfig, version, pipelineCfgFound, err := team.PipelineConfig(
		pipelineName,
	)

	if err != nil {
		return retVal, false, fmt.Errorf(
			"Error looking up pipeline %s within team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if !pipelineCfgFound {
		return retVal, false, nil
	}

	pipelineCfg, err := json.Marshal(atcConfig)
	if err != nil {
		return retVal, false, nil
	}

	pipelineCfgJSON, err := JSONToJSON(string(pipelineCfg))
	if err != nil {
		return retVal, false, fmt.Errorf(
			"Encountered error parsing pipeline %s config within team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	pipelineCfgYAML, err := JSONToYAML(pipelineCfgJSON)

	if err != nil {
		return retVal, false, fmt.Errorf(
			"Encountered error parsing pipeline %s config within team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	retVal.IsExposed = pipeline.Public
	retVal.IsPaused = pipeline.Paused
	retVal.ConfigVersion = version
	retVal.JSON = pipelineCfgJSON
	retVal.YAML = pipelineCfgYAML

	return retVal, true, nil
}

func dataPipelineRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)

	pipeline, wasFound, err := readPipeline(client, teamName, pipelineName)

	if err != nil {
		return fmt.Errorf(
			"Error reading pipeline %s from team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if wasFound {
		d.SetId(pipelineID(teamName, pipelineName))
		d.Set("is_exposed", pipeline.IsExposed)
		d.Set("is_paused", pipeline.IsPaused)
		d.Set("json", pipeline.JSON)
		d.Set("yaml", pipeline.YAML)
	} else {
		d.SetId("")
	}

	return nil
}

func resourcePipelineCreate(d *schema.ResourceData, m interface{}) error {
	return resourcePipelineUpdate(d, m)
}

func resourcePipelineRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)

	pipeline, wasFound, err := readPipeline(client, teamName, pipelineName)

	if err != nil {
		return fmt.Errorf(
			"Error reading pipeline %s from team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if wasFound {
		d.SetId(pipelineID(teamName, pipelineName))
		d.Set("is_exposed", pipeline.IsExposed)
		d.Set("is_paused", pipeline.IsPaused)
		d.Set("json", pipeline.JSON)
		d.Set("yaml", pipeline.YAML)
	} else {
		d.SetId("")
	}

	return nil
}

func resourcePipelineUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client

	if d.HasChange("team_name") && d.Id() != "" {
		// Concourse does not yet have support for moving pipeline, so we should
		// delete the pipeline from the old team

		oldTeamName := strings.SplitN(d.Id(), ":", 2)[0]
		oldPipelineName := strings.SplitN(d.Id(), ":", 2)[1]

		team := client.Team(oldTeamName)
		_, err := team.DeletePipeline(oldPipelineName)

		if err != nil {
			return fmt.Errorf(
				"Error deleting old pipeline %s in team %s: %s",
				oldPipelineName, oldTeamName, err,
			)
		}
	}

	if d.HasChange("pipeline_name") && !d.HasChange("team_name") && d.Id() != "" {
		teamName := strings.SplitN(d.Id(), ":", 2)[0]
		oldPipelineName := strings.SplitN(d.Id(), ":", 2)[1]
		newPipelineName := d.Get("pipeline_name").(string)

		team := client.Team(teamName)

		_, warnings, err := team.RenamePipeline(oldPipelineName, newPipelineName)

		if err != nil {
			return fmt.Errorf(
				"Error renaming pipeline %s to %s in team %s: %s %s",
				oldPipelineName, newPipelineName, teamName, err, SerializeWarnings(warnings),
			)
		}
	}

	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)
	d.SetId(pipelineID(teamName, pipelineName))
	team := client.Team(teamName)

	pipelineConfig := d.Get("pipeline_config").(string)
	pipelineConfigFormat := d.Get("pipeline_config_format").(string)

	pipeline, _, err := readPipeline(client, teamName, pipelineName)

	if err != nil {
		return fmt.Errorf(
			"Error looking up pipeline %s in team %s: %s",
			pipelineName, teamName, err,
		)
	}

	parsedJSON, err := ParsePipelineConfig(pipelineConfig, pipelineConfigFormat)

	if err != nil {
		return fmt.Errorf("Error parsing pipeline_config: %s", err)
	}

	_, _, configWarnings, err := team.CreateOrUpdatePipelineConfig(
		pipelineName, pipeline.ConfigVersion, []byte(parsedJSON), false,
	)

	if err != nil {
		return fmt.Errorf(
			"Encountered error setting config for pipeline %s in team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if len(configWarnings) != 0 {
		warnings := ""
		for _, w := range configWarnings {
			warnings += fmt.Sprintf("%s: %s\n", w.Type, w.Message)
		}

		return fmt.Errorf(
			"Encountered pipeline warnings (%s/%s):\n %s",
			pipelineName, teamName, warnings,
		)
	}

	if d.Get("is_exposed").(bool) {
		found, err := team.ExposePipeline(pipelineName)
		if err != nil {
			return fmt.Errorf(
				"Error exposing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return fmt.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	} else {
		found, err := team.HidePipeline(pipelineName)
		if err != nil {
			return fmt.Errorf(
				"Error hiding pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return fmt.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	}

	if d.Get("is_paused").(bool) {
		found, err := team.PausePipeline(pipelineName)
		if err != nil {
			return fmt.Errorf(
				"Error pausing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return fmt.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	} else {
		found, err := team.UnpausePipeline(pipelineName)
		if err != nil {
			return fmt.Errorf(
				"Error unpausing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return fmt.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	}

	return resourcePipelineRead(d, m)
}

func resourcePipelineDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ProviderConfig).Client
	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)
	team := client.Team(teamName)

	deleted, err := team.DeletePipeline(pipelineName)

	if err != nil {
		return fmt.Errorf(
			"Could not delete pipeline %s from team %s: %s",
			pipelineName, teamName, err,
		)
	}

	if !deleted {
		return fmt.Errorf(
			"Could not delete pipeline %s from team %s", pipelineName, teamName,
		)
	}

	d.SetId("")
	return nil
}

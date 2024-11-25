package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataPipeline() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataPipelineRead,

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
		CreateContext: resourcePipelineCreate,
		ReadContext:   resourcePipelineRead,
		UpdateContext: resourcePipelineUpdate,
		DeleteContext: resourcePipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pipeline_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"team_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"json", "yaml"}, false)),
			},

			"pipeline_config": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"vars": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
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

func parsePipelineID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected ID format (%q). Expected team_name:pipeline_name", id)
	}
	return parts[0], parts[1], nil
}

func readPipeline(
	ctx context.Context,
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

	pipeline, pipelineFound, err := team.Pipeline(
		atc.PipelineRef{Name: pipelineName},
	)

	if err != nil {
		return retVal, false, err
	}

	if !pipelineFound {
		return retVal, false, nil
	}

	atcConfig, version, pipelineCfgFound, err := team.PipelineConfig(
		atc.PipelineRef{Name: pipelineName},
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

func dataPipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)

	pipeline, wasFound, err := readPipeline(ctx, client, teamName, pipelineName)

	if err != nil {
		return diag.Errorf(
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

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourcePipelineUpdate(ctx, d, m)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	teamName, pipelineName, err := parsePipelineID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pipeline, wasFound, err := readPipeline(ctx, client, teamName, pipelineName)

	if err != nil {
		return diag.Errorf(
			"Error reading pipeline %s from team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if wasFound {
		d.SetId(pipelineID(pipeline.TeamName, pipeline.PipelineName))
		d.Set("team_name", pipeline.TeamName)
		d.Set("pipeline_name", pipeline.PipelineName)
		d.Set("is_exposed", pipeline.IsExposed)
		d.Set("is_paused", pipeline.IsPaused)
		d.Set("json", pipeline.JSON)
		d.Set("yaml", pipeline.YAML)
	} else {
		d.SetId("")
	}

	return nil
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client

	if d.HasChange("pipeline_name") && d.Id() != "" {
		teamName := strings.SplitN(d.Id(), ":", 2)[0]
		oldPipelineName := strings.SplitN(d.Id(), ":", 2)[1]
		newPipelineName := d.Get("pipeline_name").(string)

		team := client.Team(teamName)

		_, warnings, err := team.RenamePipeline(oldPipelineName, newPipelineName)

		if err != nil {
			return diag.Errorf(
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
	vars := d.Get("vars").(map[string]interface{})

	pipeline, _, err := readPipeline(ctx, client, teamName, pipelineName)

	if err != nil {
		return diag.Errorf(
			"Error looking up pipeline %s in team %s: %s",
			pipelineName, teamName, err,
		)
	}

	parsedJSON, err := ParsePipelineConfig(pipelineConfig, pipelineConfigFormat, vars)

	if err != nil {
		return diag.Errorf("Error parsing pipeline_config: %s", err)
	}

	_, _, configWarnings, err := team.CreateOrUpdatePipelineConfig(
		atc.PipelineRef{Name: pipelineName}, pipeline.ConfigVersion, []byte(parsedJSON), false,
	)

	if err != nil {
		return diag.Errorf(
			"Encountered error setting config for pipeline %s in team '%s': %s",
			pipelineName, teamName, err,
		)
	}

	if len(configWarnings) != 0 {
		warnings := ""
		for _, w := range configWarnings {
			warnings += fmt.Sprintf("%s: %s\n", w.Type, w.Message)
		}

		return diag.Errorf(
			"Encountered pipeline warnings (%s/%s):\n %s",
			pipelineName, teamName, warnings,
		)
	}

	if d.Get("is_exposed").(bool) {
		found, err := team.ExposePipeline(
			atc.PipelineRef{Name: pipelineName},
		)
		if err != nil {
			return diag.Errorf(
				"Error exposing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return diag.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	} else {
		found, err := team.HidePipeline(
			atc.PipelineRef{Name: pipelineName},
		)
		if err != nil {
			return diag.Errorf(
				"Error hiding pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return diag.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	}

	if d.Get("is_paused").(bool) {
		found, err := team.PausePipeline(
			atc.PipelineRef{Name: pipelineName},
		)
		if err != nil {
			return diag.Errorf(
				"Error pausing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return diag.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	} else {
		found, err := team.UnpausePipeline(
			atc.PipelineRef{Name: pipelineName},
		)
		if err != nil {
			return diag.Errorf(
				"Error unpausing pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
		if !found {
			return diag.Errorf(
				"Could not find pipeline %s in team '%s': %s",
				pipelineName, teamName, err,
			)
		}
	}

	return resourcePipelineRead(ctx, d, m)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ProviderConfig).Client
	pipelineName := d.Get("pipeline_name").(string)
	teamName := d.Get("team_name").(string)
	team := client.Team(teamName)

	deleted, err := team.DeletePipeline(
		atc.PipelineRef{Name: pipelineName},
	)

	if err != nil {
		return diag.Errorf(
			"Could not delete pipeline %s from team %s: %s",
			pipelineName, teamName, err,
		)
	}

	if !deleted {
		return diag.Errorf(
			"Could not delete pipeline %s from team %s", pipelineName, teamName,
		)
	}

	d.SetId("")
	return nil
}

package provider

import (
	"fmt"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/concourse/concourse/vars"
	"github.com/ghodss/yaml"
	"strings"
)

// JSONToJSON ensures that keys are ordered, etc, by double converting
func JSONToJSON(inputJSON string) (string, error) {
	intermediateYAML, err := yaml.JSONToYAML([]byte(inputJSON))

	if err != nil {
		return "", err
	}

	outputJSON, err := yaml.YAMLToJSON(intermediateYAML)

	if err != nil {
		return "", err
	}

	return string(outputJSON), nil
}

// YAMLToJSON is just a wrapper for less type boilerplate
func YAMLToJSON(inputYAML string) (string, error) {
	outputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))

	if err != nil {
		return "", err
	}

	return string(outputJSON), nil
}

// JSONToYAML is just a wrapper for less type boilerplate
func JSONToYAML(inputJSON string) (string, error) {
	outputYAML, err := yaml.JSONToYAML([]byte(inputJSON))

	if err != nil {
		return "", err
	}

	return string(outputYAML), nil
}

// ParsePipelineConfig returns parsed/validated JSON
// from either YAML or JSON
func ParsePipelineConfig(
	pipelineConfig string,
	pipelineConfigFormat string,
	inputVars map[string]interface{},
) (string, error) {
	if pipelineConfigFormat != "json" && pipelineConfigFormat != "yaml" {
		return "", fmt.Errorf("pipeline_config_format must be json or yaml")
	}

	var err error
	outputJSON := ""

	if inputVars != nil {
		params := []vars.Variables{vars.StaticVariables(inputVars)}
		evaluatedConfig, err := vars.NewTemplateResolver([]byte(pipelineConfig), params).Resolve(false, false)
		if err != nil {
			return "", err
		}

		pipelineConfig = string(evaluatedConfig[:])
	}

	if pipelineConfigFormat == "json" {
		outputJSON, err = JSONToJSON(pipelineConfig)
		if err != nil {
			return "", err
		}
	}

	if pipelineConfigFormat == "yaml" {
		outputJSON, err = YAMLToJSON(pipelineConfig)
		if err != nil {
			return "", err
		}
	}

	return outputJSON, nil
}

func SerializeWarnings(warnings []concourse.ConfigWarning) string {
	var warningsMsg strings.Builder
	if len(warnings) > 0 {
		warningsMsg.WriteString(fmt.Sprintln())
		for _, warning := range warnings {
			warningsMsg.WriteString(fmt.Sprintf("  - %v\n", warning.Message))
		}
	}

	return warningsMsg.String()
}

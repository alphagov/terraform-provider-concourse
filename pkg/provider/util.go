package provider

import (
	"fmt"
	"strings"

	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/concourse/concourse/vars"
	"github.com/ghodss/yaml"
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
	inputYAMLVars map[string]interface{},
) (string, error) {
	var err error
	outputJSON := ""

	staticVars := map[string]interface{}{}
	for name, value := range inputVars {
		staticVars[name] = value
	}
	// yaml_vars values are YAML documents (fly's --yaml-var), decoded to structured values
	for name, value := range inputYAMLVars {
		var parsed interface{}
		if err := yaml.Unmarshal([]byte(fmt.Sprintf("%v", value)), &parsed); err != nil {
			return "", fmt.Errorf("could not parse yaml_var %q: %s", name, err)
		}
		staticVars[name] = parsed
	}

	if len(staticVars) > 0 {
		params := []vars.Variables{vars.StaticVariables(staticVars)}
		evaluatedConfig, err := vars.NewTemplateResolver([]byte(pipelineConfig), params).Resolve(false)
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

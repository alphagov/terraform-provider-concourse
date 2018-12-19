package main

import (
	"fmt"
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
) (string, error) {
	if pipelineConfigFormat != "json" && pipelineConfigFormat != "yaml" {
		return "", fmt.Errorf("pipeline_config_format must be json or yaml")
	}

	var err error
	outputJSON := ""

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

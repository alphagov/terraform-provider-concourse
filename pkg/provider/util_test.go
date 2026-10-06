package provider

import (
	"strings"
	"testing"
)

func TestParsePipelineConfigSubstitutesStringVars(t *testing.T) {
	config := "jobs:\n- name: ((job_name))\n"
	out, err := ParsePipelineConfig(config, "yaml", map[string]interface{}{"job_name": "hello"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !strings.Contains(out, `"hello"`) {
		t.Fatalf("expected substituted job name in %q", out)
	}
}

func TestParsePipelineConfigSubstitutesYAMLVars(t *testing.T) {
	config := "groups: ((groups))\n"
	out, err := ParsePipelineConfig(config, "yaml", nil, map[string]interface{}{
		"groups": "- name: all\n  jobs: [a, b]\n",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !strings.Contains(out, `"name":"all"`) {
		t.Fatalf("expected yaml var expanded into structured value, got %q", out)
	}
}

func TestParsePipelineConfigSubstitutesYAMLVarList(t *testing.T) {
	config := "jobs:\n- name: matrix\n  plan:\n  - across:\n    - var: version\n      values: ((versions))\n"
	out, err := ParsePipelineConfig(config, "yaml", nil, map[string]interface{}{
		"versions": "- \"1.0\"\n- \"2.0\"\n",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !strings.Contains(out, `"1.0"`) || !strings.Contains(out, `"2.0"`) {
		t.Fatalf("expected across values expanded from yaml_var, got %q", out)
	}
}

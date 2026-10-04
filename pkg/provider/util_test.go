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

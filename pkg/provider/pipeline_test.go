package provider

import (
	"strings"
	"testing"
)

func TestPipelineValidation(t *testing.T) {
	cases := []struct {
		Config   string
		Warnings []string
		Errors   []string
	}{
		{
			Config: `
---
jobs:
- name: job
  public: true
  plan:
  - task: simple-task
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: { repository: busybox }
      run:
        path: echo
        args: ["Hello world!"]
`,
		},
		{
			Config: ``,
			Errors: []string{"pipeline must contain at least one job"},
		},
		{
			Config: `
---
jobs:
- name: Jobs
  public: true
  plan:
  - task: simple-task
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: { repository: busybox }
      run:
        path: echo
        args: ["Hello world!"]
`,
			Warnings: []string{"must start with a lowercase letter"},
		},
	}
	for _, tc := range cases {
		warnings, errors, err := validatePipelineConfig([]byte(tc.Config))
		if err != nil {
			t.Fatal(err)
		}

		if len(tc.Errors) != len(errors) {
			t.Errorf("Expected %v errors, got %v errors", len(tc.Errors), len(errors))
		}

		for i, errMsg := range tc.Errors {
			if !strings.Contains(errors[i], errMsg) {
				t.Errorf("Expected error including '%v', got '%v'", errMsg, errors[i])
			}
		}

		if len(tc.Warnings) != len(warnings) {
			t.Errorf("Expected %v warnings, got %v warnings", len(tc.Warnings), len(warnings))
		}

		for i, warning := range tc.Warnings {
			if !strings.Contains(warnings[i], warning) {
				t.Errorf("Expected warning including '%v', got '%v'", warning, warnings[i])
			}
		}

	}
}

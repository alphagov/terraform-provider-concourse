package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"team_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owners": {
				Type:     schema.TypeList,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"members": {
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pipeline_operators": {
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"viewers": {
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

func resourceTeamStateUpgradeV0(
	_ context.Context,
	rawState map[string]interface{},
	meta interface{},
) (map[string]interface{}, error) {
	isNotDigit := func(c rune) bool { return c < '0' || c > '9' }

	rawStateOut := map[string]interface{}{}

	for k, v := range rawState {
		splitKey := strings.Split(k, ".")
		if len(splitKey) == 2 && strings.IndexFunc(splitKey[1], isNotDigit) == -1 {
			switch splitKey[0] {
			case
				"owners",
				"members",
				"pipeline_operators",
				"viewers":
				rawStateOut[fmt.Sprintf("%s.%d", splitKey[0], schema.HashString(v))] = v
				continue
			}
		}
		rawStateOut[k] = v
	}
	return rawStateOut, nil
}

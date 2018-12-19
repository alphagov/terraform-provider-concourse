package main

import (
	"fmt"
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform/helper/schema"
)

type ProviderConfig struct {
	Client concourse.Client
}

func ProviderConfigurationBuilder(
	d *schema.ResourceData,
) (interface{}, error) {

	targetName := rc.TargetName(d.Get("target").(string))

	target, err := rc.LoadTarget(targetName, false)

	if err != nil {
		return nil, fmt.Errorf("Error loading target: %s", err)
	}

	client := target.Client()

	return &ProviderConfig{
		Client: interface{}(client).(concourse.Client),
	}, nil
}

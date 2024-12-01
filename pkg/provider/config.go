package provider

import (
	"fmt"

	"github.com/concourse/concourse/fly/rc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/alphagov/terraform-provider-concourse/pkg/client"
)

type ProviderConfig struct {
	Client concourse.Client
}

func ProviderConfigurationBuilder(
	d *schema.ResourceData,
) (interface{}, error) {

	targetName := rc.TargetName(d.Get("target").(string))

	if targetName != "" {
		target, err := rc.LoadTarget(targetName, false)

		if err != nil {
			return nil, fmt.Errorf("Error loading target: %s", err)
		}

		return &ProviderConfig{
			Client: target.Client(),
		}, nil
	}

	url := d.Get("url").(string)
	team := d.Get("team").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	caFile := d.Get("ca_cert_file").(string)
	skipCertificateVerification := d.Get("insecure_skip_verify").(bool)

	if url != "" && team != "" && username != "" && password != "" {
		c, err := client.NewConcourseClient(
			url,
			team,
			username,
			password,
			caFile,
			skipCertificateVerification,
		)

		if err != nil {
			return nil, fmt.Errorf("Error creating client: %s", err)
		}

		return &ProviderConfig{
			Client: c,
		}, nil
	}

	return nil, fmt.Errorf(
		`Please specify "target" or "username", "password", "team", and "url"`,
	)
}

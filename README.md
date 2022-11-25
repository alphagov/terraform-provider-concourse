# terraform-provider-concourse

## What

A terraform provider for concourse

## Why

`fly` is an amazing tool, but configuration using scripts running fly is not
ideal.

## Prerequisites

Install `go`, and `terraform`.

## How to install and use

```
make install
```

## How to build and test for development

```
make
make integration-tests
```

# Example `terraform`

## Create a provider (using target from fly)

```hcl
provider "concourse" {
  target = "target_name"
}
```

## Create a provider (using a local username and password)

Note: this is not basic authentication

```hcl
provider "concourse" {
  url  = "https://wings.pivotal.io"
  team = "main"

  username = "localuser"
  password = "very-secure-password"
  ca_cert_file = "path-to-ca-file"
  insecure_skip_verify = false
}
```

## Look up all teams

```hcl
data "concourse_teams" "teams" {
}

output "team_names" {
  value = data.concourse_teams.teams.names
}
```

## Look up a team

```hcl
data "concourse_team" "my_team" {
  team_name = "main"
}

output "my_team_name" {
  value = data.concourse_team.my_team.team_name
}

output "my_team_owners" {
  value = data.concourse_team.my_team.owners
}

output "my_team_members" {
  value = data.concourse_team.my_team.members
}

output "my_team_pipeline_operators" {
  value = data.concourse_team.my_team.pipeline_operators
}

output "my_team_viewers" {
  value = data.concourse_team.my_team.viewers
}
```

## Look up a pipeline

```hcl
data "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "pipeline"
}

output "my_pipeline_team_name" {
  value = data.concourse_pipeline.my_pipeline.team_name
}

output "my_pipeline_pipeline_name" {
  value = data.concourse_pipeline.my_pipeline.pipeline_name
}

output "my_pipeline_is_exposed" {
  value = data.concourse_pipeline.my_pipeline.is_exposed
}

output "my_pipeline_is_paused" {
  value = data.concourse_pipeline.my_pipeline.is_paused
}

output "my_pipeline_json" {
  value = data.concourse_pipeline.my_pipeline.json
}

output "my_pipeline_yaml" {
  value = data.concourse_pipeline.my_pipeline.yaml
}
```
## Create a team

Supports `owners`, `members`, `pipeline_operators`, and `viewers`.

Specify users and groups by prefixing the strings:

* `user:`
* `group:`

```hcl
resource "concourse_team" "my_team" {
  team_name = "my-team"

  owners = [
    "group:github:org-name",
    "group:github:org-name:team-name",
    "user:github:tlwr",
  ]

  viewers = [
    "user:github:samrees"
  ]
}
```

## Create a pipeline

```hcl
resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("pipeline-config.yml")
  pipeline_config_format = "yaml"
}

# OR

resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("pipeline-config.json")
  pipeline_config_format = "json"
}
```

# terraform-provider-concourse

## What

A terraform provider for concourse

## Why

`fly` is an amazing tool, but configuration using scripts running fly is not
ideal.

## Disclaimer

This is an incredibly early piece of software and should not be relied upon for production use.

## Prerequisites

Install `go`, `glide`, and `terraform`.

## How to build and use

```
cd terraform-provider-concourse
glide install
go build
cp terraform-provider-concourse /path/to/my/project
cd /path/to/my/project
terraform init
terraform apply
```

# Example `terraform`

## Create a provider (using target from fly)

```
provider "concourse" {
  target = "target_name"
}
```

## Look up a team

```
data "concourse_team" "my_team" {
  team_name = "main"
}

output "my_team_name" {
  value = "${data.concourse_team.my_team.team_name}"
}

output "my_team_groups" {
  value = "${data.concourse_team.my_team.groups}"
}

output "my_team_users" {
  value = "${data.concourse_team.my_team.users}"
}
```

## Look up a pipeline

```
data "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "pipeline"
}

output "my_pipeline_team_name" {
  value = "${data.concourse_pipeline.my_pipeline.team_name}"
}

output "my_pipeline_pipeline_name" {
  value = "${data.concourse_pipeline.my_pipeline.pipeline_name}"
}

output "my_pipeline_is_exposed" {
  value = "${data.concourse_pipeline.my_pipeline.is_exposed}"
}

output "my_pipeline_is_paused" {
  value = "${data.concourse_pipeline.my_pipeline.is_paused}"
}

output "my_pipeline_json" {
  value = "${data.concourse_pipeline.my_pipeline.json}"
}

output "my_pipeline_yaml" {
  value = "${data.concourse_pipeline.my_pipeline.yaml}"
}
```
## Create a team

```
resource "concourse_team" "my_team" {
  team_name = "my-team"

  groups = [
    "github:org-name",
    "github:org-name:team-name",
  ]

  users = [
    "github:tlwr",
  ]
}
```

## Create a pipeline
```
resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = "${file("pipeline-config.yml")}"
  pipeline_config_format = "yaml"
}

# OR

resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = "${file("pipeline-config.json")}"
  pipeline_config_format = "json"
}
```

# Concourse Provider

A terraform provider for concourse

## Why

`fly` is an amazing tool, but configuration using scripts running fly is not
ideal.

## Example Usage

### Create a provider (using target from fly)

```hcl
provider "concourse" {
  target = "target_name"
}
```

### Create a provider (using a local username and password)

Note: this is not basic authentication

```hcl
provider "concourse" {
  url  = "https://wings.pivotal.io"
  team = "main"

  username = "localuser"
  password = "very-secure-password"
}
```

### Look up all teams

```hcl
data "concourse_teams" "teams" {
}

output "team_names" {
  value = data.concourse_teams.teams.names
}
```

### Look up a team

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

### Look up a pipeline

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

### Create a team

Supports `owners`, `members`, `pipeline_operators`, and `viewers`.

Specify users and groups by prefixing the strings:

* `user:`
* `group:`

The value after the prefix is the connector-qualified identity, so any Concourse
auth connector works, e.g. `github`, `oidc`, `ldap`, `saml`:

```hcl
resource "concourse_team" "my_team" {
  team_name = "my-team"

  owners = [
    "group:github:org-name",
    "group:github:org-name:team-name",
    "user:github:tlwr",
    "group:oidc:platform-admins",
    "user:ldap:alice",
    "group:saml:sso-admins",
  ]

  viewers = [
    "user:github:samrees"
  ]
}
```

### Create a pipeline

```hcl
resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("pipeline-config.yml")
  pipeline_config_format = "yaml"

  vars = {
    foo = "bar"
  }
}

# OR

resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("pipeline-config.json")
  pipeline_config_format = "json"

  vars = {
    foo = "bar"
  }
}
```

#### Optional pipeline arguments

* `vars` — map of string values interpolated into `((var))` placeholders
  (equivalent to `fly set-pipeline --var`).
* `yaml_vars` — map whose values are YAML documents, interpolated as structured
  values (equivalent to `fly set-pipeline --yaml-var`). Useful for list/map vars
  such as a var-sourced `across` step's `values`.
* `archive_on_destroy` — when `true`, the pipeline is archived instead of deleted
  on `terraform destroy`. Defaults to `false`.

```hcl
resource "concourse_pipeline" "my_pipeline" {
  team_name     = "main"
  pipeline_name = "my-pipeline"

  is_exposed         = true
  is_paused          = true
  archive_on_destroy = true

  pipeline_config        = file("pipeline-config.yml")
  pipeline_config_format = "yaml"

  vars = {
    foo = "bar"
  }

  yaml_vars = {
    versions = "[1, 2, 3]"
  }
}
```

## Import

Concourse teams can be imported using the team name e.g.

```
 $ terraform import concourse_pipeline.my_team my-team
```

Concourse pipelines can be imported using the team name and pipeline name e.g.

```
 $ terraform import concourse_pipeline.my_app my-team:my-app
```

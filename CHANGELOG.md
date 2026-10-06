## Changelog

### 8.1.0

#### New features

- **`concourse_pipeline`: `yaml_vars`** — New optional argument accepting a map
  of YAML-document strings. Values are interpolated into `(( var ))` placeholders
  as structured types (lists, maps), equivalent to `fly set-pipeline --yaml-var`.
  Use this for var-sourced `across` step values and any pipeline variable that
  cannot be expressed as a plain string.
- **`concourse_pipeline`: `archive_on_destroy`** — New optional boolean argument.
  When `true`, `terraform destroy` archives the pipeline instead of deleting it,
  preserving its build history. Defaults to `false`.
- **`concourse_team`: auth entry validation** — The team resource now validates
  `owners`, `members`, `pipeline_operators`, and `viewers` entries at apply time.
  Any entry not prefixed with `user:` or `group:` (e.g. `user:ldap:alice`,
  `group:saml:admins`) is now rejected with an explicit error message instead of
  being silently passed to Concourse or producing an obscure API error.

#### Bug fixes

- **`concourse_pipeline`: `across` step decoding** — Pipelines that use
  var-sourced `across` steps (a common pattern for dynamic matrix builds) could
  not be decoded by the older Concourse client, causing apply failures. The
  updated client resolves this.
- **`concourse_pipeline`: `atc.PipelineRef` migration** — Internal pipeline
  references now use the typed `atc.PipelineRef` struct rather than bare name
  strings, aligning the provider with the Concourse API's current expectations.

#### Dependencies

- Concourse client bumped from Aug 2020 to
  `v1.6.1-0.20261002202909-af7536c5322a` (Oct 2026 master). The new client's
  `vars.TemplateResolver.Resolve` drops the `expectAllVarsUsed` bool parameter;
  call sites updated accordingly.
- Go minimum raised to **1.27.1** to match `concourse/concourse`'s own `go.mod`
  declaration.
- Transitive dependency updates: `golang.org/x/*`, `google.golang.org/grpc`,
  `google.golang.org/protobuf`, `github.com/onsi/gomega`, `golang.org/x/oauth2`.

#### CI / supply-chain

- GitHub Actions SHA-pinned for supply-chain hygiene:
  `actions/setup-go@v7.0.0`, `actions/checkout@v7.0.1`,
  `docker/setup-compose-action@v2.4.0`, and the release workflow equivalents.
- CI and release workflows updated to build with **Go 1.27**.

#### Testing

- Integration test matrix verified across **Concourse 6.5.1 – 8.3.1** and
  **Terraform 1.0 – 1.2**.
- `archive_on_destroy` and `yaml_vars` added to `ImportStateVerifyIgnore` in all
  relevant integration specs, as these attributes have no server-side
  representation and therefore cannot round-trip through `terraform import`.

### 8.0.0

`concourse_pipeline` resource now supports supplying (concourse)
template variables through the `vars` argument. Technically this is a
breaking change if any of your pipelines happen to have any
double-parentheses (`(( ... ))`) references that aren't intended to
be interpreted by concourse.

## Changelog

### 8.0.0

`concourse_pipeline` resource now supports supplying (concourse)
template variables through the `vars` argument. Technically this is a
breaking change if any of your pipelines happen to have any
double-parentheses (`(( ... ))`) references that aren't intended to
be interpreted by concourse.

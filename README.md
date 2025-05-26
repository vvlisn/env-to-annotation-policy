[![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/kubewarden/community/blob/main/REPOSITORIES.md#stable)

# env-to-annotation-policy

This policy converts specific container environment variables into pod annotations.

## Introduction

This repository contains a Kubewarden policy written in Go. The policy mutates Kubernetes Pods (specifically Deployments, which manage Pods) by taking a specified environment variable's value from a container and adding it as an annotation to the Pod. This is particularly useful for integrating with logging or monitoring systems that consume annotations.

The policy is configurable via runtime settings.

You can configure the policy using a JSON structure. When using `kwctl run --settings-json`, the settings should be nested under a `signatures` key:

```json
{
  "signatures": [
    {
      "env_key": "MY_LOG_PATH_ENV",
      "annotation_base": "my.company.com/log-path",
      "annotation_ext_format": "my.company.com/log-path-ext-%d"
    }
  ]
}
```

When deploying the policy to a Kubewarden cluster, the settings are typically provided directly without the `signatures` nesting:

```json
{
  "env_key": "MY_LOG_PATH_ENV",
  "annotation_base": "my.company.com/log-path",
  "annotation_ext_format": "my.company.com/log-path-ext-%d"
}
```

The available settings are:
- `env_key` (string, mandatory): The name of the container environment variable whose value will be converted into an annotation.
- `annotation_base` (string, mandatory): The base annotation key name. The value of `env_key` will be assigned to this annotation. If `env_key` contains multiple paths separated by commas, the first path will be assigned to this base annotation.
- `annotation_ext_format` (string, mandatory): The format string for extended annotation keys. If `env_key` contains multiple paths, subsequent paths will be assigned to annotations generated using this format. The string must contain `%d`, which will be replaced by sequence numbers (1, 2, 3...). Example: `my.company.com/log-path-ext-%d`.

## Code organization

The code is organized as follows:
- `settings.go`: Handles policy settings and their validation
- `validate.go`: Contains the main mutation logic that converts environment variables to annotations
- `main.go`: Registers policy entry points with the Kubewarden runtime

## Implementation details

> **DISCLAIMER:** WebAssembly is a constantly evolving area.
> This document describes the status of the Go ecosystem as of 2024.

This policy utilizes several key concepts in its implementation:

1. Environment Variable to Annotation Conversion
   - Iterates through containers in a Pod.
   - Identifies the specified `env_key` environment variable.
   - Parses the environment variable's value (which can be a comma-separated list of paths).
   - Adds these paths as annotations to the Pod, using `annotation_base` for the first path and `annotation_ext_format` for subsequent paths.

2. Configuration Management
   - All settings (`env_key`, `annotation_base`, `annotation_ext_format`) are mandatory and validated at policy load time.

3. Technical Considerations
   - Built with TinyGo for WebAssembly compatibility.
   - Uses Kubewarden's TinyGo-compatible Kubernetes types.
   - Implements Kubewarden policy interface:
     - `validate`: Main entry point for Pod mutation.
     - `validate_settings`: Entry point for settings validation.

See the [Kubewarden Policy SDK](https://github.com/kubewarden/policy-sdk-go) documentation for more details on policy development.

## Testing

The policy includes comprehensive unit tests that verify:

1. Settings validation:
   - Valid settings.
   - Invalid settings (empty `env_key`, `annotation_base`, `annotation_ext_format`, or missing `%d` in `annotation_ext_format`).
   - JSON unmarshalling of settings.

2. Deployment mutation:
   - Correctly converts a single environment variable to a base annotation.
   - Correctly converts multiple environment variables (comma-separated) to base and extended annotations.
   - Handles deployments with no target environment variable.
   - Preserves existing annotations.

The unit tests can be run via:

```console
make test
```

The policy also includes end-to-end tests that verify the WebAssembly module behavior using the `kwctl` CLI. These tests validate:

1. Mutation behavior:
   - Correct annotation addition for single and multiple paths.
   - No mutation when the target environment variable is not found.

The e2e tests are implemented in `e2e.bats` and can be run via:

```console
make e2e-tests
```

## Automation

This project has the following [GitHub Actions](https://docs.github.com/en/actions):

- `e2e-tests`: this action builds the WebAssembly policy,
installs the `bats` utility and then runs the end-to-end test.
- `unit-tests`: this action runs the Go unit tests.
- `release`: this action builds the WebAssembly policy and pushes it to a user defined OCI registry
([ghcr](https://ghcr.io) is a good candidate).
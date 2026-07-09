# tpm — Training Platform Module CLI

> ⚠️ **Beta / unsupported software.** `tpm` is published as beta, **unsupported**
> software. It is provided "as is" with no warranty or support commitment, and its
> commands and behaviour may change without notice between releases. Use at your own
> risk. See [LICENSE](LICENSE) (Apache 2.0).

`tpm` is a command-line tool for authoring and managing Training Platform modules and
labs: scaffolding new modules/labs, validating their structure against JSON schemas,
generating codes and markdown, building modules, and running labs locally.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install coreeng/public/tpm
```

> The Homebrew formula is published by CI to CECG's public tap
> ([`coreeng/homebrew-public`](https://github.com/coreeng/homebrew-public)) on each
> release. The `coreeng/public/tpm` shorthand taps that repo automatically; to browse
> the tap's other tools, run `brew tap coreeng/public`.

### Pre-built binaries

Download the archive for your OS/architecture from the
[GitHub Releases](https://github.com/coreeng/tpm/releases) page, extract it, and put the
`tpm` binary on your `PATH`.

### From source

Requires Go (see the version in [`go.mod`](go.mod)).

```bash
go install github.com/coreeng/tpm@latest
# or, from a clone:
make install
```

## Usage

```bash
tpm --help              # Show all commands
tpm --version           # Show version
```

Common commands:

```bash
tpm module list <dir>                         # List modules directly under a directory
tpm module validate <module-path>...          # Validate module source directories
tpm module compare <old> <new>                # Check for breaking code changes
tpm module generate codes <module-path>...    # Generate missing codes
tpm module generate markdown <module-path>... # Generate missing markdown files
tpm module build <module-path>... --out-root <dir>
tpm module preview <module-path> --watch
tpm artifact validate <module.yaml-or-dir>...
tpm lab init <path>                           # Scaffold a standalone lab runtime
tpm lab preview <path> --chart-uri <oci-uri>  # Run and preview a local lab
```

See [docs/cli-reference.md](docs/cli-reference.md) for the full command reference and
examples.

The JSON schemas used by `tpm module validate`, `tpm module build`, and
`tpm artifact validate`
are **embedded in the binary**, so validation works anywhere with no external files
required. To validate against a different schema set while developing `tpm`, pass
`--schema-dir <path>`.

## Examples

The [`examples/`](examples/) directory contains complete source examples:

- [`examples/modules/kubernetes-101/`](examples/modules/kubernetes-101/) is a three-chapter
  module covering Kubernetes cluster basics, workloads, and application operations.
- [`examples/labs/spring-boot-health-checks/`](examples/labs/spring-boot-health-checks/) is a
  standalone lab where learners add Spring Boot health checks and Kubernetes probes.

Preview the module locally:

```bash
tpm module preview examples/modules/kubernetes-101 --watch
```

## Authoring labs with an AI assistant

The preferred way to author labs with an AI assistant is to install the
[`authoring-labs`](https://github.com/coreeng/tpm-authoring-labs-skill) skill. It guides an
LLM-based assistant from teaching intent through to a scaffolded lab, reviewed solution,
and starter content — driving the `tpm` CLI along the way.

Clone the skill repository into your skills directory so `SKILL.md` is at the root of the
installed skill:

```bash
git clone git@github.com:coreeng/tpm-authoring-labs-skill.git ~/.config/opencode/skills/authoring-labs
```

## Example lab

[`examples/labs/spring-boot-health-checks/`](examples/labs/spring-boot-health-checks/) is a complete,
working example lab — "add Spring Boot health checks to an application" — authored with the
[`authoring-labs`](https://github.com/coreeng/tpm-authoring-labs-skill) skill. It shows the
full lab layout and a `validator/` that checks the learner's running workload.

- [Learner task](examples/labs/spring-boot-health-checks/starter-content/README.md) — what the learner starts with and has to do.
- [Reference solution](examples/labs/spring-boot-health-checks/solution/README.md) — the completed implementation.

Run it locally against a kind cluster (the lab runtime is published as an OCI Helm chart):

```bash
tpm lab preview examples/labs/spring-boot-health-checks \
  --chart-uri oci://ghcr.io/coreeng/charts/training-platform-assessment
```

> [!NOTE]
> By default, Helm uses the latest published version of the chart specified by `--chart-uri`.
> Add `--chart-version <version>` to pin a specific chart version.

## Development

```bash
make check    # Run the full local PR quality gate
make build    # Build the tpm binary
make test     # Run tests
make lint     # Run golangci-lint (falls back to go vet)
make install  # Install to GOPATH/bin
```

## Security

CI runs Trivy (vulnerability, secret, config, and license scanning) and CodeQL (Go SAST)
on every pull request and push to `main`.

## Releases

Release versions are based on Conventional Commit PR titles. See
[docs/release-policy.md](docs/release-policy.md) for the title format and version bump rules.

## Licence

Licensed under the Apache License, Version 2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).

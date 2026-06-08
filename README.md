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
tpm list                            # List all modules
tpm validate                        # Validate all modules (uses embedded schemas)
tpm validate-changes <old> <new>    # Check for removed codes between two git refs
tpm generate-codes                  # Generate missing UUID codes
tpm generate-markdown               # Generate missing markdown files
tpm build <module> -o <dir>         # Build a module into a unified module.yaml
tpm init lab <path>                 # Scaffold a new lab
tpm lab ...                         # Run and inspect local labs
```

The JSON schemas used by `tpm validate` are **embedded in the binary**, so validation
works anywhere with no external files required. To validate against a different set of
schemas, pass `--schema-dir <path>`.

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

[`examples/spring-boot-health-checks/`](examples/spring-boot-health-checks/) is a complete,
working example lab — "add Spring Boot health checks to an application" — authored with the
[`authoring-labs`](https://github.com/coreeng/tpm-authoring-labs-skill) skill. It shows the
full lab layout and a `validator/` that checks the learner's running workload.

- [Learner task](examples/spring-boot-health-checks/starter-content/README.md) — what the learner starts with and has to do.
- [Reference solution](examples/spring-boot-health-checks/solution/README.md) — the completed implementation.

Run it locally against a kind cluster (the lab runtime is published as an OCI Helm chart):

```bash
tpm lab run examples/spring-boot-health-checks \
  --chart-uri oci://ghcr.io/coreeng/charts/training-platform-assessment \
  --chart-version 0.0.249
```

> [!NOTE]
> `0.0.249` is the lab-runtime chart version used here as an example — the latest
> `training-platform-assessment` chart version may be newer. Use the most recent published tag.

## Development

```bash
make build    # Build the tpm binary
make test     # Run tests
make lint     # Run golangci-lint (falls back to go vet)
make install  # Install to GOPATH/bin
```

## Security

CI runs Trivy (vulnerability, secret, config, and license scanning) and CodeQL (Go SAST)
on every pull request and push to `main`.

## Licence

Licensed under the Apache License, Version 2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).

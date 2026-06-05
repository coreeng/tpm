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
brew install coreeng/tap/tpm
```

> The Homebrew formula is published by CI to the `coreeng/homebrew-tap` repository on
> each release. If you have not tapped it before, Homebrew resolves `coreeng/tap`
> automatically from the `coreeng/homebrew-tap` repo.

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

### Optional: richer flows with `superpowers`

The skill's discovery, planning, and execution phases are inspired by the
[`superpowers`](https://github.com/obra/superpowers) plugin (MIT, by Jesse Vincent). If you
install `superpowers`, you can use its `brainstorming`, `writing-plans`, and
`executing-plans` skills in place of the inlined guidance for a more structured experience:

```bash
# In Claude Code:
/plugin marketplace add obra/superpowers-marketplace
/plugin install superpowers@superpowers-marketplace
```

`superpowers` is **not** required and is **not** bundled with `tpm`.

## Example module

A minimal, fully-working example module lives in [`examples/`](examples/). From the repo
root you can build it — `tpm build` validates the module against the embedded schemas
before writing output:

```bash
tpm build examples/hello-module -o /tmp/hello-build
```

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

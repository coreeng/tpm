# TPM CLI Reference

Use this page when you need the exact `tpm` command syntax for module source directories, built module artifacts, and local lab previews.

> [!NOTE]
> `tpm` is beta software. Commands and flags may change between releases.

## Conventions

- All module authoring indexes are 1-based.
- Commands that accept `<module-path>` take the path to a module source directory.
- Commands that accept `<module-path>...` can process more than one module in a single invocation.
- `--set field=value` edits fields stored in YAML. Edit markdown-backed content in the markdown file itself.
- `--breaking-policy` accepts `error`, `warn`, or `ignore`. The default is `error`.
- `--allow-breaking` is a convenience alias for `--breaking-policy=warn`.

> [!TIP]
> Use `tpm module preview <module-path> --watch` while editing a module. The local web UI reloads when source YAML or markdown changes.

## Root Commands

```bash
tpm --help
tpm --version
```

| Command | Purpose |
| --- | --- |
| `tpm module` | Create, edit, validate, build, compare, and preview module source directories. |
| `tpm artifact` | Validate and inspect compiled `module.yaml` artifacts. |
| `tpm lab` | Create, preview, inspect, and clean up local labs. |
| `tpm completion` | Generate shell completion scripts. |

## Module Commands

### Create And Discover Modules

```bash
tpm module init modules/kubernetes-101
tpm module list modules
```

| Command | Purpose |
| --- | --- |
| `tpm module init <module-path>` | Create a new module skeleton. |
| `tpm module list <dir>` | List modules directly under a directory. It does not search nested directories. |

### Validate, Build, And Preview

```bash
tpm module validate modules/kubernetes-101
tpm module validate modules/kubernetes-101 modules/platform-debugging

tpm module build modules/kubernetes-101 --out-root artifacts
tpm module build modules/kubernetes-101 modules/platform-debugging --out-root artifacts

tpm module preview modules/kubernetes-101 --watch
tpm module preview modules/kubernetes-101 --addr 127.0.0.1:61231 --no-open-browser
```

| Command | Purpose |
| --- | --- |
| `tpm module validate <module-path>...` | Validate one or more module source directories. |
| `tpm module build <module-path>... --out-root <dir>` | Build each module to `<out-root>/<module-name>/module.yaml`. |
| `tpm module preview <module-path>` | Open a local browser preview of a full module. |

`module validate` flags:

| Flag | Purpose |
| --- | --- |
| `--schema-dir <dir>` | Use a source schema directory instead of the schemas embedded in `tpm`. |

`module build` flags:

| Flag | Purpose |
| --- | --- |
| `--out-root <dir>` | Required output directory root. |
| `--assessment-registry-override <registry>` | Override the registry path for all labs. |
| `--assessment-version-override <version>` | Override `imageVersion` for all labs. |

`module preview` flags:

| Flag | Purpose |
| --- | --- |
| `--addr <host:port>` | Address for the local preview server. Defaults to `127.0.0.1:0`. |
| `--watch` | Reload module metadata and markdown when source files change. |
| `--no-open-browser` | Serve the preview without opening the default browser. |

### Generate Source Helpers

```bash
tpm module generate codes modules/kubernetes-101
tpm module generate markdown modules/kubernetes-101
```

| Command | Purpose |
| --- | --- |
| `tpm module generate codes <module-path>...` | Generate missing codes for module items. |
| `tpm module generate markdown <module-path>...` | Generate missing markdown files. |

### Compare Compatibility

```bash
tpm module compare old/module.yaml new/module.yaml
tpm module compare modules/kubernetes-101@main modules/kubernetes-101@HEAD
tpm module compare artifacts/old/kubernetes-101 artifacts/new/kubernetes-101 --breaking-policy=warn
```

`module compare` accepts local paths and `path@ref` git locations. Local paths can point to module source directories, built artifact directories, or built `module.yaml` files.

| Flag | Purpose |
| --- | --- |
| `--breaking-policy <policy>` | Choose `error`, `warn`, or `ignore` for breaking code changes. Defaults to `error`. |
| `--allow-breaking` | Alias for `--breaking-policy=warn`. |

## YAML Authoring Commands

Use these commands for structured module edits that are easier and safer by index than by hand:

```bash
tpm module add <type> <module-path> --at <index> [selectors] --set field=value
tpm module edit <type> <module-path> [selectors] --set field=value
tpm module move <type> <module-path> [selectors] --from <index> --to <index>
tpm module remove <type> <module-path> [selectors] --from <index> --yes
```

> [!IMPORTANT]
> These commands edit YAML fields only. Edit markdown descriptions, challenge text, success messages, and learner instructions in the markdown files.

### Resource Types And Selectors

| Type | Add | Edit | Move | Remove | Selector Pattern |
| --- | --- | --- | --- | --- | --- |
| `module` | No | Yes | No | No | No index. Edits module-level YAML. |
| `chapter` | Yes | Yes | Yes | Yes | `--chapter <n>` for edit. `--from/--to` or `--from` for move/remove. |
| `section` | Yes | Yes | Yes | Yes | `--chapter <n>` plus `--section <n>` for edit. |
| `lab` | Yes | Yes | Yes | Yes | `--chapter <n>` plus `--lab <n>` for edit. |
| `challenge` | Yes | Yes | Yes | Yes | `--chapter <n> --lab <n>` plus `--challenge <n>` for edit. |
| `goal` | Yes | Yes | Yes | Yes | `--chapter <n> --lab <n> --challenge <n>` plus `--goal <n>` for edit. |
| `quiz` | Yes | Yes | Yes | Yes | `--chapter <n>` plus `--quiz <n>` for edit. |
| `question` | Yes | Yes | Yes | Yes | `--chapter <n> --quiz <n>` plus `--question <n>` for edit. |
| `option` | Yes | Yes | Yes | Yes | `--chapter <n> --quiz <n> --question <n>` plus `--option <n>` for edit. |

### Editable YAML Fields

| Type | Fields |
| --- | --- |
| `module` | `code`, `title`, `shortDescription`, `bannerImage`, `bannerVideo`, `tags`, `level`, `video` |
| `chapter` | `code`, `title`, `shortDescription`, `bannerImage`, `bannerVideo`, `isDraft`, `video` |
| `section` | `code`, `title`, `shortDescription`, `estimatedDuration`, `video`, `thumbnail`, `thumbnailDescription` |
| `lab` | `code`, `title`, `timeLimit`, `starterImageUri`, `validatorImageUri`, `imageVersion`, `video` |
| `challenge` | `code`, `title`, `estimatedDuration`, `video` |
| `goal` | `code`, `title`, `description` |
| `quiz` | `code`, `title`, `description`, `passingScore`, `video` |
| `question` | `code`, `question`, `type` |
| `option` | `text`, `correct` |

Values are parsed as booleans when they are `true` or `false`, integers when numeric, and strings otherwise. `tags` accepts a comma-separated list.

### Add Examples

```bash
tpm module add chapter modules/kubernetes-101 \
  --at 1 \
  --set code=cluster-fundamentals \
  --set title="Cluster fundamentals"

tpm module add section modules/kubernetes-101 \
  --chapter 1 \
  --at 2 \
  --set code=control-plane-and-nodes \
  --set title="Control plane and nodes" \
  --set estimatedDuration=18m

tpm module add quiz modules/kubernetes-101 \
  --chapter 3 \
  --at 1 \
  --set code=kubernetes-operations-check \
  --set title="Kubernetes operations check" \
  --set passingScore=70
```

### Quiz Examples

```bash
tpm module add question modules/kubernetes-101 \
  --chapter 3 \
  --quiz 1 \
  --at 1 \
  --set code=deployment-purpose \
  --set type=SINGLE \
  --set question="What does a Deployment primarily manage?"

tpm module add option modules/kubernetes-101 \
  --chapter 3 \
  --quiz 1 \
  --question 1 \
  --at 1 \
  --set text="Rollouts and replicas for Pods" \
  --set correct=true

tpm module add question modules/kubernetes-101 \
  --chapter 3 \
  --quiz 1 \
  --at 2 \
  --set code=troubleshooting-signals \
  --set type=MULTIPLE \
  --set question="Which signals are useful when troubleshooting a Kubernetes workload?"
```

### Lab Metadata Examples

Set `starterImageUri`, `validatorImageUri`, and `imageVersion` to the published lab resources you want the module to use.

```bash
tpm module add lab modules/kubernetes-101 \
  --chapter 3 \
  --at 1 \
  --set code=inspect-a-workload \
  --set title="Inspect a workload" \
  --set timeLimit=45m \
  --set starterImageUri=oci://ghcr.io/example/kubernetes-101-starter \
  --set validatorImageUri=oci://ghcr.io/example/kubernetes-101-validator \
  --set imageVersion=0.1.0

tpm module add challenge modules/kubernetes-101 \
  --chapter 3 \
  --lab 1 \
  --at 1 \
  --set code=read-workload-state \
  --set title="Read workload state"

tpm module add goal modules/kubernetes-101 \
  --chapter 3 \
  --lab 1 \
  --challenge 1 \
  --at 1 \
  --set code=pod-ready \
  --set title="Find the Ready Pod" \
  --set description="Identify the Pod that is serving traffic."
```

### Edit, Move, And Remove Examples

```bash
tpm module edit module modules/kubernetes-101 \
  --set title="Kubernetes 101" \
  --set level=BEGINNER

tpm module edit question modules/kubernetes-101 \
  --chapter 3 \
  --quiz 1 \
  --question 2 \
  --set type=MULTIPLE

tpm module move section modules/kubernetes-101 \
  --chapter 2 \
  --from 3 \
  --to 1

tpm module remove option modules/kubernetes-101 \
  --chapter 3 \
  --quiz 1 \
  --question 1 \
  --from 3 \
  --yes
```

Changing a `code` field or removing a resource is treated as breaking by default:

```bash
tpm module edit section modules/kubernetes-101 \
  --chapter 1 \
  --section 1 \
  --set code=what-kubernetes-manages \
  --breaking-policy=warn

tpm module remove quiz modules/kubernetes-101 \
  --chapter 3 \
  --from 1 \
  --yes \
  --allow-breaking
```

## Artifact Commands

```bash
tpm artifact validate artifacts/kubernetes-101/module.yaml
tpm artifact validate artifacts/kubernetes-101
tpm artifact inspect artifacts/kubernetes-101
```

| Command | Purpose |
| --- | --- |
| `tpm artifact validate <module.yaml-or-dir>...` | Validate compiled module artifacts. |
| `tpm artifact inspect <module.yaml-or-dir>...` | Print a summary of compiled module artifacts. |

`artifact validate` flags:

| Flag | Purpose |
| --- | --- |
| `--schema-dir <dir>` | Use a built-artifact schema directory instead of the schemas embedded in `tpm`. |

## Lab Commands

### Create And Outline

```bash
tpm lab init labs/pod-image-lab
tpm lab outline labs/pod-image-lab
tpm lab outline labs/pod-image-lab --codes --paths
tpm lab outline labs/pod-image-lab --json
```

| Command | Purpose |
| --- | --- |
| `tpm lab init <lab-path>` | Create a standalone lab runtime skeleton. |
| `tpm lab outline <lab-path>` | Print lab challenges and goals. |

`lab outline` flags:

| Flag | Purpose |
| --- | --- |
| `--codes` | Show lab, challenge, and goal codes. |
| `--paths` | Show metadata and runtime paths. |
| `--json` | Write JSON. |

### Preview, Status, And Cleanup

```bash
tpm lab preview labs/pod-image-lab \
  --chart-uri oci://ghcr.io/coreeng/charts/training-platform-assessment \
  --watch

tpm lab list
tpm lab status --id <run-id>
tpm lab cleanup --id <run-id>
```

`lab preview` starts the local lab runtime and opens a browser preview. The preview shows goal, challenge, and lab progress reported by the validator in the running cluster.

`lab preview` flags:

| Flag | Purpose |
| --- | --- |
| `--chart-dir <dir>` | Use a local lab runtime Helm chart directory. Mutually exclusive with `--chart-uri`. |
| `--chart-uri <uri>` | Use a lab runtime Helm chart from an OCI registry. Mutually exclusive with `--chart-dir`. |
| `--chart-version <version>` | Pin the lab runtime chart version. |
| `--validator-registry <registry>` | Registry for the locally built validator image. Defaults to `localhost`. |
| `--registry-domain <domain>` | Learner registry domain passed to the lab runtime chart. Defaults to `localhost`. |
| `--check-interval <duration>` | Validator check interval. Defaults to `5s`. |
| `--id <id>` | Use a specific lab run ID instead of a generated one. |
| `--state-dir <dir>` | Use a custom lab state directory. Defaults to `~/.config/tpm`. |
| `--addr <host:port>` | Address for the local preview server. Defaults to `127.0.0.1:0`. |
| `--watch` | Reload lab metadata and markdown when source files change. |
| `--no-open-browser` | Serve the preview without opening the default browser. |
| `--allow-non-kind` | Allow running against a kubectl context that is not a kind cluster. |
| `--assume-image-accessible` | Assume a non-kind cluster can pull the local validator image tag. |

`lab list` flags:

| Flag | Purpose |
| --- | --- |
| `--state-dir <dir>` | Use a custom lab state directory. |
| `--allow-non-kind` | Allow listing labs against a kubectl context that is not a kind cluster. |

`lab status` flags:

| Flag | Purpose |
| --- | --- |
| `--id <id>` | Show status for a specific lab run ID. |
| `--state-dir <dir>` | Use a custom lab state directory. |

`lab cleanup` flags:

| Flag | Purpose |
| --- | --- |
| `--id <id>` | Clean up a specific lab run ID. |
| `--state-dir <dir>` | Use a custom lab state directory. |
| `--allow-non-kind` | Allow cleanup against a kubectl context that is not a kind cluster. |

## Completion Commands

```bash
tpm completion bash
tpm completion fish
tpm completion powershell
tpm completion zsh
```

Use the subcommand help for shell-specific installation instructions.

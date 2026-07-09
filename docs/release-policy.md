# Release Policy

`tpm` releases are driven by Git tags. The release workflow analyzes commits since the
previous tag, then GoReleaser publishes the binaries and Homebrew formula.

## PR Titles

Use this repository's supported Conventional Commit format for every PR title:

```text
type(optional-scope): short description
```

Examples:

```text
fix: handle missing lab state
feat(preview): add module preview
ci(release): require conventional release commits
```

This repo uses squash-only merges. The PR title becomes the squash commit title on
`main`, and the PR body becomes the squash commit body. In practice, each merged PR
normally defines the next release bump.

## Version Bumps

- Patch release: use `fix:`, `perf:`, or any other supported title type. The release
  workflow falls back to patch when the PR does not request major or minor.
- Minor release: use `feat:`, such as `feat(preview): add module preview`.
- Major release: add a `BREAKING CHANGE:` section to the PR body.

> [!IMPORTANT]
> Every supported PR title publishes a release after merge. If the PR does not request a
> major or minor release, the release workflow creates a patch release.

The PR title workflow comments on each PR with the release bump expected after merge.
It also comments on invalid titles with the supported format and bump mapping.

## Manual Releases

To publish an explicit version, push the tag manually:

```bash
git tag v1.2.3
git push origin v1.2.3
```

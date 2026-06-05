# TPM Authoring Labs Skill Design

## Goal

Create a private, installable skill repository for the TPM authoring-labs skill. The repository root must be directly cloneable into a user's skills directory.

## Repository

Create a sibling repository at:

```text
/Users/chbatey/dev/cecg/tpm-authoring-labs-skill
```

Publish it to GitHub as a private repository owned by `coreeng`:

```text
coreeng/tpm-authoring-labs-skill
```

## Contents

The repository should contain only the minimal skill package:

```text
tpm-authoring-labs-skill/
  SKILL.md
  README.md
```

`SKILL.md` should be copied from:

```text
/Users/chbatey/dev/cecg/training-platform-modules/.opencode/skills/authoring-labs/SKILL.md
```

`README.md` should explain what the skill does and how to install it by cloning the private repository into a skills directory. It must not mention prior iterations or version history.

## Install Shape

The repository root must contain `SKILL.md` directly so users can install with:

```bash
git clone git@github.com:coreeng/tpm-authoring-labs-skill.git ~/.config/opencode/skills/authoring-labs
```

## Validation

Verify:

- `SKILL.md` exists at the repository root.
- The local repository has a clean git status after commit and push.
- The GitHub repository visibility is `PRIVATE`.

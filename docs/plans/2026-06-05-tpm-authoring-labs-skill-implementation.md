# TPM Authoring Labs Skill Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build and publish a private, root-installable `coreeng/tpm-authoring-labs-skill` repository.

**Architecture:** The new repository is a minimal skill package. Its root contains `SKILL.md` copied from the existing source skill and a concise README with clone-based install instructions.

**Tech Stack:** Git, GitHub CLI, Markdown, OpenCode skill layout.

---

### Task 1: Create The Skill Repository Files

**Files:**
- Create: `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill/SKILL.md`
- Create: `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill/README.md`

**Step 1: Verify target path is available**

Run: `test ! -e /Users/chbatey/dev/cecg/tpm-authoring-labs-skill`

Expected: command exits successfully. If it fails, inspect the existing path before proceeding.

**Step 2: Create the repository directory**

Run: `mkdir /Users/chbatey/dev/cecg/tpm-authoring-labs-skill`

Expected: directory is created.

**Step 3: Copy the source skill**

Copy:

```text
/Users/chbatey/dev/cecg/training-platform-modules/.opencode/skills/authoring-labs/SKILL.md
```

to:

```text
/Users/chbatey/dev/cecg/tpm-authoring-labs-skill/SKILL.md
```

Expected: root `SKILL.md` matches the source skill content.

**Step 4: Write the README**

Create `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill/README.md` with:

```markdown
# TPM Authoring Labs Skill

An agent skill for creating, designing, scaffolding, and authoring TPM labs.

## Install

Clone this repository into your skills directory:

```bash
git clone git@github.com:coreeng/tpm-authoring-labs-skill.git ~/.config/opencode/skills/authoring-labs
```

For another agent runtime, clone it into that runtime's `authoring-labs` skill directory so `SKILL.md` is at the root of the installed skill.

## Use

Use this skill when creating, designing, scaffolding, or authoring TPM labs.
```

Expected: README contains install/use instructions only and no version-history discussion.

**Step 5: Verify file layout**

Run: `test -f /Users/chbatey/dev/cecg/tpm-authoring-labs-skill/SKILL.md && test -f /Users/chbatey/dev/cecg/tpm-authoring-labs-skill/README.md`

Expected: command exits successfully.

### Task 2: Commit And Publish Privately

**Files:**
- Modify: `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill/.git/config`

**Step 1: Initialize git**

Run from `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill`: `git init`

Expected: empty git repository is initialized.

**Step 2: Commit files**

Run from `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill`: `git add SKILL.md README.md && git commit -m "Initial authoring labs skill"`

Expected: initial commit is created.

**Step 3: Create private GitHub repository**

Run from `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill`: `gh repo create coreeng/tpm-authoring-labs-skill --private --source . --remote origin --push`

Expected: private GitHub repository is created and the initial commit is pushed.

**Step 4: Verify visibility**

Run: `gh repo view coreeng/tpm-authoring-labs-skill --json visibility`

Expected: JSON reports `"visibility":"PRIVATE"`.

**Step 5: Verify clean status**

Run from `/Users/chbatey/dev/cecg/tpm-authoring-labs-skill`: `git status --short --branch`

Expected: branch tracks `origin` and there are no uncommitted changes.

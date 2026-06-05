---
name: authoring-labs
description: Use when creating, designing, scaffolding, or authoring TPM labs with the tpm CLI.
---

# Authoring Labs

## Overview

Guide an author from teaching intent to scaffolded lab, reviewed solution, and starter content.

Core rule: do not implement or scaffold until the learning objective, real-world scenario, start state, and end state are understood.

> **Note on sub-skills:** Earlier versions of this skill depended on the third-party
> `superpowers` plugin (`brainstorming`, `writing-plans`, `executing-plans`). This
> standalone version inlines the essential guidance so it works with no extra plugins.
> If you have `superpowers` installed you may use those richer flows in place of the
> inlined instructions below — see the repository README for install notes.

## When to Use

Use when the user wants to create, design, scaffold, or author a TPM lab.

Do not use for non-TPM lab formats or unrelated agent configuration work.

## Hard Gates

- Do not scaffold before teaching intent, real-world scenario, start state, and end state are understood.
- Do not run `tpm init lab` before confirming module-backed versus standalone details.
- Do not propose challenges or goals until each one has a validator-observable signal in the learner namespace, the learner registry, or both.
- Do not design a lab without a required deployment to the learner's Kubernetes namespace. Registry artifacts may be part of the lab, but they must be used by namespace resources.
- Do not dispatch solution or starter subagents until the build mechanism and one-command build are specified.
- Do not create a solution before a markdown solution plan exists and the user approves it.
- Do not create a solution plan before the scaffolded validator has been built at least once.
- Do not create starter content before the user confirms the solution is good.
- Do not dispatch the starter subagent before the starter plan is presented and approved.
- Use "lab" in user-facing wording unless schemas, commands, or runtime internals require otherwise.

## How to work through this skill

This skill is conversational and gated. Move through the phases below in order.

- **Phases 1–5 (discovery):** Ask **one question at a time**. Do not batch questions, do
  not jump ahead to topic-specific scoping, and do not scaffold. Reflect each answer back
  briefly before asking the next question. The goal is a shared, explicit understanding of
  intent, scenario, start/end state, build mechanism, and platform feasibility.
- **Phases 8 and 10 (planning):** Write a concrete, reviewable markdown plan (`PLAN.md`) at
  the lab-local path specified below, then stop and get explicit user approval before any
  implementation. A good plan lists exact files, resources, build commands, validation
  surfaces, implementation steps, and manual verification.
- **Phases 9 and 11 (implementation):** Dispatch a subagent to execute the approved
  `PLAN.md` task-by-task. Instruct it to implement only what the plan approves, preserve
  unrelated files, run the verification commands, and report back a summary, changed files,
  verification output, and blockers.

### 1. Teaching Intent

Ask one question at a time. For a new lab, the first question must be exactly:

"What should someone be able to do after completing this lab?"

Then ask:

"What real day-job scenarios, incidents, review comments, support requests, or onboarding gaps would have been helped by people having this skill?"

Do not replace these with topic-specific scope questions like "which RBAC type?" or combine them with module-placement questions. Topic scoping comes after the teaching outcome and real-world scenario are clear.

### 2. Start and End State

Ask:

"What should the learner start with?"

Then ask:

"What should the finished solution look like?"

Capture files, manifests, services, Kubernetes resources, expected behavior, and observable validation outcomes.

### 3. Build Mechanism

Ask:

"What single command should build the lab content? I recommend `make build`; do you want to use that or override it?"

Capture the build command, build files or entry points, expected artifacts or rendered output, and whether the same command should work in both solution and starter content.

Prefer a `Makefile` with a `build` target when the user does not have a specific alternative.

### 4. Platform Feasibility

State the training platform contract before suggesting lab structure:

- The learner starts from starter content pushed to a GitHub repository.
- The learner gets a Kubernetes namespace to deploy into.
- The learner gets a registry to push images or artifacts to.
- The validator can inspect the learner namespace and registry-visible state.

Every lab must require the learner to deploy something to their Kubernetes namespace. If the teaching intent is primarily about building, packaging, or pushing an artifact, convert it into an observable Kubernetes outcome by requiring a deployed workload that uses that artifact.

When a lab teaches application behavior, validate the behavior through the deployed workload, not through source files or local-only execution. Require namespace resources that expose the behavior, such as a Service plus Deployment, and have the validator observe the running workload. For HTTP behavior, prefer validator requests to the in-cluster Service and require expected status codes or responses, for example `200 OK` from health endpoints.

For each intended outcome, identify the observable validation surface:

- `namespace`: Kubernetes resources, labels, annotations, specs, status, events, image references, or runtime behavior visible in the namespace.
- `registry`: pushed images, tags, artifacts, metadata, or artifacts referenced by deployed workloads.
- `both`: registry artifact exists and namespace resources use it.

Reject or redesign goals that require checking private thoughts, local-only files never applied or pushed, source-code text as a proxy for runtime behavior, external systems the validator cannot access, subjective understanding, or registry-only outputs that are never deployed. Convert those into observable Kubernetes outcomes before proceeding.

### 5. Challenges and Goals

Suggest challenge and goal structure only after the previous phases.

For each challenge and goal, include code, title, validation outcome, observable validation surface (`namespace`, `registry`, or `both`), what the validator can check, and why it maps to the teaching objective.

Ask for confirmation before proceeding.

### 6. Scaffold Intake

Ask whether this should be part of an existing module.

If yes, ask for module path, chapter slug/name, and lab slug/name. Then run:

```bash
tpm init lab <module-path> <chapter> <lab-name> --module-backed
```

If no, ask for target path and lab slug/name. Then run:

```bash
tpm init lab <path>
```

Confirm details before running either command.

### 7. Validator Build

After the user confirms the challenge and goal structure and the lab scaffold exists, build the scaffolded validator before planning or implementing the solution.

Run:

```bash
docker build -t <lab-slug>-validator:local <scaffold-path>/validator
```

If the validator does not build, fix only scaffold or generated-validator issues needed to make the baseline validator image build. Do not implement custom validation logic or solution behavior in this phase.

Record the validator build command and result for the solution plan.

### 8. Solution Plan

Write the solution plan as `solution/PLAN.md` in the lab's solution directory.

Ask focused follow-up questions until the complete solution is precise. The plan must include:

- Learning objective and real-world scenario
- Scaffold location
- Challenge/goal structure
- Platform feasibility and observable validation surface for each goal
- Final files/resources
- Validation assumptions
- Validator build command and latest result
- One-command build mechanism, defaulting to `make build` unless overridden
- Build files and expected build output
- Implementation steps
- Manual verification

Show the plan and require user approval before implementing.

### 9. Solution Subagent

Dispatch a subagent to create the solution from the approved plan.

Prompt template:

```text
Create the lab solution from <plan-path>.

Implement the plan task-by-task. Work in <scaffold-path>. Implement only the approved solution, including the approved one-command build mechanism. Preserve unrelated files. Run the verification commands listed in the plan when feasible. Return a concise summary, changed files, verification output, and blockers.
```

Ask the user to review the solution. Iterate until they confirm it is good.

### 10. Starter Plan

Write the starter plan as `starter-content/PLAN.md` in the starter content directory before asking for confirmation.

After solution approval, present a plan to derive starter content from the solution. Include what to keep, remove, redact, replace with TODOs, what learner instructions to add, and how the same one-command build remains available in starter content. Require user confirmation.

### 11. Starter Subagent

Dispatch a subagent to create starter content from the approved starter plan.

Prompt template:

```text
Create starter content from the approved starter plan.

Implement the plan task-by-task. Preserve the reviewed solution. Modify only starter content and learner-facing instructions unless the plan explicitly says otherwise. Ensure the approved one-command build mechanism is present in starter content. Run feasible verification. Return a concise summary, changed files, verification output, and blockers.
```

## Common Mistakes

- Scaffolding too early because the user named a topic.
- Inventing goals without real workplace scenarios.
- Creating goals the validator cannot prove from namespace or registry state.
- Writing solution code before proving the scaffolded validator builds.
- Creating starter content before the solution is reviewed.
- Treating every solution directory as `kubectl apply` compatible.
- Using platform-internal terminology when "lab" is clearer.

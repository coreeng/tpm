# Spring Boot Health Checks — Lab

## The scenario

During an incident, a Spring Boot app that is hung or not-yet-ready keeps receiving traffic
because Kubernetes has no health signal — failures cascade and on-call engineers waste time
debugging symptoms instead of spotting the unhealthy pod. The common root cause: someone simply
forgot to add health checks.

This app has exactly that problem. It builds, deploys, and runs — but it has **no health
checks**, so Kubernetes can't tell whether it's ready for traffic or needs restarting.

## Your task

Add Spring Boot health checks so Kubernetes can stop routing to a not-ready pod and restart a
hung one.

1. **Expose the health endpoints.** Add the `spring-boot-starter-actuator` dependency
   (`build.gradle`) and enable the Kubernetes probe groups (`application.yaml`) so the app
   serves `GET /actuator/health/readiness` and `GET /actuator/health/liveness`.
2. **Wire the probes.** In `k8s/deployment.yaml`, add a `readinessProbe` and a `livenessProbe`,
   each an HTTP GET to the matching endpoint on port `8080`.
3. **Rebuild and redeploy** (see below).

> Keep the Deployment and Service named `health-app` on port `8080`, and use the probe paths
> `/actuator/health/readiness` and `/actuator/health/liveness` — that's what the lab checks.

## Running the lab

Start the lab locally against a kind cluster (the lab runtime is a published OCI Helm chart).
For example, from the `tpm` repo root:

```sh
tpm lab run examples/spring-boot-health-checks \
  --chart-uri oci://ghcr.io/coreeng/charts/training-platform-assessment \
  --chart-version 0.0.249
```

This prints a run id and your workspace namespace. (`0.0.249` is an example version — use the
latest published `training-platform-assessment` chart tag.)

## Build and deploy

> [!TIP]
> `<workspace-namespace>` is the workspace namespace for your lab run. It is printed by
> `tpm lab run` (the `Workspace namespace:` line) and is also shown by
> `tpm lab status --id <run-id>`. It looks like `lab-<run-id>-workspace`.

For kind:

```sh
make build
make kind-load
make deploy WORKSPACE_NAMESPACE=<workspace-namespace>
```

For a remote cluster:

```sh
make push REGISTRY=<registry>
make deploy WORKSPACE_NAMESPACE=<workspace-namespace> IMAGE=<registry>/health-app:local
```

## How you'll know you're done

The lab tracks three goals:

- **Deploy the app and reach Ready** — passes as soon as the app is deployed and running.
- **Wire up the readiness probe** — passes once the readiness probe is configured and
  `/actuator/health/readiness` serves `UP`.
- **Wire up the liveness probe** — passes once the liveness probe is configured and
  `/actuator/health/liveness` serves `UP`.

When all three are green, the challenge is complete.

> [!TIP]
> Check your progress with `tpm lab status --id <lab-id>`.

## Clean up

When you're finished, tear down the lab run (removes its namespaces and cluster resources):

```sh
tpm lab cleanup --id <lab-id>
```

> [!NOTE]
> Use `--state-dir <path>` if you started the lab with a custom state directory.

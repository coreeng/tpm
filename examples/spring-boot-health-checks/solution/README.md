# Spring Boot Health Checks — Solution

The reference solution: a Spring Boot app with Actuator, exposing the Kubernetes
liveness/readiness probe endpoints, deployed with both probes wired into the Deployment.

## What makes this the solution

- `build.gradle` includes `spring-boot-starter-actuator`.
- `src/main/resources/application.yaml` enables the health probe groups
  (`/actuator/health/liveness`, `/actuator/health/readiness`).
- `k8s/deployment.yaml` declares a `readinessProbe` and a `livenessProbe` pointing at those
  endpoints, plus restricted Pod Security settings.
- `k8s/service.yaml` is a ClusterIP Service in front of the Deployment.

## Build and deploy

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

The build uses a multi-stage `Dockerfile` (a `gradle:8.14-jdk21` builder produces the Spring
Boot jar, copied into an `eclipse-temurin:21-jre` runtime that runs as a non-root user). This
is reproducible regardless of the host JDK and works across Docker versions.

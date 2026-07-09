# Examples

This directory contains complete examples split by artifact type.

- [`modules/`](modules/) contains complete module source directories that can be validated, built, and previewed.
- [`labs/`](labs/) contains standalone lab runtimes that can be previewed locally against a lab runtime chart.

## Modules

[`modules/kubernetes-101`](modules/kubernetes-101/) is a three-chapter module covering Kubernetes cluster basics, workloads, and application operations.

```bash
tpm module validate examples/modules/kubernetes-101
tpm module build examples/modules/kubernetes-101 --out-root artifacts
tpm module preview examples/modules/kubernetes-101 --watch
```

## Labs

[`labs/spring-boot-health-checks`](labs/spring-boot-health-checks/) is a standalone lab where learners add Spring Boot health checks and Kubernetes probes.

```bash
tpm lab preview examples/labs/spring-boot-health-checks \
  --chart-uri oci://ghcr.io/coreeng/charts/training-platform-assessment
```

> [!TIP]
> Use these examples as working smoke tests when changing the CLI, schemas, preview UI, or lab runtime integration.

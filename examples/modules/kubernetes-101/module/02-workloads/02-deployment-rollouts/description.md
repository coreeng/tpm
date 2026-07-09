Deployments manage rollout and rollback behavior for replicated Pods.

When you update a Deployment, Kubernetes creates a new ReplicaSet and gradually shifts Pods toward the new template. This gives operators a controlled way to change application versions without replacing every Pod at once.

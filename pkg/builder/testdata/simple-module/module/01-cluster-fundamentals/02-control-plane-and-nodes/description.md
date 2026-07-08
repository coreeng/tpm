The control plane keeps the cluster moving toward the desired state stored in the API server.

Worker nodes run the application containers. The kubelet on each node talks to the API server, starts Pods through the container runtime, and reports status back to the cluster.

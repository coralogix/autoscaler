# Coralogix Fork of Kubernetes Autoscaler

We forked Kubernetes Autoscaler in order to solve issues with fragmentation of free resources in on our reserved capacity nodes. We used to have draining disabled for those nodes, which meant there was no mechanism that would move pods from underutilised nodes. This then made it impossible to move node-sized pods from spots to reserved. We introduced the following changes to address this:

- Bin-packing without reducing a node group's desired size.
  - nodes labeled `cluster-autoscaler.kubernetes.io/bin-packing-only=true` are treated as bin-packing nodes (we put that label on nodes in reserved-only node groups).
  - nodes labeled `cluster-autoscaler.kubernetes.io/bin-packing-only=when-on-demand` are treated as bin-packing nodes unless they have a `node-role.kubernetes.io/spot-worker=true` label. (we put that label on all nodes in node groups containing a mix of reserved and spots)
  - During scale-down planning we avoid removing empty bin-packing nodes (they are marked unremovable with reason `BinPackingEmptyNode`).
  - When simulating node removals, we only allow pods from a bin-packing source node to move onto non-empty bin-packing destinations (so that it improves bin-packing, but never moves pods into spots).
  - When deleting bin-packing nodes, the autoscaler immediately restores the node group's target size by calling `IncreaseSize` for the number of bin-packing deletions.
- Priority-aware scale-down candidate ordering.
  - Candidates are additionally sorted using the priority expander config (`ConfigMap` `cluster-autoscaler-priority-expander`, key `priorities`). Nodes with lowest priority are preferred for scaledown.
  - Within the same priority, lower utilization nodes are preferred first.

# Kubernetes Autoscaler

[![Release Charts](https://github.com/kubernetes/autoscaler/actions/workflows/release.yaml/badge.svg)](https://github.com/kubernetes/autoscaler/actions/workflows/release.yaml) [![Tests](https://github.com/kubernetes/autoscaler/actions/workflows/ci.yaml/badge.svg)](https://github.com/kubernetes/autoscaler/actions/workflows/ci.yaml) [![GoDoc Widget]][GoDoc]

This repository contains autoscaling-related components for Kubernetes.

## What's inside

[Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) - a component that automatically adjusts the size of a Kubernetes
Cluster so that all pods have a place to run and there are no unneeded nodes. Supports several public cloud providers. Version 1.0 (GA) was released with kubernetes 1.8.

[Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) - a set of components that automatically adjust the
amount of CPU and memory requested by pods running in the Kubernetes Cluster. Current state - beta.

[Addon Resizer](https://github.com/kubernetes/autoscaler/tree/master/addon-resizer) - a simplified version of vertical pod autoscaler that modifies
resource requests of a deployment based on the number of nodes in the Kubernetes Cluster. Current state - beta.

[Charts](https://github.com/kubernetes/autoscaler/tree/master/charts) - Supported Helm charts for components above.

## Contact Info

Interested in autoscaling? Want to talk? Have questions, concerns or great ideas?

Please join us on #sig-autoscaling at https://kubernetes.slack.com/, or join one
of our weekly meetings.  See [the Kubernetes Community Repo](https://github.com/kubernetes/community/blob/master/sig-autoscaling/README.md) for more information.

## Getting the Code

Fork the repository in the cloud:
1. Visit https://github.com/kubernetes/autoscaler
1. Click Fork button (top right) to establish a cloud-based fork.

The code must be checked out as a subdirectory of `k8s.io`, and not `github.com`.

```shell
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
cd autoscaler
```

Please refer to Kubernetes [Github workflow guide] for more details.

[GoDoc]: https://godoc.org/k8s.io/autoscaler
[GoDoc Widget]: https://godoc.org/k8s.io/autoscaler?status.svg
[Github workflow guide]: https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md

/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scaledown

import apiv1 "k8s.io/api/core/v1"

// BinPackingLabelKey enables bin-packing without scaling down a node group desired size.
const BinPackingLabelKey = "cluster-autoscaler.kubernetes.io/bin-packing-only"

// IsBinPacking returns true when the node should be removed without reducing desired size.
func IsBinPacking(node *apiv1.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}
	return node.Labels[BinPackingLabelKey] == "true"
}

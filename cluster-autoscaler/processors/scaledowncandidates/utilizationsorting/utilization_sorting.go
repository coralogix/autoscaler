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

package utilizationsorting

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
)

// UtilizationSorting sorts non-empty scale down candidates by utilization (lowest first)
// when priorities match.
type UtilizationSorting struct {
	nodeInfoGetter
	cloudProvider               cloudprovider.CloudProvider
	configGetter                nodegroupconfig.NodeGroupConfigProcessor
	ignoreMirrorPodsUtilization bool
}

// NewUtilizationSortingProcessor returns UtilizationSorting struct.
func NewUtilizationSortingProcessor(
	nodeInfoGetter nodeInfoGetter,
	cloudProvider cloudprovider.CloudProvider,
	configGetter nodegroupconfig.NodeGroupConfigProcessor,
	ignoreMirrorPodsUtilization bool,
) *UtilizationSorting {
	return &UtilizationSorting{
		nodeInfoGetter:              nodeInfoGetter,
		cloudProvider:               cloudProvider,
		configGetter:                configGetter,
		ignoreMirrorPodsUtilization: ignoreMirrorPodsUtilization,
	}
}

// ScaleDownEarlierThan return true if node1 should be scaled down earlier than node2.
func (u *UtilizationSorting) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	util1, ok1 := u.nodeUtilization(node1)
	util2, ok2 := u.nodeUtilization(node2)
	if !ok1 || !ok2 {
		return false
	}

	return util1 < util2
}

type nodeInfoGetter interface {
	GetNodeInfo(nodeName string) (*schedulerframework.NodeInfo, error)
}

func (u *UtilizationSorting) nodeUtilization(node *apiv1.Node) (float64, bool) {
	nodeInfo, err := u.nodeInfoGetter.GetNodeInfo(node.Name)
	if err != nil {
		return 0, false
	}

	nodeGroup, err := u.cloudProvider.NodeGroupForNode(node)
	if err != nil || nodeGroup == nil {
		return 0, false
	}

	ignoreDaemonSetsUtilization, err := u.configGetter.GetIgnoreDaemonSetsUtilization(nodeGroup)
	if err != nil {
		return 0, false
	}

	gpuConfig := u.cloudProvider.GetNodeGpuConfig(node)
	utilInfo, err := utilization.Calculate(nodeInfo, ignoreDaemonSetsUtilization, u.ignoreMirrorPodsUtilization, gpuConfig, time.Now())
	if err != nil {
		return 0, false
	}

	return utilInfo.Utilization, true
}

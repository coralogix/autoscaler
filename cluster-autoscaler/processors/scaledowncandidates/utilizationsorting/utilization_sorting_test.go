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
	"testing"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/emptycandidates"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func newTestUtilizationSorting(t *testing.T, nodes []*apiv1.Node, pods []*apiv1.Pod, provider *testprovider.TestCloudProvider) *UtilizationSorting {
	t.Helper()
	snapshot := clustersnapshot.NewBasicClusterSnapshot()
	clustersnapshot.InitializeClusterSnapshotOrDie(t, snapshot, nodes, pods)
	configGetter := nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{})
	return NewUtilizationSortingProcessor(
		emptycandidates.NewNodeInfoGetter(snapshot),
		provider,
		configGetter,
		false,
	)
}

func TestUtilizationSortingWithinPriority(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng", 0, 10, 2)

	lowUtilNode := BuildTestNode("node-low-util", 1000, 1000)
	highUtilNode := BuildTestNode("node-high-util", 1000, 1000)
	provider.AddNode("ng", lowUtilNode)
	provider.AddNode("ng", highUtilNode)

	pods := []*apiv1.Pod{
		BuildTestPod("pod-low-util", 100, 100, WithNodeName(lowUtilNode.Name)),
		BuildTestPod("pod-high-util", 300, 300, WithNodeName(highUtilNode.Name)),
	}

	comparator := newTestUtilizationSorting(t, []*apiv1.Node{lowUtilNode, highUtilNode}, pods, provider)

	assert.True(t, comparator.ScaleDownEarlierThan(lowUtilNode, highUtilNode))
	assert.False(t, comparator.ScaleDownEarlierThan(highUtilNode, lowUtilNode))
}

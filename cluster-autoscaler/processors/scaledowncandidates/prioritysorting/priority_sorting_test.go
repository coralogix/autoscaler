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

package prioritysorting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const testNamespace = "default"

func newTestPrioritySorting(t *testing.T, configYaml string, provider *testprovider.TestCloudProvider) *PrioritySorting {
	t.Helper()
	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      PriorityConfigMapName,
		},
		Data: map[string]string{
			ConfigMapKey: configYaml,
		},
	}
	lister, err := kubernetes.NewTestConfigMapLister([]*apiv1.ConfigMap{cm})
	if err != nil {
		t.Fatalf("failed to create configmap lister: %v", err)
	}
	recorder := record.NewFakeRecorder(10)
	return NewPrioritySortingProcessor(provider, lister.ConfigMaps(testNamespace), recorder)
}

func TestPrioritySortingByPriority(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-high", 0, 10, 1)
	provider.AddNodeGroup("ng-low", 0, 10, 1)

	highNode := BuildTestNode("node-high", 1000, 1000)
	lowNode := BuildTestNode("node-low", 1000, 1000)
	provider.AddNode("ng-high", highNode)
	provider.AddNode("ng-low", lowNode)

	comparator := newTestPrioritySorting(t, `
10:
  - ".*ng-high.*"
5:
  - ".*ng-low.*"
`, provider)

	assert.True(t, comparator.ScaleDownEarlierThan(lowNode, highNode))
	assert.False(t, comparator.ScaleDownEarlierThan(highNode, lowNode))
}

func TestPrioritySortingSkipsEmptyNodes(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng", 0, 10, 2)

	emptyNode := BuildTestNode("node-empty", 1000, 1000)
	nonEmptyNode := BuildTestNode("node-non-empty", 1000, 1000)
	provider.AddNode("ng", emptyNode)
	provider.AddNode("ng", nonEmptyNode)

	comparator := newTestPrioritySorting(t, `
1:
  - ".*"
`, provider)

	assert.False(t, comparator.ScaleDownEarlierThan(emptyNode, nonEmptyNode))
	assert.False(t, comparator.ScaleDownEarlierThan(nonEmptyNode, emptyNode))
}

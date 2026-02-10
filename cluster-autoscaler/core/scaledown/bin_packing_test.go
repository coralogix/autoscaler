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

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
)

func TestIsBinPacking(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected bool
	}{
		{
			name:     "nil node",
			labels:   nil,
			expected: false,
		},
		{
			name:     "node with nil labels",
			labels:   nil, // see test body: node created with Labels explicitly nil
			expected: false,
		},
		{
			name:     "enable-bin-packing true",
			labels:   map[string]string{BinPackingLabelKey: "true"},
			expected: true,
		},
		{
			name:     "enable-bin-packing true with spot-worker still bin-packing",
			labels:   map[string]string{BinPackingLabelKey: "true", SpotWorkerLabelKey: "true"},
			expected: true,
		},
		{
			name:     "when-on-demand without spot-worker",
			labels:   map[string]string{BinPackingLabelKey: BinPackingLabelValueWhenOnDemand},
			expected: true,
		},
		{
			name:     "when-on-demand with spot-worker false",
			labels:   map[string]string{BinPackingLabelKey: BinPackingLabelValueWhenOnDemand, SpotWorkerLabelKey: "false"},
			expected: true,
		},
		{
			name:     "when-on-demand with spot-worker true",
			labels:   map[string]string{BinPackingLabelKey: BinPackingLabelValueWhenOnDemand, SpotWorkerLabelKey: "true"},
			expected: false,
		},
		{
			name:     "when-on-demand and no spot-worker label",
			labels:   map[string]string{BinPackingLabelKey: BinPackingLabelValueWhenOnDemand},
			expected: true,
		},
		{
			name:     "other value not bin-packing",
			labels:   map[string]string{BinPackingLabelKey: "other"},
			expected: false,
		},
		{
			name:     "no bin-packing label",
			labels:   map[string]string{"other": "label"},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node *apiv1.Node
			if tt.name == "nil node" {
				node = nil
			} else if tt.name == "node with nil labels" {
				node = &apiv1.Node{}
				// Labels is nil by default
			} else {
				node = &apiv1.Node{}
				node.Labels = tt.labels
			}
			got := IsBinPacking(node)
			if got != tt.expected {
				t.Errorf("IsBinPacking() = %v, want %v", got, tt.expected)
			}
		})
	}
}

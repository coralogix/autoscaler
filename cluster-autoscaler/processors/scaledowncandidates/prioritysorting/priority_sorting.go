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
	"errors"
	"fmt"
	"regexp"
	"sync"

	"gopkg.in/yaml.v2"

	apiv1 "k8s.io/api/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	// PriorityConfigMapName defines a name of the ConfigMap used to store priority configuration.
	// This is shared with the priority expander to reuse its configuration.
	PriorityConfigMapName = "cluster-autoscaler-priority-expander"
	// ConfigMapKey defines the key used in the ConfigMap to configure priorities.
	ConfigMapKey = "priorities"
)

type priorities map[int][]*regexp.Regexp

type priorityConfig struct {
	cloudProvider   cloudprovider.CloudProvider
	configMapLister v1lister.ConfigMapNamespaceLister
	logRecorder     record.EventRecorder

	mu                          sync.Mutex
	cachedPriorities            priorities
	cachedConfigResourceVersion string
}

// PrioritySorting sorts non-empty scale down candidates by configured node group priorities.
type PrioritySorting struct {
	config *priorityConfig
}

// NewPrioritySortingProcessor returns PrioritySorting struct.
func NewPrioritySortingProcessor(
	cloudProvider cloudprovider.CloudProvider,
	configMapLister v1lister.ConfigMapNamespaceLister,
	logRecorder record.EventRecorder,
) *PrioritySorting {
	return &PrioritySorting{
		config: &priorityConfig{
			cloudProvider:   cloudProvider,
			configMapLister: configMapLister,
			logRecorder:     logRecorder,
		},
	}
}

// ScaleDownEarlierThan return true if node1 should be scaled down earlier than node2.
func (p *PrioritySorting) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	priorities, ok := p.config.loadPriorities()
	if !ok {
		return false
	}

	priority1 := p.config.priorityForNode(node1, priorities)
	priority2 := p.config.priorityForNode(node2, priorities)

	return priority1 < priority2
}

func (p *priorityConfig) priorityForNode(node *apiv1.Node, priorities priorities) int {
	nodeGroup, err := p.cloudProvider.NodeGroupForNode(node)
	if err != nil || nodeGroup == nil {
		return 0
	}

	id := nodeGroup.Id()
	maxPrio := 0
	for prio, nameRegexpList := range priorities {
		if !groupIDMatchesList(id, nameRegexpList) {
			continue
		}
		if prio > maxPrio {
			maxPrio = prio
		}
	}

	return maxPrio
}

func groupIDMatchesList(id string, nameRegexpList []*regexp.Regexp) bool {
	for _, re := range nameRegexpList {
		if re.FindStringIndex(id) != nil {
			return true
		}
	}
	return false
}

func (p *priorityConfig) loadPriorities() (priorities, bool) {
	cm, err := p.configMapLister.Get(PriorityConfigMapName)
	if err != nil {
		return nil, false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cachedPriorities != nil && cm.ResourceVersion == p.cachedConfigResourceVersion {
		return p.cachedPriorities, true
	}

	prioString, found := cm.Data[ConfigMapKey]
	if !found {
		msg := fmt.Sprintf("Wrong configmap for scale-down priority sorting, doesn't contain %s key. Ignoring update.",
			ConfigMapKey)
		p.logConfigWarning(cm, "ScaleDownPriorityConfigMapInvalid", msg)
		return p.cachedPriorities, p.cachedPriorities != nil
	}

	newPriorities, err := p.parsePrioritiesYAMLString(prioString)
	if err != nil {
		msg := fmt.Sprintf("Wrong configuration for scale-down priority sorting: %v. Ignoring update.", err)
		p.logConfigWarning(cm, "ScaleDownPriorityConfigMapInvalid", msg)
		return p.cachedPriorities, p.cachedPriorities != nil
	}

	p.cachedPriorities = newPriorities
	p.cachedConfigResourceVersion = cm.ResourceVersion
	return p.cachedPriorities, true
}

func (p *priorityConfig) logConfigWarning(cm *apiv1.ConfigMap, reason, msg string) {
	p.logRecorder.Event(cm, apiv1.EventTypeWarning, reason, msg)
	klog.Warning(msg)
}

func (p *priorityConfig) parsePrioritiesYAMLString(prioritiesYAML string) (priorities, error) {
	if prioritiesYAML == "" {
		return nil, fmt.Errorf("priority configuration in %s configmap is empty; please provide valid configuration",
			PriorityConfigMapName)
	}
	var config map[int][]string
	if err := yaml.Unmarshal([]byte(prioritiesYAML), &config); err != nil {
		return nil, fmt.Errorf("Can't parse YAML with priorities in the configmap: %v", err)
	}
	if len(config) == 0 {
		return nil, errors.New("no priorities entries found")
	}

	newPriorities := make(map[int][]*regexp.Regexp)
	for prio, reList := range config {
		for _, re := range reList {
			compiled, err := regexp.Compile(re)
			if err != nil {
				return nil, fmt.Errorf("Can't compile regexp rule for priority %d and rule %s: %v", prio, re, err)
			}
			newPriorities[prio] = append(newPriorities[prio], compiled)
		}
	}

	klog.V(4).Info("Successfully loaded scale-down priority configuration from configmap.")
	return newPriorities, nil
}

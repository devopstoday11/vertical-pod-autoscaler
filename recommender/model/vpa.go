/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

// Vpa (Vertical Pod Autoscaler) object is responsible for vertical scaling of
// Pods matching a given label selector.
type Vpa struct {
	ID VpaID
	// Labels selector that determines which Pods are controlled by this VPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Most recently computed recommendation. Can be nil.
	Recommendation *vpa_types.RecommendedPodResources
	// Pods controlled by this VPA object.
	Pods map[PodID]*PodState
	// Value of the Status.LastUpdateTime fetched from the VPA API object.
	LastUpdateTime metav1.Time
}

// NewVpa returns a new Vpa with a given ID and pod selector. Doesn't set the
// links to the matched pods.
func NewVpa(id VpaID, selector labels.Selector) *Vpa {
	vpa := &Vpa{
		ID:          id,
		PodSelector: selector,
		Pods:        make(map[PodID]*PodState), // Empty pods map.
	}
	return vpa
}

// MatchesPod returns true iff a given pod is matched by the Vpa pod selector.
func (vpa *Vpa) MatchesPod(pod *PodState) bool {
	if vpa.ID.Namespace != pod.ID.Namespace {
		return false
	}
	return vpa.PodSelector != nil && pod.Labels != nil && vpa.PodSelector.Matches(pod.Labels)
}

// UpdatePodLink marks the Pod as controlled or not-controlled by the VPA
// depending on whether the pod labels match the Vpa pod selector.
// If multiple VPAs match the same Pod, only one of them will effectively
// control the Pod.
func (vpa *Vpa) UpdatePodLink(pod *PodState) bool {
	_, previouslyMatched := pod.MatchingVpas[vpa.ID]
	currentlyMatching := vpa.MatchesPod(pod)

	if previouslyMatched == currentlyMatching {
		return false
	}
	if currentlyMatching {
		// Create links between VPA and pod.
		vpa.Pods[pod.ID] = pod
		pod.MatchingVpas[vpa.ID] = vpa
	} else {
		// Delete the links between VPA and pod.
		delete(vpa.Pods, pod.ID)
		delete(pod.MatchingVpas, vpa.ID)
	}
	return true
}

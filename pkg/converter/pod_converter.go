// Copyright (c) 2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"fmt"
	"reflect"

	"github.com/projectcalico/libcalico-go/lib/api"
	log "github.com/sirupsen/logrus"
	k8sApiV1 "k8s.io/client-go/pkg/api/v1"
)

// Label which represents the namespace a given pod belongs to.
const k8sNamespaceLabel = "calico/k8s_ns"

// WorkloadEndpointData is an internal struct used to store the various bits
// of information that the policy controller cares about on a workload endpoint.
type WorkloadEndpointData struct {
	Key    string
	Labels map[string]string
}

type podConverter struct {
}

// BuildWorkloadEndpointData generates the correct WorkloadEndpointData for the given
// WorkloadEndpoint, extracting fields that the policy controller is responsible for syncing.
func BuildWorkloadEndpointData(wep api.WorkloadEndpoint) WorkloadEndpointData {
	return WorkloadEndpointData{
		Key:    wep.Metadata.Workload,
		Labels: wep.Metadata.Labels,
	}
}

// MergeWorkloadEndpointData applies the given WorkloadEndpointData to the provided
// WorkloadEndpoint, updating relevant fields with new values.
func MergeWorkloadEndpointData(wep *api.WorkloadEndpoint, upd WorkloadEndpointData) {
	if wep.Metadata.Workload != upd.Key {
		log.Fatalf("Bad attempt to merge data for %s into wep %s", upd.Key, wep.Metadata.Workload)
	}
	wep.Metadata.Labels = upd.Labels
}

// NewPodConverter Constructor for podConverter
func NewPodConverter() Converter {
	return &podConverter{}
}

func (p *podConverter) Convert(k8sObj interface{}) (interface{}, error) {
	if reflect.TypeOf(k8sObj) != reflect.TypeOf(&k8sApiV1.Pod{}) {
		log.Fatalf("can not convert object %#v to workloadEndpoint. Object is not of type *v1.Pod", k8sObj)
	}

	pod := k8sObj.(*k8sApiV1.Pod)
	endpoint := api.NewWorkloadEndpoint()

	endpoint.Metadata.Workload = fmt.Sprintf("%s.%s", pod.Namespace, pod.Name)
	if pod.ObjectMeta.Labels != nil {
		endpoint.Metadata.Labels = pod.ObjectMeta.Labels
	} else {
		endpoint.Metadata.Labels = map[string]string{}
	}

	// Add a special label for the Kubernetes namespace.  This is used
	// by selector-based policies to select all pods in a given namespace.
	endpoint.Metadata.Labels[k8sNamespaceLabel] = pod.Namespace

	// Build and return a WorkloadEndpointData struct using the data.
	return BuildWorkloadEndpointData(*endpoint), nil
}

// GetKey returns workloadID of the object as the key.
// For pods, the workloadID is of the form `namespace.name`.
func (p *podConverter) GetKey(obj interface{}) string {
	e := obj.(WorkloadEndpointData)
	return e.Key
}
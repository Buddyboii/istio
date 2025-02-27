// Copyright Istio Authors
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

package echo

import (
	"sort"
	"strings"

	"istio.io/istio/pilot/pkg/model"
)

// Services is a set of Instances that share the same FQDN. While an Instance contains
// multiple deployments (a single service in a single cluster), Instances contains multiple
// deployments that may contain multiple Services.
type Services []Instances

// GetByService finds the first Instances with the given Service name. It is possible to have multiple deployments
// with the same service name but different namespaces (and therefore different FQDNs). Use caution when relying on
// Service.
func (d Services) GetByService(service string) Target {
	for _, target := range d {
		if target.Config().Service == service {
			return target
		}
	}
	return nil
}

type ServiceNameList []model.NamespacedName

func (l ServiceNameList) Names() []string {
	out := make([]string, 0, len(l))
	for _, n := range l {
		out = append(out, n.Name)
	}
	return out
}

func (l ServiceNameList) NamespacedNames() []string {
	out := make([]string, 0, len(l))
	for _, n := range l {
		out = append(out, n.Name+"."+n.Namespace)
	}
	return out
}

// ServiceNames gives the service names of each deployment in order.
func (d Services) ServiceNames() ServiceNameList {
	var out ServiceNameList
	for _, target := range d {
		out = append(out, target.NamespacedName())
	}
	return out
}

// ServiceNamesWithNamespacePrefix is similar to ServiceNames but returns namspaces prefixes rather than the full
// namespace names. This is useful for test method naming and logs.
func (d Services) ServiceNamesWithNamespacePrefix() ServiceNameList {
	var out ServiceNameList
	for _, target := range d {
		out = append(out, model.NamespacedName{
			Name:      target.Config().Service,
			Namespace: target.Config().Namespace.Prefix(),
		})
	}
	return out
}

// FQDNs gives the fully-qualified-domain-names each deployment in order.
func (d Services) FQDNs() []string {
	var out []string
	for _, target := range d {
		out = append(out, target.Config().ClusterLocalFQDN())
	}
	return out
}

func (d Services) Instances() Instances {
	var out Instances
	for _, target := range d {
		out = append(out, target.Instances()...)
	}
	return out
}

func (d Services) MatchFQDNs(fqdns ...string) Services {
	match := map[string]bool{}
	for _, fqdn := range fqdns {
		match[fqdn] = true
	}
	var out Services
	for _, target := range d {
		if match[target.Config().ClusterLocalFQDN()] {
			out = append(out, target)
		}
	}
	return out
}

// Services must be sorted to make sure tests have consistent ordering
var _ sort.Interface = Services{}

// Len returns the number of deployments
func (d Services) Len() int {
	return len(d)
}

// Less returns true if the element at i should appear before the element at j in a sorted Services
func (d Services) Less(i, j int) bool {
	return strings.Compare(d[i].Config().ClusterLocalFQDN(), d[j].Config().ClusterLocalFQDN()) < 0
}

// Swap switches the positions of elements at i and j (used for sorting).
func (d Services) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Copy this services array.
func (d Services) Copy() Services {
	return append(Services{}, d...)
}

// Append returns a new Services array with the given values appended.
func (d Services) Append(others ...Services) Services {
	out := d.Copy()
	for _, o := range others {
		out = append(out, o...)
	}
	sort.Stable(out)
	return out
}

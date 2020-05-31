package object

import (
	"fmt"

	"github.com/miekg/dns"
	api "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Ingress is a stripped down api.Ingress with only the items we need for CoreDNS.
type Ingress struct {
	// Don't add new fields to this struct without talking to the CoreDNS maintainers.
	Version   string
	Name      string
	Namespace string
	Index     string
	Hosts     []string
	// ExternalIPs we may want to export.
	ExternalIPs []string

	*Empty
}

// IngressKey return a string using for the index.
func IngressKey(name, namespace string) string { return name + "." + namespace }

// ToIngress returns a function that converts an api.Ingress to a *Ingress.
func ToIngress(skipCleanup bool) ToFunc {
	return func(obj interface{}) (interface{}, error) {
		ing, ok := obj.(*api.Ingress)
		if !ok {
			return nil, fmt.Errorf("unexpected object %v", obj)
		}
		return toIngress(skipCleanup, ing), nil
	}
}

func toIngress(skipCleanup bool, ing *api.Ingress) *Ingress {
	ingress := &Ingress{
		Version:   ing.GetResourceVersion(),
		Name:      ing.GetName(),
		Namespace: ing.GetNamespace(),
		Index:     IngressKey(ing.GetName(), ing.GetNamespace()),
		Hosts:     []string{},

		ExternalIPs: make([]string, len(ing.Status.LoadBalancer.Ingress)),
	}

	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			ingress.Hosts = append(ingress.Hosts, dns.Fqdn(rule.Host))
		}
	}

	for i, lb := range ing.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			ingress.ExternalIPs[i] = lb.IP
			continue
		}
		ingress.ExternalIPs[i] = lb.Hostname

	}

	if !skipCleanup {
		*ing = api.Ingress{}
	}

	return ingress
}

var _ runtime.Object = &Ingress{}

// DeepCopyObject implements the ObjectKind interface.
func (i *Ingress) DeepCopyObject() runtime.Object {
	i1 := &Ingress{
		Version:     i.Version,
		Name:        i.Name,
		Namespace:   i.Namespace,
		Index:       i.Index,
		Hosts:       make([]string, len(i.Hosts)),
		ExternalIPs: make([]string, len(i.ExternalIPs)),
	}
	copy(i1.Hosts, i.Hosts)
	copy(i1.ExternalIPs, i.ExternalIPs)
	return i1
}

// GetNamespace implements the metav1.Object interface.
func (i *Ingress) GetNamespace() string { return i.Namespace }

// SetNamespace implements the metav1.Object interface.
func (i *Ingress) SetNamespace(namespace string) {}

// GetName implements the metav1.Object interface.
func (i *Ingress) GetName() string { return i.Name }

// SetName implements the metav1.Object interface.
func (i *Ingress) SetName(name string) {}

// GetResourceVersion implements the metav1.Object interface.
func (i *Ingress) GetResourceVersion() string { return i.Version }

// SetResourceVersion implements the metav1.Object interface.
func (i *Ingress) SetResourceVersion(version string) {}

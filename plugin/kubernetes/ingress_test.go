package kubernetes

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/plugin/kubernetes/object"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
	api "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ingCases = []struct {
	Qname string
	Qtype uint16
	Msg   []msg.Service
	Rcode int
}{
	{
		Qname: "foo.example.org", Rcode: dns.RcodeSuccess,
		Msg: []msg.Service{
			{Host: "1.2.3.4", TTL: 5, Key: "foo.example.org"},
		},
	},
	{
		Qname: "foo6.example.org.", Rcode: dns.RcodeSuccess,
		Msg: []msg.Service{
			{Host: "1:2::5", TTL: 5, Key: "/c/org/example/testns/ing6"},
		},
	},
	{
		Qname: "*._not-udp-or-tcp.svc1.testns.example.com.", Rcode: dns.RcodeSuccess,
	},
	{
		Qname: "foo.testns.example.org.", Rcode: dns.RcodeNameError,
	},
	{
		Qname: "svc0.svc-nons.example.com.", Rcode: dns.RcodeNameError,
	},
}

func TestIngress(t *testing.T) {
	k := New([]string{"cluster.local."})
	k.APIConn = &ingressAPI{}
	k.Next = test.NextHandler(dns.RcodeSuccess, nil)
	k.Namespaces = map[string]struct{}{"testns": {}}

	for i, tc := range ingCases {
		state := testRequest(tc.Qname)

		svc, rcode := k.ExternalIngress(state)

		if x := tc.Rcode; x != rcode {
			t.Errorf("Test %d, expected rcode %d, got %d", i, x, rcode)
		}

		if len(svc) != len(tc.Msg) {
			t.Errorf("Test %d, expected %d for messages, got %d", i, len(tc.Msg), len(svc))
		}

		for j, s := range svc {
			if x := tc.Msg[j].Key; x != s.Key {
				t.Errorf("Test %d, expected key %s, got %s", i, x, s.Key)
			}
			return
		}
	}
}

type ingressAPI struct{}

func (ingressAPI) HasSynced() bool                                                   { return true }
func (ingressAPI) Run()                                                              {}
func (ingressAPI) Stop() error                                                       { return nil }
func (ingressAPI) EpIndexReverse(string) []*object.Endpoints                         { return nil }
func (ingressAPI) SvcIndexReverse(string) []*object.Service                          { return nil }
func (ingressAPI) Modified() int64                                                   { return 0 }
func (ingressAPI) EpIndex(s string) []*object.Endpoints                              { return nil }
func (ingressAPI) EndpointsList() []*object.Endpoints                                { return nil }
func (ingressAPI) GetNodeByName(ctx context.Context, name string) (*api.Node, error) { return nil, nil }
func (ingressAPI) SvcIndex(s string) []*object.Service                               { return nil }
func (ingressAPI) ServiceList() []*object.Service                                    { return nil }
func (ingressAPI) PodIndex(string) []*object.Pod                                     { return nil }
func (ingressAPI) IngIndex(s string) []*object.Ingress                               { return nil }
func (ingressAPI) IngIndexReverse(s string) []*object.Ingress                        { return ingressIndexReverse[s] }

func (ingressAPI) GetNamespaceByName(name string) (*api.Namespace, error) {
	return &api.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
	}, nil
}

var ingressIndexReverse = map[string][]*object.Ingress{
	"foo.example.org": {
		{
			Name:        "ing4",
			Namespace:   "testns",
			Hosts:       []string{"foo.example.org"},
			ExternalIPs: []string{"1.2.3.4"},
		},
	},
	"svc6.testns": {
		{
			Name:        "ing6",
			Namespace:   "testns",
			Hosts:       []string{"foo6.example.org"},
			ExternalIPs: []string{"1:2::5"},
		},
	},
}

func (ingressAPI) IngressList() []*object.Ingress {
	var ings []*object.Ingress
	for _, ing := range ingressIndexReverse {
		ings = append(ings, ing...)
	}
	return ings
}

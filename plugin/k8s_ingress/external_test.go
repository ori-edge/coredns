package ingress

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/kubernetes"
	"github.com/coredns/coredns/plugin/kubernetes/object"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
	api "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExternal(t *testing.T) {
	k := kubernetes.New([]string{"cluster.local."})
	k.Namespaces = map[string]struct{}{"testns": {}}
	k.APIConn = &APIConnTest{}

	e := New()
	e.Zones = []string{"example.com.", "example.org."}
	e.Next = test.NextHandler(dns.RcodeSuccess, nil)
	e.externalFunc = k.ExternalIngress
	e.externalAddrFunc = externalAddress // internal test function

	ctx := context.TODO()
	for i, tc := range tests {
		r := tc.Msg()
		w := dnstest.NewRecorder(&test.ResponseWriter{})

		_, err := e.ServeDNS(ctx, w, r)
		if err != tc.Error {
			t.Errorf("Test %d expected no error, got %v", i, err)
			return
		}
		if tc.Error != nil {
			continue
		}

		resp := w.Msg

		if resp == nil {
			t.Fatalf("Test %d, got nil message and no error for %q", i, r.Question[0].Name)
		}
		if err = test.SortAndCheck(resp, tc); err != nil {
			t.Error(err)
		}
	}
}

var tests = []test.Case{
	// A Service
	{
		Qname: "foo.example.com.", Qtype: dns.TypeA, Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("foo.example.com.	5	IN	A	1.2.3.4"),
		},
	},
	{
		Qname: "foo.example.org.", Qtype: dns.TypeA, Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.CNAME("foo.example.org.	5	IN	CNAME	dummy.hostname"),
		},
	},
}

type APIConnTest struct{}

func (APIConnTest) HasSynced() bool                                                   { return true }
func (APIConnTest) Run()                                                              {}
func (APIConnTest) Stop() error                                                       { return nil }
func (APIConnTest) EpIndexReverse(string) []*object.Endpoints                         { return nil }
func (APIConnTest) SvcIndexReverse(string) []*object.Service                          { return nil }
func (APIConnTest) Modified() int64                                                   { return 0 }
func (APIConnTest) EpIndex(s string) []*object.Endpoints                              { return nil }
func (APIConnTest) EndpointsList() []*object.Endpoints                                { return nil }
func (APIConnTest) GetNodeByName(ctx context.Context, name string) (*api.Node, error) { return nil, nil }
func (APIConnTest) SvcIndex(string) []*object.Service                                 { return nil }
func (APIConnTest) ServiceList() []*object.Service                                    { return nil }
func (APIConnTest) PodIndex(string) []*object.Pod                                     { return nil }
func (APIConnTest) IngIndexReverse(s string) []*object.Ingress                        { return ingIndexReverseExternal[s] }
func (APIConnTest) IngIndex(s string) []*object.Ingress                               { return nil }

func (APIConnTest) GetNamespaceByName(name string) (*api.Namespace, error) {
	return &api.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
	}, nil
}

var ingIndexReverseExternal = map[string][]*object.Ingress{
	"foo.example.com.": {
		{
			Name:        "ing1",
			Namespace:   "testns",
			Hosts:       []string{"foo.example.com", "bar.example.com"},
			ExternalIPs: []string{"1.2.3.4"},
		},
	},
	"bar.example.com.": {
		{
			Name:        "ing1",
			Namespace:   "testns",
			Hosts:       []string{"foo.example.com", "bar.example.com"},
			ExternalIPs: []string{"1.2.3.4"},
		},
	},
	"foo.example.org.": {
		{
			Name:        "ing2",
			Namespace:   "testns",
			Hosts:       []string{"foo.example.org", "bar.failed.org"},
			ExternalIPs: []string{"dummy.hostname"},
		},
	},
}

func (APIConnTest) IngressList() []*object.Ingress {
	var ings []*object.Ingress
	for _, ing := range ingIndexReverseExternal {
		ings = append(ings, ing...)
	}
	return ings
}

func externalAddress(state request.Request) []dns.RR {
	a := test.A("example.org. IN A 127.0.0.1")
	return []dns.RR{a}
}

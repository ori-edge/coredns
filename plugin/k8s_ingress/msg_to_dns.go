package ingress

import (
	"context"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func (e *Ingress) a(ctx context.Context, services []msg.Service, state request.Request) (records []dns.RR) {
	dup := make(map[string]struct{})
	for _, s := range services {

		what, ip := s.HostType()
		//log.Printf("\n Ingress Hostname: %s, Answer IP/Host: %s\n", s.Key, ip)
		if !dns.IsSubDomain(dns.Fqdn(s.Key), state.QName()) {
			//log.Println("Domain not matched, continuing")
			continue
		}

		switch what {
		case dns.TypeCNAME:
			rr := s.NewCNAME(state.QName(), s.Host)
			records = append(records, rr)
			if resp, err := e.upstream.Lookup(ctx, state, dns.Fqdn(s.Host), dns.TypeA); err == nil {
				for _, rr := range resp.Answer {
					records = append(records, rr)
				}
			}

		case dns.TypeA:

			if _, ok := dup[s.Host]; !ok {
				dup[s.Host] = struct{}{}
				rr := s.NewA(state.QName(), ip)
				rr.Hdr.Ttl = e.ttl
				records = append(records, rr)
			}

		case dns.TypeAAAA:
			// nada
		}
	}
	return records
}

func (e *Ingress) aaaa(ctx context.Context, services []msg.Service, state request.Request) (records []dns.RR) {
	dup := make(map[string]struct{})

	for _, s := range services {

		what, ip := s.HostType()

		switch what {
		case dns.TypeCNAME:
			rr := s.NewCNAME(state.QName(), s.Host)
			records = append(records, rr)
			if resp, err := e.upstream.Lookup(ctx, state, dns.Fqdn(s.Host), dns.TypeAAAA); err == nil {
				for _, rr := range resp.Answer {
					records = append(records, rr)
				}
			}

		case dns.TypeA:
			// nada

		case dns.TypeAAAA:
			if _, ok := dup[s.Host]; !ok {
				dup[s.Host] = struct{}{}
				rr := s.NewAAAA(state.QName(), ip)
				rr.Hdr.Ttl = e.ttl
				records = append(records, rr)
			}
		}
	}
	return records
}

// not sure if this is even needed.

// item holds records.
type item struct {
	name string // name of the record (either owner or something else unique).
	port uint16 // port of the record (used for address records, A and AAAA).
	addr string // address of the record (A and AAAA).
}

// isDuplicate uses m to see if the combo (name, addr, port) already exists. If it does
// not exist already IsDuplicate will also add the record to the map.
func isDuplicate(m map[item]struct{}, name, addr string, port uint16) bool {
	if addr != "" {
		_, ok := m[item{name, 0, addr}]
		if !ok {
			m[item{name, 0, addr}] = struct{}{}
		}
		return ok
	}
	_, ok := m[item{name, port, ""}]
	if !ok {
		m[item{name, port, ""}] = struct{}{}
	}
	return ok
}

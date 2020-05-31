package kubernetes

import (
	"strings"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// ExternalIngress implements the ExternalFunc call from the external ingress plugin.
// It returns any services matching in the services' ExternalIPs.
func (k *Kubernetes) ExternalIngress(state request.Request) ([]msg.Service, int) {
	//base, _ := dnsutil.TrimZone(state.Name(), state.Zone)
	//fmt.Printf("\nBase: %s, Name: %s, Zone: %s | ", base, state.Name(), state.Zone)
	//for _, i := range k.APIConn.IngressList() {
	//	fmt.Printf("\n  stored ingress : %+v\n", i)
	//}
	ingressList := k.APIConn.IngIndexReverse(state.Name())
	services := []msg.Service{}
	rcode := dns.RcodeNameError

	for _, ing := range ingressList {

		for _, ip := range ing.ExternalIPs {
			for _, host := range ing.Hosts {

				rcode = dns.RcodeSuccess
				s := msg.Service{Host: ip, TTL: k.ttl}
				s.Key = strings.Join([]string{host}, "/")

				services = append(services, s)
			}
		}
	}
	return services, rcode
}

// ExternalIngressAddress returns the external service address(es) for the CoreDNS service.
func (k *Kubernetes) ExternalIngressAddress(state request.Request) []dns.RR {
	// If CoreDNS is running inside the Kubernetes cluster: k.nsAddrs() will return the external IPs of the services
	// targeting the CoreDNS Pod.
	// If CoreDNS is running outside of the Kubernetes cluster: k.nsAddrs() will return the first non-loopback IP
	// address seen on the local system it is running on. This could be the wrong answer if coredns is using the *bind*
	// plugin to bind to a different IP address.
	return k.nsAddrs(true, state.Zone)
}

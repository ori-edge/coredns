# k8s_ingress

## Name

*k8s_ingress* - resolves ingress FQDNs from outside Kubernetes clusters.

## Description

This plugin allows an additional zone to resolve the external IP address(es) of a Kubernetes
Ingresses, similar to what `k8s_external` does for services. This plugin is only useful if the *kubernetes* plugin is also loaded.

The plugin uses an external zone to resolve in-cluster IP addresses. It only handles queries for A,
AAAA; all others result in NODATA responses. To make it a proper DNS zone, it handles SOA and NS queries for the apex of the zone.

By default the apex of the zone will look like the following (assuming the zone used is `example.org`):

~~~ dns
example.org.	5 IN	SOA ns1.dns.example.org. hostmaster.example.org. (
				12345      ; serial
				14400      ; refresh (4 hours)
				3600       ; retry (1 hour)
				604800     ; expire (1 week)
				5          ; minimum (4 hours)
				)
example.org		5 IN	NS ns1.dns.example.org.

ns1.dns.example.org.  5 IN  A    ....
ns1.dns.example.org.  5 IN  AAAA ....
~~~

Note that we use the `dns` subdomain for the records DNS needs (see the `apex` directive). Also
note the SOA's serial number is static. The IP addresses of the nameserver records are those of the
CoreDNS service.

The *k8s_ingress* plugin handles the subdomain `dns` and the apex of the zone itself; all other
queries are resolved to addresses in the cluster.

## Syntax

~~~
k8s_ingress [ZONE...]
~~~

* **ZONES** zones *k8s_ingress* should be authoritative for.

If you want to change the apex domain or use a different TTL for the returned records you can use
this extended syntax.

~~~
k8s_ingress [ZONE...] {
    apex APEX
    ttl TTL
}
~~~

* **APEX** is the name (DNS label) to use for the apex records; it defaults to `dns`.
* `ttl` allows you to set a custom **TTL** for responses. The default is 5 (seconds).

## Examples

Enable names under `example.org` to be resolved to in-cluster DNS addresses.

~~~
. {
   kubernetes cluster.local
   k8s_ingress example.org
}
~~~

With the Corefile above, the following Ingress will get an `A` record for `test.example.org` with the IP address `192.168.200.123`.

~~~
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
 name: test
 namespace: default
spec:
  rules:
  - host: test.example.org
    http:
      paths:
      - backend:
          serviceName: backend-service
          servicePort: 80
        path: /
status:
  loadBalancer:
    ingress:
    - ip: 192.168.200.123
~~~


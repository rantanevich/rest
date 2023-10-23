package rest

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-pkgz/rest/realip"
)

// OnlyFrom middleware allows access for limited list of source IPs.
// Such IPs can be defined as complete ip (like 192.168.1.12), prefix (129.168.) or CIDR (192.168.0.0/16)
func OnlyFrom(onlyIps ...string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(onlyIps) == 0 {
				// no restrictions if no ips defined
				h.ServeHTTP(w, r)
				return
			}
			matched, ip := matchSourceIP(r, onlyIps)
			if matched {
				// matched ip - allow
				h.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusForbidden)
			RenderJSON(w, JSON{"error": fmt.Sprintf("ip %s rejected", ip)})
		}
		return http.HandlerFunc(fn)
	}
}

// matchSourceIP returns true if request's ip matches any of ips
func matchSourceIP(r *http.Request, ips []string) (result bool, match string) {
	ip, err := realip.Get(r)
	if err != nil {
		return false, "" // we can't get ip, so no match
	}
	// check for ip prefix or CIDR
	for _, exclIP := range ips {
		if _, cidrnet, err := net.ParseCIDR(exclIP); err == nil {
			if cidrnet.Contains(net.ParseIP(ip)) {
				return true, ip
			}
		}
		if strings.HasPrefix(ip, exclIP) {
			return true, ip
		}
	}
	return false, ip
}

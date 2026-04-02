package scanner

import (
	"context"
	"net"
)

type DNSResult struct {
	Host  string   `json:"host"`
	IPs   []string `json:"ips,omitempty"`
	CNAME string   `json:"cname,omitempty"`
	Error string   `json:"error,omitempty"`
}

func DNSLookup(ctx context.Context, host string) *DNSResult {
	res := &DNSResult{Host: host}
	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.IPs = ips

	cname, err := net.DefaultResolver.LookupCNAME(ctx, host)
	if err == nil {
		res.CNAME = cname
	}
	return res
}

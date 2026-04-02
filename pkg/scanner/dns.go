package scanner

import (
	"context"
	"net"
	"sort"
)

func LookupDNS(ctx context.Context, host string) *DNSResult {
	result := &DNSResult{Host: host}

	ips, ipErr := net.DefaultResolver.LookupHost(ctx, host)
	if ipErr == nil {
		result.IPs = ips
	}

	cname, cnameErr := net.DefaultResolver.LookupCNAME(ctx, host)
	if cnameErr == nil {
		result.CNAME = cname
	}

	mxRecords, mxErr := net.DefaultResolver.LookupMX(ctx, host)
	if mxErr == nil {
		mx := make([]string, 0, len(mxRecords))
		for _, r := range mxRecords {
			mx = append(mx, r.Host)
		}
		sort.Strings(mx)
		result.MX = mx
	}

	nsRecords, nsErr := net.DefaultResolver.LookupNS(ctx, host)
	if nsErr == nil {
		ns := make([]string, 0, len(nsRecords))
		for _, r := range nsRecords {
			ns = append(ns, r.Host)
		}
		sort.Strings(ns)
		result.NS = ns
	}

	txtRecords, txtErr := net.DefaultResolver.LookupTXT(ctx, host)
	if txtErr == nil {
		result.TXT = txtRecords
	}

	if ipErr != nil && cnameErr != nil && mxErr != nil && nsErr != nil && txtErr != nil {
		result.Error = ipErr.Error()
	}

	return result
}

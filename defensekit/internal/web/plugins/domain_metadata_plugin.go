package plugins

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"
)

type domainMetadataPlugin struct{}

func (domainMetadataPlugin) Name() string { return "domain_metadata" }

func (domainMetadataPlugin) Description() string {
	return "Read-only domain metadata (A/AAAA, CNAME, MX, NS, TXT)"
}

func (domainMetadataPlugin) Run(ctx context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := map[string]any{
		"host": req.Target,
	}

	ips, ipErr := net.DefaultResolver.LookupHost(lookupCtx, req.Target)
	if ipErr == nil {
		sort.Strings(ips)
		result["ips"] = ips
	} else {
		result["ip_error"] = ipErr.Error()
	}

	cname, cnameErr := net.DefaultResolver.LookupCNAME(lookupCtx, req.Target)
	if cnameErr == nil {
		result["cname"] = cname
	}

	mxRecords, mxErr := net.DefaultResolver.LookupMX(lookupCtx, req.Target)
	if mxErr == nil {
		mx := make([]string, 0, len(mxRecords))
		for _, record := range mxRecords {
			mx = append(mx, record.Host)
		}
		sort.Strings(mx)
		result["mx"] = mx
	}

	nsRecords, nsErr := net.DefaultResolver.LookupNS(lookupCtx, req.Target)
	if nsErr == nil {
		ns := make([]string, 0, len(nsRecords))
		for _, record := range nsRecords {
			ns = append(ns, record.Host)
		}
		sort.Strings(ns)
		result["ns"] = ns
	}

	txt, txtErr := net.DefaultResolver.LookupTXT(lookupCtx, req.Target)
	if txtErr == nil {
		result["txt"] = txt
	}

	return result, nil
}

package bodyexecution

import "fmt"

// Validate recorded references only. Hash consistency does not re-attest an
// execution, and these rows never contribute new finite correctness cases.
func validateRecordedFaultOrigins(runtime CompositionRuntime) error {
	if len(runtime.FaultSites) > 4096 {
		return fmt.Errorf("recorded arithmetic fault site limit exceeded")
	}
	sites := map[string]CompositionFaultSite{}
	for _, site := range runtime.FaultSites {
		id, unsigned := site.ExpressionID, site
		unsigned.ExpressionID = ""
		if id != compositionDigest(unsigned) || site.ProjectionSHA256 != runtime.GeneratedSHA256 ||
			!runtimeDigest(site.ProjectionSHA256) || site.Start < 0 || site.End <= site.Start || site.End-site.Start != len(site.Expression) ||
			site.RootActivityID == "" || site.ActivityID == "" || site.Operator != "/" && site.Operator != "%" ||
			site.Assignment != "" && (site.Assignment != "/=" || site.Operator != "/") || site.Variable != "" && site.Assignment == "" {
			return fmt.Errorf("recorded fault site differs from its source operation")
		}
		if _, duplicate := sites[id]; duplicate {
			return fmt.Errorf("recorded fault site identity is duplicated")
		}
		sites[id] = site
	}
	count := 0
	for _, trace := range runtime.Traces {
		seen := map[string]bool{}
		for _, delivery := range trace.Deliveries {
			if fault := delivery.Fault; fault != nil {
				count++
				site, found := sites[fault.Site.ExpressionID]
				if !found || site != fault.Site || fault.Kind != "ZERO_DIVISOR" || fault.Right != 0 ||
					site.RootActivityID != delivery.ActivityID || seen[delivery.ActivityID] || len(delivery.BlockedBy) > 0 ||
					len(delivery.Actual) != 0 && string(delivery.Actual) != "null" || delivery.Passed != nil && *delivery.Passed {
					return fmt.Errorf("recorded fault differs from its delivery or retained site")
				}
				seen[delivery.ActivityID] = true
			}
		}
	}
	if count > 65536 {
		return fmt.Errorf("recorded arithmetic fault event limit exceeded")
	}
	return nil
}

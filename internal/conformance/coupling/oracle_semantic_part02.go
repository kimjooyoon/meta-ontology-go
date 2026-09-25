package coupling

import (
	"fmt"
	"sort"
	"strings"
)

func semanticDelta(before, after []string) (string, []string, []string) {
	left, right := make(map[string]struct{}, len(before)), make(map[string]struct{}, len(after))
	for _, fact := range before {
		left[fact] = struct{}{}
	}
	for _, fact := range after {
		right[fact] = struct{}{}
	}
	added, removed := make([]string, 0), make([]string, 0)
	for fact := range right {
		if _, ok := left[fact]; !ok {
			added = append(added, fact)
		}
	}
	for fact := range left {
		if _, ok := right[fact]; !ok {
			removed = append(removed, fact)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) == 0 && len(removed) == 0 {
		return "", added, removed
	}
	var b strings.Builder
	b.WriteString("semantic-delta-v1\n")
	for _, fact := range removed {
		b.WriteString("removed\t")
		b.WriteString(fact)
		b.WriteByte('\n')
	}
	for _, fact := range added {
		b.WriteString("added\t")
		b.WriteString(fact)
		b.WriteByte('\n')
	}
	return b.String(), added, removed
}
func bindingDigest(binding CodeBinding) string {
	return digestBytes([]byte("binding-v1\n" + binding.RegisteredSurfaceID + "\n" + binding.CodeSymbolID + "\n" + binding.SemanticOwnerID + "\n" + binding.SourceMapID + "\n"))
}
func bindingCanonical(binding CodeBinding) string {
	return binding.RegisteredSurfaceID + "\t" + binding.CodeSymbolID + "\t" + binding.SemanticOwnerID + "\t" + binding.SourceMapID + "\t" + binding.BindingDigest
}
func snapshotDigest(input Input, beforeDigest, afterDigest, registryDigest string) string {
	return digestBytes([]byte("snapshot-v1\n" + sourceDigest(input.AuthoritySourceBefore) + "\n" + sourceDigest(input.AuthoritySourceAfter) + "\n" + beforeDigest + "\n" + afterDigest + "\n" + registryDigest + "\n" + input.Config.ToolchainDigest + "\n" + input.Config.Profile.ID + "\n" + input.Config.Profile.Version + "\n" + input.Config.Profile.Digest + "\n"))
}
func stateSnapshotDigest(source string, semanticDigest, registryDigest string, config EvaluationConfig) string {
	return digestBytes([]byte("state-snapshot-v1\n" + sourceDigest(source) + "\n" + semanticDigest + "\n" + registryDigest + "\n" + config.ToolchainDigest + "\n" + config.Profile.ID + "\n" + config.Profile.Version + "\n" + config.Profile.Digest + "\n"))
}
func sourceDigest(source string) string { return digestBytes([]byte(source)) }
func resourceBindingDigest(receipt ExternalResourceReceipt) string {
	return digestBytes([]byte("resource-binding-v1\n" + receipt.ReceiptID + "\n" + receipt.Metric + "\n" + fmt.Sprint(receipt.Value) + "\n" + receipt.Unit + "\n" + receipt.ProviderDigest + "\n" + receipt.ObserverDigest + "\n" + receipt.SnapshotDigest + "\n" + receipt.SourceDigest + "\n"))
}
func resourceProviderDigest(id string) string {
	return digestBytes([]byte("resource-provider-v1\n" + id + "\n"))
}
func resourceObserverDigest(id string) string {
	return digestBytes([]byte("resource-observer-v1\n" + id + "\n"))
}
func resourceSourceDigest(providerID, observerID, snapshot string) string {
	return digestBytes([]byte("resource-source-v1\n" + providerID + "\n" + observerID + "\n" + snapshot + "\n"))
}

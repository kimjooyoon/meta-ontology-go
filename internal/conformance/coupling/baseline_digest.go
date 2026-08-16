package coupling

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func baselineBindingDigest(binding CodeBinding) string {
	return baselineHash("binding-v1\n" + binding.RegisteredSurfaceID + "\n" + binding.CodeSymbolID + "\n" + binding.SemanticOwnerID + "\n" + binding.SourceMapID + "\n")
}

func baselineBindingCanonical(binding CodeBinding) string {
	return binding.RegisteredSurfaceID + "\t" + binding.CodeSymbolID + "\t" + binding.SemanticOwnerID + "\t" + binding.SourceMapID + "\t" + binding.BindingDigest
}

func baselineSnapshot(input Input, before, after string) string {
	return baselineHash("snapshot-v1\n" + baselineHash(input.AuthoritySourceBefore) + "\n" + baselineHash(input.AuthoritySourceAfter) + "\n" + before + "\n" + after + "\n" + input.RegistryDigest + "\n" + input.Config.ToolchainDigest + "\n" + input.Config.Profile.ID + "\n" + input.Config.Profile.Version + "\n" + input.Config.Profile.Digest + "\n")
}

func baselineStateSnapshot(source, semanticDigest, registry string, config EvaluationConfig) string {
	return baselineHash("state-snapshot-v1\n" + baselineHash(source) + "\n" + semanticDigest + "\n" + registry + "\n" + config.ToolchainDigest + "\n" + config.Profile.ID + "\n" + config.Profile.Version + "\n" + config.Profile.Digest + "\n")
}

func baselineResourceDigest(receipt ExternalResourceReceipt) string {
	return baselineHash("resource-binding-v1\n" + receipt.ReceiptID + "\n" + receipt.Metric + "\n" + fmt.Sprint(receipt.Value) + "\n" + receipt.Unit + "\n" + receipt.ProviderDigest + "\n" + receipt.ObserverDigest + "\n" + receipt.SnapshotDigest + "\n" + receipt.SourceDigest + "\n")
}

func baselineProviderDigest(id string) string {
	return baselineHash("resource-provider-v1\n" + id + "\n")
}

func baselineObserverDigest(id string) string {
	return baselineHash("resource-observer-v1\n" + id + "\n")
}

func baselineSourceDigest(providerID, observerID, snapshot string) string {
	return baselineHash("resource-source-v1\n" + providerID + "\n" + observerID + "\n" + snapshot + "\n")
}

func baselineHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func baselineDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func baselineID(value string) bool {
	_, err := semantic.ParseIdentity(value)
	return err == nil
}

func baselineReceiptReason(input Input, changed []string) Reason {
	if len(input.Receipts) < len(changed) {
		return ReasonMissingReceipt
	}
	return ReasonOrphanReceipt
}

func baselineUnknown(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionUnknown, reason
	result.LocalizedSurfaces = baselineChangedLabels(input)
	return result
}

func baselineFail(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionFailClosed, reason
	result.LocalizedSurfaces = baselineChangedLabels(input)
	return result
}

func baselineChangedLabels(input Input) []string {
	result := make([]string, 0)
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			result = append(result, change.CodeSymbolID)
		}
	}
	sort.Strings(result)
	return result
}

func sameSurfaceSet(left, right []string) bool {
	left, right = append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

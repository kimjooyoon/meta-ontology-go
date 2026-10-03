package completenessdelta

import (
	"encoding/hex"
	"reflect"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

const generationProfile = "gooo/body-codegen-typed-path-v1"
const runtimeProfile = bodyexecution.RuntimeProfileV1
const ownedRuntimeProfile = bodyexecution.RuntimeProfileV2
const ownedToolchainRuntimeProfile = bodyexecution.RuntimeProfileV3

func scopeRelation(before, after *completeness.CompletenessReceipt, beforeDigest string) (string, string) {
	if !supported(before.ProfileID) || !supported(after.ProfileID) {
		return "INCOMPARABLE_SCOPE", "Unregistered observation profile; counters have no shared measurement contract."
	}
	for _, receipt := range []*completeness.CompletenessReceipt{before, after} {
		if (receipt.ProfileID == ownedRuntimeProfile || receipt.ProfileID == ownedToolchainRuntimeProfile) && bodyexecution.ValidateOwnedRuntimeScope(receipt) != nil {
			return "INCOMPARABLE_SCOPE", "Missing, unsupported or inconsistent owned runtime binding."
		}
	}
	for _, path := range [][]string{
		{"domain_scope"}, {"input_type"}, {"output_type"}, {"allowed_investment"}, {"system_budget"},
		{"typed_path", "case_evaluator"}, {"typed_path", "search_config"},
		{"typed_path", "original_source_sha256"}, {"typed_path", "document_sha256"},
		{"typed_path", "test_suite_sha256"}, {"typed_path", "search_config_sha256"},
	} {
		a, b := at(before.Scope, path...), at(after.Scope, path...)
		if a == nil || b == nil || a == "" || !reflect.DeepEqual(a, b) {
			return "INCOMPARABLE_SCOPE", "Missing or changed measurement binding: /scope/" + strings.Join(path, "/")
		}
		if strings.HasSuffix(path[len(path)-1], "sha256") && !validDigest(a) {
			return "INCOMPARABLE_SCOPE", "Invalid digest in measurement binding: /scope/" + strings.Join(path, "/")
		}
	}
	if at(before.Scope, "typed_path", "case_evaluator") != "bounded_integer_go_ast" {
		return "INCOMPARABLE_SCOPE", "Unsupported finite evaluator."
	}
	if before.ProfileID == generationProfile && runtimeProfileID(after.ProfileID) &&
		after.Scope["parent_receipt_sha256"] == beforeDigest &&
		reflect.DeepEqual(before.Scope["typed_path"], after.Scope["typed_path"]) {
		return "PARENT_RUNTIME_CONTINUATION", "Exact parent bytes link a generation observation to runtime observations; changed profiles do not authorize numeric deltas."
	}
	if before.ProfileID != after.ProfileID {
		return "INCOMPARABLE_SCOPE", "Profiles differ without an exact preserved parent observation."
	}
	if !reflect.DeepEqual(before.Scope["excluded_scope"], after.Scope["excluded_scope"]) ||
		before.Scope["excluded_scope"] == nil || !reflect.DeepEqual(before.Scope["boundary"], after.Scope["boundary"]) ||
		before.Scope["boundary"] == nil {
		return "INCOMPARABLE_SCOPE", "Excluded scope or execution boundary is missing or changed."
	}
	if runtimeProfileID(before.ProfileID) {
		for _, key := range []string{"runtime_suite_sha256", "activity_id", "suite_authority"} {
			a, b := at(before.Scope, "runtime_scope", key), at(after.Scope, "runtime_scope", key)
			if a == nil || a == "" || !reflect.DeepEqual(a, b) || key == "runtime_suite_sha256" && !validDigest(a) {
				return "INCOMPARABLE_SCOPE", "Missing or changed runtime binding: " + key
			}
		}
	}
	return "SAME_MEASUREMENT_SCOPE", "Profile, original source, alternatives, finite suite, evaluator and declared investment match. Compiler/model/context differences remain observations, not causal proof."
}

func supported(profile string) bool { return profile == generationProfile || runtimeProfileID(profile) }
func runtimeProfileID(profile string) bool {
	return profile == runtimeProfile || profile == ownedRuntimeProfile || profile == ownedToolchainRuntimeProfile
}
func validDigest(v any) bool {
	s, ok := v.(string)
	if !ok || len(s) != 71 || !strings.HasPrefix(s, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(s[7:])
	return err == nil && strings.ToLower(s) == s
}
func at(m map[string]any, path ...string) any {
	var v any = m
	for _, key := range path {
		next, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = next[key]
	}
	return v
}

func changedScope(before, after map[string]any) []string {
	changes := []string{}
	var walk func(string, any, any)
	walk = func(path string, a, b any) {
		if reflect.DeepEqual(a, b) {
			return
		}
		am, aok := a.(map[string]any)
		bm, bok := b.(map[string]any)
		if !aok || !bok {
			changes = append(changes, path)
			return
		}
		keys := make([]string, 0, len(am)+len(bm))
		for k := range am {
			keys = append(keys, k)
		}
		for k := range bm {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range slices.Compact(keys) {
			pointer := strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
			av, ae := am[k]
			bv, be := bm[k]
			if ae != be {
				changes = append(changes, path+"/"+pointer)
				continue
			}
			walk(path+"/"+pointer, av, bv)
		}
	}
	walk("/scope", before, after)
	return changes
}

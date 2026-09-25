package r4safe_workid

import (
	"crypto/sha256"
	"encoding/hex"
)

func Audit(input Input) Result {
	result := Result{ExecutionAuthorized: false, EnforcementEffect: EnforcementNoEffect}
	if input.SnapshotDigest == "" || input.ObligationID == "" || input.PathID == "" || input.PolicyDigest == "" {
		result.Decision = DecisionUnknown
		result.Reason = ReasonRequiredInputMissing
		result.FullSuiteRequired = true
		result.CanonicalDigest = canonicalDigest(result)
		return result
	}

	result.LegacyWorkID = derive(input)
	switch {
	case input.CallerWorkID == "":
		result.Decision, result.Reason = DecisionPass, ReasonDerived
	case !validLegacyWorkID(input.CallerWorkID):
		result.Decision, result.Reason = DecisionFailClosed, ReasonMalformedOverride
		result.LegacyWorkID = ""
	case input.CallerWorkID != string(result.LegacyWorkID):
		result.Decision, result.Reason = DecisionFailClosed, ReasonWorkIDMismatch
		result.LegacyWorkID = ""
	default:
		result.Decision, result.Reason = DecisionPass, ReasonMatchingCallerOverride
	}
	result.CanonicalDigest = canonicalDigest(result)
	return result
}

func validLegacyWorkID(value string) bool {
	if len(value) != 64 {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func derive(input Input) LegacyWorkID {
	digest := sha256.Sum256([]byte(input.SnapshotDigest + input.ObligationID + input.PathID + input.PolicyDigest))
	return LegacyWorkID(hex.EncodeToString(digest[:]))
}

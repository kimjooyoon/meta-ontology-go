package executioncontinuation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"

func Diagnostic(e Envelope, err error) DispatchReceipt {
	r := DispatchReceipt{Schema: Schema, State: policy.ObserveStage(false, false, false), Source: e, PolicyDigest: policy.SourceDigest()}
	return failedDispatch(r, err)
}

func BoundSource(e Envelope) DispatchReceipt {
	r := DispatchReceipt{Schema: Schema, State: "SOURCE_BOUND", Source: e, PolicyDigest: policy.SourceDigest()}
	return sealDispatch(r, "EXACT_GRANT_CONTRACT_SOURCE_BOUND")
}

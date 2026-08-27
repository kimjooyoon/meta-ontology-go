package authorization

type metricSpec struct {
	ID        string
	Class     string
	Choice    string
	Stage     string
	Operation string
	Claim     string
}

var metricSpecs = []metricSpec{
	{"execution-report", "DRIVER", "FOUNDATION", "AUTHORIZE/execution-report", "bind-execution-report", "the sealed execution report is exact"},
	{"operation", "OUTCOME", "COHERENCE", "AUTHORIZE/operation", "bind-requested-operation", "the requested operation is pinned"},
	{"subject", "DRIVER", "FOUNDATION", "AUTHORIZE/subject", "bind-subject", "the request targets the exact CI subject"},
	{"issuer", "DRIVER", "FOUNDATION", "AUTHORIZE/issuer", "bind-issuer", "the request issuer is pinned"},
	{"scope", "OUTCOME", "COHERENCE", "AUTHORIZE/scope", "bind-capability-scope", "the capability scope is exact"},
	{"policy-foundation", "DRIVER", "FOUNDATION", "AUTHORIZE/policy-foundation", "bind-policy-foundation", "the compiled Gooo policy has an immutable foundation"},
	{"default-deny", "GUARDRAIL", "REGRESSION", "AUTHORIZE/default-deny", "enforce-default-deny", "the fallback decision is deny"},
	{"invocation", "OUTCOME", "COHERENCE", "AUTHORIZE/invocation", "bind-run-attempt", "the request is bound to this run attempt"},
	{"nonce", "GUARDRAIL", "REGRESSION", "AUTHORIZE/nonce", "bind-envelope-nonce", "the envelope nonce binds its coordinates"},
	{"effect-ceiling", "GUARDRAIL", "REGRESSION", "AUTHORIZE/effect-ceiling", "enforce-zero-effect-ceiling", "the request grants no write or promotion authority"},
}

var nonClaims = []string{
	"cryptographic capability delegation is not implemented",
	"cross-run nonce reuse prevention is not implemented",
	"repository mutation and promotion authority are not granted",
	"AUTHORIZED_SHADOW is not production enforcement",
}

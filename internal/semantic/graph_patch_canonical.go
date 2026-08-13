package semantic

import "encoding/json"

type graphPatchCanonicalBase struct {
	SourceDigest string `json:"source_digest"`
	IRDigest     string `json:"ir_digest"`
}

type graphPatchCanonicalRequest struct {
	SchemaVersion        string `json:"schema_version"`
	Operation            string `json:"operation"`
	ExpectedGraphHash    string `json:"expected_graph_hash"`
	NodeID               string `json:"node_id"`
	ExpectedNodeHash     string `json:"expected_node_hash"`
	Field                string `json:"field"`
	ExpectedFieldHash    string `json:"expected_field_hash"`
	Subject              string `json:"subject"`
	Predicate            string `json:"predicate"`
	Object               string `json:"object"`
	ExpectedSourceDigest string `json:"expected_source_digest"`
	ExpectedIRDigest     string `json:"expected_ir_digest"`
	AllowedIntent        string `json:"allowed_intent"`
	Locality             string `json:"locality"`
}

type graphPatchCanonicalBinding struct {
	Base    graphPatchCanonicalBase    `json:"base"`
	Request graphPatchCanonicalRequest `json:"request"`
}

// Canonical returns deterministic JSON for the trusted source/IR base tuple.
func (b GraphPatchBase) Canonical() string {
	return marshalGraphPatch(graphPatchCanonicalBase{SourceDigest: b.SourceDigest, IRDigest: b.IRDigest})
}

func (b GraphPatchBase) StableHash() string { return StableHashString(b.Canonical()) }

// Canonical returns deterministic JSON for all graph edit preconditions.
func (r GraphPatchRequest) Canonical() string {
	return marshalGraphPatch(canonicalPatchRequest(r))
}

func (r GraphPatchRequest) StableHash() string { return StableHashString(r.Canonical()) }

// CanonicalGraphPatchBinding binds base and request in one digestable payload.
func CanonicalGraphPatchBinding(base GraphPatchBase, request GraphPatchRequest) string {
	return marshalGraphPatch(graphPatchCanonicalBinding{
		Base:    graphPatchCanonicalBase{SourceDigest: base.SourceDigest, IRDigest: base.IRDigest},
		Request: canonicalPatchRequest(request),
	})
}

func GraphPatchBindingHash(base GraphPatchBase, request GraphPatchRequest) string {
	return StableHashString(CanonicalGraphPatchBinding(base, request))
}

func canonicalPatchRequest(r GraphPatchRequest) graphPatchCanonicalRequest {
	return graphPatchCanonicalRequest{
		SchemaVersion: r.SchemaVersion, Operation: r.Operation, ExpectedGraphHash: r.ExpectedGraphHash,
		NodeID: r.NodeID.String(), ExpectedNodeHash: r.ExpectedNodeHash, Field: r.Field,
		ExpectedFieldHash: r.ExpectedFieldHash, Subject: r.Subject.String(), Predicate: r.Predicate.String(),
		Object: r.Object.String(), ExpectedSourceDigest: r.ExpectedSourceDigest, ExpectedIRDigest: r.ExpectedIRDigest,
		AllowedIntent: r.AllowedIntent, Locality: r.Locality,
	}
}

func marshalGraphPatch(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic("semantic graph patch canonicalization failed: " + err.Error())
	}
	return string(encoded)
}

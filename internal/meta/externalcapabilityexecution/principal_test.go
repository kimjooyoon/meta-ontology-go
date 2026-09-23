package externalcapabilityexecution

import "testing"

const (
	principalSourceDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	principalNonce        = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestPrincipalBindingIsStableSourceBoundIdentityEvidence(t *testing.T) {
	first, err := BindPrincipal(
		principalSourceDigest,
		"spiffe://ci.example.org",
		"spiffe://ci.example.org/workload/gooo",
		"gooo://external-capability/execute",
		principalNonce,
	)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BindPrincipal(
		principalSourceDigest,
		"spiffe://ci.example.org",
		"spiffe://ci.example.org/workload/gooo",
		"gooo://external-capability/execute",
		principalNonce,
	)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || first.IdentityDigest == "" {
		t.Fatalf("principal binding is not deterministic: %#v %#v", first, second)
	}
	if err := first.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPrincipalBindingRejectsIdentityTamperingAndMalformedScope(t *testing.T) {
	binding, err := BindPrincipal(
		principalSourceDigest,
		"spiffe://ci.example.org",
		"spiffe://ci.example.org/workload/gooo",
		"gooo://external-capability/execute",
		principalNonce,
	)
	if err != nil {
		t.Fatal(err)
	}
	binding.Subject = "spiffe://ci.example.org/workload/other"
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered subject was accepted")
	}
	if _, err := BindPrincipal(principalSourceDigest, "https://ci.example.org", "spiffe://ci.example.org/workload/gooo", "gooo://external-capability/execute", principalNonce); err == nil {
		t.Fatal("non-SPIFFE issuer was accepted")
	}
	if _, err := BindPrincipal(principalSourceDigest, "spiffe://ci.example.org", "spiffe://ci.example.org/workload/gooo", "gooo://external-capability/execute?write=true", principalNonce); err == nil {
		t.Fatal("scope query was accepted as an unbound capability")
	}
}

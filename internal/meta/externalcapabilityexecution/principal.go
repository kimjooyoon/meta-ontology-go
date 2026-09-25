package externalcapabilityexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	PrincipalBindingSchema = "gooo/external-capability-principal-binding/v1"
	PrincipalSourcePath    = "examples/external-capability-execution/authorization/policy.gooo"
)

// PrincipalBinding is a provenance-bound identity statement. It is evidence
// about who requested a capability and which scope was named, not an execution
// grant. Keeping authority out of this value prevents identity observation
// from becoming an implicit permission escalation.
type PrincipalBinding struct {
	Schema         string `json:"schema"`
	SourcePath     string `json:"source_path"`
	SourceDigest   string `json:"source_digest"`
	Issuer         string `json:"issuer"`
	Subject        string `json:"subject"`
	Scope          string `json:"scope"`
	Nonce          string `json:"nonce"`
	IdentityDigest string `json:"identity_digest"`
}

type principalDigestInput struct {
	Schema       string `json:"schema"`
	SourcePath   string `json:"source_path"`
	SourceDigest string `json:"source_digest"`
	Issuer       string `json:"issuer"`
	Subject      string `json:"subject"`
	Scope        string `json:"scope"`
	Nonce        string `json:"nonce"`
}

// BindPrincipal creates a deterministic identity evidence record from the
// language policy source and request context. It deliberately has no grant or
// execution fields; callers must still pass the binding through authorization.
func BindPrincipal(sourceDigest, issuer, subject, scope, nonce string) (PrincipalBinding, error) {
	binding := PrincipalBinding{
		Schema:       PrincipalBindingSchema,
		SourcePath:   PrincipalSourcePath,
		SourceDigest: sourceDigest,
		Issuer:       issuer,
		Subject:      subject,
		Scope:        scope,
		Nonce:        nonce,
	}
	if err := binding.validateFields(); err != nil {
		return PrincipalBinding{}, err
	}
	binding.IdentityDigest = digestPrincipal(binding)
	return binding, nil
}

// Validate checks both the identity syntax and the exact source-bound digest.
func (binding PrincipalBinding) Validate() error {
	if err := binding.validateFields(); err != nil {
		return err
	}
	if binding.IdentityDigest == "" || binding.IdentityDigest != digestPrincipal(binding) {
		return errors.New("principal identity digest does not bind the observed fields")
	}
	return nil
}

func (binding PrincipalBinding) validateFields() error {
	if binding.Schema != PrincipalBindingSchema {
		return fmt.Errorf("principal schema %q is not %q", binding.Schema, PrincipalBindingSchema)
	}
	if binding.SourcePath != PrincipalSourcePath {
		return fmt.Errorf("principal source path %q is not the authorization policy", binding.SourcePath)
	}
	if !validPrincipalDigest(binding.SourceDigest) {
		return errors.New("principal source digest is not a sha256 digest")
	}
	if err := validateSPIFFE(binding.Issuer, false); err != nil {
		return fmt.Errorf("issuer: %w", err)
	}
	if err := validateSPIFFE(binding.Subject, true); err != nil {
		return fmt.Errorf("subject: %w", err)
	}
	if err := validateScope(binding.Scope); err != nil {
		return fmt.Errorf("scope: %w", err)
	}
	if !validPrincipalDigest(binding.Nonce) {
		return errors.New("principal nonce is not a sha256 digest")
	}
	return nil
}

func validateSPIFFE(value string, subject bool) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "spiffe" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Host != strings.ToLower(parsed.Host) {
		return errors.New("must be a canonical spiffe URI")
	}
	if subject && (parsed.Path == "" || parsed.Path == "/") {
		return errors.New("subject must identify a workload path")
	}
	if !subject && parsed.Path != "" {
		return errors.New("issuer must identify a trust domain without a path")
	}
	return nil
}

func validateScope(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "gooo" || parsed.Host == "" || parsed.Path == "" || parsed.Path == "/" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Host != strings.ToLower(parsed.Host) {
		return errors.New("must be a canonical gooo capability URI")
	}
	return nil
}

func validPrincipalDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func digestPrincipal(binding PrincipalBinding) string {
	payload := principalDigestInput{
		Schema: binding.Schema, SourcePath: binding.SourcePath, SourceDigest: binding.SourceDigest,
		Issuer: binding.Issuer, Subject: binding.Subject, Scope: binding.Scope, Nonce: binding.Nonce,
	}
	encoded, _ := json.Marshal(payload)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

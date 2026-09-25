package coupling

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func inputToWire(input Input, canonical bool) wireInput {
	raw := wireInput{
		Schema: input.Schema, FixtureID: input.FixtureID, RegistryDigest: input.RegistryDigest,
		Config: wireConfig{ToolchainDigest: input.Config.ToolchainDigest, Profile: wireProfile{
			ID: input.Config.Profile.ID, Version: input.Config.Profile.Version, Digest: input.Config.Profile.Digest,
		}, ResourceBinding: input.Config.ResourceBinding}, Manifest: input.Manifest, ResourceRegistry: input.ResourceRegistry,
		AuthoritySourceBefore: input.AuthoritySourceBefore, AuthoritySourceAfter: input.AuthoritySourceAfter,
		SemanticBefore: cloneSemanticIR(input.SemanticBefore), SemanticAfter: cloneSemanticIR(input.SemanticAfter),
		Registry: append([]CodeBinding(nil), input.Registry...), Changes: append([]CodeChange(nil), input.Changes...),
		Receipts: append([]CouplingReceipt(nil), input.Receipts...), Roots: append([]string(nil), input.Roots...),
		ResourceReceipts: append([]ExternalResourceReceipt(nil), input.ResourceReceipts...),
		Path:             pathToWire(input.Path),
	}
	if canonical {
		raw.FixtureID = ""
		for i := range raw.SemanticBefore.Nodes {
			raw.SemanticBefore.Nodes[i].Name = ""
			raw.SemanticBefore.Nodes[i].Aliases = nil
		}
		for i := range raw.SemanticAfter.Nodes {
			raw.SemanticAfter.Nodes[i].Name = ""
			raw.SemanticAfter.Nodes[i].Aliases = nil
		}
		for i := range raw.Registry {
			raw.Registry[i].PackageLabel = ""
			raw.Registry[i].FileLabel = ""
			raw.Registry[i].SourceSpan = ""
		}
		normalizeWireInput(&raw)
	}
	return raw
}
func normalizeWireInput(raw *wireInput) {
	sort.Slice(raw.Registry, func(i, j int) bool {
		return raw.Registry[i].RegisteredSurfaceID+"\x00"+raw.Registry[i].CodeSymbolID <
			raw.Registry[j].RegisteredSurfaceID+"\x00"+raw.Registry[j].CodeSymbolID
	})
	sort.Slice(raw.Changes, func(i, j int) bool { return raw.Changes[i].CodeSymbolID < raw.Changes[j].CodeSymbolID })
	sort.Slice(raw.Receipts, func(i, j int) bool {
		return raw.Receipts[i].SurfaceID+"\x00"+raw.Receipts[i].ReceiptID < raw.Receipts[j].SurfaceID+"\x00"+raw.Receipts[j].ReceiptID
	})
	sort.Slice(raw.ResourceReceipts, func(i, j int) bool {
		return raw.ResourceReceipts[i].Metric+"\x00"+raw.ResourceReceipts[i].ReceiptID < raw.ResourceReceipts[j].Metric+"\x00"+raw.ResourceReceipts[j].ReceiptID
	})
	sort.Strings(raw.Roots)
	normalizeSemanticIR(&raw.SemanticBefore)
	normalizeSemanticIR(&raw.SemanticAfter)
	sort.Slice(raw.Path.Edges, func(i, j int) bool { return raw.Path.Edges[i].RecordID < raw.Path.Edges[j].RecordID })
	sort.Slice(raw.Path.Claims, func(i, j int) bool { return raw.Path.Claims[i].RecordID < raw.Path.Claims[j].RecordID })
	sort.Slice(raw.Path.Evidence, func(i, j int) bool { return raw.Path.Evidence[i].ID < raw.Path.Evidence[j].ID })
	for i := range raw.Path.Edges {
		sort.Strings(raw.Path.Edges[i].SourceRoots)
		sort.Slice(raw.Path.Edges[i].Evidence, func(a, b int) bool { return raw.Path.Edges[i].Evidence[a].ID < raw.Path.Edges[i].Evidence[b].ID })
	}
	for i := range raw.Path.Claims {
		sort.Slice(raw.Path.Claims[i].Evidence, func(a, b int) bool { return raw.Path.Claims[i].Evidence[a].ID < raw.Path.Claims[i].Evidence[b].ID })
	}
}

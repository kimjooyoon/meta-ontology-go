package languageproofartifactverifier

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

const BundleSchema = "gooo/proof-carrying-artifact-bundle/v1"

var requiredBundlePaths = []string{
	"artifact.json", "tampered.json", "coherent-tamper.json", "missing-operation.json", "byte-only.json", "wrong-recipe.json",
	"recipe-only.json", "missing-attachment.json", "wrong-attachment-digest.json", "unrelated-tamper.json", "stale-head.json", "unauthorized-consumer.json",
	"claim-proposition-tamper.json", "claim-dependency-tamper.json", "claim-proof-choice-tamper.json", "claim-target-tamper.json",
	"source.gooo", "operation-receipt.json", "recipe.json", "contract.json", "independence.json", "write-set.json", "coherent-operation-receipt.json", "checkout.json",
	"semantic-intervention-artifact.json", "semantic-intervention.gooo", "semantic-operation-receipt.json", "comment-only-intervention-artifact.json", "comment-only-intervention.gooo", "comment-operation-receipt.json",
}

func RequiredBundlePaths() []string { return append([]string(nil), requiredBundlePaths...) }

func PackBundle(head string, checkout CheckoutEvidence, inputs []BundleInput) (Bundle, error) {
	if !validHead(head) || checkout.HeadSHA != head || checkout.ActualHeadSHA != head {
		return Bundle{}, fmt.Errorf("bundle checkout identity mismatch")
	}
	if len(inputs) != len(requiredBundlePaths) {
		return Bundle{}, fmt.Errorf("bundle input count mismatch")
	}
	wanted := map[string]bool{}
	for _, path := range requiredBundlePaths {
		wanted[path] = true
	}
	files := make([]BundleFile, 0, len(inputs))
	manifest := make([]BundleManifestEntry, 0, len(inputs))
	seen := map[string]bool{}
	for _, input := range inputs {
		if !wanted[input.Path] || seen[input.Path] || input.File == "" {
			return Bundle{}, fmt.Errorf("bundle path is missing, extra, or duplicated: %s", input.Path)
		}
		seen[input.Path] = true
		raw, err := os.ReadFile(input.File)
		if err != nil {
			return Bundle{}, fmt.Errorf("read bundle input %s: %w", input.Path, err)
		}
		digest := digestBytes(raw)
		files = append(files, BundleFile{Path: input.Path, Digest: digest, Content: base64.StdEncoding.EncodeToString(raw)})
		manifest = append(manifest, BundleManifestEntry{Path: input.Path, Digest: digest, Size: len(raw), Role: input.Role})
	}
	for path := range wanted {
		if !seen[path] {
			return Bundle{}, fmt.Errorf("bundle path missing: %s", path)
		}
	}
	sortBundleFiles(files, manifest)
	bundle := Bundle{Schema: BundleSchema, Version: 1, HeadSHA: head, Checkout: checkout, Manifest: manifest, Files: files}
	bundle.Digest = bundleDigest(bundle)
	return bundle, nil
}

func DecodeBundle(raw []byte) (Bundle, error) {
	bundle, err := decodeStrict[Bundle](raw)
	if err != nil {
		return Bundle{}, err
	}
	if err := ValidateBundle(bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func ValidateBundle(bundle Bundle) error {
	if bundle.Schema != BundleSchema || bundle.Version != 1 || !validHead(bundle.HeadSHA) || len(bundle.Manifest) != len(requiredBundlePaths) || len(bundle.Files) != len(requiredBundlePaths) || bundle.Digest != bundleDigest(bundle) {
		return fmt.Errorf("bundle identity mismatch")
	}
	wanted := map[string]bool{}
	for _, path := range requiredBundlePaths {
		wanted[path] = true
	}
	manifest := map[string]BundleManifestEntry{}
	files := map[string]BundleFile{}
	for _, item := range bundle.Manifest {
		if !wanted[item.Path] || manifest[item.Path].Path != "" || !validDigest(item.Digest) || item.Size < 0 {
			return fmt.Errorf("bundle manifest path is extra or duplicated: %s", item.Path)
		}
		manifest[item.Path] = item
	}
	for _, item := range bundle.Files {
		if !wanted[item.Path] || files[item.Path].Path != "" || !validDigest(item.Digest) {
			return fmt.Errorf("bundle file path is extra or duplicated: %s", item.Path)
		}
		raw, err := base64.StdEncoding.DecodeString(item.Content)
		if err != nil || digestBytes(raw) != item.Digest || manifest[item.Path].Digest != item.Digest || manifest[item.Path].Size != len(raw) {
			return fmt.Errorf("bundle file digest mismatch: %s", item.Path)
		}
		files[item.Path] = item
	}
	if len(manifest) != len(wanted) || len(files) != len(wanted) {
		return fmt.Errorf("bundle path inventory mismatch")
	}
	if err := validateCheckout(bundle.Checkout, bundle.HeadSHA, files); err != nil {
		return err
	}
	return nil
}

func validateCheckout(checkout CheckoutEvidence, head string, files map[string]BundleFile) error {
	if checkout.HeadSHA != head || checkout.ActualHeadSHA != head || !validDigest(checkout.TreeDigest) || !validDigest(checkout.SourceDigest) || !validDigest(checkout.OperationDigest) || !validDigest(checkout.RecipeDigest) || !validDigest(checkout.ContractDigest) {
		return fmt.Errorf("bundle checkout evidence mismatch")
	}
	checkoutRaw, err := base64.StdEncoding.DecodeString(files["checkout.json"].Content)
	if err != nil {
		return err
	}
	var fromFile CheckoutEvidence
	if err := json.Unmarshal(checkoutRaw, &fromFile); err != nil || !reflect.DeepEqual(fromFile, checkout) {
		return fmt.Errorf("bundle checkout attachment mismatch")
	}
	if files["source.gooo"].Digest != checkout.SourceDigest || files["operation-receipt.json"].Digest != checkout.OperationDigest || files["recipe.json"].Digest != checkout.RecipeDigest || files["contract.json"].Digest != checkout.ContractDigest {
		return fmt.Errorf("bundle source or recipe binding mismatch")
	}
	return nil
}

func sortBundleFiles(files []BundleFile, manifest []BundleManifestEntry) {
	sort.Slice(files, func(left, right int) bool { return files[left].Path < files[right].Path })
	sort.Slice(manifest, func(left, right int) bool { return manifest[left].Path < manifest[right].Path })
}

func bundleDigest(bundle Bundle) string {
	bundle.Digest = ""
	return digestValue(bundle)
}

func InputFromBundle(bundle Bundle) (Input, error) {
	if err := ValidateBundle(bundle); err != nil {
		return Input{}, err
	}
	content := map[string][]byte{}
	for _, item := range bundle.Files {
		raw, err := base64.StdEncoding.DecodeString(item.Content)
		if err != nil {
			return Input{}, err
		}
		content[item.Path] = raw
	}
	contract, err := DecodeContract(content["contract.json"])
	if err != nil {
		return Input{}, err
	}
	var independence IndependenceEvidence
	if err := json.Unmarshal(content["independence.json"], &independence); err != nil {
		return Input{}, err
	}
	writeSet, err := DecodeWriteSet(content["write-set.json"])
	if err != nil {
		return Input{}, err
	}
	if _, err := decodeStrict[Recipe](content["recipe.json"]); err != nil {
		return Input{}, err
	}
	interventions := []InterventionInput{
		{ID: "semantic-source-intervention", Kind: "SEMANTIC", Before: SubjectInput{Artifact: content["artifact.json"], Source: content["source.gooo"], Operation: content["operation-receipt.json"], Recipe: content["recipe.json"], Checkout: bundle.Checkout}, After: SubjectInput{Artifact: content["semantic-intervention-artifact.json"], Source: content["semantic-intervention.gooo"], Operation: content["semantic-operation-receipt.json"], Recipe: content["recipe.json"], Checkout: bundle.Checkout}},
		{ID: "comment-only-intervention", Kind: "NONSEMANTIC", Before: SubjectInput{Artifact: content["artifact.json"], Source: content["source.gooo"], Operation: content["operation-receipt.json"], Recipe: content["recipe.json"], Checkout: bundle.Checkout}, After: SubjectInput{Artifact: content["comment-only-intervention-artifact.json"], Source: content["comment-only-intervention.gooo"], Operation: content["comment-operation-receipt.json"], Recipe: content["recipe.json"], Checkout: bundle.Checkout}},
	}
	return Input{Contract: contract, ContractBytes: content["contract.json"], HeadSHA: bundle.HeadSHA, ValidArtifact: content["artifact.json"], TamperedArtifact: content["tampered.json"], CoherentTamperedArtifact: content["coherent-tamper.json"], MissingArtifact: content["missing-operation.json"], ByteOnlyArtifact: content["byte-only.json"], WrongRecipe: content["wrong-recipe.json"], RecipeOnlyArtifact: content["recipe-only.json"], MissingAttachment: content["missing-attachment.json"], WrongAttachmentDigest: content["wrong-attachment-digest.json"], UnrelatedTamperedArtifact: content["unrelated-tamper.json"], StaleHeadArtifact: content["stale-head.json"], ClaimPropositionArtifact: content["claim-proposition-tamper.json"], ClaimDependencyArtifact: content["claim-dependency-tamper.json"], ClaimProofChoiceArtifact: content["claim-proof-choice-tamper.json"], ClaimTargetArtifact: content["claim-target-tamper.json"], UnauthorizedConsumer: content["unauthorized-consumer.json"], UnauthorizedBundle: bundle, Source: content["source.gooo"], Operation: content["operation-receipt.json"], Recipe: content["recipe.json"], Independence: independence, WriteSet: writeSet, CoherentOperation: content["coherent-operation-receipt.json"], Interventions: interventions, Checkout: bundle.Checkout, BundleDigest: bundle.Digest}, nil
}

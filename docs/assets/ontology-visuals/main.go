package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

const (
	width      = 960
	height     = 540
	frameCount = 36
	frameDelay = 12
	defaultDir = "docs/assets/ontology-visuals"
)

type binding struct {
	Nodes       []string `json:"nodes"`
	Edges       []string `json:"edges"`
	Transitions []string `json:"transitions"`
}

type evidenceRef struct {
	Path          string `json:"path"`
	Digest        string `json:"digest"`
	ReceiptPath   string `json:"receipt_path"`
	ReceiptDigest string `json:"receipt_digest"`
}

type manifestScene struct {
	ID                string        `json:"id"`
	Sequence          int           `json:"sequence"`
	File              string        `json:"file"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Alt               string        `json:"alt"`
	SourceRefs        []string      `json:"source_refs"`
	Binding           binding       `json:"binding"`
	InitialRelation   string        `json:"initial_relation"`
	SemanticOperation string        `json:"semantic_operation"`
	TerminalRelation  string        `json:"terminal_relation"`
	MotionEventKind   string        `json:"motion_event_kind"`
	Maturity          string        `json:"maturity"`
	InputArtifact     string        `json:"input_artifact"`
	GeneratedArtifact string        `json:"generated_or_changed_artifact"`
	TerminalDecision  string        `json:"terminal_decision"`
	NextConsumer      string        `json:"next_consumer"`
	Evidence          []evidenceRef `json:"evidence"`
}

type manifest struct {
	Schema            string          `json:"schema"`
	VisualDenominator int             `json:"visual_denominator"`
	Renderer          string          `json:"renderer"`
	Scenes            []manifestScene `json:"scenes"`
}

type assetRecord struct {
	ID        string `json:"id"`
	Sequence  int    `json:"sequence"`
	File      string `json:"file"`
	SizeBytes int    `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type assetLock struct {
	Schema            string        `json:"schema"`
	VisualDenominator int           `json:"visual_denominator"`
	VisualAssetCount  int           `json:"visual_asset_count"`
	VisualAssetBytes  int           `json:"visual_asset_bytes"`
	Assets            []assetRecord `json:"assets"`
}

var sceneIDs = []string{
	"ontology.visual.intent-ir-lowering",
	"ontology.visual.authority-boundary",
	"ontology.visual.munchausen-proof-choice",
	"ontology.visual.claim-evidence-lifecycle",
	"ontology.visual.unknown-cause-descent",
	"ontology.visual.precedence-counterexample",
	"ontology.visual.package-resolution",
	"ontology.visual.incremental-conformance",
	"ontology.visual.bootstrap-oracle",
	"ontology.visual.promotion-lineage",
}

var expectedBindings = map[string]binding{
	"ontology.visual.intent-ir-lowering":        {Nodes: []string{"gooo.intent", "semantic.ir", "generated.go", "evidence"}, Edges: []string{"lower", "normalize", "project", "record"}, Transitions: []string{"source", "ir", "structural-view", "proof"}},
	"ontology.visual.authority-boundary":        {Nodes: []string{"declaration", "semantic.ir", "typed.nodes", "backend"}, Edges: []string{"parse", "used", "wasGeneratedBy", "consume"}, Transitions: []string{"text", "node", "edge", "graph"}},
	"ontology.visual.munchausen-proof-choice":   {Nodes: []string{"activity.gooo", "entities.gooo", "canonical-order", "package-api", "entry.receipt"}, Edges: []string{"sort", "parse", "merge", "resolve", "execute"}, Transitions: []string{"unordered", "canonical", "bound", "entry", "receipt"}},
	"ontology.visual.claim-evidence-lifecycle":  {Nodes: []string{"author.agent", "intent", "compiler.agent", "generated.output", "reviewer.agent", "receipt"}, Edges: []string{"writes", "reads", "creates", "consumes", "authority-boundary"}, Transitions: []string{"intent", "handoff", "generated", "reviewed"}},
	"ontology.visual.unknown-cause-descent":     {Nodes: []string{"stage", "step", "reason", "class", "next_operation", "blocked_by"}, Edges: []string{"cause-descent", "classify", "repair", "re-evaluate"}, Transitions: []string{"unknown", "actionable", "evidence", "closed"}},
	"ontology.visual.precedence-counterexample": {Nodes: []string{"UNKNOWN", "counterexample", "REFUTED", "priority", "ledger"}, Edges: []string{"arrives", "outranks", "preserves", "appends"}, Transitions: []string{"partial", "contradiction", "refuted", "retained"}},
	"ontology.visual.package-resolution":        {Nodes: []string{"source.digest", "contract.digest", "toolchain.digest", "run.1", "run.2", "mismatch"}, Edges: []string{"pin", "replay", "compare", "block"}, Transitions: []string{"inputs", "identical", "changed-byte", "blocked"}},
	"ontology.visual.incremental-conformance":   {Nodes: []string{"changed-surface", "six-digest-identity", "EXECUTE", "REUSE", "UNKNOWN", "REFUTED"}, Edges: []string{"identify", "route", "reuse-receipt", "close-alternatives"}, Transitions: []string{"changed", "identified", "reused", "inactive"}},
	"ontology.visual.bootstrap-oracle":          {Nodes: []string{"observation", "meta-rule", "exact-head-ci", "dev", "post-adoption", "main-eligibility"}, Edges: []string{"observe", "change", "verify", "adopt", "promote"}, Transitions: []string{"bug", "rule", "ci", "dev", "receipt", "eligible"}},
	"ontology.visual.promotion-lineage":         {Nodes: []string{"service.contract", "infra.facts", "proposed.gooo", "mismatch.dossier", "UNKNOWN"}, Edges: []string{"connect", "compare", "report", "hold"}, Transitions: []string{"experimental", "projected", "mismatch", "unknown"}},
}

type motionContract struct {
	Initial, Operation, Terminal, Event string
}

var expectedMotion = map[string]motionContract{
	"ontology.visual.intent-ir-lowering":        {"source.gooo", "parse-lower-emit", "pass.receipt", "artifact-materialization"},
	"ontology.visual.authority-boundary":        {"declaration.text", "type-and-relate-semantic-ir", "typed.graph", "node-edge-materialization"},
	"ontology.visual.munchausen-proof-choice":   {"unordered.files", "canonicalize-package", "entry.receipt", "file-reorder-and-api-emission"},
	"ontology.visual.claim-evidence-lifecycle":  {"agent.intent", "handoff-by-receipt", "review.receipt", "lane-artifact-transfer"},
	"ontology.visual.unknown-cause-descent":     {"missing.evidence", "resolve-and-re-evaluate", "claim.closed", "unknown-fields-then-evidence"},
	"ontology.visual.precedence-counterexample": {"unknown.claim", "rank-counterexample", "refuted.ledger", "counterexample-precedence"},
	"ontology.visual.package-resolution":        {"pinned.digests", "replay-and-compare", "adoption.blocked", "byte-change-detection"},
	"ontology.visual.incremental-conformance":   {"changed.subject", "route-by-six-digests", "reuse.receipt", "identity-router"},
	"ontology.visual.bootstrap-oracle":          {"observed.bug", "change-rule-and-gate", "main.eligible", "gated-receipt-cascade"},
	"ontology.visual.promotion-lineage":         {"service-and-infra.contract", "project-and-diff", "unknown.dossier", "proposed-mismatch-dossier"},
}

func main() {
	outDir := flag.String("out-dir", defaultDir, "directory for generated visual assets")
	manifestPath := flag.String("manifest", filepath.Join(defaultDir, "visual-manifest.json"), "visual source manifest")
	check := flag.Bool("check", false, "verify generated assets, README references, and bindings")
	previewDir := flag.String("preview-dir", "", "optional directory for representative PNG frames")
	flag.Parse()

	m, err := loadManifest(*manifestPath)
	if err == nil {
		err = validateManifest(m)
	}
	if err != nil {
		fatal(err)
	}
	if *check {
		if err := checkAssets(*outDir, m); err != nil {
			fatal(err)
		}
		return
	}
	if err := writeAssets(*outDir, m); err != nil {
		fatal(err)
	}
	if *previewDir != "" {
		if err := writePreviews(*previewDir, m); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read visual manifest: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, fmt.Errorf("decode visual manifest: %w", err)
	}
	return m, nil
}

func validateManifest(m manifest) error {
	if m.Schema != "gooo/visual-story-manifest/v2" || m.VisualDenominator != 10 || len(m.Scenes) != 10 {
		return fmt.Errorf("visual denominator must be exactly 10")
	}
	if m.Renderer != "docs/assets/ontology-visuals" {
		return fmt.Errorf("visual renderer binding is not exact")
	}
	for i, scene := range m.Scenes {
		if scene.Sequence != i+1 || scene.ID != sceneIDs[i] || scene.File == "" || scene.Title == "" || scene.Description == "" || scene.Alt == "" {
			return fmt.Errorf("manifest scene %d is incomplete or out of order", i+1)
		}
		want, ok := expectedBindings[scene.ID]
		if !ok || !sameBinding(scene.Binding, want) || len(scene.SourceRefs) == 0 {
			return fmt.Errorf("semantic scene binding is not exact for %s", scene.ID)
		}
		if scene.Maturity != "CURRENT" && scene.Maturity != "PROPOSED" || scene.InputArtifact == "" || scene.SemanticOperation == "" || scene.GeneratedArtifact == "" || scene.TerminalDecision == "" || scene.NextConsumer == "" || len(scene.Evidence) == 0 {
			return fmt.Errorf("engineering narrative row is incomplete for %s", scene.ID)
		}
		for _, evidence := range scene.Evidence {
			if evidence.Path == "" || !strings.HasPrefix(evidence.Digest, "sha256:") {
				return fmt.Errorf("evidence path/digest is incomplete for %s", scene.ID)
			}
			if scene.Maturity == "CURRENT" && (evidence.ReceiptPath == "" || !strings.HasPrefix(evidence.ReceiptDigest, "sha256:")) {
				return fmt.Errorf("current evidence receipt is incomplete for %s", scene.ID)
			}
			if err := validateEvidenceDigest(evidence.Path, evidence.Digest); err != nil {
				return fmt.Errorf("evidence binding is not resolved for %s: %w", scene.ID, err)
			}
			if evidence.ReceiptPath != "" {
				if err := validateEvidenceDigest(evidence.ReceiptPath, evidence.ReceiptDigest); err != nil {
					return fmt.Errorf("receipt binding is not resolved for %s: %w", scene.ID, err)
				}
			}
		}
		motion, ok := expectedMotion[scene.ID]
		if !ok || scene.InitialRelation != motion.Initial || scene.SemanticOperation != motion.Operation || scene.TerminalRelation != motion.Terminal || scene.MotionEventKind != motion.Event {
			return fmt.Errorf("motion relation binding is not exact for %s", scene.ID)
		}
		if _, ok := renderers[scene.ID]; !ok {
			return fmt.Errorf("scene %s has no renderer", scene.ID)
		}
	}
	return nil
}

func validateEvidenceDigest(path, expected string) error {
	data, err := os.ReadFile(filepath.Join(repositoryRoot(), path))
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	digest := sha256.Sum256(data)
	actual := "sha256:" + hex.EncodeToString(digest[:])
	if actual != expected {
		return fmt.Errorf("digest mismatch for %s", path)
	}
	return nil
}

func sameBinding(a, b binding) bool {
	return strings.Join(a.Nodes, "\x00") == strings.Join(b.Nodes, "\x00") &&
		strings.Join(a.Edges, "\x00") == strings.Join(b.Edges, "\x00") &&
		strings.Join(a.Transitions, "\x00") == strings.Join(b.Transitions, "\x00")
}

func writeAssets(dir string, m manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create visual directory: %w", err)
	}
	assets, err := encodedAssets(m)
	if err != nil {
		return err
	}
	lock := makeLock(m, assets)
	for _, scene := range m.Scenes {
		if err := os.WriteFile(filepath.Join(dir, scene.File), assets[scene.ID], 0o644); err != nil {
			return fmt.Errorf("write %s: %w", scene.File, err)
		}
	}
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("encode asset lock: %w", err)
	}
	lockData = append(lockData, '\n')
	if err := os.WriteFile(filepath.Join(dir, "generated-asset-lock.json"), lockData, 0o644); err != nil {
		return fmt.Errorf("write asset lock: %w", err)
	}
	return nil
}

func encodedAssets(m manifest) (map[string][]byte, error) {
	assets := make(map[string][]byte, len(m.Scenes))
	for _, scene := range m.Scenes {
		animation := &gif.GIF{LoopCount: 0, Image: make([]*image.Paletted, 0, frameCount), Delay: make([]int, 0, frameCount)}
		for frame := range frameCount {
			animation.Image = append(animation.Image, renderFrame(scene, frame))
			animation.Delay = append(animation.Delay, frameDelay)
		}
		var buffer bytes.Buffer
		if err := gif.EncodeAll(&buffer, animation); err != nil {
			return nil, fmt.Errorf("encode %s: %w", scene.File, err)
		}
		assets[scene.ID] = buffer.Bytes()
	}
	return assets, nil
}

func makeLock(m manifest, assets map[string][]byte) assetLock {
	lock := assetLock{Schema: "gooo/visual-asset-lock/v1", VisualDenominator: 10, VisualAssetCount: len(m.Scenes)}
	for _, scene := range m.Scenes {
		data := assets[scene.ID]
		digest := sha256.Sum256(data)
		lock.VisualAssetBytes += len(data)
		lock.Assets = append(lock.Assets, assetRecord{ID: scene.ID, Sequence: scene.Sequence, File: scene.File, SizeBytes: len(data), SHA256: hex.EncodeToString(digest[:])})
	}
	return lock
}

func checkAssets(dir string, m manifest) error {
	first, err := encodedAssets(m)
	if err != nil {
		return err
	}
	second, err := encodedAssets(m)
	if err != nil {
		return err
	}
	for _, scene := range m.Scenes {
		if !bytes.Equal(first[scene.ID], second[scene.ID]) {
			return fmt.Errorf("repeated generation differs for %s", scene.ID)
		}
		got, err := os.ReadFile(filepath.Join(dir, scene.File))
		if err != nil {
			return fmt.Errorf("asset existence 10/10 failed: %s: %w", scene.File, err)
		}
		if !bytes.Equal(got, first[scene.ID]) {
			return fmt.Errorf("generated asset is stale: %s", scene.File)
		}
	}
	lockData, err := os.ReadFile(filepath.Join(dir, "generated-asset-lock.json"))
	if err != nil {
		return fmt.Errorf("read generated asset lock: %w", err)
	}
	var gotLock assetLock
	if err := json.Unmarshal(lockData, &gotLock); err != nil {
		return fmt.Errorf("decode generated asset lock: %w", err)
	}
	wantLock := makeLock(m, first)
	if !sameLock(gotLock, wantLock) {
		return fmt.Errorf("generated asset lock is stale")
	}
	readme, err := os.ReadFile(filepath.Join(repositoryRoot(), "README.md"))
	if err != nil {
		return fmt.Errorf("read README: %w", err)
	}
	for _, scene := range m.Scenes {
		if strings.Count(string(readme), scene.File) != 1 || strings.Count(string(readme), scene.Alt) != 1 {
			return fmt.Errorf("README reference 10/10 failed for %s", scene.ID)
		}
	}
	fmt.Println("reproducible generation: 10/10")
	fmt.Println("README references: 10/10")
	fmt.Println("asset existence and exact bytes: 10/10")
	fmt.Println("engineering narrative cells: 10/10")
	fmt.Println("provenance path and digest bindings: 10/10")
	fmt.Println("semantic motion bindings: 10/10")
	fmt.Printf("visual denominator: %d scenes, %d GIF bytes\n", wantLock.VisualAssetCount, wantLock.VisualAssetBytes)
	return nil
}

func repositoryRoot() string {
	directory, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "."
		}
		directory = parent
	}
}

func sameLock(a, b assetLock) bool {
	if a.Schema != b.Schema || a.VisualDenominator != b.VisualDenominator || a.VisualAssetCount != b.VisualAssetCount || a.VisualAssetBytes != b.VisualAssetBytes || len(a.Assets) != len(b.Assets) {
		return false
	}
	for i := range a.Assets {
		if a.Assets[i] != b.Assets[i] {
			return false
		}
	}
	return true
}

func writePreviews(dir string, m manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create preview directory: %w", err)
	}
	for _, scene := range m.Scenes {
		path := filepath.Join(dir, fmt.Sprintf("%02d.png", scene.Sequence))
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create preview %s: %w", path, err)
		}
		if err := png.Encode(file, renderFrame(scene, frameCount-1)); err != nil {
			_ = file.Close()
			return fmt.Errorf("encode preview %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close preview %s: %w", path, err)
		}
	}
	return nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageConstructUsage = "usage: gooo package construct [--json] --cases <evaluation.json> " +
	"(--construction-cases <feedback.json> --attempts <1..64> [--model <model.json>] [--fill-model <model.json>] | " +
	"--receipt <construction.json>) [--go <go-binary>] <gooo.workspace.json>"

type packageConstructionReceipt struct {
	Schema         string                                 `json:"schema"`
	Decision       string                                 `json:"decision"`
	Manifest       string                                 `json:"manifest"`
	ManifestDigest string                                 `json:"manifest_digest,omitempty"`
	CasesDigest    string                                 `json:"cases_digest,omitempty"`
	ReplayedFrom   string                                 `json:"replayed_from_sha256,omitempty"`
	Result         *workspaceexecution.ConstructionResult `json:"result,omitempty"`
	Error          string                                 `json:"error,omitempty"`
}

func runPackageConstruct(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	flags, manifestPath, err := parsePackageConstructArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	receipt, err := executePackageConstruct(ctx, reader, manifestPath, flags)
	if err != nil {
		receipt.Decision, receipt.Error = "FAIL_CLOSED", err.Error()
	}
	return writePackageConstruction(receipt, err, jsonMode, stdout, stderr)
}

func parsePackageConstructArgs(args []string) (map[string]string, string, error) {
	flags := map[string]string{"--cases": "", "--construction-cases": "", "--attempts": "", "--model": "", "--fill-model": "", "--receipt": "", "--go": ""}
	manifest := ""
	for i := 0; i < len(args); i++ {
		if previous, ok := flags[args[i]]; ok {
			if previous != "" || i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") {
				return nil, "", fmt.Errorf("%s", packageConstructUsage)
			}
			flags[args[i]] = args[i+1]
			i++
		} else if manifest == "" && strings.TrimSpace(args[i]) != "" && !strings.HasPrefix(args[i], "-") {
			manifest = args[i]
		} else {
			return nil, "", fmt.Errorf("%s", packageConstructUsage)
		}
	}
	if manifest == "" || flags["--cases"] == "" {
		return nil, "", fmt.Errorf("%s", packageConstructUsage)
	}
	if flags["--receipt"] != "" {
		for _, key := range []string{"--construction-cases", "--attempts", "--model", "--fill-model"} {
			if flags[key] != "" {
				return nil, "", fmt.Errorf("saved package construction excludes %s", key)
			}
		}
	} else {
		budget, err := strconv.Atoi(flags["--attempts"])
		if err != nil || budget < 1 || budget > 64 || flags["--construction-cases"] == "" {
			return nil, "", fmt.Errorf("%s", packageConstructUsage)
		}
	}
	return flags, manifest, nil
}

func executePackageConstruct(ctx context.Context, reader SourceReader, path string, flags map[string]string) (packageConstructionReceipt, error) {
	r := packageConstructionReceipt{Schema: "gooo/workspace-caller-construction-receipt/v1", Manifest: filepath.ToSlash(path)}
	raw, err := readSource(reader, path)
	if err != nil {
		return r, err
	}
	if int64(len(raw)) > maxInputBytes {
		return r, inputLimitError(maxInputBytes)
	}
	r.ManifestDigest = workspaceDigest(raw)
	manifest, err := decodeWorkspaceManifest(raw)
	if err != nil {
		return r, err
	}
	loaded, err := loadPackageSources(reader, path, manifest)
	if err != nil {
		return r, err
	}
	evaluation, digest, err := readPackageConstructionCases(reader, flags["--cases"])
	r.CasesDigest = digest
	if err != nil {
		return r, err
	}
	var result workspaceexecution.ConstructionResult
	if flags["--receipt"] != "" {
		saved, digest, readErr := readSavedPackageConstruction(reader, flags["--receipt"], r.ManifestDigest)
		r.ReplayedFrom = digest
		if readErr != nil {
			return r, readErr
		}
		result, err = workspaceexecution.ReplayWorkspaceConstruction(ctx, loaded, *saved.Result, evaluation, flags["--go"])
	} else {
		construction, _, readErr := readPackageConstructionCases(reader, flags["--construction-cases"])
		if readErr != nil {
			return r, readErr
		}
		budget, _ := strconv.Atoi(flags["--attempts"])
		result, err = workspaceexecution.ConstructWorkspace(ctx, loaded, construction, evaluation, workspaceexecution.ConstructOptions{
			ProgramBudget: budget, ModelPath: flags["--model"], FillModelPath: flags["--fill-model"], GoBinary: flags["--go"]})
	}
	r.Result, r.Decision = &result, packageConstructionDecision(result)
	return r, err
}

func readPackageConstructionCases(reader SourceReader, path string) (bodyexecution.CompositionCases, string, error) {
	raw, err := readSource(reader, path)
	if err != nil {
		return bodyexecution.CompositionCases{}, "", err
	}
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	return suite, workspaceDigest(raw), err
}

func readSavedPackageConstruction(reader SourceReader, path, manifestDigest string) (packageConstructionReceipt, string, error) {
	var saved packageConstructionReceipt
	raw, err := readPackageConstructionReceipt(reader, path)
	if err != nil {
		return saved, "", err
	}
	digest := workspaceDigest(raw)
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		return saved, digest, err
	}
	if saved.Schema != "gooo/workspace-caller-construction-receipt/v1" || saved.ManifestDigest != manifestDigest ||
		saved.Error != "" || saved.Result == nil || (saved.Decision != "COMPLETE_FINITE" && saved.Decision != "PARTIAL_FINITE") {
		return saved, digest, fmt.Errorf("saved package construction envelope or manifest differs")
	}
	return saved, digest, nil
}

func readPackageConstructionReceipt(reader SourceReader, path string) ([]byte, error) {
	switch reader.(type) {
	case OSFileReader, *OSFileReader:
		return readBodyExecutionFile(path, 32<<20)
	}
	raw, err := reader.ReadFile(path)
	if err == nil && len(raw) > 32<<20 {
		return nil, inputLimitError(32 << 20)
	}
	return raw, err
}

func packageConstructionDecision(r workspaceexecution.ConstructionResult) string {
	v := r.Evaluation.Runtime
	if r.Construction.Decision != "COMPLETE_FINITE" || v.Stage != "COMPLETE" || v.FiniteTotal == 0 || v.FinitePassed != v.FiniteTotal {
		return "PARTIAL_FINITE"
	}
	for _, trace := range v.Traces {
		for _, delivery := range trace.Deliveries {
			if delivery.Fault != nil || len(delivery.BlockedBy) != 0 {
				return "PARTIAL_FINITE"
			}
		}
	}
	return "COMPLETE_FINITE"
}

func writePackageConstruction(r packageConstructionReceipt, cause error, jsonMode bool, stdout, stderr io.Writer) int {
	if jsonMode {
		raw, err := json.Marshal(r)
		if err == nil && len(raw) > 32<<20 {
			err = fmt.Errorf("package construction receipt exceeds 32 MiB")
		}
		if err == nil {
			_, err = stdout.Write(append(raw, '\n'))
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitFailure
		}
	} else if cause == nil {
		c, e := r.Result.Construction, r.Result.Evaluation
		fmt.Fprintf(stdout, "%s: attempts=%d/%d evaluation=%d/%d consumed_inputs=%d other_inputs=%d replayed=%t\n",
			r.Decision, len(c.Attempts), c.ProgramBudget, e.Runtime.FinitePassed, e.Runtime.FiniteTotal,
			e.InputSeparation.ConstructionInputs, e.InputSeparation.OtherInputs, e.ConstructionReplayed)
	}
	if cause != nil {
		fmt.Fprintf(stderr, "gooo package construct: %v\n", cause)
		return exitFailure
	}
	return exitOK
}

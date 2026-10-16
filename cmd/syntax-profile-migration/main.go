// Command syntax-profile-migration exports a source-bound proposal, without applying it.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func main() {
	if err := runProfileMigration(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runProfileMigration(arguments []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("syntax-profile-migration", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	root := flags.String("root", ".", "read-only input project")
	requestPath := flags.String("request", "", "explicit before-case and target-profile request JSON")
	sourcePath := flags.String("source-after", "", "replacement Gooo source; at most 128 KiB")
	directory := flags.String("output", "", "new caller-owned directory outside the project")
	inspect := flags.Bool("inspect", false, "observe byte digests and current execution identity")
	verifyPath := flags.String("verify-candidate", "", "verify an existing candidate without applying it")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 || *requestPath == "" || *sourcePath == "" ||
		(!*inspect && *directory == "" && *verifyPath == "") ||
		(*inspect && (*directory != "" || *verifyPath != "")) || (*verifyPath != "" && *directory != "") {
		return fmt.Errorf("use --request and --source-after with exactly one of --inspect, --output, --verify-candidate")
	}
	raw, err := readProfileFile(*requestPath, 32<<10)
	if err != nil {
		return err
	}
	request, err := syntaxregistration.DecodeProfileMigrationRequest(raw)
	if err != nil {
		return err
	}
	after, err := readProfileFile(*sourcePath, 128<<10)
	if err != nil {
		return err
	}
	repository := os.DirFS(*root)
	if *inspect {
		return inspectProfileMigration(repository, request, after, output)
	}
	plan, err := syntaxregistration.CompileProfileMigration(repository, request, after)
	if err != nil {
		return err
	}
	if *verifyPath != "" {
		return verifyProfileCandidate(repository, plan, *verifyPath, output)
	}
	candidate, err := plan.Generate(repository)
	if err != nil {
		return err
	}
	return exportProfileCandidate(*root, *directory, candidate, output)
}

func readProfileFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("profile migration input exceeds %d bytes: %s", limit, path)
	}
	return raw, nil
}

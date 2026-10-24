package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDiscoverPreservesAuthoredErrorOrigins(t *testing.T) {
	const header = "package billing\nnamespace billing\nentity Order id \"billing://order\"\n"
	for _, domain := range []bool{false, true} {
		for _, tc := range []struct {
			name, source, token, code string
			column                    int
		}{
			{"input", header + "activity PayOrder(Missing) -> Order\n", "Missing", "LOWER_FAILED", 19},
			{"output", header + "activity Pay(Order) -> Missing\n", "Missing", "LOWER_FAILED", 24},
			{"syntax", header + "@\n", "@", "PARSE_FAILED", 1},
		} {
			prefix, path := "SOURCE_", "subject.gooo"
			if domain {
				prefix, path = "DOMAIN_CONTRACT_", "contract.gooo"
			}
			t.Run(prefix+tc.name, func(t *testing.T) {
				reader := runSourceReaderWithFiles{"subject.gooo": []byte(discoverFixture)}
				reader[path] = []byte(tc.source)
				args := []string{"--json", "--query", "Generate Gooo code", "subject.gooo"}
				if domain {
					args = append(args, "--domain-contract", path)
				}
				var stdout, stderr bytes.Buffer
				if code := runDiscover(args, reader, &stdout, &stderr); code != exitFailure || stderr.Len() != 0 {
					t.Fatalf("discover error result: exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
				}
				var report jsonReport
				if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
					t.Fatal(err)
				}
				if report.SchemaVersion != "gooo/diagnostics/v1" || report.Command != "discover" ||
					report.Status != "error" || report.File != path || len(report.Diagnostics) != 1 {
					t.Fatalf("discovery error envelope changed: %s", &stdout)
				}
				diagnostic := report.Diagnostics[0]
				span := diagnostic.Span
				offset := strings.Index(tc.source, tc.token)
				if diagnostic.Code != prefix+tc.code || span.File != path ||
					span.Start.Line != 4 || span.End.Line != 4 || span.Start.Column != tc.column ||
					span.End.Column != tc.column+len(tc.token) || span.Start.Offset != offset || span.End.Offset != offset+len(tc.token) {
					t.Fatalf("lost authored %q origin: %s", tc.token, &stdout)
				}
				var again, againErr bytes.Buffer
				if code := runDiscover(args, reader, &again, &againErr); code != exitFailure ||
					againErr.Len() != 0 || !bytes.Equal(stdout.Bytes(), again.Bytes()) {
					t.Fatal("error origin did not replay byte-identically")
				}
			})
		}
	}
}

func TestDiscoverReadFailureDoesNotInventCoordinates(t *testing.T) {
	for _, domain := range []bool{false, true} {
		args := []string{"--json", "--query", "Generate Gooo code", "missing.gooo"}
		reader := runSourceReaderWithFiles{}
		code := "SOURCE_READ_FAILED"
		if domain {
			args = []string{"--json", "--query", "Generate Gooo code", "--domain-contract", "missing.gooo", "subject.gooo"}
			reader["subject.gooo"] = []byte(discoverFixture)
			code = "DOMAIN_CONTRACT_READ_FAILED"
		}
		var stdout, stderr bytes.Buffer
		if exit := runDiscover(args, reader, &stdout, &stderr); exit != exitFailure || stderr.Len() != 0 {
			t.Fatalf("read error result: exit=%d stdout=%s stderr=%s", exit, &stdout, &stderr)
		}
		var report jsonReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != code ||
			report.Diagnostics[0].Span.Start != (cliPosition{}) || report.Diagnostics[0].Span.End != (cliPosition{}) {
			t.Fatalf("read failure invented a source position: %s", &stdout)
		}
	}
}

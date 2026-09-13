package publicpartialreuse

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestObserveTestOutputRequiresExactExecutedCohort(t *testing.T) {
	event := func(action, name string) string {
		data, err := json.Marshal(map[string]string{"Action": action, "Package": "partial-reuse-example", "Test": name})
		if err != nil {
			t.Fatal(err)
		}
		return string(data) + "\n"
	}
	start, finish := event("start", ""), event("pass", "")
	orders := event("run", "TestOrdersPartition") + event("pass", "TestOrdersPartition")
	inventory := event("run", "TestInventoryPartition") + event("pass", "TestInventoryPartition")
	expected := []string{"TestOrdersPartition", "TestInventoryPartition"}
	cases := []struct {
		name string
		data string
		want bool
	}{
		{"both_executed", start + orders + inventory + finish, true},
		{"inventory_not_selected", start + orders + finish, false},
		{"no_tests_selected", start + finish, false},
		{"inventory_skipped", start + orders + event("run", expected[1]) + event("skip", expected[1]) + finish, false},
		{"inventory_failed", start + orders + event("run", expected[1]) + event("fail", expected[1]), false},
		{"duplicate_pass", start + orders + inventory + event("pass", expected[1]) + finish, false},
		{"pass_without_run", start + orders + event("pass", expected[1]) + finish, false},
		{"unexpected_test", start + orders + inventory + event("run", "TestOther") + finish, false},
		{"missing_package_success", start + orders + inventory, false},
		{"malformed_trailer", start + orders + inventory + finish + "{", false},
		{"null_event", "null", false},
		{"package_identity_missing", "{}\n", false},
		{"subtest_not_extra_credit", start + event("run", expected[0]) + event("run", expected[0]+"/child") + event("pass", expected[0]+"/child") + event("pass", expected[0]) + inventory + finish, true},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			passed, err := ObserveTestOutput([]byte(item.data), expected)
			if (err == nil) != item.want {
				t.Fatalf("passed=%v error=%v, want success=%v", passed, err, item.want)
			}
			if item.want && !slices.Equal(passed, expected) {
				t.Fatalf("observed test names=%v, want %v", passed, expected)
			}
		})
	}
	for _, names := range [][]string{nil, {""}, {"TestOrdersPartition", "TestOrdersPartition"}} {
		if _, err := ObserveTestOutput([]byte(start+orders+finish), names); err == nil {
			t.Fatalf("invalid expected cohort accepted: %v", names)
		}
	}
}

func TestPartitionCommandMatchesOnlyDeclaredTest(t *testing.T) {
	for _, name := range []string{"TestOrdersPartition", "TestInventoryPartition", "TestLiteral.Partition"} {
		args := TestCommandArgs(Partition{TestName: name})
		pattern := regexp.MustCompile(args[5])
		if args[1] != "-json" || !pattern.MatchString(name) || pattern.MatchString("Prefix"+name) || pattern.MatchString(name+"Suffix") {
			t.Fatalf("command does not select the exact declared test: %v", args)
		}
	}
}

func TestReadReceiptClassifiesContentFailure(t *testing.T) {
	valid := receiptFixtureBytes(t)
	suffix := func(value string) []byte {
		return append(append([]byte(nil), valid...), value...)
	}
	input := receiptEvaluationFixture(t)
	badSeal := input.Receipts["orders"]
	badSeal.ReceiptID = cache.HashBytes([]byte("different-test-seal")).String()
	badSealJSON, err := json.Marshal(badSeal)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		data      []byte
		missing   bool
		directory bool
		wantError bool
		invalid   bool
	}{
		{name: "valid", data: valid},
		{name: "whitespace_is_not_tampering", data: suffix("\n\t ")},
		{name: "second_json_value", data: suffix("\n{}"), wantError: true, invalid: true},
		{name: "trailing_garbage", data: suffix("!"), wantError: true, invalid: true},
		{name: "malformed_json", data: []byte("{"), wantError: true, invalid: true},
		{name: "missing_content_fields", data: []byte("{}"), wantError: true, invalid: true},
		{name: "wrong_json_shape", data: []byte("[]"), wantError: true, invalid: true},
		{name: "invalid_content_seal", data: badSealJSON, wantError: true, invalid: true},
		{name: "missing_file", missing: true, wantError: true},
		{name: "non_regular_input", directory: true, wantError: true},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "receipt.json")
			if item.directory {
				if err := os.Mkdir(filename, 0o700); err != nil {
					t.Fatal(err)
				}
			} else if !item.missing {
				if err := os.WriteFile(filename, item.data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := ReadReceipt(filename)
			if (err != nil) != item.wantError {
				t.Fatalf("ReadReceipt error=%v, want error=%v", err, item.wantError)
			}
			if errors.Is(err, ErrInvalidReceipt) != item.invalid {
				t.Fatalf("content refutation=%v, want %v: %v", errors.Is(err, ErrInvalidReceipt), item.invalid, err)
			}
		})
	}
}

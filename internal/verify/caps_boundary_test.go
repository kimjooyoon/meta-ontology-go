package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoCapsAcceptExactDAMP300(t *testing.T) {
	root := t.TempDir()
	writeCapsFixture(t, root, "file.go", "package fixture\n"+strings.Repeat("var _ = 1\n", 299))
	if err := CheckGoCaps(root, []string{"file.go"}, 300, 75); err != nil {
		t.Fatal(err)
	}
}

func TestGoCapsRejectDAMP301(t *testing.T) {
	root := t.TempDir()
	writeCapsFixture(t, root, "file.go", "package fixture\n"+strings.Repeat("var _ = 1\n", 300))
	if err := CheckGoCaps(root, []string{"file.go"}, 300, 75); err == nil {
		t.Fatal("301-line Go file was accepted")
	}
}

func TestGoCapsAcceptExactDRY75(t *testing.T) {
	root := t.TempDir()
	writeCapsFixture(t, root, "function.go", exactFunction(73))
	if err := CheckGoCaps(root, []string{"function.go"}, 300, 75); err != nil {
		t.Fatal(err)
	}
}

func TestGoCapsRejectDRY76(t *testing.T) {
	root := t.TempDir()
	writeCapsFixture(t, root, "function.go", exactFunction(74))
	if err := CheckGoCaps(root, []string{"function.go"}, 300, 75); err == nil {
		t.Fatal("76-line Go function was accepted")
	}
}

func writeCapsFixture(t *testing.T, root, name, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

func exactFunction(bodyLines int) string {
	return "package fixture\nfunc Exact() {\n" + strings.Repeat("\t_ = 1\n", bodyLines) + "}\n"
}

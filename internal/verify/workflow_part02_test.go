package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIGuardianAppTokenUsesClientIDAndOneMintSecret(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci-guardian.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	if strings.Contains(text, "app-id:") || !strings.Contains(text, "client-id: ${{ env.GUARDIAN_APP_CLIENT_ID }}") {
		t.Fatal("Guardian App token action is not using the pinned client-id input")
	}
	mintStart := strings.Index(text, "id: guardian-app-token")
	if mintStart < 0 {
		t.Fatal("Guardian App mint step is missing")
	}
	mintEnd := strings.Index(text[mintStart:], "- name: Inspect changed paths from default authority")
	if mintEnd < 0 {
		t.Fatal("Guardian App mint step has no bounded end")
	}
	mintStep := text[mintStart : mintStart+mintEnd]
	if strings.Count(text, "${{ secrets.") != 1 || strings.Count(mintStep, "${{ secrets.") != 1 || !strings.Contains(mintStep, "GUARDIAN_APP_PRIVATE_KEY: ${{ secrets.GUARDIAN_APP_PRIVATE_KEY }}") {
		t.Fatal("Guardian mint step does not use exactly GUARDIAN_APP_PRIVATE_KEY")
	}
}

package collapsefixture

import "testing"

func TestCollapseFixture(t *testing.T) {
	if CollapseFixture() != 1 {
		t.Fatal("collapse fixture returned an unexpected value")
	}
}

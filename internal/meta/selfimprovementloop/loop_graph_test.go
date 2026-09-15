package selfimprovementloop

import "testing"

func TestReleasedGraphBindsExactlyOneActivityPerCell(t *testing.T) {
	bindings, err := BindActivities(testGraph())
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 12 || len(SemanticCells()) != 12 {
		t.Fatalf("bindings/cells = %d/%d, want 12/12", len(bindings), len(SemanticCells()))
	}
	for index, binding := range bindings {
		if binding.Cell != fixedCells[index] || binding.Activity != fixedCells[index] || binding.ActivityID == "" {
			t.Fatalf("binding %d = %#v", index, binding)
		}
	}
}

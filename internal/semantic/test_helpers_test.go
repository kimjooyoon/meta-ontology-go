package semantic

import "testing"

func mustNode(t *testing.T, node Node, err error) Node {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func mustEntity(t *testing.T, id ID, namespace Namespace, name string) Node {
	t.Helper()
	node, err := NewEntity(id, namespace, name)
	return mustNode(t, node, err)
}

func mustActivity(t *testing.T, id ID, namespace Namespace, name string) Node {
	t.Helper()
	node, err := NewActivity(id, namespace, name)
	return mustNode(t, node, err)
}

func mustAgent(t *testing.T, id ID, namespace Namespace, name string) Node {
	t.Helper()
	node, err := NewAgent(id, namespace, name)
	return mustNode(t, node, err)
}

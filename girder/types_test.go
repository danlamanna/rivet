package girder

import (
	"testing"
)

func TestResourceMap_Parent(t *testing.T) {
	parent, child := &Resource{Path: "/"}, &Resource{Path: "/child"}

	m := ResourceMap{
		"/":      parent,
		"/child": child,
	}

	got := m.Parent(child)

	if got != parent {
		t.Errorf("Expected parent %v, got %v", parent, got)
	}
}

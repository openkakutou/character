package character

import "testing"

func TestCharacter_ZeroValue_HasEmptyName(t *testing.T) {
	var c Character

	if c.Name != "" {
		t.Errorf("expected zero-value Character to have an empty Name, got %q", c.Name)
	}
}

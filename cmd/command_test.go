package cmd

import "testing"

func TestCommandFile_ZeroValue_AllFieldsZero(t *testing.T) {
	var f CommandFile

	if f.Remap != nil {
		t.Errorf("expected zero-value CommandFile to have nil Remap, got %v", f.Remap)
	}
	if len(f.Remap) != 0 {
		t.Errorf("expected zero-value CommandFile to have 0 Remap entries, got %d", len(f.Remap))
	}
	if f.Defaults != (CommandDefaults{}) {
		t.Errorf("expected zero-value CommandFile to have zero-value Defaults, got %+v", f.Defaults)
	}
	if f.Commands != nil {
		t.Errorf("expected zero-value CommandFile to have nil Commands, got %v", f.Commands)
	}
	if len(f.Commands) != 0 {
		t.Errorf("expected zero-value CommandFile to have 0 Commands, got %d", len(f.Commands))
	}
	if f.States != nil {
		t.Errorf("expected zero-value CommandFile to have nil States, got %v", f.States)
	}
	if len(f.States) != 0 {
		t.Errorf("expected zero-value CommandFile to have 0 States, got %d", len(f.States))
	}

	// Ranging over nil Commands/States and a nil Remap map must not panic.
	for range f.Commands {
		t.Fatal("expected no iterations over a nil Commands slice")
	}
	for range f.States {
		t.Fatal("expected no iterations over a nil States slice")
	}
	for range f.Remap {
		t.Fatal("expected no iterations over a nil Remap map")
	}
}

func TestCommandFile_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	f := CommandFile{
		Remap: map[string]string{"a": "x", "b": "y"},
		Defaults: CommandDefaults{
			Time:       15,
			BufferTime: 1,
		},
		Commands: []Command{
			{Name: "QCF_a", Input: "~D, DF, F, a", Time: 15, BufferTime: 1},
			{Name: "holdback", Input: "/$B"},
		},
	}

	if len(f.Remap) != 2 || f.Remap["a"] != "x" || f.Remap["b"] != "y" {
		t.Errorf("expected Remap {a:x, b:y}, got %v", f.Remap)
	}
	if f.Defaults.Time != 15 {
		t.Errorf("expected Defaults.Time 15, got %d", f.Defaults.Time)
	}
	if f.Defaults.BufferTime != 1 {
		t.Errorf("expected Defaults.BufferTime 1, got %d", f.Defaults.BufferTime)
	}

	if len(f.Commands) != 2 {
		t.Fatalf("expected 2 Commands, got %d", len(f.Commands))
	}
	if f.Commands[0].Name != "QCF_a" || f.Commands[0].Input != "~D, DF, F, a" || f.Commands[0].Time != 15 || f.Commands[0].BufferTime != 1 {
		t.Errorf("expected first Command to preserve its assigned fields, got %+v", f.Commands[0])
	}
	if f.Commands[1].Name != "holdback" || f.Commands[1].Input != "/$B" {
		t.Errorf("expected second Command to preserve its assigned fields, got %+v", f.Commands[1])
	}
	// A Command with no explicit Time/BufferTime override stays at zero,
	// meaning "use CommandFile.Defaults" — Command itself never resolves
	// this, mirroring cns.StateDef.Anim's own "0 means not set" convention.
	if f.Commands[1].Time != 0 || f.Commands[1].BufferTime != 0 {
		t.Errorf("expected second Command's unset Time/BufferTime to stay zero, got Time=%d BufferTime=%d", f.Commands[1].Time, f.Commands[1].BufferTime)
	}
}

func TestCommandFile_TwoInstances_DoNotShareSliceOrMapFields(t *testing.T) {
	a := CommandFile{
		Remap:    map[string]string{"a": "x"},
		Commands: []Command{{Name: "a"}},
	}
	b := CommandFile{
		Remap:    map[string]string{"a": "a"},
		Commands: []Command{{Name: "b"}},
	}

	a.Remap["a"] = "mutated"
	a.Commands[0].Name = "mutated"

	if b.Remap["a"] != "a" {
		t.Errorf("expected b.Remap to be unaffected by mutating a, got %v", b.Remap)
	}
	if b.Commands[0].Name != "b" {
		t.Errorf("expected b.Commands to be unaffected by mutating a, got %v", b.Commands)
	}
}

func TestCommandFile_MissingOptionalSections_LeaveThemEmptyWithoutError(t *testing.T) {
	// A minimal command file might define only a single command, with no
	// Remap and no linked states at all — every other field should simply
	// stay at its zero value rather than requiring a placeholder.
	f := CommandFile{
		Commands: []Command{{Name: "a", Input: "a"}},
	}

	if len(f.Commands) != 1 || f.Commands[0].Name != "a" {
		t.Errorf("expected the single assigned Command to be preserved, got %v", f.Commands)
	}
	if len(f.Remap) != 0 {
		t.Errorf("expected empty Remap to remain empty, got %v", f.Remap)
	}
	if len(f.States) != 0 {
		t.Errorf("expected empty States to remain empty, got %v", f.States)
	}
}

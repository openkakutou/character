package def

import "testing"

func TestCharacterInfo_ZeroValue_AllFieldsZero(t *testing.T) {
	var info CharacterInfo

	if info.Name != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty Name, got %q", info.Name)
	}
	if info.Author != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty Author, got %q", info.Author)
	}
	if info.SpriteFile != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty SpriteFile, got %q", info.SpriteFile)
	}
	if info.AnimationFile != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty AnimationFile, got %q", info.AnimationFile)
	}
	if info.SoundFile != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty SoundFile, got %q", info.SoundFile)
	}
	if info.CommandFile != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty CommandFile, got %q", info.CommandFile)
	}
	if info.ConstantsFile != "" {
		t.Errorf("expected zero-value CharacterInfo to have empty ConstantsFile, got %q", info.ConstantsFile)
	}
	if info.StateFiles != nil {
		t.Errorf("expected zero-value CharacterInfo to have nil StateFiles, got %v", info.StateFiles)
	}
	if len(info.StateFiles) != 0 {
		t.Errorf("expected zero-value CharacterInfo to have 0 StateFiles, got %d", len(info.StateFiles))
	}
	if info.Palettes != nil {
		t.Errorf("expected zero-value CharacterInfo to have nil Palettes, got %v", info.Palettes)
	}
	if len(info.Palettes) != 0 {
		t.Errorf("expected zero-value CharacterInfo to have 0 Palettes, got %d", len(info.Palettes))
	}

	// Ranging over nil StateFiles/Palettes slices must not panic.
	for range info.StateFiles {
		t.Fatal("expected no iterations over a nil StateFiles slice")
	}
	for range info.Palettes {
		t.Fatal("expected no iterations over a nil Palettes slice")
	}
}

func TestCharacterInfo_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	info := CharacterInfo{
		Name:          "Kung Fu Man",
		Author:        "Elecbyte",
		SpriteFile:    "kfm.sff",
		AnimationFile: "kfm.air",
		SoundFile:     "kfm.snd",
		CommandFile:   "kfm.cmd",
		ConstantsFile: "kfm.cns",
		StateFiles:    []string{"kfm_extra1.st", "kfm_extra2.st"},
		Palettes:      []string{"kfm1.act", "kfm2.act", "kfm3.act"},
	}

	if info.Name != "Kung Fu Man" {
		t.Errorf("expected Name %q, got %q", "Kung Fu Man", info.Name)
	}
	if info.Author != "Elecbyte" {
		t.Errorf("expected Author %q, got %q", "Elecbyte", info.Author)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "kfm.sff", info.SpriteFile)
	}
	if info.AnimationFile != "kfm.air" {
		t.Errorf("expected AnimationFile %q, got %q", "kfm.air", info.AnimationFile)
	}
	if info.SoundFile != "kfm.snd" {
		t.Errorf("expected SoundFile %q, got %q", "kfm.snd", info.SoundFile)
	}
	if info.CommandFile != "kfm.cmd" {
		t.Errorf("expected CommandFile %q, got %q", "kfm.cmd", info.CommandFile)
	}
	if info.ConstantsFile != "kfm.cns" {
		t.Errorf("expected ConstantsFile %q, got %q", "kfm.cns", info.ConstantsFile)
	}

	if len(info.StateFiles) != 2 {
		t.Fatalf("expected 2 StateFiles, got %d", len(info.StateFiles))
	}
	if info.StateFiles[0] != "kfm_extra1.st" || info.StateFiles[1] != "kfm_extra2.st" {
		t.Errorf("expected StateFiles in assigned order, got %v", info.StateFiles)
	}

	if len(info.Palettes) != 3 {
		t.Fatalf("expected 3 Palettes, got %d", len(info.Palettes))
	}
	if info.Palettes[0] != "kfm1.act" || info.Palettes[1] != "kfm2.act" || info.Palettes[2] != "kfm3.act" {
		t.Errorf("expected Palettes in assigned order, got %v", info.Palettes)
	}
}

func TestCharacterInfo_TwoInstances_DoNotShareSliceFields(t *testing.T) {
	a := CharacterInfo{StateFiles: []string{"a.st"}, Palettes: []string{"a1.act"}}
	b := CharacterInfo{StateFiles: []string{"b.st"}, Palettes: []string{"b1.act"}}

	a.StateFiles[0] = "mutated.st"
	a.Palettes[0] = "mutated.act"

	if b.StateFiles[0] != "b.st" {
		t.Errorf("expected b.StateFiles to be unaffected by mutating a, got %v", b.StateFiles)
	}
	if b.Palettes[0] != "b1.act" {
		t.Errorf("expected b.Palettes to be unaffected by mutating a, got %v", b.Palettes)
	}
}

func TestCharacterInfo_MissingOptionalFiles_LeavesThemEmptyWithoutError(t *testing.T) {
	// A minimal character definition only needs a name and a sprite file;
	// every other reference is optional and should simply stay at its zero
	// value rather than requiring a placeholder.
	info := CharacterInfo{
		Name:       "Minimal",
		SpriteFile: "minimal.sff",
	}

	if info.Name != "Minimal" || info.SpriteFile != "minimal.sff" {
		t.Errorf("expected Name %q and SpriteFile %q, got Name %q SpriteFile %q", "Minimal", "minimal.sff", info.Name, info.SpriteFile)
	}
	if info.Author != "" {
		t.Errorf("expected empty Author to remain empty, got %q", info.Author)
	}
	if info.AnimationFile != "" || info.SoundFile != "" || info.CommandFile != "" || info.ConstantsFile != "" {
		t.Errorf("expected unassigned file references to remain empty, got AnimationFile %q SoundFile %q CommandFile %q ConstantsFile %q",
			info.AnimationFile, info.SoundFile, info.CommandFile, info.ConstantsFile)
	}
	if len(info.StateFiles) != 0 || len(info.Palettes) != 0 {
		t.Errorf("expected unassigned StateFiles/Palettes to remain empty, got StateFiles %v Palettes %v", info.StateFiles, info.Palettes)
	}
}

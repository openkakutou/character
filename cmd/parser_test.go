package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestParse_SampleFixture_ProducesExpectedCommandFile(t *testing.T) {
	f, err := os.Open("testdata/sample.cmd")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	file, err := Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantRemap := map[string]string{"a": "a", "b": "b", "x": "y"}
	if len(file.Remap) != len(wantRemap) {
		t.Fatalf("expected %d Remap entries, got %d: %v", len(wantRemap), len(file.Remap), file.Remap)
	}
	for k, v := range wantRemap {
		if file.Remap[k] != v {
			t.Errorf("Remap[%q]: expected %q, got %q", k, v, file.Remap[k])
		}
	}

	if file.Defaults.Time != 15 {
		t.Errorf("expected Defaults.Time 15, got %d", file.Defaults.Time)
	}
	if file.Defaults.BufferTime != 1 {
		t.Errorf("expected Defaults.BufferTime 1, got %d", file.Defaults.BufferTime)
	}

	if len(file.Commands) != 2 {
		t.Fatalf("expected 2 Commands, got %d: %v", len(file.Commands), file.Commands)
	}
	if file.Commands[0].Name != "a" || file.Commands[0].Input != "a" || file.Commands[0].Time != 1 {
		t.Errorf("Commands[0]: expected {Name: a, Input: a, Time: 1}, got %+v", file.Commands[0])
	}
	if file.Commands[1].Name != "QCF_a" || file.Commands[1].Input != "~D, DF, F, a" {
		t.Errorf("Commands[1]: expected {Name: QCF_a, Input: \"~D, DF, F, a\"}, got %+v", file.Commands[1])
	}
	// Commands[1] never set its own time; it must stay at zero rather than
	// inheriting Defaults.Time — resolving that fallback is a consumer's
	// concern, not Parse's.
	if file.Commands[1].Time != 0 {
		t.Errorf("expected Commands[1].Time to stay 0 (unset), got %d", file.Commands[1].Time)
	}

	if len(file.States) != 1 {
		t.Fatalf("expected 1 linked State (Statedef -1), got %d: %v", len(file.States), file.States)
	}
	if file.States[0].Number != -1 {
		t.Errorf("expected the linked state's Number to be -1, got %d", file.States[0].Number)
	}
	if len(file.States[0].Controllers) != 2 {
		t.Fatalf("expected 2 Controllers under Statedef -1, got %d", len(file.States[0].Controllers))
	}
	// The second controller is the actual command-to-state link: its
	// trigger references the Command's Name via the same unevaluated
	// cns.Controller.Triggers mechanism cns.Parse already provides — Parse
	// does not need any dedicated "link" field.
	linkCtrl := file.States[0].Controllers[1]
	if linkCtrl.Type != "ChangeState" {
		t.Errorf("expected linked controller Type ChangeState, got %q", linkCtrl.Type)
	}
	if len(linkCtrl.Triggers) == 0 || !strings.Contains(linkCtrl.Triggers[0], `command = "QCF_a"`) {
		t.Errorf("expected linked controller's first trigger to reference command \"QCF_a\", got %v", linkCtrl.Triggers)
	}
}

func TestParse_NoRemapSection_LeavesRemapNil(t *testing.T) {
	src := `[Defaults]
command.time = 10

[Command]
name = "a"
command = a
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Remap != nil {
		t.Errorf("expected nil Remap when no [Remap] section is present, got %v", file.Remap)
	}
}

func TestParse_CaseInsensitiveKeysAndSections_StillResolve(t *testing.T) {
	// Real MUGEN-authored .cmd files mix key/section case freely
	// (e.g. "Command.Time" vs "command.time", "[COMMAND]" vs "[Command]").
	src := `[COMMAND]
Name = "b"
Command = b
Time = 2
Buffer.Time = 3
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Commands) != 1 {
		t.Fatalf("expected 1 Command, got %d", len(file.Commands))
	}
	got := file.Commands[0]
	if got.Name != "b" || got.Input != "b" || got.Time != 2 || got.BufferTime != 3 {
		t.Errorf("expected {Name: b, Input: b, Time: 2, BufferTime: 3}, got %+v", got)
	}
}

// Real-world input syntax: GGX_chipp's "Command.cmd" (Guilty Gear, Ikemen GO
// corpus) uses charge-then-release notation combining "~" (release), "$"
// (any direction of travel counts), and "+" (simultaneous buttons) in a
// single input string — Parse must store it verbatim, not attempt to
// evaluate or decompose it, mirroring cns.Controller's unevaluated
// Triggers/Parameters.
func TestParse_RealChargeInputSyntax_IsStoredVerbatim(t *testing.T) {
	src := `[Command]
name = "charge_hs"
command = ~$D, /$U, a+b~
time = 15
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Commands) != 1 {
		t.Fatalf("expected 1 Command, got %d", len(file.Commands))
	}
	if file.Commands[0].Input != "~$D, /$U, a+b~" {
		t.Errorf("expected Input to be preserved verbatim, got %q", file.Commands[0].Input)
	}
}

func TestParse_NonKeyValueContentLine_IsIgnoredNotError(t *testing.T) {
	// A decorative separator or truncated leftover key inside a [Command]
	// block is tolerated the same way cns.Parse/def.Parse already tolerate
	// it elsewhere in this repo (no "=" character, so not a valid
	// key=value pair).
	src := `[Command]
;----------------------
name = "a"
------------------------
command = a
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Commands) != 1 || file.Commands[0].Name != "a" || file.Commands[0].Input != "a" {
		t.Errorf("expected the decorative line to be ignored, got %+v", file.Commands)
	}
}

func TestParse_UnrecognizedSection_IsSkippedWithoutAbortingKnownSections(t *testing.T) {
	src := `[SomeUnrelatedSection]
foo = bar

[Command]
name = "a"
command = a
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Commands) != 1 || file.Commands[0].Name != "a" {
		t.Errorf("expected the unrecognized section to be skipped and [Command] still parsed, got %+v", file.Commands)
	}
}

func TestParse_EmptyInput_ReturnsZeroValueCommandFile(t *testing.T) {
	file, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Commands) != 0 || len(file.Remap) != 0 || len(file.States) != 0 {
		t.Errorf("expected an empty CommandFile, got %+v", file)
	}
}

func TestParse_MalformedCommandSectionHeader_ReturnsDescriptiveError(t *testing.T) {
	// "[Command" is missing its closing "]" and is recognizable as an
	// attempt at this package's own "[Command]" header (as opposed to a
	// genuinely unrelated section, which Parse tolerates without error) —
	// mirrors cns.Parse's own attempted-header-vs-unrelated-content
	// distinction (.vibe/decisions/012).
	src := `[Command
name = "a"
`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a malformed [Command] header, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected the error to name line 1, got: %v", err)
	}
}

// A "[State ...]" header missing its closing "]" is a real-world .cmd
// authoring pattern (the same typo backlog item 042 documented for .cns,
// since a .cmd file's "[Statedef -1]" block shares the exact same syntax)
// — it must not be misreported as a malformed [Command]/[Defaults]/[Remap]
// attempt; cns.Parse (delegated to for the States field) already recovers
// it on its own.
func TestParse_StateHeaderMissingClosingBracket_DoesNotErrorInSectionScan(t *testing.T) {
	src := `[Statedef -1]

[State -1, Roll Forward
type = ChangeState
value = 715
trigger1 = command = "holdback"
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.States) != 1 || len(file.States[0].Controllers) != 1 {
		t.Fatalf("expected the recovered [State ...] block to still be parsed via cns.Parse, got %+v", file.States)
	}
}

// Real-world .cmd files sometimes omit the "[Statedef -1]" header entirely
// before their "[State ...]" controllers, since -1 is the only Statedef
// number a .cmd file's "always" section can ever use — the header adds
// nothing a reader doesn't already know. Reproduces the exact structure
// found via a 520-file real-character corpus scan (Marvel's "Jean Grey" and
// "Nova"): every "[State -1, ...]" controller sits directly after the last
// "[Command]" block, with no "[Statedef -1]" line anywhere in the file.
// cns.Parse alone rejects this ("[State ...] block found outside of any
// Statedef") since a bare .cns file has no such implicit-numbering
// convention; Parse must compensate before delegating.
func TestParse_MissingImplicitStatedefHeader_IsToleratedForCmdFiles(t *testing.T) {
	src := `[Command]
name = "a"
command = a

[State -1, AI]
type = VarSet
trigger1 = 1
value = 1
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.States) != 1 {
		t.Fatalf("expected the implicit Statedef -1 to be recovered as 1 State, got %d: %+v", len(file.States), file.States)
	}
	if file.States[0].Number != -1 {
		t.Errorf("expected the recovered State's Number to be -1, got %d", file.States[0].Number)
	}
	if len(file.States[0].Controllers) != 1 {
		t.Fatalf("expected 1 Controller under the recovered State, got %d", len(file.States[0].Controllers))
	}
}

// A file that already declares "[Statedef -1]" explicitly must not have a
// second, synthetic one injected — that would silently produce two States
// (both numbered -1) instead of one.
func TestParse_ExplicitStatedefHeader_IsNotDuplicated(t *testing.T) {
	src := `[Statedef -1]

[State -1, First]
type = VarSet
trigger1 = 1
value = 1

[State -1, Second]
type = VarSet
trigger1 = 1
value = 2
`
	file, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.States) != 1 {
		t.Fatalf("expected a single Statedef -1 (not duplicated), got %d: %+v", len(file.States), file.States)
	}
	if len(file.States[0].Controllers) != 2 {
		t.Fatalf("expected both [State -1, ...] blocks under the one Statedef, got %d Controllers", len(file.States[0].Controllers))
	}
}

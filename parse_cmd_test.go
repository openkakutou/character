package character

import (
	"strings"
	"testing"
)

// TestParseCmd_MugenStyleFixture_ProducesExpectedCommandFile covers a
// classic MUGEN-style .cmd file: an explicit [Statedef -1] header, a
// [Remap] section, and a [Defaults] section — the read-path counterpart to
// SerializeCmd, wrapping cmd.Parse for the WASM entrypoint (item 056).
func TestParseCmd_MugenStyleFixture_ProducesExpectedCommandFile(t *testing.T) {
	src := []byte(`[Remap]
a = a
b = b

[Defaults]
command.time = 15
command.buffer.time = 1

[Command]
name = "QCF_a"
command = ~D, DF, F, a

[Statedef -1]

[State -1, QCF Special]
type = ChangeState
value = 1000
trigger1 = command = "QCF_a"
`)

	file, err := ParseCmd(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Remap) != 2 || file.Remap["a"] != "a" || file.Remap["b"] != "b" {
		t.Errorf("expected Remap {a: a, b: b}, got %v", file.Remap)
	}
	if file.Defaults.Time != 15 || file.Defaults.BufferTime != 1 {
		t.Errorf("expected Defaults {Time: 15, BufferTime: 1}, got %+v", file.Defaults)
	}
	if len(file.Commands) != 1 || file.Commands[0].Name != "QCF_a" || file.Commands[0].Input != "~D, DF, F, a" {
		t.Errorf("expected 1 Command {Name: QCF_a, Input: \"~D, DF, F, a\"}, got %+v", file.Commands)
	}
	if len(file.States) != 1 || file.States[0].Number != -1 {
		t.Fatalf("expected 1 linked State (Statedef -1), got %+v", file.States)
	}
}

// TestParseCmd_IkemenStyleFixture_OmittingStatedefHeader_StillLinksState
// covers an Ikemen GO-style .cmd file that omits the "[Statedef -1]" header
// entirely (a real-world quirk cmd.Parse already tolerates by synthesizing
// it — see .vibe/decisions/026) and uses charge-input notation ("~$", "/$")
// combining release/any-direction modifiers, real-world syntax found in the
// Ikemen GO corpus.
func TestParseCmd_IkemenStyleFixture_OmittingStatedefHeader_StillLinksState(t *testing.T) {
	src := []byte(`[Command]
name = "charge_hs"
command = ~$D, /$U, a+b~
time = 15

[State -1, Charge Special]
type = ChangeState
value = 2000
trigger1 = command = "charge_hs"
`)

	file, err := ParseCmd(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Commands) != 1 || file.Commands[0].Input != "~$D, /$U, a+b~" {
		t.Errorf("expected the charge input to be preserved verbatim, got %+v", file.Commands)
	}
	if len(file.States) != 1 || file.States[0].Number != -1 {
		t.Fatalf("expected the implicit Statedef -1 to still be synthesized and linked, got %+v", file.States)
	}
}

// TestParseCmd_NoRemapOrCommandsSection_NormalizesToEmptyNotNil verifies the
// same nil-slice/map-to-empty normalization LoadBytes's JSON contract
// already guarantees (see normalizeCommandFileForJSON), since ParseCmd
// feeds the same WASM/JSON boundary via loadCmd.
func TestParseCmd_NoRemapOrCommandsSection_NormalizesToEmptyNotNil(t *testing.T) {
	src := []byte(`[Defaults]
command.time = 10
`)

	file, err := ParseCmd(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if file.Remap == nil {
		t.Error("expected Remap to be normalized to a non-nil empty map, got nil")
	}
	if file.Commands == nil {
		t.Error("expected Commands to be normalized to a non-nil empty slice, got nil")
	}
	if file.States == nil {
		t.Error("expected States to be normalized to a non-nil empty slice, got nil")
	}
}

// TestParseCmd_EmptyInput_ReturnsNormalizedZeroValue mirrors
// cmd.Parse's own "empty input is a valid, empty result" contract, still
// normalized for the JSON boundary.
func TestParseCmd_EmptyInput_ReturnsNormalizedZeroValue(t *testing.T) {
	file, err := ParseCmd(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Remap) != 0 || len(file.Commands) != 0 || len(file.States) != 0 {
		t.Errorf("expected an empty CommandFile, got %+v", file)
	}
}

// TestParseCmd_MalformedInput_ReturnsDescriptiveError verifies a malformed
// .cmd (an unclosed attempt at this package's own [Command] header) returns
// a descriptive error rather than panicking — the error path the WASM
// entrypoint's own never-throws guarantee relies on.
func TestParseCmd_MalformedInput_ReturnsDescriptiveError(t *testing.T) {
	src := []byte("[Command\nname = \"a\"\n")

	_, err := ParseCmd(src)
	if err == nil {
		t.Fatal("expected an error for a malformed [Command] header, got nil")
	}
	if !strings.Contains(err.Error(), "cmd:") {
		t.Errorf("expected error to mention the cmd package's own diagnostic, got: %v", err)
	}
}

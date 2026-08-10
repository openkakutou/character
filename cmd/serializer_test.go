package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/openkakutou/character/cns"
)

func TestSerialize_WithValues_ProducesReparsableOutput(t *testing.T) {
	file := CommandFile{
		Remap: map[string]string{"a": "x", "b": "y"},
		Defaults: CommandDefaults{
			Time:       15,
			BufferTime: 1,
		},
		Commands: []Command{
			{Name: "a", Input: "a", Time: 1},
			{Name: "QCF_a", Input: "~D, DF, F, a"},
		},
		States: []cns.StateDef{
			{
				Number: -1,
				Controllers: []cns.Controller{
					{
						Type:       "ChangeState",
						Triggers:   []string{`command = "QCF_a"`},
						Parameters: map[string]string{"value": "1000"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, file); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parsing serialized output: %v\noutput:\n%s", err, buf.String())
	}

	if len(got.Remap) != 2 || got.Remap["a"] != "x" || got.Remap["b"] != "y" {
		t.Errorf("expected Remap to round-trip, got %v", got.Remap)
	}
	if got.Defaults != file.Defaults {
		t.Errorf("expected Defaults %+v to round-trip, got %+v", file.Defaults, got.Defaults)
	}
	if len(got.Commands) != 2 {
		t.Fatalf("expected 2 Commands to round-trip, got %d", len(got.Commands))
	}
	if got.Commands[0].Name != "a" || got.Commands[0].Input != "a" || got.Commands[0].Time != 1 {
		t.Errorf("expected Commands[0] to round-trip, got %+v", got.Commands[0])
	}
	if got.Commands[1].Name != "QCF_a" || got.Commands[1].Input != "~D, DF, F, a" {
		t.Errorf("expected Commands[1] to round-trip, got %+v", got.Commands[1])
	}
	if len(got.States) != 1 || got.States[0].Number != -1 {
		t.Fatalf("expected 1 State (Statedef -1) to round-trip, got %+v", got.States)
	}
	if len(got.States[0].Controllers) != 1 || got.States[0].Controllers[0].Type != "ChangeState" {
		t.Fatalf("expected the linked Controller to round-trip, got %+v", got.States[0].Controllers)
	}
}

func TestSerialize_ZeroValueCommandFile_ProducesEmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Serialize(&buf, CommandFile{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("expected empty output for a zero-value CommandFile, got %q", buf.String())
	}
}

func TestSerialize_ZeroValueDefaults_OmitsDefaultsSection(t *testing.T) {
	file := CommandFile{
		Commands: []Command{{Name: "a", Input: "a"}},
	}
	var buf bytes.Buffer
	if err := Serialize(&buf, file); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(strings.ToLower(buf.String()), "[defaults]") {
		t.Errorf("expected no [Defaults] section for zero-value Defaults, got:\n%s", buf.String())
	}
}

func TestSerialize_UnsetCommandTimeOverrides_AreOmittedNotWrittenAsZero(t *testing.T) {
	// A Command that never overrides Time/BufferTime must not have "time =
	// 0"/"buffer.time = 0" written — that would round-trip back into an
	// explicit zero override rather than staying "unset" (falls back to
	// Defaults).
	file := CommandFile{
		Commands: []Command{{Name: "a", Input: "a"}},
	}
	var buf bytes.Buffer
	if err := Serialize(&buf, file); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.ToLower(buf.String())
	if strings.Contains(out, "time") {
		t.Errorf("expected no time/buffer.time key for an unset Command override, got:\n%s", buf.String())
	}
}

func TestSerialize_NoRemap_OmitsRemapSection(t *testing.T) {
	file := CommandFile{Commands: []Command{{Name: "a", Input: "a"}}}
	var buf bytes.Buffer
	if err := Serialize(&buf, file); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(strings.ToLower(buf.String()), "[remap]") {
		t.Errorf("expected no [Remap] section when Remap is empty, got:\n%s", buf.String())
	}
}

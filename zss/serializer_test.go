package zss

import (
	"strings"
	"testing"
)

func TestSerialize_EmptyScript_WritesNothing(t *testing.T) {
	var buf strings.Builder
	if err := Serialize(&buf, Script{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestSerialize_StatedefBlock_WritesSortedHeaderParamsAndBody(t *testing.T) {
	script := Script{
		Blocks: []Block{
			{
				Kind:   BlockKindStatedef,
				Number: 200,
				HeaderParams: map[string]string{
					"type":    "S",
					"physics": "N",
					"anim":    "200",
				},
				Body: "if AnimElem = 1 {\n\tmapAdd{map:\"chain_5a\";value:1}\n}\n",
			},
		},
	}

	var buf strings.Builder
	if err := Serialize(&buf, script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "[Statedef 200; anim: 200; physics: N; type: S]\n" +
		"if AnimElem = 1 {\n\tmapAdd{map:\"chain_5a\";value:1}\n}\n"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

func TestSerialize_StatedefBlock_NoHeaderParams_WritesBareHeader(t *testing.T) {
	script := Script{Blocks: []Block{{Kind: BlockKindStatedef, Number: 0, Body: "noop{}\n"}}}

	var buf strings.Builder
	if err := Serialize(&buf, script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[Statedef 0]\nnoop{}\n"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

func TestSerialize_FunctionBlock_WritesParamsAndRet(t *testing.T) {
	script := Script{
		Blocks: []Block{
			{
				Kind:   BlockKindFunction,
				Name:   "AngleConversion",
				Params: []string{"alpha", "beta"},
				Ret:    []string{"xangl", "yangl"},
				Body:   "let xangl = 0;\n",
			},
		},
	}

	var buf strings.Builder
	if err := Serialize(&buf, script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[Function AngleConversion(alpha, beta) xangl, yangl]\nlet xangl = 0;\n"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

func TestSerialize_FunctionBlock_NoParamsNoRet_WritesEmptyParens(t *testing.T) {
	script := Script{Blocks: []Block{{Kind: BlockKindFunction, Name: "AirGuardLand", Body: "noop{}\n"}}}

	var buf strings.Builder
	if err := Serialize(&buf, script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[Function AirGuardLand()]\nnoop{}\n"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

func TestSerialize_Preamble_WrittenBeforeFirstBlock(t *testing.T) {
	script := Script{
		Preamble: "# banner\n",
		Blocks:   []Block{{Kind: BlockKindStatedef, Number: 0, Body: "noop{}\n"}},
	}

	var buf strings.Builder
	if err := Serialize(&buf, script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "# banner\n[Statedef 0]\nnoop{}\n"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

// Serialize's output must always reparse into an equivalent Script — the
// round-trip contract Parse/Serialize give every other format in this repo
// (air, def, cns, cmd).
func TestSerialize_RoundTripsThroughParse(t *testing.T) {
	original := Script{
		Preamble: "# banner\n",
		Blocks: []Block{
			{
				Kind:         BlockKindStatedef,
				Number:       200,
				HeaderParams: map[string]string{"type": "S", "physics": "N"},
				Body:         "if AnimElem = 1 {\n\tmapAdd{map:\"chain_5a\";value:1}\n}\n",
			},
			{
				Kind:   BlockKindFunction,
				Name:   "Eff_5a",
				Params: []string{"px", "py"},
				Body:   "helper{stateno: 6721;}\n",
			},
		},
	}

	var buf strings.Builder
	if err := Serialize(&buf, original); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	reparsed, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("unexpected error reparsing: %v", err)
	}

	if reparsed.Preamble != original.Preamble {
		t.Errorf("expected Preamble %q, got %q", original.Preamble, reparsed.Preamble)
	}
	if len(reparsed.Blocks) != len(original.Blocks) {
		t.Fatalf("expected %d blocks, got %d", len(original.Blocks), len(reparsed.Blocks))
	}
	for i := range original.Blocks {
		want, got := original.Blocks[i], reparsed.Blocks[i]
		if want.Kind != got.Kind || want.Number != got.Number || want.Name != got.Name || want.Body != got.Body {
			t.Errorf("block %d: expected %+v, got %+v", i, want, got)
		}
	}
}

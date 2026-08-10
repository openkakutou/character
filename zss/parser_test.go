package zss

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_EmptyInput_ReturnsEmptyScript(t *testing.T) {
	script, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(script.Blocks))
	}
	if script.Preamble != "" {
		t.Errorf("expected empty preamble, got %q", script.Preamble)
	}
}

func TestParse_StatedefBlock_ParsesNumberAndHeaderParams(t *testing.T) {
	src := `[Statedef 200; type: S; physics: N; movetype:A; sprpriority: 2; anim: 200;velSet:0,0;ctrl: 0;]
if AnimElem = 1 {
	mapAdd{map:"chain_5a";value:1}
}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(script.Blocks))
	}

	b := script.Blocks[0]
	if b.Kind != BlockKindStatedef {
		t.Errorf("expected Kind %q, got %q", BlockKindStatedef, b.Kind)
	}
	if b.Number != 200 {
		t.Errorf("expected Number 200, got %d", b.Number)
	}
	wantParams := map[string]string{
		"type":        "S",
		"physics":     "N",
		"movetype":    "A",
		"sprpriority": "2",
		"anim":        "200",
		"velset":      "0,0",
		"ctrl":        "0",
	}
	if len(b.HeaderParams) != len(wantParams) {
		t.Fatalf("expected %d header params, got %d: %v", len(wantParams), len(b.HeaderParams), b.HeaderParams)
	}
	for k, v := range wantParams {
		if got := b.HeaderParams[k]; got != v {
			t.Errorf("header param %q: expected %q, got %q", k, v, got)
		}
	}

	wantBody := "if AnimElem = 1 {\n\tmapAdd{map:\"chain_5a\";value:1}\n}\n"
	if b.Body != wantBody {
		t.Errorf("expected body %q, got %q", wantBody, b.Body)
	}
}

func TestParse_FunctionBlock_ParsesNameParamsAndRet(t *testing.T) {
	src := `[Function AngleConversion(alpha,beta,gamma,delta) xangl,yangl,zangl]
let xangl = 0;
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(script.Blocks))
	}

	b := script.Blocks[0]
	if b.Kind != BlockKindFunction {
		t.Errorf("expected Kind %q, got %q", BlockKindFunction, b.Kind)
	}
	if b.Name != "AngleConversion" {
		t.Errorf("expected Name %q, got %q", "AngleConversion", b.Name)
	}
	wantParams := []string{"alpha", "beta", "gamma", "delta"}
	if strings.Join(b.Params, ",") != strings.Join(wantParams, ",") {
		t.Errorf("expected Params %v, got %v", wantParams, b.Params)
	}
	wantRet := []string{"xangl", "yangl", "zangl"}
	if strings.Join(b.Ret, ",") != strings.Join(wantRet, ",") {
		t.Errorf("expected Ret %v, got %v", wantRet, b.Ret)
	}
}

func TestParse_FunctionBlock_NoParamsNoRet_ParsesEmpty(t *testing.T) {
	src := `[Function AirGuardLand()]
helper{stateno: 6721;}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := script.Blocks[0]
	if b.Name != "AirGuardLand" {
		t.Errorf("expected Name %q, got %q", "AirGuardLand", b.Name)
	}
	if len(b.Params) != 0 {
		t.Errorf("expected no Params, got %v", b.Params)
	}
	if len(b.Ret) != 0 {
		t.Errorf("expected no Ret, got %v", b.Ret)
	}
}

func TestParse_ContentBeforeFirstBlock_PreservedAsPreamble(t *testing.T) {
	src := `# Zantei State Script
# Syntax highlighter for Notepad++
#-------------------------------------------------------------------------------
[StateDef 195; type: S; ctrl: 1;]
if time = 0 {
	changeState{value: 0;}
}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantPreamble := "# Zantei State Script\n# Syntax highlighter for Notepad++\n#-------------------------------------------------------------------------------\n"
	if script.Preamble != wantPreamble {
		t.Errorf("expected preamble %q, got %q", wantPreamble, script.Preamble)
	}
	if len(script.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(script.Blocks))
	}
}

func TestParse_CaseInsensitiveKeywords_StateDefAndLowercaseFunction(t *testing.T) {
	src := `[StateDef 20; type: S; physics: S; sprPriority: 0;]
noop{}
[function changeStates()]
noop{}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(script.Blocks))
	}
	if script.Blocks[0].Kind != BlockKindStatedef || script.Blocks[0].Number != 20 {
		t.Errorf("expected first block to be Statedef 20, got %+v", script.Blocks[0])
	}
	if script.Blocks[1].Kind != BlockKindFunction || script.Blocks[1].Name != "changeStates" {
		t.Errorf("expected second block to be Function changeStates, got %+v", script.Blocks[1])
	}
}

func TestParse_MultipleBlocks_BodyStopsAtNextHeader(t *testing.T) {
	src := `[Function CancelNormal() ret]
let ret = 0;
if (stateNo = 100 && AnimElem = 3,>= 2) {
	let ret = 1;
}
[Function CancelSpecial() ret]
let ret = 0;
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(script.Blocks))
	}
	firstBody := script.Blocks[0].Body
	if strings.Contains(firstBody, "CancelSpecial") {
		t.Errorf("expected first block's body to stop before the second header, got %q", firstBody)
	}
	wantFirstBody := "let ret = 0;\nif (stateNo = 100 && AnimElem = 3,>= 2) {\n\tlet ret = 1;\n}\n"
	if firstBody != wantFirstBody {
		t.Errorf("expected first body %q, got %q", wantFirstBody, firstBody)
	}
}

func TestParse_MalformedBracketLine_ReturnsLineNumberedError(t *testing.T) {
	src := `[Statedef 200; type: S;]
noop{}
[NotAKnownHeader something]
`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the unrecognized bracket line, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to identify line 3, got: %v", err)
	}
}

// Real .zss files (e.g. Melty Blood Type Lumina's "Akiha_TL", surfaced via a
// corpus scan while implementing this package) sometimes wrap a Statedef
// header's semicolon-separated parameters across several physical lines,
// closing the bracket only on a later line — real Ikemen GO engines
// tolerate this. Parse now recovers it, producing the same Block data a
// single-line header would have.
func TestParse_StatedefHeaderSpanningMultipleLines_ParsesAsOneHeader(t *testing.T) {
	src := `[StateDef 801;
type: S; movetype: A; physics: N; VelSet: 0,0;sprpriority: 0;
anim: 801; poweradd: 0;ctrl: 0;]
if AnimElem = 1 {
	targetPowerAdd{value:25}
}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(script.Blocks))
	}
	b := script.Blocks[0]
	if b.Number != 801 {
		t.Errorf("expected Number 801, got %d", b.Number)
	}
	wantParams := map[string]string{
		"type": "S", "movetype": "A", "physics": "N", "velset": "0,0",
		"sprpriority": "0", "anim": "801", "poweradd": "0", "ctrl": "0",
	}
	for k, v := range wantParams {
		if got := b.HeaderParams[k]; got != v {
			t.Errorf("header param %q: expected %q, got %q", k, v, got)
		}
	}
	if !strings.Contains(b.Body, "targetPowerAdd") {
		t.Errorf("expected body to contain the block's own content, got %q", b.Body)
	}
}

func TestParse_UnterminatedHeader_ReturnsLineNumberedError(t *testing.T) {
	src := `[Statedef 200; type: S;]
noop{}
[Statedef 300; type: S
still no closing bracket
`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the unterminated header, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to identify line 3 (the header's start), got: %v", err)
	}
}

func TestParse_ReaderError_ReturnsWrappedError(t *testing.T) {
	_, err := Parse(&failingReader{})
	if err == nil {
		t.Fatal("expected an error from a failing reader, got nil")
	}
	var target *failingReaderError
	if !errors.As(err, &target) {
		t.Errorf("expected wrapped failingReaderError, got: %v", err)
	}
}

// failingReaderError is a sentinel error type used to verify Parse wraps
// (rather than discards) the underlying reader's error.
type failingReaderError struct{}

func (e *failingReaderError) Error() string { return "simulated read failure" }

// failingReader always fails on Read.
type failingReader struct{}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, &failingReaderError{}
}

func TestParse_StatedefWithoutHeaderParams_ParsesWithEmptyParams(t *testing.T) {
	src := `[Statedef 0]
noop{}
`
	script, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := script.Blocks[0]
	if b.Number != 0 {
		t.Errorf("expected Number 0, got %d", b.Number)
	}
	if len(b.HeaderParams) != 0 {
		t.Errorf("expected no header params, got %v", b.HeaderParams)
	}
}

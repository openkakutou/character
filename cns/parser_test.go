package cns

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParse_SampleFile_ProducesExpectedStateDefsAndControllers(t *testing.T) {
	f, err := os.Open("testdata/sample.cns")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	states, err := Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(states) != 3 {
		t.Fatalf("expected 3 StateDefs, got %d", len(states))
	}

	// [Statedef 0, Standing]
	s0 := states[0]
	if s0.Number != 0 {
		t.Errorf("expected Number 0, got %d", s0.Number)
	}
	if s0.Type != StateTypeStanding {
		t.Errorf("expected Type %q, got %q", StateTypeStanding, s0.Type)
	}
	if s0.MoveType != MoveTypeIdle {
		t.Errorf("expected MoveType %q, got %q", MoveTypeIdle, s0.MoveType)
	}
	if s0.Physics != PhysicsStanding {
		t.Errorf("expected Physics %q, got %q", PhysicsStanding, s0.Physics)
	}
	if s0.Anim != 0 {
		t.Errorf("expected Anim 0, got %d", s0.Anim)
	}
	if !s0.Ctrl {
		t.Errorf("expected Ctrl true, got %v", s0.Ctrl)
	}
	if s0.FaceP2 || s0.HitDefPersist || s0.MoveHitPersist || s0.HitCountPersist {
		t.Errorf("expected all persist/facing flags false, got %+v", s0)
	}

	if len(s0.Controllers) != 2 {
		t.Fatalf("expected 2 controllers in Statedef 0, got %d", len(s0.Controllers))
	}

	c0 := s0.Controllers[0]
	if c0.Type != "ChangeState" {
		t.Errorf("expected controller Type %q, got %q", "ChangeState", c0.Type)
	}
	wantTriggers := []string{`Command = "holdback"`, "StateType = S"}
	if len(c0.Triggers) != len(wantTriggers) {
		t.Fatalf("expected %d triggers, got %d: %v", len(wantTriggers), len(c0.Triggers), c0.Triggers)
	}
	for i, want := range wantTriggers {
		if c0.Triggers[i] != want {
			t.Errorf("Triggers[%d]: expected %q, got %q (file order expected)", i, want, c0.Triggers[i])
		}
	}
	if c0.Parameters["value"] != "10" {
		t.Errorf("expected Parameters[\"value\"] = %q, got %q", "10", c0.Parameters["value"])
	}
	if c0.Parameters["ctrl"] != "0" {
		t.Errorf("expected Parameters[\"ctrl\"] = %q, got %q", "0", c0.Parameters["ctrl"])
	}

	c1 := s0.Controllers[1]
	if c1.Type != "VelSet" {
		t.Errorf("expected controller Type %q, got %q", "VelSet", c1.Type)
	}
	if c1.Triggers != nil {
		t.Errorf("expected no triggers (unconditional controller), got %v", c1.Triggers)
	}
	if c1.Parameters["x"] != "0" || c1.Parameters["y"] != "0" {
		t.Errorf("expected Parameters x=0,y=0, got %v", c1.Parameters)
	}

	// [Statedef -1]
	sNeg1 := states[1]
	if sNeg1.Number != -1 {
		t.Errorf("expected Number -1, got %d", sNeg1.Number)
	}
	if len(sNeg1.Controllers) != 1 {
		t.Fatalf("expected 1 controller in Statedef -1, got %d", len(sNeg1.Controllers))
	}
	cHit := sNeg1.Controllers[0]
	if cHit.Type != "HitDef" {
		t.Errorf("expected controller Type %q, got %q", "HitDef", cHit.Type)
	}
	wantHitTriggers := []string{"MoveType != H", `Command = "x"`}
	if len(cHit.Triggers) != len(wantHitTriggers) {
		t.Fatalf("expected %d triggers, got %d: %v", len(wantHitTriggers), len(cHit.Triggers), cHit.Triggers)
	}
	for i, want := range wantHitTriggers {
		if cHit.Triggers[i] != want {
			t.Errorf("Triggers[%d]: expected %q, got %q", i, want, cHit.Triggers[i])
		}
	}
	if cHit.Parameters["attr"] != "S, NA" {
		t.Errorf("expected Parameters[\"attr\"] = %q, got %q", "S, NA", cHit.Parameters["attr"])
	}
	if cHit.Parameters["damage"] != "30" {
		t.Errorf("expected Parameters[\"damage\"] = %q, got %q", "30", cHit.Parameters["damage"])
	}

	// [Statedef 200, Attack] has no [State N] blocks at all.
	s200 := states[2]
	if s200.Number != 200 {
		t.Errorf("expected Number 200, got %d", s200.Number)
	}
	if s200.MoveType != MoveTypeAttack {
		t.Errorf("expected MoveType %q, got %q", MoveTypeAttack, s200.MoveType)
	}
	if s200.Anim != 200 {
		t.Errorf("expected Anim 200, got %d", s200.Anim)
	}
	if s200.Ctrl {
		t.Errorf("expected Ctrl false, got %v", s200.Ctrl)
	}
	if s200.Controllers != nil {
		t.Errorf("expected no controllers, got %v", s200.Controllers)
	}
}

func TestParse_EmptyInput_ReturnsEmptyResultWithoutError(t *testing.T) {
	states, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("expected no StateDefs, got %v", states)
	}
}

func TestParse_UnrecognizedSectionBetweenStatedefs_IsSkippedWithoutError(t *testing.T) {
	src := `[Statedef 0]
type = S

[Clsn1Default]
Clsn1: 1
 Clsn1[0] = 0,0,0,0

[Statedef 1]
type = C
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 StateDefs (unrecognized section skipped), got %d", len(states))
	}
	if states[0].Number != 0 || states[1].Number != 1 {
		t.Errorf("expected Statedefs 0 and 1, got %v, %v", states[0].Number, states[1].Number)
	}
	if states[1].Type != StateTypeCrouching {
		t.Errorf("expected second Statedef Type %q, got %q", StateTypeCrouching, states[1].Type)
	}
}

func TestParse_MalformedSectionHeader_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Statedef 0
type = S
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the malformed section header, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to identify line 1, got: %v", err)
	}
}

func TestParse_MalformedStatedefHeader_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Statedef abc]
type = S
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the malformed Statedef header, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to identify line 1, got: %v", err)
	}
}

// Real MUGEN/Ikemen .cns files widely write "[State <label>]" with no state
// number at all (e.g. "[State removeexplod]") instead of "[State N]" — real
// engines tolerate this. cns.Controller never stored the header's own number
// in the first place (Parse discards it unconditionally), so this parser
// tolerates it too. See .vibe/decisions/022-cns-parse-state-header-accepts-any-label.md.
func TestParse_StateHeaderWithoutNumber_AttachesToEnclosingStatedef(t *testing.T) {
	src := `[Statedef 0]
type = S

[State removeexplod]
type = RemoveExplod
trigger1 = Time = 0
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if len(states[0].Controllers) != 1 {
		t.Fatalf("expected the controller under the label-only [State] header to attach to Statedef 0, got %d controllers", len(states[0].Controllers))
	}
	ctrl := states[0].Controllers[0]
	if ctrl.Type != "RemoveExplod" {
		t.Errorf("expected controller Type %q, got %q", "RemoveExplod", ctrl.Type)
	}
	if len(ctrl.Triggers) != 1 || ctrl.Triggers[0] != "Time = 0" {
		t.Errorf("expected Triggers [\"Time = 0\"], got %v", ctrl.Triggers)
	}
}

// A bracket line that starts with the "state" keyword but has no closing
// bracket is still a genuine error — the label-only relaxation only widens
// what counts as valid content between the brackets, not whether the
// brackets themselves are well-formed (already covered generically by
// TestParse_MalformedSectionHeader_ReturnsErrorNamingTheLine, exercised here
// specifically for a "[State" line to guard the relaxation above).
func TestParse_UnclosedStateHeader_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Statedef 0]
type = S

[State removeexplod
type = RemoveExplod
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the unclosed State header, got nil")
	}
	if !strings.Contains(err.Error(), "line 4") {
		t.Errorf("expected error to identify line 4, got: %v", err)
	}
}

func TestParse_StateBlockOutsideAnyStatedef_ReturnsError(t *testing.T) {
	src := `[State 0]
type = VelSet
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a [State N] block with no enclosing Statedef, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to identify line 1, got: %v", err)
	}
}

func TestParse_MalformedKeyValueLine_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Statedef 0]
type = S
ctrl 1
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the malformed key=value line, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to identify line 3, got: %v", err)
	}
}

func TestParse_InvalidBooleanFieldValue_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Statedef 0]
type = S
ctrl = maybe
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for an invalid boolean value, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to identify line 3, got: %v", err)
	}
}

// Real MUGEN/Ikemen .cns files sometimes give a numeric header field a
// trigger expression instead of a literal integer (e.g.
// "anim = IfElse(ceil(lifemax/2) < life ,181,182)"). Parse cannot tell a
// genuine expression apart from a typo without an expression evaluator this
// codebase deliberately doesn't have (ADR 011), so it stores the raw text
// unevaluated instead of erroring. See
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md.
func TestParse_NumericHeaderFieldWithExpression_StoresRawTextInHeaderExprs(t *testing.T) {
	src := `[Statedef 0]
type = S
anim = IfElse(ceil(lifemax/2) < life ,181,182)
poweradd = ifelse(PrevStateNo = 9000, 0, 20)
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}

	s := states[0]
	if s.Anim != 0 {
		t.Errorf("expected Anim to stay at its zero value when the source is an expression, got %d", s.Anim)
	}
	if got, want := s.HeaderExprs["anim"], "IfElse(ceil(lifemax/2) < life ,181,182)"; got != want {
		t.Errorf("expected HeaderExprs[\"anim\"] = %q, got %q", want, got)
	}
	if s.PowerAdd != 0 {
		t.Errorf("expected PowerAdd to stay at its zero value when the source is an expression, got %d", s.PowerAdd)
	}
	if got, want := s.HeaderExprs["poweradd"], "ifelse(PrevStateNo = 9000, 0, 20)"; got != want {
		t.Errorf("expected HeaderExprs[\"poweradd\"] = %q, got %q", want, got)
	}
}

// A numeric header field's source text that isn't even a plausible
// expression (plain garbage) is stored the same way as a real expression,
// not rejected — Parse has no expression grammar to tell the two apart, so
// it treats anything that fails strconv.Atoi identically. See ADR 023.
func TestParse_NumericHeaderFieldWithGarbageValue_StoresRawTextInsteadOfErroring(t *testing.T) {
	src := `[Statedef 0]
type = S
anim = abc
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if states[0].Anim != 0 {
		t.Errorf("expected Anim to stay at its zero value, got %d", states[0].Anim)
	}
	if got, want := states[0].HeaderExprs["anim"], "abc"; got != want {
		t.Errorf("expected HeaderExprs[\"anim\"] = %q, got %q", want, got)
	}
}

// A literal integer value for a numeric header field is completely
// unaffected by the expression escape hatch: the typed field is set and no
// HeaderExprs entry is created for it.
func TestParse_NumericHeaderFieldWithLiteralInteger_DoesNotPopulateHeaderExprs(t *testing.T) {
	src := `[Statedef 0]
type = S
anim = 200
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if states[0].Anim != 200 {
		t.Errorf("expected Anim 200, got %d", states[0].Anim)
	}
	if _, ok := states[0].HeaderExprs["anim"]; ok {
		t.Errorf("expected no HeaderExprs entry for a literal integer field, got %v", states[0].HeaderExprs)
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

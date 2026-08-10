package cns

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSerialize_SingleStatedefWithController_ProducesExpectedText(t *testing.T) {
	states := []StateDef{
		{
			Number:   0,
			Type:     StateTypeStanding,
			MoveType: MoveTypeIdle,
			Physics:  PhysicsStanding,
			Ctrl:     true,
			Controllers: []Controller{
				{
					Type:       "ChangeState",
					Triggers:   []string{"Time = 0", "Command = \"holdback\""},
					Parameters: map[string]string{"value": "10"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, states); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `[Statedef 0]
type = S
movetype = I
physics = S
anim = 0
ctrl = 1
poweradd = 0
juggle = 0
facep2 = 0
hitdefpersist = 0
movehitpersist = 0
hitcountpersist = 0
sprpriority = 0

[State 0]
type = ChangeState
trigger1 = Time = 0
trigger2 = Command = "holdback"
value = 10

`
	if buf.String() != want {
		t.Errorf("unexpected output.\n--- want ---\n%s\n--- got ---\n%s", want, buf.String())
	}
}

func TestSerialize_RoundTripsThroughParse_ProducesEquivalentStateDefs(t *testing.T) {
	states := []StateDef{
		{
			Number:          0,
			Type:            StateTypeStanding,
			MoveType:        MoveTypeIdle,
			Physics:         PhysicsStanding,
			Anim:            0,
			Ctrl:            true,
			PowerAdd:        5,
			Juggle:          2,
			FaceP2:          true,
			HitDefPersist:   true,
			MoveHitPersist:  false,
			HitCountPersist: true,
			SprPriority:     3,
			Controllers: []Controller{
				{
					Type:     "ChangeState",
					Triggers: []string{`Command = "holdback"`, "StateType = S"},
					Parameters: map[string]string{
						"value": "10",
						"ctrl":  "0",
					},
				},
				{
					// Unconditional controller: no triggers at all.
					Type:       "VelSet",
					Parameters: map[string]string{"x": "0", "y": "0"},
				},
			},
		},
		{
			// Negative state number, no controllers at all.
			Number:   -1,
			Type:     StateTypeAir,
			MoveType: MoveTypeHit,
			Physics:  PhysicsAir,
		},
		{
			Number:   200,
			Type:     StateTypeStanding,
			MoveType: MoveTypeAttack,
			Physics:  PhysicsStanding,
			Anim:     200,
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, states); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	reparsed, err := Parse(&buf)
	if err != nil {
		t.Fatalf("unexpected error re-parsing serialized output: %v\n--- output ---\n%s", err, buf.String())
	}

	if !reflect.DeepEqual(reparsed, states) {
		t.Errorf("round trip did not produce an equivalent structure.\n--- original ---\n%+v\n--- reparsed ---\n%+v", states, reparsed)
	}
}

// A StateDef carrying a numeric header field's unevaluated expression text
// (HeaderExprs) round-trips through Serialize/Parse: the raw expression is
// written out verbatim in place of the typed field's formatted value, and
// re-parsing recovers the same HeaderExprs entry with the typed field left
// at its zero value. See
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md.
func TestSerialize_NumericHeaderFieldExpression_RoundTripsThroughParse(t *testing.T) {
	states := []StateDef{
		{
			Number: 0,
			Type:   StateTypeStanding,
			HeaderExprs: map[string]string{
				"anim":     "IfElse(ceil(lifemax/2) < life ,181,182)",
				"poweradd": "ifelse(PrevStateNo = 9000, 0, 20)",
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, states); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	if !strings.Contains(buf.String(), "anim = IfElse(ceil(lifemax/2) < life ,181,182)") {
		t.Errorf("expected the raw anim expression to be written verbatim, got:\n%s", buf.String())
	}

	reparsed, err := Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error re-parsing serialized output: %v\n--- output ---\n%s", err, buf.String())
	}
	if !reflect.DeepEqual(reparsed, states) {
		t.Errorf("round trip did not produce an equivalent structure.\n--- original ---\n%+v\n--- reparsed ---\n%+v", states, reparsed)
	}
}

// A StateDef carrying a boolean header field's unevaluated expression text
// (HeaderExprs) round-trips through Serialize/Parse the same way a numeric
// one does: the raw expression is written out verbatim in place of the
// typed field's "1"/"0" formatted value, and re-parsing recovers the same
// HeaderExprs entry with the typed field left at its zero value (false).
// Mirrors TestSerialize_NumericHeaderFieldExpression_RoundTripsThroughParse.
// See item 046.
func TestSerialize_BooleanHeaderFieldExpression_RoundTripsThroughParse(t *testing.T) {
	states := []StateDef{
		{
			Number: 9000,
			Type:   StateTypeStanding,
			HeaderExprs: map[string]string{
				"ctrl":   "0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))",
				"facep2": "1-(prevstateno=[100,119])",
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, states); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	if !strings.Contains(buf.String(), "ctrl = 0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))") {
		t.Errorf("expected the raw ctrl expression to be written verbatim, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "facep2 = 1-(prevstateno=[100,119])") {
		t.Errorf("expected the raw facep2 expression to be written verbatim, got:\n%s", buf.String())
	}

	reparsed, err := Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error re-parsing serialized output: %v\n--- output ---\n%s", err, buf.String())
	}
	if !reflect.DeepEqual(reparsed, states) {
		t.Errorf("round trip did not produce an equivalent structure.\n--- original ---\n%+v\n--- reparsed ---\n%+v", states, reparsed)
	}
}

func TestSerialize_StatedefWithNoControllers_OmitsStateSections(t *testing.T) {
	states := []StateDef{
		{Number: 5, Type: StateTypeStanding},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, states); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(buf.String(), "[State ") {
		t.Errorf("expected no [State N] section for a Statedef with no controllers, got:\n%s", buf.String())
	}
}

func TestSerialize_EmptyInput_ProducesEmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Serialize(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for no StateDefs, got:\n%s", buf.String())
	}
}

func TestSerialize_WriterError_ReturnsWrappedError(t *testing.T) {
	states := []StateDef{{Number: 0, Type: StateTypeStanding}}

	err := Serialize(&failingWriter{}, states)
	if err == nil {
		t.Fatal("expected an error from a failing writer, got nil")
	}
	var target *failingWriterError
	if !errors.As(err, &target) {
		t.Errorf("expected wrapped failingWriterError, got: %v", err)
	}
}

// failingWriterError is a sentinel error type used to verify Serialize wraps
// (rather than discards) the underlying writer's error.
type failingWriterError struct{}

func (e *failingWriterError) Error() string { return "simulated write failure" }

// failingWriter always fails on Write.
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (int, error) {
	return 0, &failingWriterError{}
}

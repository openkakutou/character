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

// A bracket line missing its closing "]" that isn't a recognizable
// Statedef/State header attempt (no "statedef"/"state" keyword) is genuinely
// unrelated content, not a typo in either of this package's own two known
// headers — it is skipped without erroring, the same way a *closed*
// unrecognized bracket line (e.g. "[foo]") already is. See backlog item 053.
func TestParse_UnclosedBracketLineNotAHeaderAttempt_IsSkippedWithoutError(t *testing.T) {
	src := `[Clsn1Default
Clsn1: 1
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("expected no StateDefs (no Statedef header present), got %d", len(states))
	}
}

// The real-world case backlog item 053 was filed from: King of Fighters'
// "Shun Ei" character's Command.cmd has a lone "[" on its own line, used
// purely as a section-banner decoration (surrounded by ";===..." comment
// lines), with no closing "]" anywhere nearby, ahead of the file's
// "[Statedef -1]" block. Real MUGEN/Ikemen engines tolerate this; it must
// not abort parsing of the Statedef that follows it. (cns.Parse runs
// against a .cmd file's entire source — see cmd.Parse — so it encounters
// banner lines like this one even though they aren't its own concern.)
func TestParse_LoneUnclosedBracketBannerLine_IsSkippedAndParsingContinues(t *testing.T) {
	src := `;===============================================================
[
;===============================================================

[Statedef -1]
[State -1]
type = ChangeState
trigger1 = command = "holdback"
value = 100
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if len(states[0].Controllers) != 1 {
		t.Fatalf("expected 1 controller in the Statedef following the decorative banner line, got %d", len(states[0].Controllers))
	}
	if got := states[0].Controllers[0].Type; got != "ChangeState" {
		t.Errorf("expected controller Type %q, got %q", "ChangeState", got)
	}
}

// Real MUGEN/Ikemen .cns files widely omit a "[Statedef N]" header's closing
// "]" (a typo real engines tolerate) — a full corpus scan found this is the
// single most common character.Load failure (~15% of a 717-file corpus).
// Parse now recovers it, producing the same StateDef data a correctly
// bracketed header would have.
func TestParse_StatedefHeaderMissingClosingBracket_RecoversSameAsClosedHeader(t *testing.T) {
	src := `[Statedef 0, Standing
type = S
movetype = I
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if states[0].Number != 0 {
		t.Errorf("expected Number 0, got %d", states[0].Number)
	}
	if states[0].Type != StateTypeStanding {
		t.Errorf("expected Type %q, got %q", StateTypeStanding, states[0].Type)
	}
	if states[0].MoveType != MoveTypeIdle {
		t.Errorf("expected MoveType %q, got %q", MoveTypeIdle, states[0].MoveType)
	}
}

// The real-world case backlog item 042 was filed from: a real character file
// ("Bardock", surfaced via character-viewer-web) has "[State 110, 1"
// (missing "]") instead of "[State 110, 1]", and the whole file failed to
// load because of it. Also matches a second real fixture (Capcom's
// "Commando (CVS)": "[State -2, Blocking-12" in cvscommando.cns) — same root
// cause, not scoped to one character's file.
func TestParse_StateHeaderMissingClosingBracket_RecoversAndAttachesController(t *testing.T) {
	src := `[Statedef 110]
type = S

[State 110, 1
type = ChangeState
trigger1 = Time = 0
value = 100
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if len(states[0].Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(states[0].Controllers))
	}
	c := states[0].Controllers[0]
	if c.Type != "ChangeState" {
		t.Errorf("expected controller Type %q, got %q", "ChangeState", c.Type)
	}
	if len(c.Triggers) != 1 || c.Triggers[0] != "Time = 0" {
		t.Errorf("expected triggers [%q], got %v", "Time = 0", c.Triggers)
	}
	if c.Parameters["value"] != "100" {
		t.Errorf("expected Parameters[\"value\"] = %q, got %q", "100", c.Parameters["value"])
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

// A bare-label "[State <label>]" header (no state number, see
// .vibe/decisions/022) missing its closing "]" is recovered the same way a
// numbered "[State N]" header is (backlog item 042) — the label shape isn't
// special-cased differently by the recovery.
func TestParse_UnclosedStateHeaderWithoutNumber_RecoversAndAttachesController(t *testing.T) {
	src := `[Statedef 0]
type = S

[State removeexplod
type = RemoveExplod
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 || len(states[0].Controllers) != 1 {
		t.Fatalf("expected 1 StateDef with 1 controller, got %+v", states)
	}
	if got := states[0].Controllers[0].Type; got != "RemoveExplod" {
		t.Errorf("expected controller Type %q, got %q", "RemoveExplod", got)
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

// A non-header content line inside a Statedef/State block that has no "="
// (and thus isn't a valid key=value pair) used to be a hard error. Item 043's
// corpus scan found this is the largest failure class after item 042 (49 of
// 717 real characters): a truncated leftover key, a decorative separator, or
// a comment missing its leading ";" all produce such a line, and real
// engines tolerate it the same way Parse already tolerates an unrecognized
// key. Parse now ignores it instead, and keeps reading the rest of the block
// normally.
func TestParse_NonKeyValueLineInsideBlock_IsIgnoredAndParsingContinues(t *testing.T) {
	src := `[Statedef 0]
type = S
ctrl 1

[State 0]
type = VelSet
getpower
trigger1 = 1
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if states[0].Type != StateTypeStanding {
		t.Errorf("expected Type %q, got %q", StateTypeStanding, states[0].Type)
	}
	if len(states[0].Controllers) != 1 {
		t.Fatalf("expected 1 Controller, got %d", len(states[0].Controllers))
	}
	ctrl := states[0].Controllers[0]
	if ctrl.Type != "VelSet" {
		t.Errorf("expected Controller Type %q, got %q", "VelSet", ctrl.Type)
	}
	if len(ctrl.Triggers) != 1 || ctrl.Triggers[0] != "1" {
		t.Errorf("expected Triggers [%q], got %v — the line after the ignored one was not parsed", "1", ctrl.Triggers)
	}
}

// The real-world case backlog item 043 was filed from: a real character file
// ("Aino 2", from Arcana Heart, surfaced via a 717-file corpus scan) has a
// bare "getpower" line (no "=") inside a [State ...] block, and the whole
// file failed to load because of it.
func TestParse_BareWordLineInsideStateBlock_IsIgnored(t *testing.T) {
	src := `[Statedef 0]
type = S

[State 0]
type = VarSet
getpower
v = 50
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctrl := states[0].Controllers[0]
	if ctrl.Parameters["v"] != "50" {
		t.Errorf("expected Parameters[%q] = %q, got %q", "v", "50", ctrl.Parameters["v"])
	}
}

// Real MUGEN/Ikemen .cns files sometimes give a boolean header field a
// trigger expression instead of a literal true/false/0/1 (e.g. Mai's
// "ctrl = 0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))", Avdol's
// "facep2 = 1-(prevstateno=[100,119])"). Parse cannot tell a genuine
// expression apart from a typo without an expression evaluator this
// codebase deliberately doesn't have (ADR 011), so — mirroring the numeric
// header fields' own escape hatch (ADR 023) — it stores the raw text
// unevaluated instead of erroring. See item 046.
func TestParse_BooleanHeaderFieldWithExpression_StoresRawTextInHeaderExprs(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		get   func(StateDef) bool
	}{
		{"ctrl (real Mai.cns line)", "ctrl", `0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))`, func(s StateDef) bool { return s.Ctrl }},
		{"facep2 (real avdul.cns line)", "facep2", `1-(prevstateno=[100,119])`, func(s StateDef) bool { return s.FaceP2 }},
		{"hitdefpersist", "hitdefpersist", `ifelse(time=0,1,0)`, func(s StateDef) bool { return s.HitDefPersist }},
		{"movehitpersist", "movehitpersist", `var(10)=1`, func(s StateDef) bool { return s.MoveHitPersist }},
		{"hitcountpersist", "hitcountpersist", `stateno=200`, func(s StateDef) bool { return s.HitCountPersist }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "[Statedef 0]\ntype = S\n" + tt.key + " = " + tt.value + "\n"

			states, err := Parse(strings.NewReader(src))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(states) != 1 {
				t.Fatalf("expected 1 StateDef, got %d", len(states))
			}

			s := states[0]
			if got := tt.get(s); got {
				t.Errorf("expected the typed field to stay at its zero value (false) when the source is an expression, got %v", got)
			}
			if got, want := s.HeaderExprs[tt.key], tt.value; got != want {
				t.Errorf("expected HeaderExprs[%q] = %q, got %q", tt.key, want, got)
			}
		})
	}
}

// A boolean header field's source text that isn't even a plausible
// expression (plain garbage) is stored the same way as a real expression,
// not rejected — Parse has no expression grammar to tell the two apart, so
// it treats anything that fails strconv.ParseBool identically. Mirrors
// TestParse_NumericHeaderFieldWithGarbageValue_StoresRawTextInsteadOfErroring.
func TestParse_BooleanHeaderFieldWithGarbageValue_StoresRawTextInsteadOfErroring(t *testing.T) {
	src := `[Statedef 0]
type = S
ctrl = maybe
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 StateDef, got %d", len(states))
	}
	if states[0].Ctrl {
		t.Errorf("expected Ctrl to stay at its zero value (false), got %v", states[0].Ctrl)
	}
	if got, want := states[0].HeaderExprs["ctrl"], "maybe"; got != want {
		t.Errorf("expected HeaderExprs[\"ctrl\"] = %q, got %q", want, got)
	}
}

// A literal boolean value for a boolean header field is completely
// unaffected by the expression escape hatch: the typed field is set and no
// HeaderExprs entry is created for it. Mirrors
// TestParse_NumericHeaderFieldWithLiteralInteger_DoesNotPopulateHeaderExprs.
func TestParse_BooleanHeaderFieldWithLiteralValue_DoesNotPopulateHeaderExprs(t *testing.T) {
	src := `[Statedef 0]
type = S
ctrl = 1
facep2 = 0
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := states[0]
	if !s.Ctrl {
		t.Errorf("expected Ctrl true, got %v", s.Ctrl)
	}
	if s.FaceP2 {
		t.Errorf("expected FaceP2 false, got %v", s.FaceP2)
	}
	if _, ok := s.HeaderExprs["ctrl"]; ok {
		t.Errorf("expected no HeaderExprs entry for a literal boolean field, got %v", s.HeaderExprs)
	}
	if _, ok := s.HeaderExprs["facep2"]; ok {
		t.Errorf("expected no HeaderExprs entry for a literal boolean field, got %v", s.HeaderExprs)
	}
}

// Reproduces the exact real-world failures reported in backlog item 046:
// Mai (BlazBlue) fails on a ctrl expression, Avdol (JoJo's Bizarre
// Adventure) fails on a facep2 expression — both real community character
// files.
func TestParse_RealMaiAndAvdolBooleanHeaderExpressions_ParsedSuccessfully(t *testing.T) {
	src := `[Statedef 9000]
type = A
movetype = I
physics = N
anim = 9999
ctrl = 0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))

[Statedef 200]
type = S
movetype = A
physics = N
ctrl = 0
anim = ifelse((prevstateno=[100,105]),210,200)
poweradd = 0
sprpriority = 1-(var(0)>0)*2
facep2 = 1-(prevstateno=[100,119])
`

	states, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error parsing real Mai/Avdol boolean header expressions: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 StateDefs, got %d", len(states))
	}

	mai := states[0]
	if got, want := mai.HeaderExprs["ctrl"], "0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))"; got != want {
		t.Errorf("expected Mai HeaderExprs[\"ctrl\"] = %q, got %q", want, got)
	}

	avdol := states[1]
	if got, want := avdol.HeaderExprs["facep2"], "1-(prevstateno=[100,119])"; got != want {
		t.Errorf("expected Avdol HeaderExprs[\"facep2\"] = %q, got %q", want, got)
	}
	// Avdol's ctrl and sprpriority are also expressions on the same block —
	// confirm they coexist fine alongside a boolean-field expression.
	if got, want := avdol.HeaderExprs["sprpriority"], "1-(var(0)>0)*2"; got != want {
		t.Errorf("expected Avdol HeaderExprs[\"sprpriority\"] = %q, got %q", want, got)
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

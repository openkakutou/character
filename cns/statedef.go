// Package cns defines the pure-data model for MUGEN/Ikemen GO combat logic
// (.cns) files: StateDef and Controller.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (editor, engine) works with. It carries no INI/expression parsing, file
// I/O, or write-only (format-preservation) logic; per CLAUDE.md's read/write
// separation constraint, that lives elsewhere so importing this data model
// alone never pulls in write-only dependencies.
//
// This is deliberately minimal scaffolding, not a full CNS expression
// engine: a Controller's trigger conditions and parameters are stored as
// unevaluated data (plain strings and a string-to-string map), not resolved
// or type-checked against MUGEN/Ikemen's trigger expression language. See
// .vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md.
// StateDef's own typed numeric header fields carry the same escape hatch,
// one layer up: HeaderExprs holds the raw source text for one of those
// fields whenever it held a trigger expression rather than a literal
// integer. See
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md.
package cns

// StateType is a .cns [Statedef N] block's "type" parameter: the state
// classification MUGEN/Ikemen uses to pick default behavior (e.g. gravity)
// when a controller doesn't explicitly override it.
type StateType string

const (
	// StateTypeStanding is a standing state ("S").
	StateTypeStanding StateType = "S"
	// StateTypeCrouching is a crouching state ("C").
	StateTypeCrouching StateType = "C"
	// StateTypeAir is an airborne state ("A").
	StateTypeAir StateType = "A"
	// StateTypeLiedown is a lying-down state ("L").
	StateTypeLiedown StateType = "L"
	// StateTypeUnchanged leaves the character's current state type
	// unchanged when entering this state ("U").
	StateTypeUnchanged StateType = "U"
)

// MoveType is a .cns [Statedef N] block's "movetype" parameter: whether
// this state represents an attack, an idle/neutral action, or a hit
// reaction, used by other characters' trigger conditions (e.g. "MoveType =
// H" to detect being hit).
type MoveType string

const (
	// MoveTypeAttack is an attacking state ("A").
	MoveTypeAttack MoveType = "A"
	// MoveTypeIdle is an idle/neutral state ("I").
	MoveTypeIdle MoveType = "I"
	// MoveTypeHit is a hit-reaction state ("H").
	MoveTypeHit MoveType = "H"
	// MoveTypeUnchanged leaves the character's current move type
	// unchanged when entering this state ("U").
	MoveTypeUnchanged MoveType = "U"
)

// PhysicsType is a .cns [Statedef N] block's "physics" parameter: which
// built-in physics (gravity, friction, air drag) apply while this state is
// active.
type PhysicsType string

const (
	// PhysicsStanding applies standing physics ("S").
	PhysicsStanding PhysicsType = "S"
	// PhysicsCrouching applies crouching physics ("C").
	PhysicsCrouching PhysicsType = "C"
	// PhysicsAir applies airborne physics ("A").
	PhysicsAir PhysicsType = "A"
	// PhysicsNone applies no built-in physics ("N").
	PhysicsNone PhysicsType = "N"
	// PhysicsUnchanged leaves the character's current physics unchanged
	// when entering this state ("U").
	PhysicsUnchanged PhysicsType = "U"
)

// StateDef is a .cns [Statedef N] block: the state's header parameters plus
// the state controllers ([State N] blocks) that run while it is active.
type StateDef struct {
	// Number is the state number this block defines (the N in
	// [Statedef N]).
	Number int `json:"number"`
	// Type is the state's classification ("type" parameter).
	Type StateType `json:"type"`
	// MoveType is the state's move classification ("movetype" parameter).
	MoveType MoveType `json:"moveType"`
	// Physics is the built-in physics applied while this state is active
	// ("physics" parameter).
	Physics PhysicsType `json:"physics"`
	// Anim is the animation number played on entering this state ("anim"
	// parameter); 0 means "not set", in which case MUGEN/Ikemen defaults
	// it to Number.
	Anim int `json:"anim"`
	// Ctrl reports whether the player has control while this state is
	// active ("ctrl" parameter).
	Ctrl bool `json:"ctrl"`
	// PowerAdd is the power meter gain applied on entering this state
	// ("poweradd" parameter).
	PowerAdd int `json:"powerAdd"`
	// Juggle is the juggle points this state costs to use against an
	// already-airborne opponent ("juggle" parameter).
	Juggle int `json:"juggle"`
	// FaceP2 reports whether the character turns to face the opponent on
	// entering this state ("facep2" parameter).
	FaceP2 bool `json:"faceP2"`
	// HitDefPersist reports whether an active hit definition survives
	// into this state instead of being cleared ("hitdefpersist"
	// parameter).
	HitDefPersist bool `json:"hitDefPersist"`
	// MoveHitPersist reports whether "MoveHit"-triggered conditions
	// survive into this state instead of being cleared
	// ("movehitpersist" parameter).
	MoveHitPersist bool `json:"moveHitPersist"`
	// HitCountPersist reports whether the hit counter survives into this
	// state instead of being reset ("hitcountpersist" parameter).
	HitCountPersist bool `json:"hitCountPersist"`
	// SprPriority is the sprite drawing priority (layering order) for
	// this state ("sprpriority" parameter).
	SprPriority int `json:"sprPriority"`
	// HeaderExprs holds the raw, unevaluated source text of a numeric
	// header field ("anim", "poweradd", "juggle", "sprpriority") or
	// boolean header field ("ctrl", "facep2", "hitdefpersist",
	// "movehitpersist", "hitcountpersist") whose value did not parse as a
	// plain literal integer/bool — real MUGEN/Ikemen .cns files sometimes
	// give these fields a trigger expression instead (e.g.
	// "anim = IfElse(ceil(lifemax/2) < life ,181,182)",
	// "facep2 = 1-(prevstateno=[100,119])"). Keyed by lowercase field
	// name. A field with an entry here has its corresponding typed field
	// (Anim, PowerAdd, Juggle, SprPriority, Ctrl, FaceP2, HitDefPersist,
	// MoveHitPersist, HitCountPersist) left at its zero value; a field
	// without an entry here was a literal value and its typed field holds
	// it as usual. See
	// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md
	// (numeric fields) and item 046 (boolean fields, extending the same
	// pattern).
	HeaderExprs map[string]string `json:"headerExprs"`
	// Controllers are the state controllers ([State N] blocks) that run
	// while this state is active, in file order.
	Controllers []Controller `json:"controllers"`
}

// Controller is a .cns [State N] block: a single state controller, stored
// as unevaluated data rather than resolved or type-checked. A controller's
// effect (e.g. which state to transition to for a "ChangeState" controller)
// is just another entry in Parameters, exactly like any other controller
// type's parameters — see the package doc comment.
type Controller struct {
	// Type is the controller type ("type" parameter of the [State N]
	// block, e.g. "ChangeState", "VelSet", "HitDef").
	Type string `json:"type"`
	// Triggers are this controller's trigger condition expressions
	// ("trigger1", "trigger2", ... lines), verbatim and unevaluated, in
	// file order. A nil or empty Triggers means the controller has no
	// trigger lines at all — it runs unconditionally whenever its state
	// is active, not "never runs".
	Triggers []string `json:"triggers"`
	// Parameters are this controller's remaining key/value parameters
	// (every non-trigger line of the block), verbatim and unevaluated,
	// keyed by parameter name.
	Parameters map[string]string `json:"parameters"`
}

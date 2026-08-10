package cns

import (
	"fmt"
	"io"
	"sort"
)

// Serialize writes states to w as MUGEN/Ikemen GO .cns text, emitting one
// "[Statedef N]" block per StateDef in order, each followed by its
// "[State N]" controller blocks in order.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting, comments, or unrecognized
// sections (a separate, format-preserving concern — see Document) — it only
// guarantees valid, readable output that Parse reads back into an
// equivalent []StateDef. Every recognized Statedef header field is always
// written, even at its zero value, since a StateDef cannot distinguish "not
// present in the original file" from "explicitly set to the zero value" —
// omitting zero-value fields would be no more faithful and would only add
// complexity. A Controller's Triggers are written as "trigger1",
// "trigger2", ... in slice order; Parse only ever appends to Triggers by a
// "trigger"-prefixed key without recording which one was used (see
// .vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md),
// so this always reparses back into the same Triggers regardless of which
// key names were written. Parameters (an unordered map) are written sorted
// by key for deterministic, reviewable output; map order carries no
// meaning Parse relies on. A "[State N]" block's N is always its enclosing
// StateDef's Number — Parse discards this number entirely (see
// .vibe/decisions/012-cns-parse-header-detection-strategy.md), so any value
// reparses identically.
func Serialize(w io.Writer, states []StateDef) error {
	for _, state := range states {
		if err := writeStatedefHeader(w, state); err != nil {
			return err
		}
		for _, ctrl := range state.Controllers {
			if err := writeController(w, state.Number, ctrl); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeStatedefHeader writes a "[Statedef N]" line followed by every
// recognized header field, then a blank line separating it from what
// follows (the first controller block, or the next Statedef). A numeric or
// boolean field with a HeaderExprs entry (see
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md
// and item 046) writes that raw, unevaluated text verbatim instead of the
// typed field's (necessarily zero) formatted value.
func writeStatedefHeader(w io.Writer, state StateDef) error {
	if _, err := fmt.Fprintf(w, "[Statedef %d]\n", state.Number); err != nil {
		return fmt.Errorf("cns: writing Statedef %d header: %w", state.Number, err)
	}

	fields := []struct {
		key   string
		value string
	}{
		{"type", string(state.Type)},
		{"movetype", string(state.MoveType)},
		{"physics", string(state.Physics)},
		{"anim", intOrExprValue(state, "anim", state.Anim)},
		{"ctrl", boolOrExprValue(state, "ctrl", state.Ctrl)},
		{"poweradd", intOrExprValue(state, "poweradd", state.PowerAdd)},
		{"juggle", intOrExprValue(state, "juggle", state.Juggle)},
		{"facep2", boolOrExprValue(state, "facep2", state.FaceP2)},
		{"hitdefpersist", boolOrExprValue(state, "hitdefpersist", state.HitDefPersist)},
		{"movehitpersist", boolOrExprValue(state, "movehitpersist", state.MoveHitPersist)},
		{"hitcountpersist", boolOrExprValue(state, "hitcountpersist", state.HitCountPersist)},
		{"sprpriority", intOrExprValue(state, "sprpriority", state.SprPriority)},
	}
	for _, f := range fields {
		if _, err := fmt.Fprintf(w, "%s = %s\n", f.key, f.value); err != nil {
			return fmt.Errorf("cns: writing Statedef %d %s: %w", state.Number, f.key, err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cns: writing Statedef %d separator: %w", state.Number, err)
	}
	return nil
}

// intOrExprValue returns state.HeaderExprs[name] when present (a numeric
// header field parsed as an unevaluated expression rather than a literal
// integer), otherwise literal's formatted decimal text.
func intOrExprValue(state StateDef, name string, literal int) string {
	if expr, ok := state.HeaderExprs[name]; ok {
		return expr
	}
	return fmt.Sprintf("%d", literal)
}

// boolOrExprValue returns state.HeaderExprs[name] when present (a boolean
// header field parsed as an unevaluated expression rather than a literal
// bool, see item 046), otherwise literal's conventional "1"/"0" text.
func boolOrExprValue(state StateDef, name string, literal bool) string {
	if expr, ok := state.HeaderExprs[name]; ok {
		return expr
	}
	return boolToStr(literal)
}

// writeController writes a "[State statedefNumber]" block for ctrl: its
// type, triggers (in order), and remaining parameters (sorted by key), then
// a blank line separating it from what follows.
func writeController(w io.Writer, statedefNumber int, ctrl Controller) error {
	if _, err := fmt.Fprintf(w, "[State %d]\n", statedefNumber); err != nil {
		return fmt.Errorf("cns: writing State %d header: %w", statedefNumber, err)
	}
	if _, err := fmt.Fprintf(w, "type = %s\n", ctrl.Type); err != nil {
		return fmt.Errorf("cns: writing State %d type: %w", statedefNumber, err)
	}
	for i, trigger := range ctrl.Triggers {
		if _, err := fmt.Fprintf(w, "trigger%d = %s\n", i+1, trigger); err != nil {
			return fmt.Errorf("cns: writing State %d trigger%d: %w", statedefNumber, i+1, err)
		}
	}

	keys := make([]string, 0, len(ctrl.Parameters))
	for key := range ctrl.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := fmt.Fprintf(w, "%s = %s\n", key, ctrl.Parameters[key]); err != nil {
			return fmt.Errorf("cns: writing State %d %s: %w", statedefNumber, key, err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cns: writing State %d separator: %w", statedefNumber, err)
	}
	return nil
}

// boolToStr renders b the way .cns files conventionally write a boolean
// field ("1"/"0"), matching strconv.ParseBool's accepted vocabulary so
// Parse reads it back unchanged.
func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

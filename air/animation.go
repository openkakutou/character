// Package air defines the pure-data model for MUGEN/Ikemen GO animation
// (.air) files: Animation, Frame, and ClsnBox.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (editor, engine) works with. It carries no parsing, file I/O, or
// write-only (format-preservation) logic; per CLAUDE.md's read/write
// separation constraint, that lives elsewhere so importing this data model
// alone never pulls in write-only dependencies.
//
// Collision boxes on a Frame are already resolved to the boxes active on
// that specific frame — the .air format's Clsn1Default/Clsn2Default
// fallback mechanism is a parsing concern, not part of this data model. See
// .vibe/decisions/001-frame-clsn-boxes-pre-resolved.md.
package air

// Flip is a frame's mirroring axis, matching the tokens used in .air files.
type Flip string

const (
	// FlipNone means the frame is drawn without mirroring.
	FlipNone Flip = ""
	// FlipH mirrors the frame horizontally.
	FlipH Flip = "H"
	// FlipV mirrors the frame vertically.
	FlipV Flip = "V"
	// FlipHV mirrors the frame both horizontally and vertically.
	FlipHV Flip = "HV"
)

// BlendMode is a frame's blending mode token as found in .air files (e.g.
// "A" for additive, "S" for subtractive). The zero value means normal
// blending (no special blend mode).
type BlendMode string

// ClsnBox is an axis-aligned collision box, expressed in the coordinate
// system used by Clsn1 (attack) and Clsn2 (vulnerability) boxes in .air
// files: Left/Top is one corner, Right/Bottom the opposite corner.
type ClsnBox struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

// Frame is a single displayed image within an Animation: which sprite to
// show (Group, Image), where to show it (X, Y), how long to hold it (Time),
// how to mirror and blend it, and the collision boxes active while it is
// displayed.
//
// Group and Image are normally non-negative sprite indices, but MUGEN/Ikemen
// GO ".air" files also use -1 on either field — most commonly -1,-1 — as a
// sentinel meaning "show no sprite on this frame". Such a frame is not
// malformed; see IsBlank.
type Frame struct {
	Group int
	Image int
	X     int
	Y     int
	Time  int
	Flip  Flip
	Blend BlendMode
	Clsn1 []ClsnBox
	Clsn2 []ClsnBox
}

// IsBlank reports whether frame uses the ".air" "no sprite shown" sentinel:
// a Group or Image of -1. A blank frame intentionally has no sprite to
// resolve, distinct from a frame whose sprite reference is simply missing
// from the loaded sprite collection.
func (f Frame) IsBlank() bool {
	return f.Group < 0 || f.Image < 0
}

// Animation is a MUGEN/Ikemen "[Begin Action N]" block: an ordered sequence
// of Frames plus the loop point frames return to once they have all played
// through once.
//
// LoopStart is the index into Frames the animation loops back to. Its zero
// value (0) matches .air's own default: an animation with no Loopstart
// marker loops back to its very first frame.
type Animation struct {
	Number    int
	Frames    []Frame
	LoopStart int
}

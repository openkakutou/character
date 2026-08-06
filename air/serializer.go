package air

import (
	"fmt"
	"io"
)

// Serialize writes animations to w as MUGEN/Ikemen GO .air text, emitting
// one "[Begin Action N]" block per Animation in order.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting or comments (a separate,
// not-yet-implemented concern) — it only guarantees valid, readable output
// that Parse reads back into an equivalent []Animation. Because Frame's
// Clsn1/Clsn2 already hold each frame's fully resolved collision boxes (see
// .vibe/decisions/001-frame-clsn-boxes-pre-resolved.md), boxes are always
// written per frame rather than reconstructing a
// Clsn1Default/Clsn2Default an original file might have used. A Loopstart
// marker is written only when LoopStart is non-zero, since the zero value
// already matches .air's own default of looping to the first frame.
func Serialize(w io.Writer, animations []Animation) error {
	for _, animation := range animations {
		if _, err := fmt.Fprintf(w, "[Begin Action %d]\n", animation.Number); err != nil {
			return fmt.Errorf("air: writing action %d header: %w", animation.Number, err)
		}

		for i, frame := range animation.Frames {
			if animation.LoopStart != 0 && animation.LoopStart == i {
				if _, err := fmt.Fprintln(w, "Loopstart"); err != nil {
					return fmt.Errorf("air: writing Loopstart marker for action %d: %w", animation.Number, err)
				}
			}
			if err := writeClsnBoxes(w, "Clsn1", frame.Clsn1); err != nil {
				return fmt.Errorf("air: writing action %d frame %d: %w", animation.Number, i, err)
			}
			if err := writeClsnBoxes(w, "Clsn2", frame.Clsn2); err != nil {
				return fmt.Errorf("air: writing action %d frame %d: %w", animation.Number, i, err)
			}
			if err := writeFrameLine(w, frame); err != nil {
				return fmt.Errorf("air: writing action %d frame %d line: %w", animation.Number, i, err)
			}
		}

		// LoopStart may point past the last frame (the animation never
		// visibly loops); that marker has no frame line to precede, so it
		// is written once the frame loop above is done.
		if animation.LoopStart != 0 && animation.LoopStart == len(animation.Frames) {
			if _, err := fmt.Fprintln(w, "Loopstart"); err != nil {
				return fmt.Errorf("air: writing Loopstart marker for action %d: %w", animation.Number, err)
			}
		}

		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("air: writing action %d separator: %w", animation.Number, err)
		}
	}

	return nil
}

// writeClsnBoxes writes a "declType: count" declaration followed by count
// indexed box lines (declType is "Clsn1" or "Clsn2"). It writes nothing
// when boxes is empty.
func writeClsnBoxes(w io.Writer, declType string, boxes []ClsnBox) error {
	if len(boxes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "%s: %d\n", declType, len(boxes)); err != nil {
		return fmt.Errorf("writing %s declaration: %w", declType, err)
	}
	for i, box := range boxes {
		if _, err := fmt.Fprintf(w, " %s[%d] = %d,%d,%d,%d\n", declType, i, box.Left, box.Top, box.Right, box.Bottom); err != nil {
			return fmt.Errorf("writing %s[%d] box: %w", declType, i, err)
		}
	}
	return nil
}

// writeFrameLine writes a single frame line: "Group,Image, X,Y, Time[,
// Flip, Blend]". The Flip and Blend fields are only written when at least
// one of them is set, mirroring Parse's treatment of a missing field as
// the zero value.
func writeFrameLine(w io.Writer, frame Frame) error {
	if frame.Flip == FlipNone && frame.Blend == "" {
		_, err := fmt.Fprintf(w, "%d,%d, %d,%d, %d\n", frame.Group, frame.Image, frame.X, frame.Y, frame.Time)
		return err
	}
	_, err := fmt.Fprintf(w, "%d,%d, %d,%d, %d, %s, %s\n", frame.Group, frame.Image, frame.X, frame.Y, frame.Time, frame.Flip, frame.Blend)
	return err
}

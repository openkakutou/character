package air

import (
	"strings"
	"testing"

	"github.com/openkakutou/sff"
)

// sampleSpriteGroups builds a small, multi-group sprite collection with
// distinguishing metadata (Width/Height/AxisX/AxisY/Palette) per sprite, so
// a test can tell a wrong (but plausible) lookup apart from the correct one.
func sampleSpriteGroups() []sff.SpriteGroup {
	return []sff.SpriteGroup{
		{
			Index: 0,
			Sprites: []sff.Sprite{
				{Group: 0, Image: 0, Width: 64, Height: 128, AxisX: 32, AxisY: 128, Palette: 1},
				{Group: 0, Image: 1, Width: 66, Height: 130, AxisX: 33, AxisY: 130, Palette: 1},
			},
		},
		{
			Index: 5,
			Sprites: []sff.Sprite{
				{Group: 5, Image: 0, Width: 40, Height: 40, AxisX: 20, AxisY: 40, Palette: 2},
			},
		},
	}
}

func TestSpriteResolver_Resolve_ReturnsMatchingSpriteForEveryFrame(t *testing.T) {
	src := `[Begin Action 0]
0,0, 0,0, 5
0,1, 5,-2, 4, H
5,0, 0,0, 3
`
	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	resolver := NewSpriteResolver(sampleSpriteGroups())

	want := []sff.Sprite{
		{Group: 0, Image: 0, Width: 64, Height: 128, AxisX: 32, AxisY: 128, Palette: 1},
		{Group: 0, Image: 1, Width: 66, Height: 130, AxisX: 33, AxisY: 130, Palette: 1},
		{Group: 5, Image: 0, Width: 40, Height: 40, AxisX: 20, AxisY: 40, Palette: 2},
	}

	frames := animations[0].Frames
	if len(frames) != len(want) {
		t.Fatalf("expected %d frames, got %d", len(want), len(frames))
	}

	for i, frame := range frames {
		got, err := resolver.Resolve(frame)
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		if got != want[i] {
			t.Errorf("frame %d: expected sprite %+v, got %+v", i, want[i], got)
		}
	}
}

func TestSpriteResolver_Resolve_DistinguishesGroupFromImageIndex(t *testing.T) {
	// Group 0 has images 0 and 1; group 5 has image 0. A frame referencing
	// (group 5, image 1) must NOT resolve to (group 0, image 1) or to
	// (group 5, image 0) even though each half of the key matches something.
	resolver := NewSpriteResolver(sampleSpriteGroups())

	_, err := resolver.Resolve(Frame{Group: 5, Image: 1})
	if err == nil {
		t.Fatal("expected an error for (group 5, image 1), got nil")
	}
}

func TestSpriteResolver_Resolve_ReturnsDescriptiveErrorOnMissingSprite(t *testing.T) {
	resolver := NewSpriteResolver(sampleSpriteGroups())

	_, err := resolver.Resolve(Frame{Group: 99, Image: 7})
	if err == nil {
		t.Fatal("expected an error for a non-existent (group, image) reference, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "99") || !strings.Contains(msg, "7") {
		t.Errorf("expected error to mention the missing (group, image) = (99, 7), got: %q", msg)
	}
}

func TestSpriteResolver_Resolve_ReturnsErrorWhenNoSpriteGroupsLoaded(t *testing.T) {
	resolver := NewSpriteResolver(nil)

	_, err := resolver.Resolve(Frame{Group: 0, Image: 0})
	if err == nil {
		t.Fatal("expected an error when resolving against an empty sprite collection, got nil")
	}
}

func TestSpriteResolver_Resolve_ReturnsZeroSpriteAndNoErrorForBlankFrame(t *testing.T) {
	resolver := NewSpriteResolver(sampleSpriteGroups())

	got, err := resolver.Resolve(Frame{Group: -1, Image: -1, X: 10, Y: 20, Time: 3})
	if err != nil {
		t.Fatalf("expected no error resolving a blank frame, got: %v", err)
	}
	if got != (sff.Sprite{}) {
		t.Errorf("expected a zero-value Sprite for a blank frame, got %+v", got)
	}
}

func TestSpriteResolver_Resolve_TreatsPartialSentinelAsBlankToo(t *testing.T) {
	resolver := NewSpriteResolver(sampleSpriteGroups())

	// Group 0 legitimately has an Image 0 sprite, but Group -1 is the blank
	// sentinel: the whole frame must be treated as blank, not partially
	// resolved against group 0's data.
	got, err := resolver.Resolve(Frame{Group: -1, Image: 0})
	if err != nil {
		t.Fatalf("expected no error resolving a partially-sentinel frame, got: %v", err)
	}
	if got != (sff.Sprite{}) {
		t.Errorf("expected a zero-value Sprite, got %+v", got)
	}
}

func TestSpriteResolver_Resolve_WorksTheSameRegardlessOfSFFVersionOrigin(t *testing.T) {
	// sff.Sprite/sff.SpriteGroup are already version-agnostic (populated
	// identically by ParseV1+DecodePCX and ParseV2+DecodeV2Sprite), so the
	// resolver needs no version-specific branching. This test exercises the
	// resolver against a sprite collection using v2-only concepts (a
	// non-zero Palette bank index unrelated to v1's shared-palette flag) to
	// document that no special-casing is required.
	groups := []sff.SpriteGroup{
		{Index: 3, Sprites: []sff.Sprite{{Group: 3, Image: 2, Width: 10, Height: 10, Palette: 4}}},
	}
	resolver := NewSpriteResolver(groups)

	got, err := resolver.Resolve(Frame{Group: 3, Image: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sff.Sprite{Group: 3, Image: 2, Width: 10, Height: 10, Palette: 4}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

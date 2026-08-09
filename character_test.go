package character

import (
	"os"
	"testing"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/sff"
)

func TestCharacter_ZeroValue_HasEmptyName(t *testing.T) {
	var c Character

	if c.Name != "" {
		t.Errorf("expected zero-value Character to have an empty Name, got %q", c.Name)
	}
}

// realFixtureCharacter assembles a Character from the air package's own
// realistic .air fixture (loaded via air.Parse, exactly as a real caller
// would) and a hand-built sprite collection matching every (Group, Image)
// the fixture's frames reference — group 0, images 0 through 3 — with
// distinguishing metadata per sprite so a test can tell a wrong (but
// plausible) resolution apart from the correct one.
func realFixtureCharacter(t *testing.T) *Character {
	t.Helper()

	f, err := os.Open("air/testdata/sample.air")
	if err != nil {
		t.Fatalf("failed to open .air fixture: %v", err)
	}
	defer f.Close()

	animations, err := air.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error parsing .air fixture: %v", err)
	}

	sprites := []sff.SpriteGroup{
		{
			Index: 0,
			Sprites: []sff.Sprite{
				{Group: 0, Image: 0, Width: 64, Height: 128, AxisX: 32, AxisY: 128, Palette: 1},
				{Group: 0, Image: 1, Width: 66, Height: 130, AxisX: 33, AxisY: 130, Palette: 1},
				{Group: 0, Image: 2, Width: 68, Height: 132, AxisX: 34, AxisY: 132, Palette: 1},
				{Group: 0, Image: 3, Width: 70, Height: 134, AxisX: 35, AxisY: 134, Palette: 1},
			},
		},
	}

	return &Character{
		Name:       "Sample",
		Animations: animations,
		Sprites:    sprites,
	}
}

func TestCharacter_Assembled_ExposesAnimationsAndSpritesFromReadPathTypes(t *testing.T) {
	c := realFixtureCharacter(t)

	if len(c.Animations) != 2 {
		t.Fatalf("expected 2 animations from the fixture, got %d", len(c.Animations))
	}
	if c.Animations[0].Number != 200 {
		t.Errorf("expected first animation to be action 200, got %d", c.Animations[0].Number)
	}
	if len(c.Sprites) != 1 || len(c.Sprites[0].Sprites) != 4 {
		t.Fatalf("expected 1 sprite group with 4 sprites, got %+v", c.Sprites)
	}
}

func TestCharacter_ResolveSprite_ReturnsMatchingSpriteForEveryFrame(t *testing.T) {
	c := realFixtureCharacter(t)

	want := []sff.Sprite{
		{Group: 0, Image: 0, Width: 64, Height: 128, AxisX: 32, AxisY: 128, Palette: 1},
		{Group: 0, Image: 1, Width: 66, Height: 130, AxisX: 33, AxisY: 130, Palette: 1},
		{Group: 0, Image: 2, Width: 68, Height: 132, AxisX: 34, AxisY: 132, Palette: 1},
		{Group: 0, Image: 3, Width: 70, Height: 134, AxisX: 35, AxisY: 134, Palette: 1},
	}

	frames := c.Animations[0].Frames
	if len(frames) != len(want) {
		t.Fatalf("expected %d frames in action 200, got %d", len(want), len(frames))
	}

	for i, frame := range frames {
		got, err := c.ResolveSprite(frame)
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		if got != want[i] {
			t.Errorf("frame %d: expected sprite %+v, got %+v", i, want[i], got)
		}
	}
}

func TestCharacter_ResolveSprite_ReturnsDescriptiveErrorForFrameNotInLoadedSprites(t *testing.T) {
	c := realFixtureCharacter(t)

	// (group 9, image 9) exists in no sprite group loaded into c.
	_, err := c.ResolveSprite(air.Frame{Group: 9, Image: 9})
	if err == nil {
		t.Fatal("expected an error for a frame referencing a sprite absent from the loaded sprites, got nil")
	}
}

func TestCharacter_ResolveSprite_ReturnsErrorOnZeroValueCharacter(t *testing.T) {
	var c Character

	_, err := c.ResolveSprite(air.Frame{Group: 0, Image: 0})
	if err == nil {
		t.Fatal("expected an error resolving a sprite on a Character with no sprites loaded, got nil")
	}
}

func TestCharacter_ResolveSprite_ReturnsZeroSpriteNoErrorForBlankFrame(t *testing.T) {
	c := realFixtureCharacter(t)

	got, err := c.ResolveSprite(air.Frame{Group: -1, Image: -1})
	if err != nil {
		t.Fatalf("expected no error resolving a blank (-1,-1) frame, got: %v", err)
	}
	if got != (sff.Sprite{}) {
		t.Errorf("expected a zero-value Sprite for a blank frame, got %+v", got)
	}
}

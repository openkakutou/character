package character

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/sff"
)

// writeFixtureCharacter creates a .def, .air, and .sff file inside a fresh
// temp directory, referencing each other exactly as a real character folder
// would, and returns the .def file's path.
func writeFixtureCharacter(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	defContent := `[Info]
name = Test Character
author = Test Author

[Files]
sprite = char.sff
anim = char.air
cns = char.cns
`
	if err := os.WriteFile(filepath.Join(dir, "char.def"), []byte(defContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .def fixture: %v", err)
	}

	cnsContent := `[Statedef 200, Attack]
type = S
movetype = A
physics = S
anim = 200
ctrl = 0

[State 200, ChangeState]
type = ChangeState
trigger1 = Time = 0
value = 0
`
	if err := os.WriteFile(filepath.Join(dir, "char.cns"), []byte(cnsContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .cns fixture: %v", err)
	}

	airContent := `[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`
	if err := os.WriteFile(filepath.Join(dir, "char.air"), []byte(airContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .air fixture: %v", err)
	}

	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, AxisX: 32, AxisY: 128, PixelData: mustEncodePCXFixture(t, 4, 2), Palette: make([]byte, sff.V1PaletteBlockSize)},
		{Group: 0, Image: 1, AxisX: 33, AxisY: 130, PixelData: mustEncodePCXFixture(t, 6, 3), Palette: make([]byte, sff.V1PaletteBlockSize)},
	}
	var sffBuf bytes.Buffer
	if err := sff.SerializeV1(&sffBuf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "char.sff"), sffBuf.Bytes(), 0o644); err != nil {
		t.Fatalf("test setup: writing .sff fixture: %v", err)
	}

	return filepath.Join(dir, "char.def")
}

// mustEncodePCXFixture builds valid PCX-encoded pixel data of the given
// dimensions for use as fixture sprite data.
func mustEncodePCXFixture(t *testing.T, width, height int) []byte {
	t.Helper()
	pixels := make([]byte, width*height)
	for i := range pixels {
		pixels[i] = byte(i)
	}
	data, err := sff.EncodePCX(&sff.PCXImage{Width: width, Height: height, Pixels: pixels})
	if err != nil {
		t.Fatalf("test setup: EncodePCX failed: %v", err)
	}
	return data
}

func TestLoad_ValidDefFile_ProducesFullyPopulatedCharacter(t *testing.T) {
	defPath := writeFixtureCharacter(t)

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if c.Name != "Test Character" {
		t.Errorf("expected Name %q, got %q", "Test Character", c.Name)
	}

	if len(c.Animations) != 1 {
		t.Fatalf("expected 1 animation, got %d", len(c.Animations))
	}
	if c.Animations[0].Number != 200 {
		t.Errorf("expected action 200, got %d", c.Animations[0].Number)
	}
	if len(c.Animations[0].Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(c.Animations[0].Frames))
	}

	if len(c.Sprites) != 1 || len(c.Sprites[0].Sprites) != 2 {
		t.Fatalf("expected 1 sprite group with 2 sprites, got %+v", c.Sprites)
	}
	if w := c.Sprites[0].Sprites[0].Width; w != 4 {
		t.Errorf("expected sprite (0,0) width 4, got %d", w)
	}

	// The loaded animations/sprites must actually resolve against each
	// other through the existing ResolveSprite path.
	for _, frame := range c.Animations[0].Frames {
		if _, err := c.ResolveSprite(frame); err != nil {
			t.Errorf("frame (group %d, image %d): unexpected ResolveSprite error: %v", frame.Group, frame.Image, err)
		}
	}

	if len(c.StateDefs) != 1 {
		t.Fatalf("expected 1 state def, got %d", len(c.StateDefs))
	}
	state := c.StateDefs[0]
	if state.Number != 200 {
		t.Errorf("expected state number 200, got %d", state.Number)
	}
	if state.MoveType != cns.MoveTypeAttack {
		t.Errorf("expected move type %q, got %q", cns.MoveTypeAttack, state.MoveType)
	}
	if len(state.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(state.Controllers))
	}
	if state.Controllers[0].Type != "ChangeState" {
		t.Errorf("expected controller type %q, got %q", "ChangeState", state.Controllers[0].Type)
	}
}

func TestLoad_DefReferencesMissingAnimationFile_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if err := os.Remove(filepath.Join(dir, "char.air")); err != nil {
		t.Fatalf("test setup: removing .air fixture: %v", err)
	}

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced .air file is missing, got nil")
	}
}

func TestLoad_DefReferencesMissingSpriteFile_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if err := os.Remove(filepath.Join(dir, "char.sff")); err != nil {
		t.Fatalf("test setup: removing .sff fixture: %v", err)
	}

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced .sff file is missing, got nil")
	}
}

func TestLoad_DefReferencesMissingCnsFile_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if err := os.Remove(filepath.Join(dir, "char.cns")); err != nil {
		t.Fatalf("test setup: removing .cns fixture: %v", err)
	}

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced .cns file is missing, got nil")
	}
}

func TestLoad_DefReferencesConstantsFileWithBackslashPath_ResolvesNestedFile(t *testing.T) {
	dir := t.TempDir()

	defContent := `[Info]
name = Test Character
author = Test Author

[Files]
sprite = char.sff
anim = char.air
cns = states\constants.cns
`
	if err := os.WriteFile(filepath.Join(dir, "char.def"), []byte(defContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .def fixture: %v", err)
	}

	statesDir := filepath.Join(dir, "states")
	if err := os.Mkdir(statesDir, 0o755); err != nil {
		t.Fatalf("test setup: creating states subdirectory: %v", err)
	}

	cnsContent := `[Statedef 200, Attack]
type = S
movetype = A
physics = S
anim = 200
ctrl = 0

[State 200, ChangeState]
type = ChangeState
trigger1 = Time = 0
value = 0
`
	if err := os.WriteFile(filepath.Join(statesDir, "constants.cns"), []byte(cnsContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .cns fixture: %v", err)
	}

	airContent := `[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`
	if err := os.WriteFile(filepath.Join(dir, "char.air"), []byte(airContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .air fixture: %v", err)
	}

	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, AxisX: 32, AxisY: 128, PixelData: mustEncodePCXFixture(t, 4, 2), Palette: make([]byte, sff.V1PaletteBlockSize)},
		{Group: 0, Image: 1, AxisX: 33, AxisY: 130, PixelData: mustEncodePCXFixture(t, 6, 3), Palette: make([]byte, sff.V1PaletteBlockSize)},
	}
	var sffBuf bytes.Buffer
	if err := sff.SerializeV1(&sffBuf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "char.sff"), sffBuf.Bytes(), 0o644); err != nil {
		t.Fatalf("test setup: writing .sff fixture: %v", err)
	}

	c, err := Load(filepath.Join(dir, "char.def"))
	if err != nil {
		t.Fatalf("Load returned error for a backslash-separated .def path: %v", err)
	}

	if len(c.StateDefs) != 1 {
		t.Fatalf("expected 1 state def loaded from the nested states/constants.cns file, got %d", len(c.StateDefs))
	}
	if c.StateDefs[0].Number != 200 {
		t.Errorf("expected state number 200, got %d", c.StateDefs[0].Number)
	}
}

func TestLoad_DefReferencesFileWithDifferentCasing_ResolvesCaseInsensitively(t *testing.T) {
	dir := t.TempDir()

	// The .def references the animation file with different letter casing
	// than the actual file on disk — a real-world authoring pattern left
	// over from Windows' case-insensitive filesystem (backlog item 050).
	defContent := `[Info]
name = Test Character
author = Test Author

[Files]
sprite = char.sff
anim = Char.AIR
cns = char.cns
`
	if err := os.WriteFile(filepath.Join(dir, "char.def"), []byte(defContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .def fixture: %v", err)
	}

	cnsContent := `[Statedef 200, Attack]
type = S
movetype = A
physics = S
anim = 200
ctrl = 0

[State 200, ChangeState]
type = ChangeState
trigger1 = Time = 0
value = 0
`
	if err := os.WriteFile(filepath.Join(dir, "char.cns"), []byte(cnsContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .cns fixture: %v", err)
	}

	airContent := `[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`
	// Actual on-disk filename is lowercase "char.air", but the .def above
	// references "Char.AIR".
	if err := os.WriteFile(filepath.Join(dir, "char.air"), []byte(airContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .air fixture: %v", err)
	}

	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, AxisX: 32, AxisY: 128, PixelData: mustEncodePCXFixture(t, 4, 2), Palette: make([]byte, sff.V1PaletteBlockSize)},
		{Group: 0, Image: 1, AxisX: 33, AxisY: 130, PixelData: mustEncodePCXFixture(t, 6, 3), Palette: make([]byte, sff.V1PaletteBlockSize)},
	}
	var sffBuf bytes.Buffer
	if err := sff.SerializeV1(&sffBuf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "char.sff"), sffBuf.Bytes(), 0o644); err != nil {
		t.Fatalf("test setup: writing .sff fixture: %v", err)
	}

	c, err := Load(filepath.Join(dir, "char.def"))
	if err != nil {
		t.Fatalf("Load returned error for a case-mismatched referenced file: %v", err)
	}

	if len(c.Animations) != 1 {
		t.Fatalf("expected 1 animation loaded via case-insensitive fallback, got %d", len(c.Animations))
	}
	if c.Animations[0].Number != 200 {
		t.Errorf("expected action 200, got %d", c.Animations[0].Number)
	}
}

func TestLoad_DefReferencesFileWithNoCaseInsensitiveMatch_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if err := os.Remove(filepath.Join(dir, "char.air")); err != nil {
		t.Fatalf("test setup: removing .air fixture: %v", err)
	}

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced file has no case-insensitive match either, got nil")
	}
	if !os.IsNotExist(errors.Unwrap(err)) {
		t.Errorf("expected a not-exist error, got: %v", err)
	}
}

func TestLoad_DefFileItselfMissing_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.def"))
	if err == nil {
		t.Fatal("expected an error when the .def file itself does not exist, got nil")
	}
}

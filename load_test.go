package character

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// writeFixtureCharacterWithSoundFile is writeFixtureCharacter plus a
// "sound = char.snd" [Files] entry, with sndBytes written as char.snd's
// content — used to test Load's SoundFile wiring without duplicating the
// rest of the fixture setup.
func writeFixtureCharacterWithSoundFile(t *testing.T, sndBytes []byte) string {
	t.Helper()

	dir := t.TempDir()

	defContent := `[Info]
name = Test Character
author = Test Author

[Files]
sprite = char.sff
anim = char.air
cns = char.cns
sound = char.snd
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

	if sndBytes != nil {
		if err := os.WriteFile(filepath.Join(dir, "char.snd"), sndBytes, 0o644); err != nil {
			t.Fatalf("test setup: writing .snd fixture: %v", err)
		}
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

// TestLoad_ValidDefFile_ThreadsAuthorAndReferencedFilePathsToCharacter pins
// backlog item 038 for the filesystem-based loader too: Load and LoadBytes
// share the same Character type, so a field threaded through one must be
// threaded through the other, not just the WASM-facing entrypoint.
func TestLoad_ValidDefFile_ThreadsAuthorAndReferencedFilePathsToCharacter(t *testing.T) {
	defPath := writeFixtureCharacter(t)

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if c.Author != "Test Author" {
		t.Errorf("expected Author %q, got %q", "Test Author", c.Author)
	}
	if c.SpriteFile != "char.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "char.sff", c.SpriteFile)
	}
	if c.AnimationFile != "char.air" {
		t.Errorf("expected AnimationFile %q, got %q", "char.air", c.AnimationFile)
	}
	if c.ConstantsFile != "char.cns" {
		t.Errorf("expected ConstantsFile %q, got %q", "char.cns", c.ConstantsFile)
	}
	// The fixture's .def has no sound/cmd/st/pal entries: those fields
	// must stay empty/zero rather than erroring.
	if c.SoundFile != "" || c.CommandFile != "" {
		t.Errorf("expected empty SoundFile/CommandFile, got %+v", c)
	}
	if len(c.StateFiles) != 0 {
		t.Errorf("expected empty StateFiles, got %v", c.StateFiles)
	}
	if len(c.Palettes) != 0 {
		t.Errorf("expected empty Palettes, got %v", c.Palettes)
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

// TestLoad_DefWithoutFilesSection_ReturnsClearDiagnosticNotFilesystemError
// covers backlog item 051: an Ikemen GO storyboard/intro/ending screen .def
// (a [SceneDef]/[Scene N]-based file, an entirely different sub-format) has
// no [Files] section at all. def.Parse correctly skips those unrecognized
// sections per its own documented contract, leaving CharacterInfo's file
// fields empty — Load must catch that before attempting to open an empty
// referenced path (which resolves to the containing directory itself) and
// name the real problem, instead of surfacing a raw, confusing
// "is a directory" filesystem error.
// TestLoad_DefReferencesSoundFile_DecodesAndExposesSoundGroups covers
// backlog item 057's nominal path: Load resolves and decodes SoundFile via
// the snd library, exposing the result as Character.Sounds keyed by
// (Group, Sample) the same way sff-resolved Sprites are keyed by
// (Group, Image).
func TestLoad_DefReferencesSoundFile_DecodesAndExposesSoundGroups(t *testing.T) {
	sample0 := []int16{100, -100, 200}
	sample1 := []int16{1, 2, 3, 4}
	raw0 := make([]byte, len(sample0)*2)
	for i, s := range sample0 {
		binary.LittleEndian.PutUint16(raw0[i*2:i*2+2], uint16(s))
	}
	raw1 := make([]byte, len(sample1)*2)
	for i, s := range sample1 {
		binary.LittleEndian.PutUint16(raw1[i*2:i*2+2], uint16(s))
	}

	sndBytes := buildV1SndFile(t, []sndFixtureEntry{
		{group: 0, sample: 0, payload: buildWAVFixture(t, 1, 44100, 16, raw0)},
		{group: 0, sample: 1, payload: buildWAVFixture(t, 1, 22050, 16, raw1)},
		{group: 1, sample: 5, payload: buildWAVFixture(t, 1, 44100, 16, raw0)},
	})

	defPath := writeFixtureCharacterWithSoundFile(t, sndBytes)

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if c.SoundFile != "char.snd" {
		t.Errorf("expected SoundFile %q, got %q", "char.snd", c.SoundFile)
	}
	if len(c.Sounds) != 2 {
		t.Fatalf("expected 2 sound groups (0 and 1), got %d: %+v", len(c.Sounds), c.Sounds)
	}

	group0 := c.Sounds[0]
	if group0.Index != 0 {
		t.Errorf("expected first group index 0, got %d", group0.Index)
	}
	if len(group0.Sounds) != 2 {
		t.Fatalf("expected 2 sounds in group 0, got %d", len(group0.Sounds))
	}
	if group0.Sounds[0].Sample != 0 || group0.Sounds[1].Sample != 1 {
		t.Errorf("expected samples 0 and 1 in group 0, got %d and %d", group0.Sounds[0].Sample, group0.Sounds[1].Sample)
	}
	if got := group0.Sounds[0].PCM; !equalInt16Slices(got, sample0) {
		t.Errorf("expected group 0 sample 0 PCM %v, got %v", sample0, got)
	}
	if group0.Sounds[0].SampleRate != 44100 {
		t.Errorf("expected sample rate 44100, got %d", group0.Sounds[0].SampleRate)
	}
	if group0.Sounds[1].SampleRate != 22050 {
		t.Errorf("expected sample rate 22050, got %d", group0.Sounds[1].SampleRate)
	}
	if group0.Sounds[0].Channels != 1 {
		t.Errorf("expected 1 channel, got %d", group0.Sounds[0].Channels)
	}

	group1 := c.Sounds[1]
	if group1.Index != 1 || len(group1.Sounds) != 1 || group1.Sounds[0].Sample != 5 {
		t.Fatalf("expected group 1 with a single sample 5, got %+v", group1)
	}
}

func equalInt16Slices(a, b []int16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestLoad_DefWithoutSoundFile_LoadsSuccessfullyWithNoSounds covers backlog
// item 057's optionality requirement: unlike AnimationFile/SpriteFile/
// ConstantsFile, an empty SoundFile is not an error — every other file kind
// still loads normally, and Sounds is simply empty.
func TestLoad_DefWithoutSoundFile_LoadsSuccessfullyWithNoSounds(t *testing.T) {
	defPath := writeFixtureCharacter(t)

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load returned error for a character with no SoundFile: %v", err)
	}
	if len(c.Sounds) != 0 {
		t.Errorf("expected no sounds for a character with no SoundFile, got %+v", c.Sounds)
	}
	if len(c.Animations) == 0 || len(c.Sprites) == 0 || len(c.StateDefs) == 0 {
		t.Errorf("expected every other file kind to still load, got %+v", c)
	}
}

// TestLoad_DefReferencesMissingSoundFile_ReturnsDescriptiveErrorNotPanic
// covers the "once referenced" half of item 057's optionality decision
// (.vibe/decisions/029): a declared but missing SoundFile is a hard error,
// matching the existing convention for the other referenced files.
func TestLoad_DefReferencesMissingSoundFile_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacterWithSoundFile(t, nil) // sound = char.snd, but char.snd is never written

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced .snd file is missing, got nil")
	}
	if !strings.Contains(err.Error(), "char.snd") {
		t.Errorf("expected error to name the missing sound file, got: %v", err)
	}
}

// TestLoad_DefReferencesMalformedSoundFile_ReturnsDescriptiveErrorNotPanic
// covers the error path for a SoundFile that exists but isn't a valid .snd
// file (e.g. truncated or corrupted).
func TestLoad_DefReferencesMalformedSoundFile_ReturnsDescriptiveErrorNotPanic(t *testing.T) {
	defPath := writeFixtureCharacterWithSoundFile(t, []byte("not a sound file"))

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the referenced .snd file is malformed, got nil")
	}
	if !strings.Contains(err.Error(), "char.snd") {
		t.Errorf("expected error to name the malformed sound file, got: %v", err)
	}
}

func TestLoad_DefWithoutFilesSection_ReturnsClearDiagnosticNotFilesystemError(t *testing.T) {
	dir := t.TempDir()

	defContent := `[SceneDef]
spr = system.sff
fadein.time = 30
fadeout.time = 30

[Scene 0]
type = layerinfo
`
	defPath := filepath.Join(dir, "Ending.def")
	if err := os.WriteFile(defPath, []byte(defContent), 0o644); err != nil {
		t.Fatalf("test setup: writing .def fixture: %v", err)
	}

	_, err := Load(defPath)
	if err == nil {
		t.Fatal("expected an error when the .def file has no [Files] section, got nil")
	}
	if !strings.Contains(err.Error(), defPath) {
		t.Errorf("expected error to name the .def path %q, got: %v", defPath, err)
	}
	if strings.Contains(err.Error(), "is a directory") {
		t.Errorf("expected a clear diagnostic instead of a raw filesystem error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "[Files]") {
		t.Errorf("expected error to mention the missing [Files] section, got: %v", err)
	}
}

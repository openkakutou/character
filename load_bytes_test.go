package character

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/openkakutou/sff"
)

// fixtureCharacterBytes builds valid, minimal .def/.air/.sff/.cns content
// entirely in memory (no filesystem I/O), the shape a browser caller would
// hand to LoadBytes after fetching/selecting each file's bytes itself.
func fixtureCharacterBytes(t *testing.T) (defBytes, airBytes, sffBytes, cnsBytes []byte) {
	t.Helper()

	defBytes = []byte(`[Info]
name = Test Character
author = Test Author
`)

	airBytes = []byte(`[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`)

	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, AxisX: 32, AxisY: 128, PixelData: mustEncodePCXFixture(t, 4, 2), Palette: make([]byte, sff.V1PaletteBlockSize)},
		{Group: 0, Image: 1, AxisX: 33, AxisY: 130, PixelData: mustEncodePCXFixture(t, 6, 3), Palette: make([]byte, sff.V1PaletteBlockSize)},
	}
	var sffBuf bytes.Buffer
	if err := sff.SerializeV1(&sffBuf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}
	sffBytes = sffBuf.Bytes()

	cnsBytes = []byte(`[Statedef 200, Attack]
type = S
movetype = A
physics = S
anim = 200
ctrl = 0

[State 200, ChangeState]
type = ChangeState
trigger1 = Time = 0
value = 0
`)

	return defBytes, airBytes, sffBytes, cnsBytes
}

func TestLoadBytes_ValidBuffers_ProducesFullyPopulatedCharacter(t *testing.T) {
	defBytes, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
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
	if len(state.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(state.Controllers))
	}
	if state.Controllers[0].Type != "ChangeState" {
		t.Errorf("expected controller type %q, got %q", "ChangeState", state.Controllers[0].Type)
	}
}

// TestLoadBytes_FrameAndStateWithNoOptionalCollections_MarshalToEmptyArraysNotNull
// pins LoadBytes's JSON-facing contract (see
// .vibe/decisions/019-wasm-entrypoint-byte-buffer-loading-and-json-contract.md):
// a frame with no Clsn boxes and a state with no controllers must still
// marshal to "[]", never "null" — a JS caller should never have to
// null-check before iterating.
func TestLoadBytes_FrameAndStateWithNoOptionalCollections_MarshalToEmptyArraysNotNull(t *testing.T) {
	defBytes := []byte(`[Info]
name = Bare
`)
	// A frame line with no Clsn1/Clsn2 declared anywhere in the file.
	airBytes := []byte(`[Begin Action 0]
0,0, 0,0, 5
`)
	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, PixelData: mustEncodePCXFixture(t, 2, 2), Palette: make([]byte, sff.V1PaletteBlockSize)},
	}
	var sffBuf bytes.Buffer
	if err := sff.SerializeV1(&sffBuf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}
	// A Statedef with no [State N] blocks under it at all.
	cnsBytes := []byte(`[Statedef 0]
type = S
movetype = I
physics = S
`)

	c, err := LoadBytes(defBytes, airBytes, sffBuf.Bytes(), cnsBytes)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	out := string(data)

	if strings.Contains(out, "null") {
		t.Errorf("expected no \"null\" anywhere in marshaled output, got: %s", out)
	}
	if !strings.Contains(out, `"clsn1":[]`) {
		t.Errorf("expected frame's clsn1 to marshal as an empty array, got: %s", out)
	}
	if !strings.Contains(out, `"clsn2":[]`) {
		t.Errorf("expected frame's clsn2 to marshal as an empty array, got: %s", out)
	}
	if !strings.Contains(out, `"controllers":[]`) {
		t.Errorf("expected state's controllers to marshal as an empty array, got: %s", out)
	}
	if !strings.Contains(out, `"stateFiles":[]`) {
		t.Errorf("expected stateFiles to marshal as an empty array, got: %s", out)
	}
	if !strings.Contains(out, `"palettes":[]`) {
		t.Errorf("expected palettes to marshal as an empty array, got: %s", out)
	}
}

// TestLoadBytes_DefWithFullMetadata_ThreadsAllCharacterInfoFieldsToCharacter
// pins backlog item 038: every field def.CharacterInfo already parses (not
// just Name) must reach the Character LoadBytes returns, since that's the
// JSON contract WASM consumers (e.g. character-viewer-web's characteristics
// panel) read from.
func TestLoadBytes_DefWithFullMetadata_ThreadsAllCharacterInfoFieldsToCharacter(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	defBytes := []byte(`[Info]
name = Test Character
author = Test Author

[Files]
sprite = char.sff
anim = char.air
sound = char.snd
cmd = char.cmd
cns = char.cns
st1 = extra1.st
st2 = extra2.st
pal1 = char1.act
pal2 = char2.act
`)

	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
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
	if c.SoundFile != "char.snd" {
		t.Errorf("expected SoundFile %q, got %q", "char.snd", c.SoundFile)
	}
	if c.CommandFile != "char.cmd" {
		t.Errorf("expected CommandFile %q, got %q", "char.cmd", c.CommandFile)
	}
	if c.ConstantsFile != "char.cns" {
		t.Errorf("expected ConstantsFile %q, got %q", "char.cns", c.ConstantsFile)
	}
	wantStateFiles := []string{"extra1.st", "extra2.st"}
	if !reflect.DeepEqual(c.StateFiles, wantStateFiles) {
		t.Errorf("expected StateFiles %v, got %v", wantStateFiles, c.StateFiles)
	}
	wantPalettes := []string{"char1.act", "char2.act"}
	if !reflect.DeepEqual(c.Palettes, wantPalettes) {
		t.Errorf("expected Palettes %v, got %v", wantPalettes, c.Palettes)
	}
}

// TestLoadBytes_DefMissingOptionalMetadataFields_LoadsSuccessfullyWithEmptyFields
// pins the acceptance criterion that a .def missing optional metadata still
// loads successfully, with those fields empty/zero rather than erroring.
func TestLoadBytes_DefMissingOptionalMetadataFields_LoadsSuccessfullyWithEmptyFields(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	// No "author" key, no [Files] entries beyond what LoadBytes itself
	// needs to parse the other three buffers.
	defBytes := []byte(`[Info]
name = Bare Character
`)

	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}

	if c.Author != "" {
		t.Errorf("expected empty Author, got %q", c.Author)
	}
	if c.SpriteFile != "" || c.AnimationFile != "" || c.SoundFile != "" || c.CommandFile != "" || c.ConstantsFile != "" {
		t.Errorf("expected all optional file fields empty, got %+v", c)
	}
	if len(c.StateFiles) != 0 {
		t.Errorf("expected empty StateFiles, got %v", c.StateFiles)
	}
	if len(c.Palettes) != 0 {
		t.Errorf("expected empty Palettes, got %v", c.Palettes)
	}
}

func TestLoadBytes_MalformedDefBytes_ReturnsError(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	_, err := LoadBytes([]byte("[Info\nname=Test\n"), airBytes, sffBytes, cnsBytes)
	if err == nil {
		t.Fatal("expected an error for malformed .def bytes, got nil")
	}
}

func TestLoadBytes_MalformedAirBytes_ReturnsError(t *testing.T) {
	defBytes, _, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	_, err := LoadBytes(defBytes, []byte("[Begin Action abc]\n0,0, 0,0, 5\n"), sffBytes, cnsBytes)
	if err == nil {
		t.Fatal("expected an error for malformed .air bytes, got nil")
	}
}

func TestLoadBytes_MalformedSffBytes_ReturnsError(t *testing.T) {
	defBytes, airBytes, _, cnsBytes := fixtureCharacterBytes(t)

	_, err := LoadBytes(defBytes, airBytes, []byte("not a sprite sheet"), cnsBytes)
	if err == nil {
		t.Fatal("expected an error for malformed .sff bytes, got nil")
	}
}

func TestLoadBytes_MalformedCnsBytes_ReturnsError(t *testing.T) {
	defBytes, airBytes, sffBytes, _ := fixtureCharacterBytes(t)

	_, err := LoadBytes(defBytes, airBytes, sffBytes, []byte("[Statedef\ntype = S\n"))
	if err == nil {
		t.Fatal("expected an error for malformed .cns bytes, got nil")
	}
}

func TestLoadBytes_NilInputBuffers_ReturnsErrorNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LoadBytes panicked on nil buffers: %v", r)
		}
	}()

	_, err := LoadBytes(nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected an error for nil input buffers, got nil")
	}
}

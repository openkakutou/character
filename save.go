package character

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cmd"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/character/zss"
)

// SerializeDef produces .def bytes for info, the write-path counterpart to
// def.Parse/def.ParseDocument. original is the .def file's own previously
// loaded bytes — empty when info describes a brand new character with no
// original file yet.
//
// When original is non-empty and info, once normalized the same way a JSON
// round trip already normalizes it (see normalizeCharacterInfoForJSON), is
// unchanged from what parsing original itself produces, the original bytes
// are written back out verbatim — a byte-exact round trip, matching
// def.Document's existing guarantee. Otherwise (info was edited, or
// original is empty) fresh text is generated via def.Serialize, reflecting
// info's current values without preserving original's comments/ordering —
// see .vibe/decisions/028-wasm-save-path-per-format-diff-or-serialize.md.
//
// A malformed original returns a descriptive error (from def.ParseDocument)
// rather than silently falling back to a fresh serialize.
func SerializeDef(original []byte, info def.CharacterInfo) ([]byte, error) {
	edited := info
	normalizeCharacterInfoForJSON(&edited)

	if len(original) > 0 {
		doc, err := def.ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("character: parsing original character definition for save: %w", err)
		}

		baseline := doc.Info
		normalizeCharacterInfoForJSON(&baseline)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("character: writing unmodified character definition: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := def.Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("character: serializing edited character definition: %w", err)
	}
	return buf.Bytes(), nil
}

// SerializeAir produces .air bytes for animations, the write-path
// counterpart to air.Parse/air.ParseDocument. original is the .air file's
// own previously loaded bytes — empty when animations describes a brand
// new file with no original yet.
//
// Same byte-exact-when-unmodified / fresh-serialize-when-edited strategy as
// SerializeDef — see that function's doc comment and
// .vibe/decisions/028-wasm-save-path-per-format-diff-or-serialize.md.
func SerializeAir(original []byte, animations []air.Animation) ([]byte, error) {
	edited := normalizeAnimationsForJSON(animations)

	if len(original) > 0 {
		doc, err := air.ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("character: parsing original animation file for save: %w", err)
		}

		baseline := normalizeAnimationsForJSON(doc.Animations)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("character: writing unmodified animation file: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := air.Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("character: serializing edited animation file: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeAnimationsForJSON returns animations with every nil slice
// (itself, and each Frame's Clsn1/Clsn2) replaced by its non-nil empty
// equivalent — reusing normalizeAnimationForJSON's existing per-animation
// normalization (see load_bytes.go) across a whole slice, for the same
// reason normalizeCharacterInfoForJSON exists.
func normalizeAnimationsForJSON(animations []air.Animation) []air.Animation {
	if animations == nil {
		animations = []air.Animation{}
	}
	for i := range animations {
		normalizeAnimationForJSON(&animations[i])
	}
	return animations
}

// SerializeCns produces .cns bytes for states, the write-path counterpart
// to cns.Parse/cns.ParseDocument. original is the .cns file's own
// previously loaded bytes — empty when states describes a brand new file
// with no original yet.
//
// Same byte-exact-when-unmodified / fresh-serialize-when-edited strategy as
// SerializeDef — see that function's doc comment and
// .vibe/decisions/028-wasm-save-path-per-format-diff-or-serialize.md.
func SerializeCns(original []byte, states []cns.StateDef) ([]byte, error) {
	edited := normalizeStateDefsForJSON(states)

	if len(original) > 0 {
		doc, err := cns.ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("character: parsing original combat logic file for save: %w", err)
		}

		baseline := normalizeStateDefsForJSON(doc.StateDefs)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("character: writing unmodified combat logic file: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := cns.Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("character: serializing edited combat logic file: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeStateDefsForJSON returns states with every nil slice/map (itself,
// and each StateDef's Controllers/HeaderExprs/Triggers/Parameters) replaced
// by its non-nil empty equivalent — reusing normalizeStateDefForJSON's
// existing per-StateDef normalization (see load_bytes.go) across a whole
// slice, for the same reason normalizeCharacterInfoForJSON exists.
func normalizeStateDefsForJSON(states []cns.StateDef) []cns.StateDef {
	if states == nil {
		states = []cns.StateDef{}
	}
	for i := range states {
		normalizeStateDefForJSON(&states[i])
	}
	return states
}

// SerializeCmd produces .cmd bytes for file, the write-path counterpart to
// cmd.Parse/cmd.ParseDocument. original is the .cmd file's own previously
// loaded bytes — empty when file describes a brand new file with no
// original yet.
//
// Same byte-exact-when-unmodified / fresh-serialize-when-edited strategy as
// SerializeDef — see that function's doc comment and
// .vibe/decisions/028-wasm-save-path-per-format-diff-or-serialize.md.
func SerializeCmd(original []byte, file cmd.CommandFile) ([]byte, error) {
	edited := normalizeCommandFileForJSON(file)

	if len(original) > 0 {
		doc, err := cmd.ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("character: parsing original command file for save: %w", err)
		}

		baseline := normalizeCommandFileForJSON(doc.File)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("character: writing unmodified command file: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := cmd.Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("character: serializing edited command file: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeCommandFileForJSON returns file with every nil slice/map
// (Remap, Commands, and States — via normalizeStateDefsForJSON) replaced by
// its non-nil empty equivalent, for the same reason
// normalizeCharacterInfoForJSON exists.
func normalizeCommandFileForJSON(file cmd.CommandFile) cmd.CommandFile {
	if file.Remap == nil {
		file.Remap = map[string]string{}
	}
	if file.Commands == nil {
		file.Commands = []cmd.Command{}
	}
	file.States = normalizeStateDefsForJSON(file.States)
	return file
}

// SerializeZss produces .zss bytes for script, the write-path counterpart
// to zss.Parse/zss.ParseDocument. original is the .zss file's own
// previously loaded bytes — empty when script describes a brand new file
// with no original yet.
//
// Same byte-exact-when-unmodified / fresh-serialize-when-edited strategy as
// SerializeDef — see that function's doc comment and
// .vibe/decisions/028-wasm-save-path-per-format-diff-or-serialize.md.
func SerializeZss(original []byte, script zss.Script) ([]byte, error) {
	edited := normalizeScriptForJSON(script)

	if len(original) > 0 {
		doc, err := zss.ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("character: parsing original state script file for save: %w", err)
		}

		baseline := normalizeScriptForJSON(doc.Script)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("character: writing unmodified state script file: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := zss.Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("character: serializing edited state script file: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeScriptForJSON returns script with every nil slice/map (Blocks,
// and each Block's HeaderParams/Params/Ret) replaced by its non-nil empty
// equivalent, for the same reason normalizeCharacterInfoForJSON exists.
func normalizeScriptForJSON(script zss.Script) zss.Script {
	if script.Blocks == nil {
		script.Blocks = []zss.Block{}
	}
	for i := range script.Blocks {
		if script.Blocks[i].HeaderParams == nil {
			script.Blocks[i].HeaderParams = map[string]string{}
		}
		if script.Blocks[i].Params == nil {
			script.Blocks[i].Params = []string{}
		}
		if script.Blocks[i].Ret == nil {
			script.Blocks[i].Ret = []string{}
		}
	}
	return script
}

// normalizeCharacterInfoForJSON replaces every nil slice on info with its
// non-nil empty equivalent, in place — the same normalization LoadBytes
// already applies to Character's own JSON contract (see
// normalizeForJSON), needed here so a caller round-tripping an unmodified
// CharacterInfo through JSON (where an absent/empty list becomes "[]", not
// "null") is never spuriously treated as an edit.
func normalizeCharacterInfoForJSON(info *def.CharacterInfo) {
	if info.StateFiles == nil {
		info.StateFiles = []string{}
	}
	if info.Palettes == nil {
		info.Palettes = []string{}
	}
}

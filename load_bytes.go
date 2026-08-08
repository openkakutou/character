package character

import (
	"bytes"
	"fmt"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/character/sff"
)

// LoadBytes assembles a fully populated Character from .def/.air/.sff/.cns
// content already held in memory as byte buffers, rather than read from a
// filesystem path the way Load does.
//
// It exists for consumers with no filesystem access — chiefly the WASM
// entrypoint (cmd/wasm), whose JS caller has already fetched or selected
// each file's bytes itself. Unlike Load, LoadBytes does not resolve or
// follow the .def file's own referenced file paths: the caller supplies the
// .air/.sff/.cns content directly, already knowing which bytes belong to
// which file.
//
// The returned Character is JSON-marshal-ready for that same JS boundary:
// every slice and map reachable from it is guaranteed non-nil (encoding/json
// renders empty as "[]"/"{}", never "null" — see
// .vibe/decisions/019-wasm-entrypoint-byte-buffer-loading-and-json-contract.md).
//
// A malformed or truncated buffer for any of the four inputs returns a
// descriptive error naming which one failed, rather than panicking; on any
// error the returned Character is always nil, never partially populated.
func LoadBytes(defBytes, airBytes, sffBytes, cnsBytes []byte) (*Character, error) {
	info, err := def.Parse(bytes.NewReader(defBytes))
	if err != nil {
		return nil, fmt.Errorf("character: parsing character definition bytes: %w", err)
	}

	animations, err := air.Parse(bytes.NewReader(airBytes))
	if err != nil {
		return nil, fmt.Errorf("character: parsing animation bytes: %w", err)
	}

	sprites, err := sff.Load(bytes.NewReader(sffBytes))
	if err != nil {
		return nil, fmt.Errorf("character: loading sprite bytes: %w", err)
	}

	stateDefs, err := cns.Parse(bytes.NewReader(cnsBytes))
	if err != nil {
		return nil, fmt.Errorf("character: parsing combat logic bytes: %w", err)
	}

	c := &Character{
		Name:       info.Name,
		Animations: animations,
		Sprites:    sprites,
		StateDefs:  stateDefs,
	}
	normalizeForJSON(c)
	return c, nil
}

// normalizeForJSON replaces every nil slice/map reachable from c with its
// non-nil, empty equivalent, in place. Parsers are free to leave a nil slice
// or map when a file declares nothing for it (idiomatic, zero-cost Go); this
// function exists solely so LoadBytes's JSON output never surprises a JS
// caller with "null" where an empty, iterable collection was expected.
func normalizeForJSON(c *Character) {
	if c.Animations == nil {
		c.Animations = []air.Animation{}
	}
	for i := range c.Animations {
		normalizeAnimationForJSON(&c.Animations[i])
	}

	if c.Sprites == nil {
		c.Sprites = []sff.SpriteGroup{}
	}
	for i := range c.Sprites {
		if c.Sprites[i].Sprites == nil {
			c.Sprites[i].Sprites = []sff.Sprite{}
		}
	}

	if c.StateDefs == nil {
		c.StateDefs = []cns.StateDef{}
	}
	for i := range c.StateDefs {
		normalizeStateDefForJSON(&c.StateDefs[i])
	}
}

func normalizeAnimationForJSON(a *air.Animation) {
	if a.Frames == nil {
		a.Frames = []air.Frame{}
	}
	for i := range a.Frames {
		if a.Frames[i].Clsn1 == nil {
			a.Frames[i].Clsn1 = []air.ClsnBox{}
		}
		if a.Frames[i].Clsn2 == nil {
			a.Frames[i].Clsn2 = []air.ClsnBox{}
		}
	}
}

func normalizeStateDefForJSON(s *cns.StateDef) {
	if s.Controllers == nil {
		s.Controllers = []cns.Controller{}
	}
	for i := range s.Controllers {
		if s.Controllers[i].Triggers == nil {
			s.Controllers[i].Triggers = []string{}
		}
		if s.Controllers[i].Parameters == nil {
			s.Controllers[i].Parameters = map[string]string{}
		}
	}
}

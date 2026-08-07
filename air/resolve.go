package air

import (
	"fmt"

	"github.com/openkakutou/character/sff"
)

// SpriteResolver resolves a Frame's (Group, Image) reference against a
// loaded set of sprite groups. sff.Sprite/sff.SpriteGroup are the same
// version-agnostic shape regardless of whether they came from a .sff v1 or
// v2 file, so Resolve needs no version-specific branching.
//
// This is the first cross-package integration point between air and sff:
// per CLAUDE.md, .air and .sff are interdependent formats, not independent
// ones — a Frame reference is meaningless without a sprite collection to
// resolve it against. See
// .vibe/decisions/008-air-sprite-resolution-lives-in-air-package.md.
type SpriteResolver struct {
	sprites map[spriteKey]sff.Sprite
}

// spriteKey is the (Group, Image) pair identifying a sprite, shared by
// sff.Sprite and air.Frame.
type spriteKey struct {
	group int
	image int
}

// NewSpriteResolver indexes the given sprite groups by (Group, Image) so
// Resolve can look up any Frame's sprite without rescanning groups on every
// call. Passing nil or an empty slice is valid and produces a resolver for
// which every Resolve call fails.
func NewSpriteResolver(groups []sff.SpriteGroup) *SpriteResolver {
	sprites := make(map[spriteKey]sff.Sprite)
	for _, group := range groups {
		for _, sprite := range group.Sprites {
			sprites[spriteKey{group: sprite.Group, image: sprite.Image}] = sprite
		}
	}
	return &SpriteResolver{sprites: sprites}
}

// Resolve returns the Sprite that frame's (Group, Image) reference names.
// It returns a descriptive error, rather than a zero Sprite, when no sprite
// with a matching (Group, Image) exists in the resolver — a missing
// reference must fail explicitly rather than silently rendering blank.
//
// A blank frame (frame.IsBlank(), the ".air" -1 "no sprite" sentinel) is not
// a missing reference: Resolve recognizes it directly and returns a zero
// Sprite with a nil error, without looking it up.
func (r *SpriteResolver) Resolve(frame Frame) (sff.Sprite, error) {
	if frame.IsBlank() {
		return sff.Sprite{}, nil
	}
	sprite, ok := r.sprites[spriteKey{group: frame.Group, image: frame.Image}]
	if !ok {
		return sff.Sprite{}, fmt.Errorf("air: frame references sprite (group %d, image %d), which was not found in the loaded sprite groups", frame.Group, frame.Image)
	}
	return sprite, nil
}

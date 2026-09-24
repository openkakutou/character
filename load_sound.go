package character

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/openkakutou/snd"
)

// sndSignature is the fixed 12-byte signature every MUGEN/Ikemen GO .snd
// file (v1 or v2) starts with, and sndVersionByteOffset is where the byte
// distinguishing the two versions lives — mirrors
// github.com/openkakutou/snd's own (unexported) detection logic. This
// package needs its own version peek, distinct from snd's, because snd only
// exposes a way to decode one already-known (group, sample) entry at a time
// (ParseV1/ParseV2 plus DecodeV1Sound/DecodeV2Sound) — there is no exported
// "decode every entry" function yet, but Load/LoadBytes need every entry a
// character's .snd file declares, not just one picked in advance. This is
// the same "duplicate a small, well-documented file-format detail rather
// than depend on an unexported helper" precedent the sibling sff module
// already established for .sff's own version detection (see sff's
// load.go:detectVersion).
const (
	sndSignature         = "ElecbyteSnd\x00"
	sndVersionByteOffset = 12
	sndVersion2Marker    = 2
	sndHeaderPeekSize    = sndVersionByteOffset + 1
)

// decodeSoundGroups decodes every sound entry declared in a MUGEN/Ikemen GO
// .snd file's data (v1 or v2, auto-detected the same way snd's own decode
// entry point does), grouping the results by Group the same way loadSprites
// groups sff.Sprite values via groupSprites. context names the .snd file (a
// path, or a fixed label for LoadBytes) for descriptive errors.
//
// openExternal resolves a v2 entry's Ikemen GO external-file-reference
// extension (see snd.ExternalAudioOpener); pass nil when there is no way to
// resolve one (e.g. from LoadBytes, which has no filesystem access) — an
// entry that actually needs it then reports a descriptive error rather than
// silently losing data.
func decodeSoundGroups(data []byte, context string, openExternal snd.ExternalAudioOpener) ([]SoundGroup, error) {
	if len(data) < sndHeaderPeekSize {
		return nil, fmt.Errorf("character: sound file %q: too short to be a valid .snd file", context)
	}
	if sig := string(data[0:sndVersionByteOffset]); sig != sndSignature {
		return nil, fmt.Errorf("character: sound file %q: not a .snd file (unexpected signature %q)", context, sig)
	}

	r := bytes.NewReader(data)
	if data[sndVersionByteOffset] == sndVersion2Marker {
		return decodeV2SoundGroups(r, context, openExternal)
	}
	return decodeV1SoundGroups(r, context)
}

// decodeV1SoundGroups parses and decodes every entry in a .snd v1 file.
func decodeV1SoundGroups(r io.ReaderAt, context string) ([]SoundGroup, error) {
	table, err := snd.ParseV1(r)
	if err != nil {
		return nil, fmt.Errorf("character: parsing sound file %q: %w", context, err)
	}

	sounds := make([]Sound, len(table.Sounds))
	for i, e := range table.Sounds {
		decoded, err := snd.DecodeV1Sound(r, table, e.Group, e.Sample)
		if err != nil {
			return nil, fmt.Errorf("character: decoding sound file %q: %w", context, err)
		}
		sounds[i] = toSound(e.Group, e.Sample, decoded)
	}
	return groupSounds(sounds), nil
}

// decodeV2SoundGroups parses and decodes every entry in a .snd v2 file,
// resolving an Ikemen GO external-file-reference entry via openExternal
// when one is present.
func decodeV2SoundGroups(r io.ReaderAt, context string, openExternal snd.ExternalAudioOpener) ([]SoundGroup, error) {
	table, err := snd.ParseV2(r)
	if err != nil {
		return nil, fmt.Errorf("character: parsing sound file %q: %w", context, err)
	}

	sounds := make([]Sound, len(table.Sounds))
	for i, e := range table.Sounds {
		decoded, err := snd.DecodeV2Sound(r, table, e.Group, e.Sample, openExternal)
		if err != nil {
			return nil, fmt.Errorf("character: decoding sound file %q: %w", context, err)
		}
		sounds[i] = toSound(e.Group, e.Sample, decoded)
	}
	return groupSounds(sounds), nil
}

// toSound converts one decoded snd entry to this package's own read-path
// Sound type.
func toSound(group, sample int, decoded *snd.DecodedSound) Sound {
	return Sound{
		Group:         group,
		Sample:        sample,
		SampleRate:    decoded.SampleRate,
		Channels:      decoded.Channels,
		BitsPerSample: decoded.BitsPerSample,
		PCM:           decoded.PCM,
	}
}

// groupSounds buckets sounds by Group into SoundGroup values, preserving
// each sound's relative order within its group, and returns the groups
// sorted by ascending group index — mirrors sff's own groupSprites.
func groupSounds(sounds []Sound) []SoundGroup {
	var order []int
	byGroup := make(map[int][]Sound)
	for _, s := range sounds {
		if _, ok := byGroup[s.Group]; !ok {
			order = append(order, s.Group)
		}
		byGroup[s.Group] = append(byGroup[s.Group], s)
	}

	sort.Ints(order)
	groups := make([]SoundGroup, len(order))
	for i, g := range order {
		groups[i] = SoundGroup{Index: g, Sounds: byGroup[g]}
	}
	return groups
}

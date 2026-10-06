package character

import (
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
)

// sndSignature mirrors github.com/openkakutou/snd's own (unexported)
// v1Signature constant: the fixed 12-byte signature every .snd file (v1 or
// v2) starts with. Duplicated here purely for building test fixtures — this
// package's own production code never assumes a specific literal value
// beyond what sound.go's decodeSoundGroups already documents.
const testSndSignature = "ElecbyteSnd\x00"

// buildWAVFixture assembles a minimal RIFF/WAVE blob carrying uncompressed
// PCM, the shape a real .snd entry's embedded audio data must have for
// snd.DecodeV1Sound/DecodeV2Sound to decode it.
func buildWAVFixture(t *testing.T, channels, sampleRate, bitsPerSample int, data []byte) []byte {
	t.Helper()

	fmtChunk := make([]byte, 16)
	binary.LittleEndian.PutUint16(fmtChunk[0:2], 1) // PCM format tag
	binary.LittleEndian.PutUint16(fmtChunk[2:4], uint16(channels))
	binary.LittleEndian.PutUint32(fmtChunk[4:8], uint32(sampleRate))
	blockAlign := channels * bitsPerSample / 8
	binary.LittleEndian.PutUint32(fmtChunk[8:12], uint32(sampleRate*blockAlign))
	binary.LittleEndian.PutUint16(fmtChunk[12:14], uint16(blockAlign))
	binary.LittleEndian.PutUint16(fmtChunk[14:16], uint16(bitsPerSample))

	var buf []byte
	buf = append(buf, []byte("RIFF")...)
	buf = append(buf, 0, 0, 0, 0) // filled in below
	buf = append(buf, []byte("WAVE")...)

	buf = append(buf, []byte("fmt ")...)
	buf = appendUint32Fixture(buf, uint32(len(fmtChunk)))
	buf = append(buf, fmtChunk...)

	buf = append(buf, []byte("data")...)
	buf = appendUint32Fixture(buf, uint32(len(data)))
	buf = append(buf, data...)
	if len(data)%2 == 1 {
		buf = append(buf, 0)
	}

	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)-8))
	return buf
}

func appendUint32Fixture(b []byte, v uint32) []byte {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], v)
	return append(b, tmp[:]...)
}

// sndFixtureEntry is one sound entry for buildV1SndFile/buildV2SndFile.
type sndFixtureEntry struct {
	group, sample int
	payload       []byte
}

// buildV1SndFile assembles a minimal, well-formed .snd v1 file in memory
// from a list of (group, sample, payload) entries, chaining each
// subheader's NextSubHeaderOffset the way a real file does. The subheader
// itself is byte-for-byte identical to buildV2SndFile's (Group/Sample as
// 4-byte fields, no reserved region) — real files, Ikemen GO's own reference
// loader, and github.com/openkakutou/snd's ParseV1 all agree v1 and v2 share
// this exact layout; only the header's version stamp differs. See
// github.com/openkakutou/snd's own decision
// 005-v1-group-sample-fields-are-4-bytes-not-2.md.
func buildV1SndFile(t *testing.T, entries []sndFixtureEntry) []byte {
	t.Helper()

	const headerSize = 24
	const subheaderSize = 16
	var buf []byte

	header := make([]byte, headerSize)
	copy(header[0:12], []byte(testSndSignature))
	header[12], header[13], header[14], header[15] = 1, 0, 1, 0
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(entries)))
	binary.LittleEndian.PutUint32(header[20:24], uint32(headerSize))
	buf = append(buf, header...)

	offsets := make([]int, len(entries))
	pos := headerSize
	for i, e := range entries {
		offsets[i] = pos
		pos += subheaderSize + len(e.payload)
	}

	for i, e := range entries {
		sub := make([]byte, subheaderSize)
		if i+1 < len(entries) {
			binary.LittleEndian.PutUint32(sub[0:4], uint32(offsets[i+1]))
		}
		binary.LittleEndian.PutUint32(sub[4:8], uint32(len(e.payload)))
		binary.LittleEndian.PutUint32(sub[8:12], uint32(int32(e.group)))
		binary.LittleEndian.PutUint32(sub[12:16], uint32(int32(e.sample)))
		buf = append(buf, sub...)
		buf = append(buf, e.payload...)
	}

	return buf
}

// buildV2SndFile is buildV1SndFile's v2 counterpart: a v2 subheader stores
// Group and Sample as 4-byte fields rather than 2-byte ones, and the header
// declares version byte 2.
func buildV2SndFile(t *testing.T, entries []sndFixtureEntry) []byte {
	t.Helper()

	const headerSize = 24
	const subheaderSize = 16
	var buf []byte

	header := make([]byte, headerSize)
	copy(header[0:12], []byte(testSndSignature))
	header[12], header[13], header[14], header[15] = 2, 0, 0, 0
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(entries)))
	binary.LittleEndian.PutUint32(header[20:24], uint32(headerSize))
	buf = append(buf, header...)

	offsets := make([]int, len(entries))
	pos := headerSize
	for i, e := range entries {
		offsets[i] = pos
		pos += subheaderSize + len(e.payload)
	}

	for i, e := range entries {
		sub := make([]byte, subheaderSize)
		if i+1 < len(entries) {
			binary.LittleEndian.PutUint32(sub[0:4], uint32(offsets[i+1]))
		}
		binary.LittleEndian.PutUint32(sub[4:8], uint32(len(e.payload)))
		binary.LittleEndian.PutUint32(sub[8:12], uint32(int32(e.group)))
		binary.LittleEndian.PutUint32(sub[12:16], uint32(int32(e.sample)))
		buf = append(buf, sub...)
		buf = append(buf, e.payload...)
	}

	return buf
}

// TestDecodeSoundGroups_OneUndecodableEntry_KeepsOthersAndFlagsIt covers
// backlog item 058 for both .snd versions: a single corrupt entry stays in
// its group with a descriptive Error while every other entry still decodes.
func TestDecodeSoundGroups_OneUndecodableEntry_KeepsOthersAndFlagsIt(t *testing.T) {
	good := buildWAVFixture(t, 1, 44100, 16, []byte{1, 0, 2, 0})
	entries := []sndFixtureEntry{
		{group: 1, sample: 0, payload: good},
		{group: 1, sample: 1, payload: []byte("not a wav at all")},
		{group: 2, sample: 0, payload: good},
	}
	builders := map[string]func(*testing.T, []sndFixtureEntry) []byte{
		"v1": buildV1SndFile,
		"v2": buildV2SndFile,
	}
	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			groups, err := decodeSoundGroups(build(t, entries), "x.snd", nil)
			if err != nil {
				t.Fatalf("expected no hard error, got %v", err)
			}
			if len(groups) != 2 || len(groups[0].Sounds) != 2 || len(groups[1].Sounds) != 1 {
				t.Fatalf("unexpected grouping: %+v", groups)
			}
			bad := groups[0].Sounds[1]
			if bad.Group != 1 || bad.Sample != 1 || bad.PCM != nil {
				t.Errorf("bad sound should keep its key and have no PCM: %+v", bad)
			}
			for _, want := range []string{"group 1", "sample 1"} {
				if !strings.Contains(bad.Error, want) {
					t.Errorf("Error %q should mention %q", bad.Error, want)
				}
			}
			for _, ok := range []Sound{groups[0].Sounds[0], groups[1].Sounds[0]} {
				if ok.Error != "" || len(ok.PCM) != 2 {
					t.Errorf("good sound corrupted: %+v", ok)
				}
			}
		})
	}
}

// TestSound_JSON_OmitsErrorForGoodSounds keeps the wire format unchanged for
// every consumer of good sounds.
func TestSound_JSON_OmitsErrorForGoodSounds(t *testing.T) {
	data, err := json.Marshal(Sound{Group: 1, PCM: []int16{1}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"error"`) {
		t.Errorf("good sound should not carry an error field: %s", data)
	}
	data, _ = json.Marshal(Sound{Group: 1, Error: "boom"})
	if !strings.Contains(string(data), `"error":"boom"`) {
		t.Errorf("bad sound should marshal its error: %s", data)
	}
}

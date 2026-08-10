package character

import (
	"strings"
	"testing"

	"github.com/openkakutou/character/zss"
)

func TestSerializeZss_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[Statedef 200; type: S; physics: N; movetype: A; ctrl: 0;]
if AnimElem = 1 {
	mapAdd{map:"chain_5a";value:1}
}

[Function Eff_5a()]
helper{stateno: 6721; id: 6721;}
`)

	script, err := zss.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: zss.Parse failed: %v", err)
	}

	out, err := SerializeZss(original, script)
	if err != nil {
		t.Fatalf("SerializeZss returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeZss_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[Statedef 200; type: S; physics: N; movetype: A; ctrl: 0;]
if AnimElem = 1 {
	mapAdd{map:"chain_5a";value:1}
}
`)

	script, err := zss.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: zss.Parse failed: %v", err)
	}
	script.Blocks[0].HeaderParams["ctrl"] = "1"

	out, err := SerializeZss(original, script)
	if err != nil {
		t.Fatalf("SerializeZss returned an error: %v", err)
	}
	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := zss.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.Blocks[0].HeaderParams["ctrl"] != "1" {
		t.Fatalf("expected edited ctrl header param to survive round trip, got %q", roundTripped.Blocks[0].HeaderParams["ctrl"])
	}
}

func TestSerializeZss_EmptyOriginal_SerializesFreshForNewScript(t *testing.T) {
	script := zss.Script{
		Blocks: []zss.Block{
			{Kind: zss.BlockKindFunction, Name: "NewHelper", Body: "helper{stateno: 1;}\n"},
		},
	}

	out, err := SerializeZss(nil, script)
	if err != nil {
		t.Fatalf("SerializeZss returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for a brand new script")
	}

	roundTripped, err := zss.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if len(roundTripped.Blocks) != 1 || roundTripped.Blocks[0].Name != "NewHelper" {
		t.Fatalf("expected the new block to survive round trip, got %+v", roundTripped.Blocks)
	}
}

func TestSerializeZss_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[NotABlock]\nsomething\n")

	_, err := SerializeZss(malformed, zss.Script{})
	if err == nil {
		t.Fatalf("expected an error for a malformed original .zss, got nil")
	}
	if !strings.Contains(err.Error(), "zss:") {
		t.Fatalf("expected error to mention the zss package's own diagnostic, got: %v", err)
	}
}

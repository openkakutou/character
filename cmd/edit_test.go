package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func editedWith(t *testing.T, src string, mutate func(f *CommandFile)) string {
	t.Helper()
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	file, err := Parse(strings.NewReader(src)) // independent copy to mutate
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	mutate(&file)
	var buf bytes.Buffer
	if err := doc.SerializeEdited(&buf, file); err != nil {
		t.Fatalf("SerializeEdited: %v", err)
	}
	return buf.String()
}

func expectEdit(t *testing.T, src, want string, mutate func(f *CommandFile)) {
	t.Helper()
	if got := editedWith(t, src, mutate); got != want {
		t.Errorf("edited output mismatch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

const editSrc = `; header comment
[Remap]
A = a ; keep me
b = b

[Defaults]
Command.Time  = 15 ; window
command.buffer.time = 1

; the first
[Command]
name = "a"
Command = a ; plain
time = 1

; the second
[Command]
name = "QCF_a"
command = ~D, DF, F, a

[Statedef -1]

[State -1, QCF Special]
type = ChangeState
value = 1000
trigger1 = command = "QCF_a"
trigger1 = Ctrl ; control
`

func TestSerializeEdited_ChangedCommandInput_RewritesOnlyItsCommandLine(t *testing.T) {
	want := strings.Replace(editSrc, "Command = a ; plain", "Command = ~a ; plain", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.Commands[0].Input = "~a" })
}

func TestSerializeEdited_NewlyNonZeroTime_AddsKeyAfterLastKeyLine(t *testing.T) {
	want := strings.Replace(editSrc, "command = ~D, DF, F, a\n", "command = ~D, DF, F, a\ntime = 20\nbuffer.time = 3\n", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.Commands[1].Time = 20; f.Commands[1].BufferTime = 3 })
}

func TestSerializeEdited_TimeResetToZero_DeletesItsLine(t *testing.T) {
	want := strings.Replace(editSrc, "\ntime = 1\n", "\n", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.Commands[0].Time = 0 })
}

func TestSerializeEdited_AddedCommand_GoesAfterItsPredecessor(t *testing.T) {
	want := strings.Replace(editSrc, "time = 1\n\n; the second", "time = 1\n\n[Command]\nname = \"b\"\ncommand = b\n\n; the second", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) {
		f.Commands = []Command{f.Commands[0], {Name: "b", Input: "b"}, f.Commands[1]}
	})
}

func TestSerializeEdited_AddedCommandAtEnd_GoesBeforeStates(t *testing.T) {
	want := strings.Replace(editSrc, "command = ~D, DF, F, a\n\n[Statedef -1]", "command = ~D, DF, F, a\n\n[Command]\nname = \"z\"\ncommand = z\ntime = 4\n\n[Statedef -1]", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) {
		f.Commands = append(f.Commands, Command{Name: "z", Input: "z", Time: 4})
	})
}

func TestSerializeEdited_RemovedCommand_DeletesItWithItsLeadingComment(t *testing.T) {
	want := strings.Replace(editSrc, "; the first\n[Command]\nname = \"a\"\nCommand = a ; plain\ntime = 1\n\n", "", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.Commands = f.Commands[1:] })
}

func TestSerializeEdited_DuplicateNames_EditingSecondLeavesFirstUntouched(t *testing.T) {
	src := "[Command]\nname = \"x\"\ncommand = a\n\n[Command]\nname = \"x\"\ncommand = b ; alt\n"
	want := "[Command]\nname = \"x\"\ncommand = a\n\n[Command]\nname = \"x\"\ncommand = c ; alt\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Commands[1].Input = "c" })
}

func TestSerializeEdited_DefaultsTimeChange_PatchesInPlaceKeepingSpellingAndComment(t *testing.T) {
	want := strings.Replace(editSrc, "Command.Time  = 15 ; window", "Command.Time  = 12 ; window", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.Defaults.Time = 12 })
}

func TestSerializeEdited_MissingDefaults_CreatedBeforeFirstCommand(t *testing.T) {
	src := "[Remap]\na = a\n\n[Command]\nname = \"a\"\ncommand = a\n"
	want := "[Remap]\na = a\n\n[Defaults]\ncommand.time = 15\ncommand.buffer.time = 1\n\n[Command]\nname = \"a\"\ncommand = a\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Defaults = CommandDefaults{Time: 15, BufferTime: 1} })
}

func TestSerializeEdited_RemapChange_PatchesValueAddsSortedKeyAndDeletesRemoved(t *testing.T) {
	want := strings.Replace(editSrc, "A = a ; keep me\nb = b\n", "A = c ; keep me\nx = y\n", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) {
		f.Remap = map[string]string{"a": "c", "x": "y"}
	})
}

func TestSerializeEdited_RemapSeveralNewKeys_InsertedSortedAfterLastKey(t *testing.T) {
	src := "[Remap]\na = a\n\n[Command]\nname = \"a\"\ncommand = a\n"
	want := "[Remap]\na = a\nb = 1\nm = 2\nz = 3\n\n[Command]\nname = \"a\"\ncommand = a\n"
	expectEdit(t, src, want, func(f *CommandFile) {
		f.Remap = map[string]string{"a": "a", "z": "3", "b": "1", "m": "2"}
	})
}

func TestSerializeEdited_StateTriggerEdit_PatchedInPlace(t *testing.T) {
	want := strings.Replace(editSrc, "trigger1 = Ctrl ; control", "trigger1 = Alive ; control", 1)
	expectEdit(t, editSrc, want, func(f *CommandFile) { f.States[0].Controllers[0].Triggers[1] = "Alive" })
}

func TestSerializeEdited_ImplicitStatedef_EditKeepsNoSyntheticHeader(t *testing.T) {
	src := "[Command]\nname = \"a\"\ncommand = a\n\n; go\n[State -1, Go]\ntype = ChangeState\nvalue = 200\ntrigger1 = command = \"a\"\n"
	want := strings.Replace(src, "value = 200", "value = 300", 1)
	expectEdit(t, src, want, func(f *CommandFile) { f.States[0].Controllers[0].Parameters["value"] = "300" })
	// combined with a command edit
	want2 := strings.Replace(want, "command = a", "command = b", 1)
	expectEdit(t, src, want2, func(f *CommandFile) {
		f.States[0].Controllers[0].Parameters["value"] = "300"
		f.Commands[0].Input = "b"
	})
}

func TestSerializeEdited_ImplicitStatedef_NoStateEdit_ByteExact(t *testing.T) {
	src := "[Command]\nname = \"a\"\ncommand = a\n\n[State -1, Go]\ntype = Null\ntrigger1 = 1\n"
	want := strings.Replace(src, "command = a", "command = b", 1)
	expectEdit(t, src, want, func(f *CommandFile) { f.Commands[0].Input = "b" })
}

func TestSerializeEdited_CRLF_KeepsLineEndingsOnPatchedAndGeneratedLines(t *testing.T) {
	src := "[Defaults]\r\ncommand.time = 15\r\n\r\n[Command]\r\nname = \"a\"\r\ncommand = a\r\n"
	want := "[Defaults]\r\ncommand.time = 15\r\n\r\n[Command]\r\nname = \"a\"\r\ncommand = b\r\ntime = 2\r\n\r\n[Command]\r\nname = \"c\"\r\ncommand = c\r\n"
	expectEdit(t, src, want, func(f *CommandFile) {
		f.Commands[0].Input = "b"
		f.Commands[0].Time = 2
		f.Commands = append(f.Commands, Command{Name: "c", Input: "c"})
	})
}

func TestSerializeEdited_Unchanged_IsByteExact(t *testing.T) {
	srcs := []string{editSrc}
	b, err := os.ReadFile("testdata/sample.cmd")
	if err != nil {
		t.Fatal(err)
	}
	srcs = append(srcs, string(b), "[Command]\r\nname = \"a\"\r\ncommand = a\r\n[State -1]\r\ntype = Null\r\ntrigger1 = 1", "")
	for _, src := range srcs {
		if got := editedWith(t, src, func(*CommandFile) {}); got != src {
			t.Errorf("no-op edit changed the source\n got: %q\nwant: %q", got, src)
		}
	}
}

type brokenWriter struct{}

var errBroken = errors.New("disk full")

func (brokenWriter) Write([]byte) (int, error) { return 0, errBroken }

func TestSerializeEdited_FailingWriter_ReturnsWrappedError(t *testing.T) {
	doc, _ := ParseDocument(strings.NewReader(editSrc))
	file := doc.File
	file.Commands = append([]Command(nil), file.Commands...)
	file.Commands[0].Input = "x"
	if err := doc.SerializeEdited(brokenWriter{}, file); !errors.Is(err, errBroken) {
		t.Errorf("want wrapped errBroken, got %v", err)
	}
}

func TestSerializeEdited_MissingRemap_CreatedAtEndOfFileWhenNoOtherSection(t *testing.T) {
	src := "; only a comment"
	want := "; only a comment\n\n[Remap]\nx = y\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Remap = map[string]string{"x": "y"} })
}

func TestSerializeEdited_ImplicitStatedef_HeaderFieldChange_EmitsRealHeaderBeforeKeys(t *testing.T) {
	src := "[Command]\nname = \"a\"\ncommand = a\n\n[State -1, x]\ntype = VarSet\ntrigger1 = 1\nvalue = 1\n"
	want := "[Command]\nname = \"a\"\ncommand = a\n\n[Statedef -1]\nanim = 1\n[State -1, x]\ntype = VarSet\ntrigger1 = 1\nvalue = 1\n"
	got := editedWith(t, src, func(f *CommandFile) { f.States[0].Anim = 1 })
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	f, err := Parse(strings.NewReader(got))
	if err != nil || f.States[0].Anim != 1 {
		t.Errorf("reparse: anim=%v err=%v", f.States, err)
	}
}

func TestSerializeEdited_DefaultsMissingOnlyBufferTime_AppendsItAfterLastKeyLine(t *testing.T) {
	src := "[Defaults]\ncommand.time = 15 ; w\n\n[Command]\nname = \"a\"\ncommand = a\n"
	want := "[Defaults]\ncommand.time = 15 ; w\ncommand.buffer.time = 2\n\n[Command]\nname = \"a\"\ncommand = a\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Defaults.BufferTime = 2 })
}

func TestSerializeEdited_CommandWithoutNameLine_NameAddedAfterLastKeyLine(t *testing.T) {
	src := "[Command]\ncommand = a ; x\ntime = 3\n\n[Statedef -1]\n"
	want := "[Command]\ncommand = a ; x\ntime = 3\nname = \"hold\"\n\n[Statedef -1]\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Commands[0].Name = "hold" })
}

func TestSerializeEdited_CommandWithNoPredecessor_InsertedBeforeFirstCommand(t *testing.T) {
	src := "[Defaults]\ncommand.time = 15\n\n; first\n[Command]\nname = \"a\"\ncommand = a\n"
	want := "[Defaults]\ncommand.time = 15\n\n[Command]\nname = \"n\"\ncommand = b\n\n; first\n[Command]\nname = \"a\"\ncommand = a\n"
	expectEdit(t, src, want, func(f *CommandFile) {
		f.Commands = append([]Command{{Name: "n", Input: "b"}}, f.Commands...)
	})
}

func TestSerializeEdited_CommandAddedToFileWithOnlyStates_InsertedBeforeFirstState(t *testing.T) {
	src := "[Statedef -1]\n\n[State -1, x]\ntype = Null\ntrigger1 = 1\n"
	want := "[Command]\nname = \"n\"\ncommand = b\n\n[Statedef -1]\n\n[State -1, x]\ntype = Null\ntrigger1 = 1\n"
	expectEdit(t, src, want, func(f *CommandFile) { f.Commands = []Command{{Name: "n", Input: "b"}} })
}

func TestSerializeEdited_TimeWrittenWithBothSpellings_PatchesTheLineParseUses(t *testing.T) {
	src := "[Command]\nname = \"a\"\ncommand.time = 5\ncommand = a\ntime = 7\nbuffer.time = 1\ncommand.buffer.time = 2\n"
	want := "[Command]\nname = \"a\"\ncommand.time = 5\ncommand = a\ntime = 9\nbuffer.time = 1\ncommand.buffer.time = 4\n"
	expectEdit(t, src, want, func(f *CommandFile) {
		if f.Commands[0].Time != 7 || f.Commands[0].BufferTime != 2 {
			t.Fatalf("parse precondition: %+v", f.Commands[0])
		}
		f.Commands[0].Time, f.Commands[0].BufferTime = 9, 4
	})
	src2 := "[Command]\nname = \"a\"\ntime = 7\ncommand = a\ncommand.time = 5\n"
	want2 := "[Command]\nname = \"a\"\ntime = 7\ncommand = a\ncommand.time = 9\n"
	expectEdit(t, src2, want2, func(f *CommandFile) { f.Commands[0].Time = 9 })
}

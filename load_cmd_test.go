package character

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cmdFixtureContent is a small .cmd file with comments, a [Remap] section, a
// [Defaults] section, two commands, and a command-linked state section, so a
// byte-for-byte round trip proves comments and ordering survive.
const cmdFixtureContent = `; fixture .cmd with comments

[Remap]
x = y

[Defaults]
command.time = 15
command.buffer.time = 1

; a plain button
[Command]
name = "a"
command = a
time = 1

[Command]
name = "QCF_x"
command = ~D, DF, F, x

[Statedef -1]

[State -1, Kick]
type = ChangeState
value = 200
triggerall = command = "a"
trigger1 = ctrl
`

// writeCmdFixtureCharacter returns the path of a .def referencing char.cmd,
// whose content is cmdContent (when write is true; otherwise no file is
// created, so the reference dangles). declare=false omits the cmd entry.
func writeCmdFixtureCharacter(t *testing.T, declare, write bool, cmdContent string) string {
	t.Helper()
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if declare {
		f, err := os.OpenFile(defPath, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("test setup: %v", err)
		}
		if _, err := f.WriteString("cmd = char.cmd\n"); err != nil {
			t.Fatalf("test setup: %v", err)
		}
		f.Close()
	}
	if write {
		if err := os.WriteFile(filepath.Join(dir, "char.cmd"), []byte(cmdContent), 0o644); err != nil {
			t.Fatalf("test setup: %v", err)
		}
	}
	return defPath
}

func TestLoad_DeclaredCmd_ExposesParsedCommands(t *testing.T) {
	c, err := Load(writeCmdFixtureCharacter(t, true, true, cmdFixtureContent))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.CommandError != "" {
		t.Fatalf("CommandError = %q, want empty", c.CommandError)
	}
	if len(c.Commands.Commands) != 2 || c.Commands.Commands[1].Name != "QCF_x" {
		t.Errorf("Commands.Commands = %+v, want 2 commands, second QCF_x", c.Commands.Commands)
	}
	if c.Commands.Remap["x"] != "y" {
		t.Errorf("Remap = %v, want x=y", c.Commands.Remap)
	}
	if len(c.Commands.States) != 1 {
		t.Errorf("Commands.States = %d, want 1", len(c.Commands.States))
	}
}

func TestLoad_MissingCmdFile_NonFatalWithCommandError(t *testing.T) {
	c, err := Load(writeCmdFixtureCharacter(t, true, false, ""))
	if err != nil {
		t.Fatalf("Load must succeed when .cmd is missing, got: %v", err)
	}
	if !strings.Contains(c.CommandError, "char.cmd") || !strings.Contains(c.CommandError, "opening") {
		t.Errorf("CommandError = %q, want it to name char.cmd and the opening step", c.CommandError)
	}
	if c.Commands.Remap == nil || c.Commands.Commands == nil {
		t.Errorf("Commands must be the normalized empty value, got %+v", c.Commands)
	}
	if len(c.Animations) == 0 {
		t.Error("other files must still be loaded")
	}
}

func TestLoad_MalformedCmdFile_NonFatalWithCommandError(t *testing.T) {
	c, err := Load(writeCmdFixtureCharacter(t, true, true, "[Command\nname = \"a\"\n"))
	if err != nil {
		t.Fatalf("Load must succeed when .cmd is malformed, got: %v", err)
	}
	if !strings.Contains(c.CommandError, "char.cmd") || !strings.Contains(c.CommandError, "parsing") {
		t.Errorf("CommandError = %q, want it to name char.cmd and the parsing step", c.CommandError)
	}
	if len(c.Commands.Commands) != 0 || c.Commands.Remap == nil {
		t.Errorf("Commands must be the normalized empty value, got %+v", c.Commands)
	}
}

func TestLoad_NoCmdDeclared_NoErrorAndEmptyCommands(t *testing.T) {
	c, err := Load(writeCmdFixtureCharacter(t, false, false, ""))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.CommandError != "" {
		t.Errorf("CommandError = %q, want empty when no cmd declared", c.CommandError)
	}
	if c.Commands.Remap == nil || c.Commands.Commands == nil || c.Commands.States == nil {
		t.Errorf("Commands must be the normalized empty value, got %+v", c.Commands)
	}
}

func TestLoad_UnmodifiedCmd_RoundTripsByteForByte(t *testing.T) {
	c, err := Load(writeCmdFixtureCharacter(t, true, true, cmdFixtureContent))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	out, err := SerializeCmd([]byte(cmdFixtureContent), c.Commands)
	if err != nil {
		t.Fatalf("SerializeCmd: %v", err)
	}
	if string(out) != cmdFixtureContent {
		t.Errorf("round trip differs:\n%s", out)
	}
}

func TestLoadBytes_ValidCmdBytes_ExposesCommandsAndRoundTrips(t *testing.T) {
	defBytes, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, []byte(cmdFixtureContent), nil)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if c.CommandError != "" || len(c.Commands.Commands) != 2 {
		t.Fatalf("CommandError=%q commands=%d, want empty and 2", c.CommandError, len(c.Commands.Commands))
	}
	out, err := SerializeCmd([]byte(cmdFixtureContent), c.Commands)
	if err != nil {
		t.Fatalf("SerializeCmd: %v", err)
	}
	if string(out) != cmdFixtureContent {
		t.Errorf("round trip differs:\n%s", out)
	}
}

func TestLoadBytes_NoCmdBytes_NoErrorAndNonNullJSON(t *testing.T) {
	defBytes, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	for _, in := range [][]byte{nil, {}} {
		c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, in, nil)
		if err != nil {
			t.Fatalf("LoadBytes: %v", err)
		}
		if c.CommandError != "" {
			t.Errorf("CommandError = %q, want empty", c.CommandError)
		}
		data, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		cm, ok := m["commands"].(map[string]any)
		if !ok {
			t.Fatalf("commands = %v, want an object", m["commands"])
		}
		for _, k := range []string{"remap", "commands", "states"} {
			if cm[k] == nil {
				t.Errorf("commands.%s is null in JSON", k)
			}
		}
		if _, present := m["commandError"]; present {
			t.Error("commandError must be omitted when empty")
		}
	}
}

func TestLoadBytes_MalformedCmdBytes_NonFatalWithCommandError(t *testing.T) {
	defBytes, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, []byte("[Command\nname = \"a\"\n"), nil)
	if err != nil {
		t.Fatalf("LoadBytes must not fail on malformed cmd: %v", err)
	}
	if !strings.Contains(c.CommandError, "command") {
		t.Errorf("CommandError = %q, want a descriptive message", c.CommandError)
	}
	if c.Commands.Commands == nil || c.Commands.Remap == nil || c.Commands.States == nil {
		t.Errorf("Commands must be normalized empty, got %+v", c.Commands)
	}
}

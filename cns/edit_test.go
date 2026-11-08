package cns

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

const editSrc = "[Data]\nlife = 1000\n\n" +
	"; stand\n[Statedef 0, Standing]\ntype = S\nmovetype = I\nphysics = S\nanim = 0 ; idle anim\nctrl = 1\npoweradd = 0\njuggle = 0\nfacep2 = 0\nhitdefpersist = 0\nmovehitpersist = 0\nhitcountpersist = 0\nsprpriority = 0\n\n" +
	"; go to walk\n[State 0, walk]\ntype = ChangeState\ntrigger1 = Command = \"holdfwd\" ; walk\nvalue = 20 ; target\n\n" +
	"[State 0, x]\ntype = VelSet\nTriggerAll = 1\ntrigger1 = Time = 0\nX = 3\n\n" +
	"[Size]\nxscale = 1\n"

func editSerialize(t *testing.T, src string, mutate func(*Document)) string {
	t.Helper()
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	mutate(doc)
	var buf bytes.Buffer
	if err := doc.SerializeEdited(&buf, doc.StateDefs); err != nil {
		t.Fatalf("SerializeEdited: %v", err)
	}
	return buf.String()
}

func TestSerializeEdited_ChangedParameter_RewritesOnlyThatLineKeepingComment(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers[0].Parameters["value"] = "21"
	})
	want := strings.Replace(editSrc, "value = 20 ; target", "value = 21 ; target", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_ChangedHeaderField_PatchesValueKeepingTrailingComment(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) { d.StateDefs[0].Anim = 5 })
	want := strings.Replace(editSrc, "anim = 0 ; idle anim", "anim = 5 ; idle anim", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_RenumberedStatedef_KeepsLeadingCommentAndHeaderSpelling(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) { d.StateDefs[0].Number = 5 })
	want := strings.Replace(editSrc, "[Statedef 0, Standing]", "[Statedef 5, Standing]", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_AddedController_GoesAfterLastControllerBeforeNextSection(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers = append(d.StateDefs[0].Controllers,
			Controller{Type: "Null", Triggers: []string{"1"}})
	})
	want := strings.Replace(editSrc, "X = 3\n", "X = 3\n\n[State 0]\ntype = Null\ntrigger1 = 1\n", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_RemovedController_DropsItsLeadingCommentAndBlankLine(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers = d.StateDefs[0].Controllers[1:]
	})
	want := strings.Replace(editSrc, "\n; go to walk\n[State 0, walk]\ntype = ChangeState\ntrigger1 = Command = \"holdfwd\" ; walk\nvalue = 20 ; target\n", "", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_ChangedTriggerall_KeepsOriginalKeySpelling(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers[1].Triggers[0] = "Life > 0"
	})
	want := strings.Replace(editSrc, "TriggerAll = 1", "TriggerAll = Life > 0", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_AddedTrigger_ContinuesNumberingAfterLastTriggerLine(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		c := &d.StateDefs[0].Controllers[1]
		c.Triggers = append(c.Triggers, "Vel X > 0")
	})
	want := strings.Replace(editSrc, "trigger1 = Time = 0\n", "trigger1 = Time = 0\ntrigger2 = Vel X > 0\n", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_RemovedTriggerAndParameter_DeleteTheirLines(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		c := &d.StateDefs[0].Controllers[1]
		c.Triggers = c.Triggers[:1]
		delete(c.Parameters, "x")
	})
	want := strings.Replace(editSrc, "trigger1 = Time = 0\nX = 3\n", "", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_NewParameter_GoesAfterLastParameterLine(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers[1].Parameters["y"] = "4"
	})
	want := strings.Replace(editSrc, "X = 3\n", "X = 3\ny = 4\n", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_DuplicateParameterKey_RegeneratesOnlyThatController(t *testing.T) {
	src := "[Statedef 0]\ntype = S\n\n; keep me\n[State 0, a]\ntype = VelSet\ntrigger1 = 1\nx = 1\nx = 2 ; dup\n\n[State 0, b]\ntype = Null\ntrigger1 = 1 ; untouched\n"
	got := editSerialize(t, src, func(d *Document) {
		d.StateDefs[0].Controllers[0].Parameters["x"] = "9"
	})
	want := "[Statedef 0]\ntype = S\n\n; keep me\n[State 0]\ntype = VelSet\ntrigger1 = 1\nx = 9\n\n[State 0, b]\ntype = Null\ntrigger1 = 1 ; untouched\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_AddedStatedef_GoesAfterPredecessorBeforeUnrelatedSection(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs = append(d.StateDefs, StateDef{Number: 10, Type: StateTypeAir})
	})
	add := "\n[Statedef 10]\ntype = A\nmovetype = \nphysics = \nanim = 0\nctrl = 0\npoweradd = 0\njuggle = 0\nfacep2 = 0\nhitdefpersist = 0\nmovehitpersist = 0\nhitcountpersist = 0\nsprpriority = 0\n"
	want := strings.Replace(editSrc, "X = 3\n", "X = 3\n"+add, 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_RemovedStatedef_DropsItWithControllersAndLeavesOtherSections(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) { d.StateDefs = nil })
	want := "[Data]\nlife = 1000\n\n[Size]\nxscale = 1\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_CRLFSource_KeepsCRLFOnPatchedAndGeneratedLines(t *testing.T) {
	src := strings.ReplaceAll(editSrc, "\n", "\r\n")
	got := editSerialize(t, src, func(d *Document) {
		d.StateDefs[0].Controllers[0].Parameters["value"] = "21"
		d.StateDefs[0].Controllers[1].Parameters["y"] = "4"
		d.StateDefs[0].Controllers = append(d.StateDefs[0].Controllers, Controller{Type: "Null"})
	})
	want := strings.ReplaceAll(editSrc, "\n", "\r\n")
	want = strings.Replace(want, "value = 20 ; target", "value = 21 ; target", 1)
	want = strings.Replace(want, "X = 3\r\n", "X = 3\r\ny = 4\r\n\r\n[State 0]\r\ntype = Null\r\n", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_RealisticQuirks_UntouchedLinesStayVerbatim(t *testing.T) {
	src := "[Statedef 200, Stand\r\nType=S ;t\r\nanim = IfElse(life<5,1,2)\r\n\r\njunk line\r\n[State 200, lbl\r\ntype = Null\r\ntrigger1 = 1\r\n"
	got := editSerialize(t, src, func(d *Document) {
		d.StateDefs[0].Controllers[0].Triggers[0] = "0"
	})
	want := strings.Replace(src, "trigger1 = 1", "trigger1 = 0", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_NoChange_IsByteExact(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.cns")
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{string(data), editSrc, "[Statedef 0]\ntype = S"} {
		if got := editSerialize(t, src, func(*Document) {}); got != src {
			t.Errorf("not byte exact:\n%q\nvs\n%q", got, src)
		}
	}
}

func TestSerializeEdited_WriterFails_ReturnsError(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader(editSrc))
	if err != nil {
		t.Fatal(err)
	}
	doc.StateDefs[0].Anim = 3
	err = doc.SerializeEdited(&failingWriter{}, doc.StateDefs)
	var want *failingWriterError
	if !errors.As(err, &want) {
		t.Fatalf("expected wrapped failingWriterError, got %v", err)
	}
	if !strings.HasPrefix(err.Error(), "cns:") {
		t.Errorf("error %q lacks the cns: prefix", err)
	}
}

func TestSerializeEdited_SeveralNewParameters_InsertedSortedAfterLastParameterLine(t *testing.T) {
	src := "[Statedef 0]\ntype = S\n\n[State 0]\ntype = VelSet\ntrigger1 = 1\nx = 1 ; last\n\n[State 1]\ntype = Null\ntrigger1 = 1\n"
	got := editSerialize(t, src, func(d *Document) {
		ps := d.StateDefs[0].Controllers[0].Parameters
		ps["z"], ps["b"], ps["m"] = "3", "1", "2"
	})
	want := "[Statedef 0]\ntype = S\n\n[State 0]\ntype = VelSet\ntrigger1 = 1\nx = 1 ; last\nb = 1\nm = 2\nz = 3\n\n[State 1]\ntype = Null\ntrigger1 = 1\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_ChangedHeaderFieldWithoutSourceLine_AddedAfterLastKeyLine(t *testing.T) {
	src := "[Statedef 0]\ntype = S\nphysics = S\n\n[State 0]\ntype = Null\ntrigger1 = 1\n"
	got := editSerialize(t, src, func(d *Document) { d.StateDefs[0].Anim = 5 })
	want := "[Statedef 0]\ntype = S\nphysics = S\nanim = 5\n\n[State 0]\ntype = Null\ntrigger1 = 1\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_ChangedControllerType_PatchedInPlaceKeepingComment(t *testing.T) {
	got := editSerialize(t, editSrc, func(d *Document) {
		d.StateDefs[0].Controllers[1].Type = "PosSet"
	})
	want := strings.Replace(editSrc, "type = VelSet\n", "type = PosSet\n", 1)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_StatedefWithNoSurvivingPredecessor_PrependedBeforeFirstStatedef(t *testing.T) {
	src := "[Data]\nlife = 1\n\n; first\n[Statedef 5]\ntype = S\n"
	got := editSerialize(t, src, func(d *Document) {
		d.StateDefs = append([]StateDef{{Number: 1, Type: StateTypeStanding}}, d.StateDefs...)
	})
	add := "[Statedef 1]\ntype = S\nmovetype = \nphysics = \nanim = 0\nctrl = 0\npoweradd = 0\njuggle = 0\nfacep2 = 0\nhitdefpersist = 0\nmovehitpersist = 0\nhitcountpersist = 0\nsprpriority = 0\n"
	want := "[Data]\nlife = 1\n" + add + "\n\n; first\n[Statedef 5]\ntype = S\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_StatedefAddedToFileWithoutStatesAndToEmptySource(t *testing.T) {
	add := "[Statedef 3]\ntype = S\nmovetype = \nphysics = \nanim = 0\nctrl = 0\npoweradd = 0\njuggle = 0\nfacep2 = 0\nhitdefpersist = 0\nmovehitpersist = 0\nhitcountpersist = 0\nsprpriority = 0\n"
	mutate := func(d *Document) { d.StateDefs = []StateDef{{Number: 3, Type: StateTypeStanding}} }
	src := "[Data]\r\nlife = 1\r\n"
	got := editSerialize(t, src, mutate)
	want := src + "\r\n" + strings.ReplaceAll(add, "\n", "\r\n")
	if got != want {
		t.Errorf("no states: got:\n%q\nwant:\n%q", got, want)
	}
	if got := editSerialize(t, "", mutate); got != add {
		t.Errorf("empty: got:\n%q\nwant:\n%q", got, add)
	}
}

func TestSerializeEdited_TriggerAddedWhenOnlyTriggerLineIsTriggerall_NumberedTrigger1(t *testing.T) {
	src := "[Statedef 0]\ntype = S\n\n[State 0]\ntype = Null\ntriggerall = 1 ; all\nvalue = 2\n"
	got := editSerialize(t, src, func(d *Document) {
		c := &d.StateDefs[0].Controllers[0]
		c.Triggers = append(c.Triggers, "Time = 0")
	})
	want := "[Statedef 0]\ntype = S\n\n[State 0]\ntype = Null\ntriggerall = 1 ; all\ntrigger1 = Time = 0\nvalue = 2\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

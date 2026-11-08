package def

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// srcLine is one source line split into its text and its own line ending
// ("\r\n", "\n", or "" for an unterminated last line).
type srcLine struct {
	body, eol string
}

// sourceLayout records, for the recognized keys of [Info] and [Files], the
// index of the source line holding each one. Scalar keys map "section.key"
// (lowercase) to every line carrying it, in source order (Parse takes the
// value from the last); stateFiles and palettes list the lines in
// CharacterInfo order, with maxStateNumber and maxPaletteNumber the highest
// key numbers seen ("st" counts as 0; -1 and 0 when there are none).
type sourceLayout struct {
	lines                            []srcLine
	scalars                          map[string][]int
	stateFiles                       []int
	palettes                         []int
	maxStateNumber, maxPaletteNumber int
	sectionEnd                       map[string]int // last key line (or header) of the section's last occurrence
}

func splitLines(source []byte) []srcLine {
	var lines []srcLine
	s := string(source)
	for len(s) > 0 {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			lines = append(lines, srcLine{body: s})
			break
		}
		body, eol := s[:i], "\n"
		if strings.HasSuffix(body, "\r") {
			body, eol = body[:len(body)-1], "\r\n"
		}
		lines = append(lines, srcLine{body, eol})
		s = s[i+1:]
	}
	return lines
}

// indexSource mirrors Parse's line classification, recording line numbers.
func indexSource(source []byte) sourceLayout {
	l := sourceLayout{lines: splitLines(source), scalars: map[string][]int{}, sectionEnd: map[string]int{}, maxStateNumber: -1}
	type entry struct{ number, line int }
	var sts, pals []entry
	section := ""
	for n, ln := range l.lines {
		text := stripComment(ln.body)
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "[") {
			if !strings.HasSuffix(text, "]") {
				continue // Parse rejects this; ParseDocument never gets here.
			}
			section = strings.ToLower(strings.TrimSpace(text[1 : len(text)-1]))
			if section == "info" || section == "files" {
				l.sectionEnd[section] = n
			}
			continue
		}
		if section != "info" && section != "files" {
			continue
		}
		key, _, ok := parseKeyValueLine(text)
		if !ok {
			continue
		}
		lk := strings.ToLower(key)
		switch {
		case section == "info" && (lk == "name" || lk == "author"):
			l.scalars[section+"."+lk] = append(l.scalars[section+"."+lk], n)
		case section == "files" && scalarFileKeys[lk]:
			l.scalars[section+"."+lk] = append(l.scalars[section+"."+lk], n)
		case section == "files" && stateFileKeyPattern.MatchString(key):
			number := -1
			if m := stateFileKeyPattern.FindStringSubmatch(key); m[1] != "" {
				number, _ = strconv.Atoi(m[1])
			}
			sts = append(sts, entry{number, n})
			l.maxStateNumber = max(l.maxStateNumber, number, 0)
		case section == "files" && paletteKeyPattern.MatchString(key):
			number, _ := strconv.Atoi(paletteKeyPattern.FindStringSubmatch(key)[1])
			pals = append(pals, entry{number, n})
			l.maxPaletteNumber = max(l.maxPaletteNumber, number)
		default:
			// unrecognized key
		}
		l.sectionEnd[section] = n
	}
	sort.SliceStable(sts, func(i, j int) bool { return sts[i].number < sts[j].number })
	sort.SliceStable(pals, func(i, j int) bool { return pals[i].number < pals[j].number })
	for _, e := range sts {
		l.stateFiles = append(l.stateFiles, e.line)
	}
	for _, e := range pals {
		l.palettes = append(l.palettes, e.line)
	}
	return l
}

var scalarFileKeys = map[string]bool{"sprite": true, "anim": true, "sound": true, "cmd": true, "cns": true, "stcommon": true}

func cloneInfo(i CharacterInfo) CharacterInfo {
	i.StateFiles = append([]string(nil), i.StateFiles...)
	i.Palettes = append([]string(nil), i.Palettes...)
	return i
}

// scalarField names one single-valued key and how to reach it in a CharacterInfo.
type scalarField struct {
	section, key string
	get          func(*CharacterInfo) *string
}

// scalarFields is in Serialize's emission order, which fixes the order of
// inserted lines.
var scalarFields = []scalarField{
	{"info", "name", func(i *CharacterInfo) *string { return &i.Name }},
	{"info", "author", func(i *CharacterInfo) *string { return &i.Author }},
	{"files", "sprite", func(i *CharacterInfo) *string { return &i.SpriteFile }},
	{"files", "anim", func(i *CharacterInfo) *string { return &i.AnimationFile }},
	{"files", "sound", func(i *CharacterInfo) *string { return &i.SoundFile }},
	{"files", "cmd", func(i *CharacterInfo) *string { return &i.CommandFile }},
	{"files", "cns", func(i *CharacterInfo) *string { return &i.ConstantsFile }},
	{"files", "stcommon", func(i *CharacterInfo) *string { return &i.StateCommonFile }},
}

// patchValue replaces only the value of a "key = value ; comment" line body,
// keeping the key, spacing, quoting style and trailing comment. A quoted
// original stays quoted.
func patchValue(body, value string) string {
	end := len(body)
	comment := ""
	if i := strings.IndexByte(body, ';'); i >= 0 {
		end, comment = i, body[i:]
	}
	eq := strings.IndexByte(body[:end], '=')
	rest := body[eq+1 : end]
	trimmed := strings.TrimSpace(rest)
	lead := rest[:len(rest)-len(strings.TrimLeft(rest, " \t"))]
	trail := rest[len(strings.TrimRight(rest, " \t")):]
	if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		value = `"` + value + `"`
	}
	return body[:eq+1] + lead + value + trail + comment
}

// SerializeEdited writes the Document's source to w with info applied,
// rewriting only what changed relative to the CharacterInfo as parsed
// (edits made to Document.Info itself are not the baseline). Rules:
//
//   - A field equal to its parsed value leaves its line byte-for-byte as is;
//     with no change at all the output equals the source exactly. Comments,
//     blank lines, ordering and unknown sections/keys are never touched.
//   - A changed value replaces only the value of its line: key spelling and
//     casing, spacing around "=", surrounding quotes (kept if the original
//     was quoted), trailing ";" comment and the line's own ending (CRLF/LF)
//     are kept. When a key repeats, the line Parse takes the value from
//     (the last one) is the one edited.
//   - A field cleared to its zero value has its line removed, every line
//     of the key if it repeats (nothing else).
//   - A field newly set gets a new line placed right after the last key line
//     of its section ([Info] or [Files]), in Serialize's field order. Name
//     and Author are quoted as Serialize does; file paths are not. If the
//     section is absent it is appended at the end of the file, preceded by a
//     blank line, with the header text Serialize writes. New lines use the
//     file's dominant line ending (LF on a tie or an empty file).
//   - StateFiles and Palettes are aligned by index with the existing
//     st/stN and palN lines in CharacterInfo order: differing entries are
//     patched, surplus lines removed, extra entries appended after the last
//     key line of [Files], each numbered from the highest existing number
//     plus one: "st" (number 0) then "stN", and "palN" starting at pal1.
//     Contiguous files get the numbers Serialize writes; sparse ones never
//     collide or reorder on reparse.
func (d *Document) SerializeEdited(w io.Writer, info CharacterInfo) error {
	l := d.layout
	eol := dominantEOL(l.lines)
	replace := map[int]string{}
	remove := map[int]bool{}
	adds := map[string][]string{}

	for _, f := range scalarFields {
		oldV, newV := *f.get(&d.baseline), *f.get(&info)
		if oldV == newV {
			continue
		}
		idxs := l.scalars[f.section+"."+f.key]
		present := len(idxs) > 0
		switch {
		case present && newV == "":
			for _, idx := range idxs {
				remove[idx] = true
			}
		case present:
			idx := idxs[len(idxs)-1]
			replace[idx] = patchValue(l.lines[idx].body, newV)
		case newV != "":
			value := newV
			if f.section == "info" {
				value = fmt.Sprintf("%q", newV)
			}
			adds[f.section] = append(adds[f.section], f.key+" = "+value)
		}
	}

	align := func(lines []int, oldV, newV []string, maxNumber int, keyFor func(int) string) {
		if slices.Equal(oldV, newV) {
			return
		}
		for i := range lines {
			switch {
			case i >= len(newV):
				remove[lines[i]] = true
			case i < len(oldV) && oldV[i] != newV[i]:
				replace[lines[i]] = patchValue(l.lines[lines[i]].body, newV[i])
			}
		}
		for i := len(lines); i < len(newV); i++ {
			maxNumber++
			adds["files"] = append(adds["files"], keyFor(maxNumber)+" = "+newV[i])
		}
	}
	align(l.stateFiles, d.baseline.StateFiles, info.StateFiles, l.maxStateNumber, func(i int) string {
		if i == 0 {
			return "st"
		}
		return fmt.Sprintf("st%d", i)
	})
	align(l.palettes, d.baseline.Palettes, info.Palettes, l.maxPaletteNumber, func(n int) string { return fmt.Sprintf("pal%d", n) })

	headers := map[string]string{"info": "[Info]", "files": "[Files]"}
	var out bytes.Buffer
	insertAfter := map[int][]string{}
	var tail []string
	for _, sec := range []string{"info", "files"} {
		if len(adds[sec]) == 0 {
			continue
		}
		if anchor, ok := l.sectionEnd[sec]; ok {
			insertAfter[anchor] = adds[sec]
		} else {
			tail = append(tail, headers[sec])
			tail = append(tail, adds[sec]...)
		}
	}

	for n, ln := range l.lines {
		if !remove[n] {
			body := ln.body
			if r, ok := replace[n]; ok {
				body = r
			}
			out.WriteString(body)
			out.WriteString(ln.eol)
		}
		if ins := insertAfter[n]; len(ins) > 0 {
			if ln.eol == "" && !remove[n] {
				out.WriteString(eol)
			}
			for _, s := range ins {
				out.WriteString(s + eol)
			}
		}
	}
	if len(tail) > 0 {
		if len(l.lines) > 0 {
			if last := l.lines[len(l.lines)-1]; last.eol == "" {
				out.WriteString(eol)
			}
			out.WriteString(eol)
		}
		for _, s := range tail {
			out.WriteString(s + eol)
		}
	}

	if _, err := w.Write(out.Bytes()); err != nil {
		return fmt.Errorf("def: writing edited document: %w", err)
	}
	return nil
}

func dominantEOL(lines []srcLine) string {
	crlf, lf := 0, 0
	for _, ln := range lines {
		switch ln.eol {
		case "\r\n":
			crlf++
		case "\n":
			lf++
		}
	}
	if crlf > lf {
		return "\r\n"
	}
	return "\n"
}

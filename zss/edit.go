package zss

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// blockSpan locates one baseline block in the source, in line indexes
// (lines keep their own terminators). Never stored on Script/Block.
type blockSpan struct {
	triviaStart int // first line of the comments/blank lines directly above the header
	headerStart int // first header line
	headerEnd   int // one past the last header line (multi-line wrapped headers)
	end         int // one past the last body line (next header's headerStart, or EOF)
}

// splitLines cuts src after every "\n", keeping the terminators.
func splitLines(src []byte) []string {
	var lines []string
	for len(src) > 0 {
		i := bytes.IndexByte(src, '\n')
		if i < 0 {
			lines = append(lines, string(src))
			break
		}
		lines = append(lines, string(src[:i+1]))
		src = src[i+1:]
	}
	return lines
}

func stripEOL(s string) string {
	return strings.TrimSuffix(strings.TrimSuffix(s, "\n"), "\r")
}

func isTrivia(line string) bool {
	t := strings.TrimSpace(line)
	return t == "" || strings.HasPrefix(t, "#")
}

// indexBlocks finds every block header the way Parse does (a line starting
// with "[" runs until a line whose trimmed text ends in "]") and records its
// span. Trivia is the run of blank/"#" lines right above a header, bounded
// by the end of the previous header.
func indexBlocks(lines []string) []blockSpan {
	var spans []blockSpan
	prevHeaderEnd := 0
	for i := 0; i < len(lines); {
		if !strings.HasPrefix(lines[i], "[") {
			i++
			continue
		}
		end := i + 1
		for !strings.HasSuffix(strings.TrimSpace(lines[end-1]), "]") && end < len(lines) {
			end++
		}
		ts := i
		for ts > prevHeaderEnd && isTrivia(lines[ts-1]) {
			ts--
		}
		if n := len(spans); n > 0 {
			spans[n-1].end = i
		}
		spans = append(spans, blockSpan{triviaStart: ts, headerStart: i, headerEnd: end})
		prevHeaderEnd = end
		i = end
	}
	if n := len(spans); n > 0 {
		spans[n-1].end = len(lines)
	}
	return spans
}

func dominantEOL(src []byte) string {
	crlf := bytes.Count(src, []byte("\r\n"))
	if crlf > bytes.Count(src, []byte("\n"))-crlf {
		return "\r\n"
	}
	return "\n"
}

func blockKey(b Block) string {
	if b.Kind == BlockKindFunction {
		return string(b.Kind) + "\x00" + b.Name
	}
	return fmt.Sprintf("%s\x00%d", b.Kind, b.Number)
}

func headerEqual(a, b Block) bool {
	return a.Kind == b.Kind && a.Number == b.Number && a.Name == b.Name &&
		maps.Equal(a.HeaderParams, b.HeaderParams) &&
		slices.Equal(a.Params, b.Params) && slices.Equal(a.Ret, b.Ret)
}

// alignBlocks returns, for each baseline block, the index of the edited
// block it maps to (-1 when deleted) and, for each edited block, whether it
// is matched to a baseline one. Anchors come from an LCS on identity keys;
// the unmatched runs between anchors are paired positionally.
func alignBlocks(base, edited []Block) (pair []int, matched []bool) {
	n, m := len(base), len(edited)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if blockKey(base[i]) == blockKey(edited[j]) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	pair = make([]int, n)
	matched = make([]bool, m)
	for i := range pair {
		pair[i] = -1
	}
	pairRun := func(i0, i1, j0, j1 int) {
		for i, j := i0, j0; i < i1 && j < j1; i, j = i+1, j+1 {
			pair[i], matched[j] = j, true
		}
	}
	i, j, ri, rj := 0, 0, 0, 0
	for i < n && j < m {
		switch {
		case blockKey(base[i]) == blockKey(edited[j]):
			pairRun(ri, i, rj, j)
			pair[i], matched[j] = j, true
			i, j = i+1, j+1
			ri, rj = i, j
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	pairRun(ri, n, rj, m)
	return pair, matched
}

// diffLines patches base (raw lines) into want (terminator-less lines):
// lines kept by an LCS are emitted verbatim from base, the rest is deleted
// or generated with eol.
func diffLines(base, want []string, eol string) []string {
	cmp := make([]string, len(base))
	for i, l := range base {
		cmp[i] = stripEOL(l)
	}
	pre := 0
	for pre < len(cmp) && pre < len(want) && cmp[pre] == want[pre] {
		pre++
	}
	suf := 0
	for suf < len(cmp)-pre && suf < len(want)-pre && cmp[len(cmp)-1-suf] == want[len(want)-1-suf] {
		suf++
	}
	a, b := cmp[pre:len(cmp)-suf], want[pre:len(want)-suf]
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	out := append([]string{}, base[:pre]...)
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b) && a[i] == b[j]:
			out = append(out, base[pre+i])
			i, j = i+1, j+1
		case i < len(a) && (j == len(b) || lcs[i+1][j] >= lcs[i][j+1]):
			i++
		default:
			out = append(out, b[j]+eol)
			j++
		}
	}
	return append(out, base[len(base)-suf:]...)
}

// splitText turns Body/Preamble text into terminator-less lines.
func splitText(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

// patchRegion rewrites the baseline lines lines[lo:hi] (a preamble or a
// body) to newText. The last t lines are the trivia of the next block; when
// newText still ends with them they are split off and returned apart so the
// caller emits them with that block (or drops them if it is deleted).
func patchRegion(lines []string, lo, hi, t int, nextDeleted bool, newText, eol string) (out, trivia []string) {
	region := lines[lo:hi]
	t = min(t, len(region))
	trivia = region[len(region)-t:]
	main := region[:len(region)-t]
	want := splitText(newText)
	hasSuffix := t > 0 && len(want) >= t
	for k := 0; hasSuffix && k < t; k++ {
		hasSuffix = want[len(want)-t+k] == stripEOL(trivia[k])
	}
	switch {
	case hasSuffix:
		return diffLines(main, want[:len(want)-t], eol), trivia
	case nextDeleted:
		return diffLines(main, want, eol), trivia
	default:
		return diffLines(region, want, eol), nil
	}
}

func withEOL(s, eol string) string {
	return strings.ReplaceAll(s, "\n", eol)
}

// SerializeEdited writes script to w, rewriting only what differs from the
// source this Document was parsed from. Untouched text (comments, blank
// lines, spacing, wrapped headers, order, line endings) stays byte-exact;
// when script deep-equals the parsed Script the source is written verbatim.
//
// Rules (see .vibe/decisions/033-format-preserving-edits-per-format-alignment.md):
//   - Blocks are aligned by LCS on identity (Statedef: number; Function:
//     name); unmatched runs between two anchors are paired positionally as
//     modifications, so a renumbered Statedef keeps its comments.
//   - A modified block whose header fields are unchanged keeps its header
//     lines verbatim; otherwise the header is regenerated as Serialize
//     writes it. A changed Body is patched line by line (LCS), so a one-line
//     edit rewrites one line.
//   - Comment/blank ("#") lines directly above a header belong to that block
//     and are deleted with it. Since they sit at the end of the previous
//     body in Script, a deleted block's trivia is removed from that body.
//   - A surplus edited block is inserted right after the nearest surviving
//     predecessor's body (before the next block's comments), or after the
//     preamble when it has none, rendered as Serialize would.
//   - A changed Preamble is patched line by line like a body.
//   - Generated lines use the file's dominant line ending (CRLF or LF).
func (d *Document) SerializeEdited(w io.Writer, script Script) error {
	if reflect.DeepEqual(script, d.baseline) {
		return d.Serialize(w)
	}
	lines := splitLines(d.source)
	eol := dominantEOL(d.source)
	spans := d.spans
	pair, matched := alignBlocks(d.baseline.Blocks, script.Blocks)

	// inserts[k+1] holds blocks to add after baseline block k (k = -1: after the preamble).
	inserts := make([][]string, len(spans)+1)
	last := -1
	paired := make(map[int]int)
	for k, j := range pair {
		if j >= 0 {
			paired[j] = k
		}
	}
	for j, b := range script.Blocks {
		if matched[j] {
			last = paired[j]
			continue
		}
		var buf bytes.Buffer
		if err := Serialize(&buf, Script{Blocks: []Block{b}}); err != nil {
			return err
		}
		inserts[last+1] = append(inserts[last+1], withEOL(buf.String(), eol))
	}

	var out []string
	emitInserts := func(k int) {
		out = append(out, inserts[k+1]...)
	}
	deleted := func(k int) bool { return k < len(spans) && pair[k] < 0 }

	preEnd, nextTrivia := len(lines), 0
	if len(spans) > 0 {
		preEnd, nextTrivia = spans[0].headerStart, spans[0].headerStart-spans[0].triviaStart
	}
	body, pending := patchRegion(lines, 0, preEnd, nextTrivia, deleted(0), script.Preamble, eol)
	out = append(out, body...)
	emitInserts(-1)

	for k, sp := range spans {
		nextTrivia = 0
		if k+1 < len(spans) {
			nextTrivia = spans[k+1].headerStart - spans[k+1].triviaStart
		}
		if pair[k] < 0 {
			region := lines[sp.headerEnd:sp.end]
			nextTrivia = min(nextTrivia, len(region))
			pending = region[len(region)-nextTrivia:]
			continue
		}
		nb, ob := script.Blocks[pair[k]], d.baseline.Blocks[k]
		out = append(out, pending...)
		if headerEqual(nb, ob) {
			out = append(out, lines[sp.headerStart:sp.headerEnd]...)
		} else {
			var buf bytes.Buffer
			if err := writeBlockHeader(&buf, nb); err != nil {
				return err
			}
			out = append(out, withEOL(buf.String(), eol))
		}
		body, pending = patchRegion(lines, sp.headerEnd, sp.end, nextTrivia, deleted(k+1), nb.Body, eol)
		out = append(out, body...)
		emitInserts(k)
	}

	for i, l := range out {
		if i < len(out)-1 && !strings.HasSuffix(l, "\n") {
			out[i] = l + eol
		}
	}
	if _, err := io.WriteString(w, strings.Join(out, "")); err != nil {
		return fmt.Errorf("zss: writing edited document: %w", err)
	}
	return nil
}

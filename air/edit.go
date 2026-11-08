package air

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// blockSpan locates an explicit per-frame "ClsnN: count" declaration; its
// count box lines follow on the next lines.
type blockSpan struct{ decl, count int }

// frameSpan locates one frame in the source.
type frameSpan struct {
	line   int           // the frame line
	start  int           // first line of the frame: its explicit Clsn blocks, else the frame line
	blocks [2]*blockSpan // explicit Clsn1/Clsn2 blocks right above the frame, if any
	def    [2][]ClsnBox  // Clsn1Default/Clsn2Default boxes in effect at the frame
}

// actionSpan records where one recognized "[Begin Action N]" block sits in
// the source of a Document. Line numbers are 0-based indexes into the
// source's lines. It lives on Document, never on Animation/Frame (decision
// 033).
type actionSpan struct {
	header int // line of the "[Begin Action N]" header
	start  int // first line owned: the leading comment/blank lines above the header
	end    int // exclusive: the line after the last non-trivia line of the block

	frames    []frameSpan
	loopLine  int  // line of the Loopstart marker, -1 when none
	loopPos   int  // number of frames before the marker
	ambiguous bool // several markers or stacked/dangling explicit blocks: only regeneration is safe
}

// splitLines splits source into lines that keep their line ending.
func splitLines(source []byte) []string {
	var lines []string
	for len(source) > 0 {
		i := bytes.IndexByte(source, '\n')
		if i < 0 {
			lines = append(lines, string(source))
			break
		}
		lines = append(lines, string(source[:i+1]))
		source = source[i+1:]
	}
	return lines
}

// lineContent returns line without its line ending.
func lineContent(line string) string {
	return strings.TrimRight(line, "\r\n")
}

// isTrivia reports whether line is blank or comment-only (';' or ':').
func isTrivia(line string) bool {
	s := stripComment(lineContent(line))
	return s == "" || strings.HasPrefix(s, ":")
}

// indexSource scans source with the same line classification Parse uses and
// returns the span of each recognized action, in file order.
func indexSource(source []byte) []actionSpan {
	lines := splitLines(source)

	// headers lists every bracket header line, recognized or not.
	type header struct {
		line   int
		action bool
	}
	var headers []header
	for i := 0; i < len(lines); i++ {
		s := stripComment(lineContent(lines[i]))
		if strings.HasPrefix(s, "[") {
			headers = append(headers, header{i, actionHeaderPattern.MatchString(s)})
			continue
		}
		if _, count, ok := parseClsnDeclarationHeader(s); ok {
			i += boundedSkip(count, i, len(lines)) // the box lines are consumed with their declaration
		}
	}

	// trivStart walks back from line over the comment/blank lines directly above it.
	trivStart := func(line, floor int) int {
		for line > floor && isTrivia(lines[line-1]) {
			line--
		}
		return line
	}

	var spans []actionSpan
	for h, hd := range headers {
		if !hd.action {
			continue
		}
		floor := 0
		if h > 0 {
			floor = headers[h-1].line + 1
		}
		limit := len(lines)
		if h+1 < len(headers) {
			limit = headers[h+1].line
		}
		sp := actionSpan{header: hd.line, start: trivStart(hd.line, floor), loopLine: -1}
		sp.end = trivStart(limit, hd.line+1)
		var pending [2]*blockSpan
		var def [2][]ClsnBox
		markers := 0
		for i := hd.line + 1; i < sp.end; i++ {
			s := stripComment(lineContent(lines[i]))
			if s == "" || strings.HasPrefix(s, ":") || isInterpolateDirective(s) {
				continue
			}
			if strings.EqualFold(s, "loopstart") {
				markers++
				sp.loopLine, sp.loopPos = i, len(sp.frames)
				continue
			}
			if declType, count, ok := parseClsnDeclarationHeader(s); ok {
				idx := int(declType[4] - '1')
				count = boundedSkip(count, i, sp.end)
				if strings.HasSuffix(declType, "Default") {
					def[idx] = scanBoxes(lines, i+1, count)
				} else {
					if pending[idx] != nil {
						sp.ambiguous = true
					}
					pending[idx] = &blockSpan{i, count}
				}
				i += count
				continue
			}
			fs := frameSpan{line: i, start: i, blocks: pending, def: def}
			for _, b := range pending {
				if b != nil && b.decl < fs.start {
					fs.start = b.decl
				}
			}
			pending = [2]*blockSpan{}
			sp.frames = append(sp.frames, fs)
		}
		if pending[0] != nil || pending[1] != nil || markers > 1 {
			sp.ambiguous = true
		}
		spans = append(spans, sp)
	}
	return spans
}

// boundedSkip clamps the box-line count of a ClsnN declaration at line i to
// what the source can hold below line limit: the count is untrusted input, a
// negative one counts as zero and a huge one must not run past the source.
func boundedSkip(count, i, limit int) int {
	if count < 0 {
		return 0
	}
	if remaining := limit - i - 1; count > remaining {
		if remaining < 0 {
			return 0
		}
		return remaining
	}
	return count
}

// scanBoxes reads count box lines starting at line from; unparsable lines are skipped.
func scanBoxes(lines []string, from, count int) []ClsnBox {
	boxes := []ClsnBox{}
	for i := from; i < from+count && i < len(lines); i++ {
		if m := clsnBoxLinePattern.FindStringSubmatch(stripComment(lineContent(lines[i]))); m != nil {
			var b ClsnBox
			fmt.Sscan(m[1], &b.Left)
			fmt.Sscan(m[2], &b.Top)
			fmt.Sscan(m[3], &b.Right)
			fmt.Sscan(m[4], &b.Bottom)
			boxes = append(boxes, b)
		}
	}
	return boxes
}

func frameEqual(a, b Frame) bool {
	return a.Group == b.Group && a.Image == b.Image && a.X == b.X && a.Y == b.Y &&
		a.Time == b.Time && a.Flip == b.Flip && a.Blend == b.Blend &&
		slices.Equal(a.Clsn1, b.Clsn1) && slices.Equal(a.Clsn2, b.Clsn2)
}

func animationEqual(a, b Animation) bool {
	if a.Number != b.Number || a.LoopStart != b.LoopStart || len(a.Frames) != len(b.Frames) {
		return false
	}
	for i := range a.Frames {
		if !frameEqual(a.Frames[i], b.Frames[i]) {
			return false
		}
	}
	return true
}

// textEdit replaces source[start:end] with text.
type textEdit struct {
	start, end int
	text       string
}

// headerNumberPattern captures the action number of a header line.
var headerNumberPattern = regexp.MustCompile(`(?i)^\[\s*begin\s+action\s+(-?\d+)`)

// SerializeEdited writes the source the Document was parsed from to w with
// animations applied as an edit, rewriting only what changed so untouched
// comments, blank lines, ordering and formatting stay byte-exact
// (.vibe/decisions/033-format-preserving-edits-per-format-alignment.md).
// The baseline is the Animations parsed from the retained source, not the
// current value of d.Animations, so mutating d.Animations in place is safe.
//
// Rules:
//   - If animations equals the baseline, the source is written verbatim.
//   - Baseline and edited actions are aligned by longest common subsequence
//     on Number. Unmatched runs between two anchors are paired positionally
//     as modifications: a renumbered action keeps its comments, only the
//     number in its header changes (header spelling and casing are kept).
//   - Surplus edited actions are inserted right after the nearest surviving
//     predecessor's block (a blank line, then the action as Serialize writes
//     it); with no predecessor they go before the first baseline action's
//     block (or at the start of the file when it has none), followed by a
//     blank line. Surplus baseline actions are deleted with the comment and
//     blank lines directly above their header, which belong to them.
//   - Inside a modified action, frames are aligned by LCS on full Frame
//     equality (unmatched runs are paired positionally; equal frame counts
//     pair index-wise). Equal frames keep their lines. A changed frame gets
//     only its changed comma-separated fields rewritten in its frame line:
//     unchanged fields, fields past Blend (Ikemen GO xscale, yscale, angle),
//     the trailing ";comment" and the line ending are kept. Flip/Blend
//     missing from the line are appended after its last field. A box value change in an explicit per-frame
//     Clsn1:/Clsn2: block rewrites only that "ClsnN[i] = ..." line in place;
//     a box count change rewrites only that block (the declaration keeps its
//     trailing comment). A deleted frame loses its frame line and its own
//     explicit blocks, never shared Default lines. An inserted frame goes
//     right after its predecessor frame's line (or before the first frame),
//     with explicit blocks only where its boxes differ from the Default in
//     effect. A LoopStart change moves only the Loopstart marker line.
//     Interpolate lines and comments stay where they are.
//   - The whole action is regenerated with Serialize's text (keeping its
//     leading comments) when the patch is ambiguous: boxes inherited from a
//     Clsn1Default/Clsn2Default change, the Loopstart marker must move while
//     the frame count also changes, or the action has several markers or
//     stacked/dangling explicit blocks. Interpolate lines and Clsn defaults
//     inside it are then not carried over.
//   - Preamble and embedded non-action bracket sections stay verbatim.
//
// Generated lines use the file's dominant line ending (CRLF or LF).
func (d *Document) SerializeEdited(w io.Writer, animations []Animation) error {
	base, err := Parse(bytes.NewReader(d.source))
	if err != nil {
		return err
	}
	if len(base) != len(d.spans) {
		return fmt.Errorf("air: document index out of sync: %d actions, %d spans", len(base), len(d.spans))
	}

	var out []byte
	if slices.EqualFunc(base, animations, animationEqual) {
		out = d.source
	} else {
		out = d.applyEdits(base, animations)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("air: writing edited document: %w", err)
	}
	return nil
}

// lcsPairs returns the (a, b) index pairs of a longest common subsequence
// of two lists of lengths n and m, where same(i, j) tells whether a[i] and
// b[j] are the same element.
func lcsPairs(n, m int, same func(i, j int) bool) [][2]int {
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if same(i, j) {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	var pairs [][2]int
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case same(i, j):
			pairs = append(pairs, [2]int{i, j})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			i++
		default:
			j++
		}
	}
	return pairs
}

// alignment is the result of aligning a baseline list with an edited one.
type alignment struct {
	pairs    [][2]int      // (baseline, edited) kept elements: anchors and positional modifications
	deleted  []int         // surplus baseline indexes
	inserted map[int][]int // nearest surviving predecessor baseline index (-1: none) -> surplus edited indexes
}

// align aligns two lists by LCS; unmatched runs between two anchors are
// paired positionally, the surplus is inserted or deleted.
func align(n, m int, same func(i, j int) bool) alignment {
	al := alignment{inserted: map[int][]int{}}
	anchors := append(lcsPairs(n, m, same), [2]int{n, m})
	prevB, prevE, last := 0, 0, -1
	for _, p := range anchors {
		gapB, gapE := p[0]-prevB, p[1]-prevE
		common := min(gapB, gapE)
		for k := 0; k < common; k++ {
			al.pairs = append(al.pairs, [2]int{prevB + k, prevE + k})
			last = prevB + k
		}
		for e := prevE + common; e < p[1]; e++ {
			al.inserted[last] = append(al.inserted[last], e)
		}
		for b := prevB + common; b < p[0]; b++ {
			al.deleted = append(al.deleted, b)
		}
		if p[0] < n {
			al.pairs = append(al.pairs, p)
			last = p[0]
		}
		prevB, prevE = p[0]+1, p[1]+1
	}
	return al
}

// dominantEOL returns "\r\n" when CRLF endings outnumber bare LF, else "\n".
func dominantEOL(source []byte) string {
	crlf := bytes.Count(source, []byte("\r\n"))
	if crlf > bytes.Count(source, []byte("\n"))-crlf {
		return "\r\n"
	}
	return "\n"
}

// editEnv carries what the action patching needs to compute byte edits.
type editEnv struct {
	source  []byte
	lines   []string
	offsets []int // byte offset of each line start, plus the source length
	eol     string
}

// off returns the byte offset of the start of line i.
func (v *editEnv) off(i int) int { return v.offsets[i] }

// lead returns eol when the byte before offset is not a line end, so text
// inserted there starts on its own line.
func (v *editEnv) lead(offset int) string {
	if offset > 0 && v.source[offset-1] != '\n' {
		return v.eol
	}
	return ""
}

// splitTail splits a line into its value part and the trailing
// whitespace + ";comment" + line ending, which an in-place rewrite keeps.
func splitTail(line string) (head, tail string) {
	content := lineContent(line)
	end := len(content)
	if idx := strings.IndexByte(content, ';'); idx >= 0 {
		end = len(strings.TrimRight(content[:idx], " \t"))
	} else {
		end = len(strings.TrimRight(content, " \t"))
	}
	return content[:end], content[end:] + line[len(content):]
}

// patchFrameFields rewrites only the comma-separated fields of a frame line
// head (no comment, no line ending) that differ between o and n, keeping the
// surrounding whitespace of each patched field, every unchanged field's text
// and every field after the last modeled one (Ikemen GO xscale, yscale,
// angle...). Flip and Blend missing from the line are appended after the
// last field, so nothing following them exists to be displaced.
func patchFrameFields(head string, o, n Frame) string {
	fields := strings.Split(head, ",")
	vals := []string{strconv.Itoa(n.Group), strconv.Itoa(n.Image), strconv.Itoa(n.X),
		strconv.Itoa(n.Y), strconv.Itoa(n.Time), string(n.Flip), string(n.Blend)}
	changed := []bool{o.Group != n.Group, o.Image != n.Image, o.X != n.X, o.Y != n.Y,
		o.Time != n.Time, o.Flip != n.Flip, o.Blend != n.Blend}
	for i, val := range vals {
		if !changed[i] {
			continue
		}
		for len(fields) <= i {
			fields = append(fields, " ")
		}
		f := fields[i]
		lead := f[:len(f)-len(strings.TrimLeft(f, " \t"))]
		trail := f[len(strings.TrimRight(f, " \t")):]
		if strings.TrimSpace(f) == "" {
			lead, trail = " ", ""
		}
		fields[i] = lead + val + trail
	}
	return strings.Join(fields, ",")
}

// blockText is the "ClsnN: count" declaration and box lines for boxes, or
// nothing when boxes is empty and no default needs overriding.
func (v *editEnv) blockText(name string, boxes, def []ClsnBox) string {
	if len(boxes) == 0 {
		if len(def) == 0 {
			return ""
		}
		return name + ": 0" + v.eol
	}
	var buf bytes.Buffer
	_ = writeClsnBoxes(&buf, name, boxes) // writing to a bytes.Buffer cannot fail
	return strings.ReplaceAll(buf.String(), "\n", v.eol)
}

func frameBoxes(f Frame, t int) []ClsnBox {
	if t == 0 {
		return f.Clsn1
	}
	return f.Clsn2
}

func frameLineEqual(a, b Frame) bool {
	a.Clsn1, a.Clsn2, b.Clsn1, b.Clsn2 = nil, nil, nil, nil
	return frameEqual(a, b)
}

// defaultBetween reports whether a Clsn1Default/Clsn2Default declaration sits
// on a line strictly between lines from and to.
func (v *editEnv) defaultBetween(from, to int) bool {
	for i := from + 1; i < to; i++ {
		if d, _, ok := parseClsnDeclarationHeader(stripComment(lineContent(v.lines[i]))); ok && strings.HasSuffix(d, "Default") {
			return true
		}
	}
	return false
}

// patchAction computes in-place edits turning action old into nu. It
// reports false when the change is ambiguous and the action must be
// regenerated instead.
func (v *editEnv) patchAction(sp actionSpan, old, nu Animation) ([]textEdit, bool) {
	if sp.ambiguous || len(sp.frames) != len(old.Frames) {
		return nil, false
	}
	var edits []textEdit

	if old.Number != nu.Number {
		hl := v.lines[sp.header]
		trimmed := strings.TrimLeft(hl, " \t")
		if loc := headerNumberPattern.FindStringSubmatchIndex(trimmed); loc != nil {
			shift := v.off(sp.header) + len(hl) - len(trimmed)
			edits = append(edits, textEdit{shift + loc[2], shift + loc[3], fmt.Sprint(nu.Number)})
		}
	}

	var al alignment
	if len(old.Frames) == len(nu.Frames) {
		al = alignment{inserted: map[int][]int{}}
		for i := range old.Frames {
			al.pairs = append(al.pairs, [2]int{i, i})
		}
	} else {
		al = align(len(old.Frames), len(nu.Frames), func(i, j int) bool { return frameEqual(old.Frames[i], nu.Frames[j]) })
	}

	// The Loopstart marker stays put only if the frame it precedes keeps
	// that position; otherwise it moves, which is only well defined when
	// the frame count is unchanged.
	newIdx := map[int]int{}
	for _, p := range al.pairs {
		newIdx[p[0]] = p[1]
	}
	markerStays := nu.LoopStart == 0
	if sp.loopLine >= 0 {
		if sp.loopPos == len(old.Frames) {
			markerStays = nu.LoopStart == len(nu.Frames)
		} else {
			j, ok := newIdx[sp.loopPos]
			markerStays = ok && nu.LoopStart == j
		}
	}
	if !markerStays {
		if len(old.Frames) != len(nu.Frames) || nu.LoopStart > len(nu.Frames) {
			return nil, false
		}
		if sp.loopLine >= 0 {
			edits = append(edits, textEdit{v.off(sp.loopLine), v.off(sp.loopLine + 1), ""})
		}
		if nu.LoopStart > 0 {
			var at int
			if nu.LoopStart < len(sp.frames) {
				at = v.off(sp.frames[nu.LoopStart].start)
			} else {
				at = v.off(sp.frames[len(sp.frames)-1].line + 1)
			}
			edits = append(edits, textEdit{at, at, v.lead(at) + "Loopstart" + v.eol})
		}
	}

	// Inserted frames first, so they precede edits at the same offset.
	for pred, idxs := range al.inserted {
		var at int
		var def [2][]ClsnBox
		switch {
		case pred >= 0:
			at, def = v.off(sp.frames[pred].line+1), sp.frames[pred].def
		case len(sp.frames) > 0:
			first := sp.frames[0].start
			if sp.loopLine >= 0 && sp.loopPos == 0 && sp.loopLine < first {
				// shortcut: a Clsn*Default between the marker and frame 0 would
				// not apply to a frame placed above it, so regenerate; upgrade by
				// computing the defaults in effect at the insertion line.
				if v.defaultBetween(sp.loopLine, sp.frames[0].line) {
					return nil, false
				}
				first = sp.loopLine // stay above the marker that precedes the old first frame
			}
			at, def = v.off(first), sp.frames[0].def
		default:
			at = v.off(sp.header + 1)
		}
		text := v.lead(at)
		for _, e := range idxs {
			f := nu.Frames[e]
			for t, name := range [2]string{"Clsn1", "Clsn2"} {
				if !slices.Equal(frameBoxes(f, t), def[t]) {
					text += v.blockText(name, frameBoxes(f, t), def[t])
				}
			}
			var fb bytes.Buffer
			_ = writeFrameLine(&fb, f)
			text += strings.TrimSuffix(fb.String(), "\n") + v.eol
		}
		edits = append(edits, textEdit{at, at, text})
	}

	for _, b := range al.deleted {
		fs := sp.frames[b]
		edits = append(edits, textEdit{v.off(fs.line), v.off(fs.line + 1), ""})
		for _, blk := range fs.blocks {
			if blk != nil {
				edits = append(edits, textEdit{v.off(blk.decl), v.off(blk.decl + blk.count + 1), ""})
			}
		}
	}

	for _, p := range al.pairs {
		o, n := old.Frames[p[0]], nu.Frames[p[1]]
		if frameEqual(o, n) {
			continue
		}
		fs := sp.frames[p[0]]
		if !frameLineEqual(o, n) {
			head, tail := splitTail(v.lines[fs.line])
			edits = append(edits, textEdit{v.off(fs.line), v.off(fs.line + 1), patchFrameFields(head, o, n) + tail})
		}
		for t, name := range [2]string{"Clsn1", "Clsn2"} {
			oldB, newB := frameBoxes(o, t), frameBoxes(n, t)
			if slices.Equal(oldB, newB) {
				continue
			}
			blk := fs.blocks[t]
			switch {
			case blk == nil && len(fs.def[t]) > 0:
				// shortcut: boxes inherited from a Clsn*Default change as a
				// regeneration of the action, upgrade by splitting the default.
				return nil, false
			case blk == nil:
				at := v.off(fs.line)
				edits = append(edits, textEdit{at, at, v.blockText(name, newB, fs.def[t])})
			case len(oldB) == len(newB):
				for i := range newB {
					if oldB[i] == newB[i] {
						continue
					}
					ln := blk.decl + 1 + i
					head, tail := splitTail(v.lines[ln])
					eq := strings.IndexByte(head, '=')
					if eq < 0 {
						return nil, false
					}
					after := head[eq+1:]
					space := after[:len(after)-len(strings.TrimLeft(after, " \t"))]
					b := newB[i]
					text := fmt.Sprintf("%s%s%d,%d,%d,%d%s", head[:eq+1], space, b.Left, b.Top, b.Right, b.Bottom, tail)
					edits = append(edits, textEdit{v.off(ln), v.off(ln + 1), text})
				}
			default:
				text := v.blockText(name, newB, fs.def[t])
				if _, tail := splitTail(v.lines[blk.decl]); text != "" {
					// keep the declaration line's trailing ";comment"
					if comment := strings.TrimRight(tail, "\r\n"); comment != "" {
						first := strings.Index(text, v.eol)
						text = text[:first] + comment + text[first:]
					}
				}
				edits = append(edits, textEdit{v.off(blk.decl), v.off(blk.decl + blk.count + 1), text})
			}
		}
	}
	return edits, true
}

func (d *Document) applyEdits(base, edited []Animation) []byte {
	lines := splitLines(d.source)
	offsets := make([]int, len(lines)+1)
	for i, l := range lines {
		offsets[i+1] = offsets[i] + len(l)
	}
	env := &editEnv{source: d.source, lines: lines, offsets: offsets, eol: dominantEOL(d.source)}
	eol := env.eol

	// actionText is Serialize's text for one action, without a trailing
	// line ending, using the file's line ending.
	actionText := func(a Animation) string {
		var buf bytes.Buffer
		_ = Serialize(&buf, []Animation{a}) // writing to a bytes.Buffer cannot fail
		text := strings.TrimRight(buf.String(), "\n")
		return strings.ReplaceAll(text, "\n", eol)
	}

	var edits []textEdit
	modify := func(b, e int) {
		sp, old, nu := d.spans[b], base[b], edited[e]
		if animationEqual(old, nu) {
			return
		}
		if patch, ok := env.patchAction(sp, old, nu); ok {
			edits = append(edits, patch...)
			return
		}
		edits = append(edits, textEdit{offsets[sp.header], offsets[sp.end], actionText(nu) + eol})
	}

	al := align(len(base), len(edited), func(i, j int) bool { return base[i].Number == edited[j].Number })
	for _, p := range al.pairs {
		modify(p[0], p[1])
	}
	for _, b := range al.deleted {
		sp := d.spans[b]
		edits = append(edits, textEdit{offsets[sp.start], offsets[sp.end], ""})
	}
	for after, idxs := range al.inserted {
		var sb strings.Builder
		var offset int
		if after < 0 {
			if len(d.spans) > 0 {
				offset = offsets[d.spans[0].start]
			}
			for _, e := range idxs {
				sb.WriteString(actionText(edited[e]) + eol + eol)
			}
		} else {
			offset = offsets[d.spans[after].end]
			sb.WriteString(env.lead(offset))
			for _, e := range idxs {
				sb.WriteString(eol + actionText(edited[e]) + eol)
			}
		}
		edits = append(edits, textEdit{offset, offset, sb.String()})
	}

	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start < edits[j].start
		}
		return edits[i].end-edits[i].start < edits[j].end-edits[j].start
	})
	var out bytes.Buffer
	pos := 0
	for _, e := range edits {
		out.Write(d.source[pos:e.start])
		out.WriteString(e.text)
		pos = e.end
	}
	out.Write(d.source[pos:])
	return out.Bytes()
}

package pine

import (
	"fmt"
	"strconv"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

type Render struct {
	termH, termW       int
	viewMaxH, viewMaxW int
	buf                *Buffer
	viewCursor         *Pos
	viewAnchor         *Pos
	mouseCursor        *Pos
	sett               *Setting
}

type RenderContent struct {
	mod      tm.Modifier
	key      tm.Key
	ch       rune
	bufIdx   int
	bufDirty bool
	msg      string
}

func (r *Render) Init() {
	r.viewCursor = &Pos{0, 0}
	r.viewAnchor = &Pos{0, 0}
	r.mouseCursor = &Pos{0, 0}
}

func (r *Render) Clear() {
	if err := tm.Clear(tm.ColorDefault, tm.ColorDefault); err != nil {
		log.Warn("failed to clear screen")
	}
}

func (r *Render) DrawScreen(mode Mode, content RenderContent) {
	r.Clear()
	defer tm.Flush()

	r.termW, r.termH = tm.Size()
	r.viewMaxH = r.termH - 1
	r.viewMaxW = r.termW

	r.drawStatusline(mode, getFilename(r.buf.filePath), content)

	if r.buf == nil {
		log.Fatalln("buffer is null while in edit/file mode")
	}

	r.drawBuffer()
	r.drawCursor()

	if len(content.msg) > 0 {
		r.drawMessage(content.msg)
	}
}

func (r *Render) drawMessage(msg string) {
	tbprint(r.termH-1, 0, tm.ColorRed, tm.ColorDefault, msg)
}

func (r *Render) drawStatusline(
	mode Mode,
	filePath string,
	content RenderContent,
) {
	for i := 0; i < r.termW-1; i++ {
		tm.SetCell(i, 0, rune(' '), tm.ColorBlack, tm.ColorWhite)
	}
	if mode == FileOpenMode {
		tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, " INPUT A FILE NAME | Press Enter to open")
		return
	}
	if mode == FileSaveMode {
		tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, " INPUT A FILE NAME | Press Enter to save")
		return
	}
	tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("Pine Editor v%s", VERSION))
	bufId := strconv.Itoa(content.bufIdx)
	bufDirtyMark := " "
	if content.bufDirty {
		bufDirtyMark = "*"
	}
	tbprint(0, r.termW-len(filePath)-len(bufId)-3, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("%s%s: %s", bufDirtyMark, bufId, filePath))

	for i := 0; i < r.termW-1; i++ {
		tm.SetCell(i, r.termH-2, rune(' '), tm.ColorCyan, tm.ColorCyan)
	}
	tbprint(r.termH-2, 0, tm.ColorBlack, tm.ColorCyan, fmt.Sprintf("%06d,%06d %4d%%  %x-%s:%x", r.buf.cursor.x, r.buf.cursor.y, getLinePer(r.buf), int(content.mod), string(content.ch), int(content.key)))
	statusTailMsg := "^/ Help    ^X Exit"
	tbprint(r.termH-2, r.termW-len(statusTailMsg), tm.ColorBlack, tm.ColorCyan, statusTailMsg)

	/*
		if r.sett.IsDebug {
			tbprint(r.termH-1, 0, tm.ColorDefault, tm.ColorDefault, fmt.Sprintf("%d:%d | vc %d:%d | va %d:%d | mc %d:%d | op:%s | tr: %d; crc: %d;", r.cursor.x, r.cursor.y, r.viewCursor.x, r.viewCursor.y, r.viewAnchor.x, r.viewAnchor.y, r.mouseCursor.x, r.mouseCursor.y, string(r.buf.lastModifiedCh), len(r.buf.lines), currLineLen))
		   }
	*/
}

func (r *Render) drawBuffer() {
	viewIndex := 0
	for i := 0; i < len(r.buf.lines); i++ {
		if i < r.viewAnchor.x {
			continue
		}
		if viewIndex == r.termH-BUFFER_CONTENT_END_OFFSET-1 {
			break
		}
		viewIndex++
		r.drawBufferLine(i, viewIndex)
	}
}

func (r *Render) drawBufferLine(i int, viewIndex int) {
	renderedData := ""
	y := 0
	for _, ch := range r.buf.lines[i].txt {
		if ch == rune('\t') {
			renderedData += r.drawTab(runeRenderedWidth(y, ch))
		} else {
			renderedData += string(ch)
		}
		y += runeRenderedWidth(y, ch)
	}
	if y < r.viewAnchor.y {
		return
	}
	// TODO: it can print out of the screen. but termbox-go handles
	// this misbehavior. need to clean up this mess.
	tbprint(i-r.viewAnchor.x+BUFFER_CONTENT_START_OFFSET, 0, tm.ColorDefault, tm.ColorDefault, renderedData[r.viewAnchor.y:])
	if len(renderedData[r.viewAnchor.y:]) >= r.termW {
		tbprint(i-r.viewAnchor.x+BUFFER_CONTENT_START_OFFSET, r.termW-1, tm.ColorCyan, tm.ColorDefault, ">")
	}
}

func (r *Render) drawTab(tablength int) string {
	tab := ""
	for i := 0; i < tablength; i++ {
		tab += " "
	}
	return tab
}

func (r *Render) drawCursor() {
	tm.SetCursor(r.viewCursor.y-r.viewAnchor.y, r.viewCursor.x-r.viewAnchor.x+1)
}

/*
 * MoveCursor
 *
 * Update cursor position and view based on cursor movement
 * Cursor x,y are determined by view x,y
 */
func (r *Render) moveCursorUp() {
	if r.buf.isEmpty() {
		return
	}
	if r.buf.cursor.x == 0 {
		return
	}
	r.buf.cursor.x--
	if r.buf.cursor.y > len(r.buf.lines[r.buf.cursor.x].txt) {
		r.buf.cursor.y = len(r.buf.lines[r.buf.cursor.x].txt)
	}
}

func (r *Render) moveCursorDown() {
	if r.buf.isEmpty() {
		return
	}
	if r.buf.cursor.x == len(r.buf.lines)-1 {
		return
	}
	r.buf.cursor.x++
	if r.buf.cursor.y > len(r.buf.lines[r.buf.cursor.x].txt) {
		r.buf.cursor.y = len(r.buf.lines[r.buf.cursor.x].txt)
	}
}

func (r *Render) moveCursorLeft() {
	if r.buf.isEmpty() {
		return
	}
	if r.buf.cursor.y == 0 && r.viewAnchor.y > 0 {
		log.Fatalf("cursor %d and viewpoint %d out of sync", r.buf.cursor.y, r.viewAnchor.y)
	}
	if r.buf.cursor.y == 0 {
		return
	}
	r.buf.cursor.y--
}

func (r *Render) moveCursorRight() {
	if r.buf.isEmpty() {
		return
	}
	if r.buf.cursor.y == len(r.buf.lines[r.buf.cursor.x].txt) {
		return
	}
	r.buf.cursor.y++
}

/*
 * Mouse position is limited by terminal height and weight
 * It should be one to one mapped to view cursor position
 * But there are multi-length rune that we cannot position cursor
 * in between. So we convert back to cursor first.
 *
 */
func (r *Render) MoveCursorByMouse(p Pos) {
	r.syncViewCursorToCursor(p)
}

func (r *Render) syncViewCursorToCursor(p Pos) {
	if p.x < 0 || ((p.x+r.viewAnchor.x) >= len(r.buf.lines) || p.x >= r.termH-1) {
		return
	}
	currLine := r.buf.lines[p.x+r.viewAnchor.x]
	lineIndex := 0
	viewLineIndex := 0
	for viewLineIndex < p.y+r.viewAnchor.y && lineIndex < len(currLine.txt) {
		viewLineIndex += runeRenderedWidth(viewLineIndex, currLine.txt[lineIndex])
		lineIndex++
	}
	r.buf.cursor.x = p.x + r.viewAnchor.x
	r.buf.cursor.y = lineIndex
}

func (r *Render) SyncCursorToView() {
	r.buf.cursor = r.buf.cursor
	r.viewCursor.x = r.buf.cursor.x
	r.viewCursor.y = 0
	if len(r.buf.lines) > 0 {
		currLine := &r.buf.lines[r.buf.cursor.x]
		if len(currLine.txt) > 0 {
			for j := 0; j < r.buf.cursor.y; j++ {
				r.viewCursor.y += runeRenderedWidth(r.viewCursor.y, currLine.txt[j])
			}
		}
	}
	r.offsetView()
}

func (r *Render) offsetView() {
	if r.viewAnchor.x > 0 && r.viewCursor.x < r.viewAnchor.x {
		r.viewAnchor.x = r.viewCursor.x
	}
	if r.viewCursor.x > r.viewAnchor.x+r.viewMaxH-2 {
		r.viewAnchor.x = r.viewCursor.x - r.viewMaxH + 2
	}
	if r.viewAnchor.y > 0 && r.viewCursor.y < r.viewAnchor.y {
		r.viewAnchor.y = r.viewCursor.y
	}
	if r.viewCursor.y > r.viewAnchor.y+r.viewMaxW-1 {
		r.viewAnchor.y = r.viewCursor.y - r.viewMaxW + 1
	}
}

// y width, x height
func tbprint(x, y int, fg, bg tm.Attribute, msg string) {
	for _, c := range msg {
		tm.SetCell(y, x, c, fg, bg)
		y += runewidth.RuneWidth(c)
	}
}

func (r *Render) moveCursorToNextHalfScreen() {
	if r.buf.cursor.x+r.termH/2 >= len(r.buf.lines) {
		r.buf.cursor.x = len(r.buf.lines) - 1
	} else {
		r.buf.cursor.x += r.termH / 2
	}
	if r.buf.cursor.y > len(r.buf.lines[r.buf.cursor.x].txt) {
		r.buf.cursor.y = len(r.buf.lines[r.buf.cursor.x].txt)
	}
}

func (r *Render) moveCursorToPrevHalfScreen() {
	if r.buf.cursor.x-r.termH/2 < 0 {
		r.buf.cursor.x = 0
	} else {
		r.buf.cursor.x -= r.termH / 2
	}
	if r.buf.cursor.y > len(r.buf.lines[r.buf.cursor.x].txt) {
		r.buf.cursor.y = len(r.buf.lines[r.buf.cursor.x].txt)
	}
}

func getLinePer(buf *Buffer) int {
	if len(buf.lines) > 0 {
		return int((buf.cursor.x + 1) * 100 / len(buf.lines))
	}
	return 0
}

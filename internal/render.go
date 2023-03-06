package pine

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

type Render struct {
	termH, termW       int
	viewMaxH, viewMaxW int
	buf                *Buffer
	cursor             *Pos
	viewCursor         *Pos
	viewAnchor         *Pos
	mouseCursor        *Pos
}

func (r *Render) Init(b *Buffer, c *Pos) {
	r.buf = b
	r.cursor = c
	r.viewCursor = &Pos{0, 0}
	r.viewAnchor = &Pos{0, 0}
	r.mouseCursor = &Pos{0, 0}
}

func (r *Render) Clear() {
	if err := tm.Clear(tm.ColorDefault, tm.ColorDefault); err != nil {
		log.Warn("failed to clear screen")
	}
}

func (r *Render) DrawScreen(mode Mode, msg string) {
	r.Clear()
	defer tm.Flush()

	r.termW, r.termH = tm.Size()
	r.viewMaxH = r.termH - 1
	r.viewMaxW = r.termW

	r.drawStatusline(mode)

	if mode == HelpMode {
		r.drawHelpPage()
		tm.HideCursor()
		return
	}

	if mode == WelcomeMode {
		r.drawWelcomePage()
		tm.HideCursor()
		return
	}

	if r.buf == nil {
		panic("buffer is null while in edit/file mode")
	}

	r.drawBuffer()
	r.drawCursor()

	if len(msg) > 0 {
		r.drawMessage(msg)
	}
}

func (r *Render) drawMessage(msg string) {
	tbprint(r.termH-1, 0, tm.ColorRed, tm.ColorDefault, msg)
}

func (r *Render) drawWelcomePage() {
	tbprint(2, 0, tm.ColorDefault, tm.ColorDefault, "Press any key to start editing a new file")
}

func (r *Render) drawHelpPage() {
	tbprint(2, 0, tm.ColorDefault, tm.ColorDefault, "Files")
	tbprint(3, 0, tm.ColorDefault, tm.ColorDefault, "^R: open a file")
	tbprint(4, 0, tm.ColorDefault, tm.ColorDefault, "^O: save a file")
	tbprint(5, 0, tm.ColorDefault, tm.ColorDefault, "Navigation")
	tbprint(6, 0, tm.ColorDefault, tm.ColorDefault, "^A: go to start of the line")
	tbprint(7, 0, tm.ColorDefault, tm.ColorDefault, "^E: go to end of the line")
	tbprint(8, 0, tm.ColorDefault, tm.ColorDefault, "^V: go to next half screen")
	tbprint(9, 0, tm.ColorDefault, tm.ColorDefault, "^Z: go to previous half screen")
}

func (r *Render) drawStatusline(mode Mode) {
	for i := 0; i < r.termW-1; i++ {
		tm.SetCell(i, 0, rune(' '), tm.ColorBlack, tm.ColorWhite)
	}
	if mode == HelpMode {
		tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, " HELP PAGE | Press ^X to return")
		return
	}
	if mode == FileOpenMode {
		tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, " INPUT A FILE NAME | Press Enter to open")
		return
	}
	if mode == FileSaveMode {
		tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, " INPUT A FILE NAME | Press Enter to save")
		return
	}
	currLineLen := 0
	if len(r.buf.lines) > 0 {
		currLineLen = len(r.buf.lines[r.cursor.x].txt)
	}
	tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("Pine Editor v%s", VERSION))
	filePath := "New Buffer"
	if r.buf.filePath != "" {
		filePath = r.buf.filePath
	}
	tbprint(0, r.termW-len(filePath), tm.ColorBlack, tm.ColorWhite, filePath)
	tbprint(r.termH-1, r.termW-16, tm.ColorDefault, tm.ColorDefault, "^/ Help; ^X Exit")
	tbprint(r.termH-1, 0, tm.ColorDefault, tm.ColorDefault, fmt.Sprintf("%d:%d | vc %d:%d | va %d:%d | mc %d:%d | op:%s | tr: %d; crc: %d", r.cursor.x, r.cursor.y, r.viewCursor.x, r.viewCursor.y, r.viewAnchor.x, r.viewAnchor.y, r.mouseCursor.x, r.mouseCursor.y, string(r.buf.lastModifiedCh), len(r.buf.lines), currLineLen))
}

func (r *Render) drawBuffer() {
	viewIndex := 0
	for i := 0; i < len(r.buf.lines); i++ {
		if i < r.viewAnchor.x {
			continue
		}
		if viewIndex == r.termH-2 {
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
	tbprint(i-r.viewAnchor.x+1, 0, tm.ColorDefault, tm.ColorDefault, renderedData[r.viewAnchor.y:])
	if len(renderedData[r.viewAnchor.y:]) >= r.termW {
		tbprint(i-r.viewAnchor.x+1, r.termW-1, tm.ColorCyan, tm.ColorDefault, ">")
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
func (r *Render) MoveCursor(
	keyType tm.Key,
) {
	if len(r.buf.lines) == 0 {
		return
	}
	switch keyType {
	case tm.KeyArrowUp:
		r.moveCursorUp()
	case tm.KeyArrowDown:
		r.moveCursorDown()
	case tm.KeyArrowLeft:
		r.moveCursorLeft()
	case tm.KeyArrowRight:
		r.moveCursorRight()
	}
}

func (r *Render) moveCursorUp() {
	if r.cursor.x == 0 {
		return
	}
	r.cursor.x--
	if r.cursor.y > len(r.buf.lines[r.cursor.x].txt) {
		r.cursor.y = len(r.buf.lines[r.cursor.x].txt)
	}
}

func (r *Render) moveCursorDown() {
	if r.cursor.x == len(r.buf.lines)-1 {
		return
	}
	r.cursor.x++
	if r.cursor.y > len(r.buf.lines[r.cursor.x].txt) {
		r.cursor.y = len(r.buf.lines[r.cursor.x].txt)
	}
}

func (r *Render) moveCursorLeft() {
	if r.cursor.y == 0 && r.viewAnchor.y > 0 {
		panic("cursor and viewpoint out of sync")
	}
	if r.cursor.y == 0 {
		return
	}
	r.cursor.y--
}

func (r *Render) moveCursorRight() {
	if r.cursor.y == len(r.buf.lines[r.cursor.x].txt) {
		return
	}
	r.cursor.y++
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
	r.cursor.x = p.x + r.viewAnchor.x
	r.cursor.y = lineIndex
}

func (r *Render) SyncCursorToView() {
	r.viewCursor.x = r.cursor.x
	r.viewCursor.y = 0
	currLine := &r.buf.lines[r.cursor.x]
	if len(r.buf.lines) > 0 && len(currLine.txt) > 0 {
		for j := 0; j < r.cursor.y; j++ {
			r.viewCursor.y += runeRenderedWidth(r.viewCursor.y, currLine.txt[j])
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

// func (r *Render) printInfo(msg string) {
// 	tbprint(0, r.termH-1, tm.ColorRed, tm.ColorDefault, msg)
// }

func (r *Render) moveCursorToNextHalfScreen() {
	if r.cursor.x+r.termH/2 >= len(r.buf.lines) {
		r.cursor.x = len(r.buf.lines) - 1
	} else {
		r.cursor.x += r.termH / 2
	}
	if r.cursor.y > len(r.buf.lines[r.cursor.x].txt) {
		r.cursor.y = len(r.buf.lines[r.cursor.x].txt)
	}
}

func (r *Render) moveCursorToPrevHalfScreen() {
	if r.cursor.x-r.termH/2 < 0 {
		r.cursor.x = 0
	} else {
		r.cursor.x -= r.termH / 2
	}
	if r.cursor.y > len(r.buf.lines[r.cursor.x].txt) {
		r.cursor.y = len(r.buf.lines[r.cursor.x].txt)
	}
}

package pine

import (
	"fmt"
	"strconv"

	tm "github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

const (
	HeaderlineOffset = 0
	StatuslineOffset = -1
	FileOpenInfo     = "Open file (^X ^X to cancel): "
	FileSaveInfo     = "Save file (^X ^X to cancel): "
)

type Render struct {
	termH, termW  int
	bufRender     *BufRender // bufRender renders content of the main buffer
	miscBufRender *BufRender // miscBufRender renders content of the buffer for misc usage, e.g. open/save files
	sett          *Setting
	log           *log.Logger
}

type RenderContent struct {
	buf      *Buffer
	miscBuf  *Buffer
	mode     Mode
	mod      tm.Modifier
	key      tm.Key
	ch       rune
	bufIdx   int
	bufDirty bool
}

// BufRender renders content of a buffer
type BufRender struct {
	viewStartPos *Pos
	viewEndPos   *Pos
	viewCursor   *Pos
	viewAnchor   *Pos
	log          *log.Logger
}

func (r *Render) Init(sett *Setting, logger *log.Logger) {
	r.sett = sett
	r.log = logger
	r.bufRender = &BufRender{}
	r.bufRender.Reset()
	r.miscBufRender = &BufRender{}
	r.miscBufRender.Reset()
}

/*
Update viewpoint of each rendering components
1   Headline
2   Main Buffer
.
.
End-1 Statusline
End   Misc Buffer
*/
func (r *Render) updateViewPos(mode Mode) {
	r.termW, r.termH = tm.Size()
	r.bufRender.viewStartPos = &Pos{1, 0}
	r.bufRender.viewEndPos = &Pos{r.termH - 2, r.termW}
	offset := 0
	if mode == FileOpenMode {
		offset = len(FileOpenInfo)
	}
	if mode == FileSaveMode {
		offset = len(FileSaveInfo)
	}
	r.miscBufRender.viewStartPos = &Pos{r.termH - 1, offset}
	r.miscBufRender.viewEndPos = &Pos{r.termH, r.termW}
}

func (r *Render) Clear() {
	if err := tm.Clear(tm.ColorDefault, tm.ColorDefault); err != nil {
		r.log.Error("failed to clear screen")
	}
}

func (r *Render) Draw(content RenderContent) {
	r.Clear()
	defer tm.Flush()

	r.updateViewPos(content.mode)
	r.bufRender.SyncCursorToView(content.buf)
	if content.mode == FileOpenMode || content.mode == FileSaveMode {
		r.miscBufRender.SyncCursorToView(content.miscBuf)
	}

	r.drawHeadline(content)

	miscMode := false
	if content.mode == FileOpenMode || content.mode == FileSaveMode {
		miscMode = true
	}
	r.bufRender.Draw(content.buf, !miscMode)
	r.miscBufRender.Draw(content.miscBuf, miscMode)
	if miscMode {
		r.drawMiscInfo(content.mode)
	}

	r.drawStatusline(content)
}

func (r *Render) MoveCursor(mode Mode, buf *Buffer, op KeyOps) {
	bufRender := r.bufRender
	if mode == FileOpenMode || mode == FileSaveMode {
		bufRender = r.miscBufRender
	}
	if buf.isEmpty() {
		r.log.Warnf("try to move cursor in an empty buffer: %s", buf.filePath)
		return
	}
	switch op {
	case MoveCursorUpOp:
		bufRender.moveCursorUp(buf)
	case MoveCursorDownOp:
		bufRender.moveCursorDown(buf)
	case MoveCursorLeftOp:
		bufRender.moveCursorLeft(buf)
	case MoveCursorRightOp:
		bufRender.moveCursorRight(buf)
	}
}

func (r *Render) MoveCursorByMouse(buf *Buffer, p Pos) {
	r.bufRender.MoveCursorByMouse(buf, p)
}

func (r *Render) drawHeadline(content RenderContent) {
	for i := 0; i < r.termW; i++ {
		tm.SetCell(i, 0, rune(' '), tm.ColorBlack, tm.ColorWhite)
	}
	tbprint(0, 0, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("Pine Editor v%s", VERSION))
	bufId := strconv.Itoa(content.bufIdx)
	bufDirtyMark := " "
	if content.buf.dirty {
		bufDirtyMark = "*"
	}
	bufName := getFilename(content.buf.filePath)
	tbprint(0, r.termW-len(bufName)-len(bufId)-3, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("%s%s: %s", bufDirtyMark, bufId, bufName))
}

func (r *Render) drawStatusline(content RenderContent) {
	x := r.termH - 1 + StatuslineOffset
	for i := 0; i < r.termW; i++ {
		tm.SetCell(i, x, rune(' '), tm.ColorCyan, tm.ColorCyan)
	}
	buf := content.buf
	tbprint(x, 0, tm.ColorBlack, tm.ColorCyan, fmt.Sprintf("%06d,%06d %4d%%  %x-%s:%x", buf.cursor.x, buf.cursor.y, getLinePer(buf), int(content.mod), string(content.ch), int(content.key)))
	statusTailMsg := "^/ Help    ^X Exit"
	tbprint(x, r.termW-len(statusTailMsg), tm.ColorBlack, tm.ColorCyan, statusTailMsg)
}

func (r *Render) drawMiscInfo(mode Mode) {
	info := ""
	if mode == FileOpenMode {
		info = FileOpenInfo
	}
	if mode == FileSaveMode {
		info = FileSaveInfo
	}
	tbprint(r.miscBufRender.viewStartPos.x, 0, tm.ColorCyan, tm.ColorDefault, info)
}

func (r *BufRender) Reset() {
	r.viewStartPos = &Pos{0, 0}
	r.viewEndPos = &Pos{0, 0}
	r.viewCursor = &Pos{0, 0}
	r.viewAnchor = &Pos{0, 0}
}

func (r *BufRender) Draw(buf *Buffer, hasCursor bool) {
	drawBuffer(buf, r.viewStartPos, r.viewEndPos, r.viewAnchor, r.viewCursor)
	if hasCursor {
		drawCursor(r.viewStartPos, r.viewAnchor, r.viewCursor)
	}
}

// Calculate cursor colume position
// Cursor buffer position is different than terminal view
// since runes can have multiple width
func (r *BufRender) SyncCursorToView(buf *Buffer) {
	r.viewCursor.x = buf.cursor.x
	r.viewCursor.y = 0
	if len(buf.lines) > 0 {
		currLine := &buf.lines[buf.cursor.x]
		if len(currLine.txt) > 0 {
			for j := 0; j < buf.cursor.y; j++ {
				r.viewCursor.y += runeRenderedWidth(r.viewCursor.y, currLine.txt[j])
			}
		}
	}
	offsetView(r.viewCursor, r.viewAnchor, r.viewStartPos, r.viewEndPos)
}

/*
 * MoveCursor
 *
 * Update cursor position and view based on cursor movement
 * Cursor x,y are determined by view x,y
 */
func (r *BufRender) moveCursorUp(buf *Buffer) {
	if buf.cursor.x == 0 {
		return
	}
	buf.cursor.x--
	if buf.cursor.y > len(buf.lines[buf.cursor.x].txt) {
		buf.cursor.y = len(buf.lines[buf.cursor.x].txt)
	}
}

func (r *BufRender) moveCursorDown(buf *Buffer) {
	if buf.cursor.x == len(buf.lines)-1 {
		return
	}
	buf.cursor.x++
	if buf.cursor.y > len(buf.lines[buf.cursor.x].txt) {
		buf.cursor.y = len(buf.lines[buf.cursor.x].txt)
	}
}

func (r *BufRender) moveCursorLeft(buf *Buffer) {
	if buf.cursor.y == 0 && r.viewAnchor.y > 0 {
		r.log.Fatalf("cursor %d and viewpoint %d out of sync", buf.cursor.y, r.viewAnchor.y)
	}
	if buf.cursor.y == 0 {
		return
	}
	buf.cursor.y--
}

func (r *BufRender) moveCursorRight(buf *Buffer) {
	if buf.cursor.y == len(buf.lines[buf.cursor.x].txt) {
		return
	}
	buf.cursor.y++
}

/*
 * Mouse position is limited by terminal height and weight
 * It should be one to one mapped to view cursor position
 * But there are multi-length rune that we cannot position cursor
 * in between. So we convert back to cursor first.
 *
 */
func (r *BufRender) MoveCursorByMouse(buf *Buffer, p Pos) {
	r.syncViewCursorToCursor(buf, p)
}

func (r *BufRender) syncViewCursorToCursor(buf *Buffer, p Pos) {
	if p.x < r.viewStartPos.x || ((p.x+r.viewAnchor.x) >= len(buf.lines) || p.x >= r.viewEndPos.x-1) {
		return
	}
	currLine := buf.lines[p.x+r.viewAnchor.x]
	lineIndex := 0
	viewLineIndex := 0
	for viewLineIndex < p.y+r.viewAnchor.y && lineIndex < len(currLine.txt) {
		viewLineIndex += runeRenderedWidth(viewLineIndex, currLine.txt[lineIndex])
		lineIndex++
	}
	buf.cursor.x = p.x + r.viewAnchor.x
	buf.cursor.y = lineIndex
}

func (r *BufRender) moveCursorToNextHalfScreen(buf *Buffer) {
	h := r.viewEndPos.x - r.viewStartPos.x
	if buf.cursor.x+h/2 >= len(buf.lines) {
		buf.cursor.x = len(buf.lines) - 1
	} else {
		buf.cursor.x += h / 2
	}
	if buf.cursor.y > len(buf.lines[buf.cursor.x].txt) {
		buf.cursor.y = len(buf.lines[buf.cursor.x].txt)
	}
}

func (r *BufRender) moveCursorToPrevHalfScreen(buf *Buffer) {
	h := r.viewEndPos.x - r.viewStartPos.x
	if buf.cursor.x-h/2 < 0 {
		buf.cursor.x = 0
	} else {
		buf.cursor.x -= h / 2
	}
	if buf.cursor.y > len(buf.lines[buf.cursor.x].txt) {
		buf.cursor.y = len(buf.lines[buf.cursor.x].txt)
	}
}

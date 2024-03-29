package pine

import (
	"fmt"
	"strconv"

	tm "github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

const (
	FileOpenInfo = "Open file (^G to cancel): "
	FileSaveInfo = "Save file (^G to cancel): "
	SearchInfo   = "Search (^G to cancel): "
)

type Render struct {
	termH, termW  int
	bufRender     *BufRender // bufRender renders content of the main buffer
	miscBufRender *BufRender // miscBufRender renders content of the buffer for misc usage, e.g. open/save files
	sett          *Setting
	log           *log.Logger
	event         tm.Event
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
// ViewStartPos and ViewEndPos are the absolute coordinate of the view on terminal screen
// ViewCursor is the absolute coordinate of the cusor
// ViewAnchor is the coordinate of buffer content, used to calculate content outside of the screen
// HlStartPos and hlEndPos are the view coordinate of the highlight area
type BufRender struct {
	viewStartPos   *Pos
	viewEndPos     *Pos
	viewCursor     *Pos
	viewAnchor     *Pos
	hlViewStartPos *Pos
	hlViewEndPos   *Pos
	log            *log.Logger
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
	r.bufRender.viewStartPos = &Pos{BUFFER_CONTENT_START_OFFSET, 0}
	if mode == DirMode {
		r.bufRender.viewStartPos = &Pos{BUFFER_DIR_CONTENT_START_OFFSET, 0}
	}
	r.bufRender.viewEndPos = &Pos{r.termH + BUFFER_END_OFFSET, r.termW}
	offset := 0
	switch mode {
	case FileOpenMode:
		offset = len(FileOpenInfo)
	case FileSaveMode:
		offset = len(FileSaveInfo)
	case SearchMode:
		offset = len(SearchInfo)
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
	if content.mode == FileOpenMode || content.mode == FileSaveMode || content.mode == SearchMode {
		r.miscBufRender.SyncCursorToView(content.miscBuf)
	}

	r.drawHeadline(content)

	if content.mode == DirMode {
		r.drawDir(content.buf.filePath)
	}

	miscMode := false
	if content.mode == FileOpenMode || content.mode == FileSaveMode || content.mode == SearchMode {
		miscMode = true
	}
	r.bufRender.Draw(content.buf, !miscMode, content.mode == SearchMode)
	r.miscBufRender.Draw(content.miscBuf, miscMode, false)
	if miscMode {
		r.drawMiscInfo(content.mode)
	}

	r.drawStatusline(content)
}

func (r *Render) MoveCursor(mode Mode, buf *Buffer, op KeyOps) {
	bufRender := r.bufRender
	if mode == FileOpenMode || mode == FileSaveMode || mode == SearchMode {
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

func (r *Render) MoveCursorByMouse(buf *Buffer, p Pos, mode Mode) {
	r.bufRender.MoveCursorByMouse(buf, p, mode)
}

func (r *Render) drawHeadline(content RenderContent) {
	for i := 0; i < r.termW; i++ {
		tm.SetCell(i, 0, rune(' '), tm.ColorBlack, tm.ColorWhite)
	}
	tbprint(HEADLINE_OFFSET, 0, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("Pine Editor v%s", VERSION))
	bufId := strconv.Itoa(content.bufIdx)
	bufDirtyMark := " "
	if content.buf.dirty {
		bufDirtyMark = "*"
	}
	bufName := getFilename(content.buf.filePath)
	tbprint(HEADLINE_OFFSET, r.termW-len(bufName)-len(bufId)-3, tm.ColorBlack, tm.ColorWhite, fmt.Sprintf("%s%s: %s", bufDirtyMark, bufId, bufName))
}

func (r *Render) drawStatusline(content RenderContent) {
	x := r.termH - 1 + STATUSLINE_OFFSET
	for i := 0; i < r.termW; i++ {
		tm.SetCell(i, x, rune(' '), tm.ColorCyan, tm.ColorCyan)
	}
	buf := content.buf
	tbprint(x, 0, tm.ColorBlack, tm.ColorCyan, fmt.Sprintf("%06d,%06d %4d%%  %x-%s:%x %d:%d", buf.cursor.x, buf.cursor.y, getLinePer(buf), int(content.mod), string(content.ch), int(content.key), r.bufRender.hlViewStartPos.x, r.bufRender.hlViewStartPos.y))
	statusTailMsg := "^/ Help    ^X Exit"
	tbprint(x, r.termW-len(statusTailMsg), tm.ColorBlack, tm.ColorCyan, statusTailMsg)
}

func (r *Render) drawMiscInfo(mode Mode) {
	info := ""
	switch mode {
	case FileOpenMode:
		info = FileOpenInfo
	case FileSaveMode:
		info = FileSaveInfo
	case SearchMode:
		info = SearchInfo
	}
	tbprint(r.miscBufRender.viewStartPos.x, 0, tm.ColorCyan, tm.ColorDefault, info)
}

func (r *Render) drawDir(path string) {
	tbprint(1, 0, tm.ColorDefault, tm.ColorDefault, fmt.Sprintf("Files under %s", path))
}

func (r *Render) getBufNamePos(filePath string, bufIdx int) (Pos, Pos) {
	bufName := getFilename(filePath)
	bufId := strconv.Itoa(bufIdx)
	return Pos{HEADLINE_OFFSET, r.termW - len(bufName) - len(bufId) - 3}, Pos{HEADLINE_OFFSET + 1, r.termW}
}

func (r *Render) IsMousePointerOnBufferName(mousePos Pos, filePath string, bufIdx int) bool {
	bufStartPos, bufEndPos := r.getBufNamePos(filePath, bufIdx)
	return isOnArea(mousePos, bufStartPos, bufEndPos)
}

func (r *Render) IsMousePointerOnBuffer(mousePos Pos) bool {
	return isOnArea(mousePos, *r.bufRender.viewStartPos, *r.bufRender.viewEndPos)
}

func (r *BufRender) Reset() {
	r.viewStartPos = &Pos{0, 0}
	r.viewEndPos = &Pos{0, 0}
	r.viewCursor = &Pos{0, 0}
	r.viewAnchor = &Pos{0, 0}
	r.hlViewStartPos = &Pos{-1, -1}
	r.hlViewEndPos = &Pos{-1, -1}
}

func (r *BufRender) Draw(buf *Buffer, hasCursor, hasHighlight bool) {
	drawBuffer(buf, r.viewStartPos, r.viewEndPos, r.viewAnchor, r.viewCursor)
	if hasCursor {
		drawCursor(r.viewStartPos, r.viewAnchor, r.viewCursor)
	}
	r.updateHighlight(buf)
	if hasHighlight {
		drawHighlight(r.hlViewStartPos, r.hlViewEndPos, r.viewAnchor, r.viewStartPos, r.viewEndPos)
	}
}

// Calculate cursor colume position
// Cursor buffer position is different than terminal view
// since runes can have multiple width
func (r *BufRender) SyncCursorToView(buf *Buffer) {
	convertBufPosToViewPos(r.viewCursor, buf.cursor, r.viewAnchor, r.viewStartPos, r.viewEndPos, buf.lines)
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
	r.viewCursor.x -= 1
	r.syncViewPosToCursor(buf, Pos{r.viewCursor.x - r.viewAnchor.x, r.viewCursor.y - r.viewAnchor.y})
}

func (r *BufRender) moveCursorDown(buf *Buffer) {
	if buf.cursor.x == len(buf.lines)-1 {
		return
	}
	r.viewCursor.x += 1
	r.syncViewPosToCursor(buf, Pos{r.viewCursor.x - r.viewAnchor.x, r.viewCursor.y - r.viewAnchor.y})
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
func (r *BufRender) MoveCursorByMouse(buf *Buffer, p Pos, mode Mode) {
	if p.x < r.viewStartPos.x || p.x >= r.viewEndPos.x {
		return
	}
	offset := BUFFER_CONTENT_START_OFFSET
	if mode == DirMode {
		offset += 1
	}
	viewCursorPos := Pos{p.x - offset, p.y}
	r.syncViewPosToCursor(buf, viewCursorPos)
}

// Sync view position to cursor
// View position is the absolute coordinate of terminal
func (r *BufRender) syncViewPosToCursor(buf *Buffer, viewPos Pos) {
	if len(buf.lines) <= 0 {
		return
	}
	if (viewPos.x + r.viewAnchor.x) >= len(buf.lines) {
		viewPos.x = len(buf.lines) - 1
	}
	currLine := buf.lines[viewPos.x+r.viewAnchor.x]
	lineIndex := 0
	viewLineIndex := 0
	for viewLineIndex < viewPos.y+r.viewAnchor.y && lineIndex < len(currLine.txt) {
		viewLineIndex += runeRenderedWidth(viewLineIndex, currLine.txt[lineIndex])
		lineIndex++
	}
	buf.cursor.x = viewPos.x + r.viewAnchor.x
	buf.cursor.y = lineIndex
}

func (r *BufRender) moveCursorToNextHalfScreen(buf *Buffer) {
	h := r.viewEndPos.x - r.viewStartPos.x
	if buf.cursor.x+h/2 >= len(buf.lines) {
		buf.cursor.x = len(buf.lines) - 1
	} else {
		buf.cursor.x += h / 2
	}
	r.syncViewPosToCursor(buf, Pos{buf.cursor.x - r.viewAnchor.x, buf.cursor.y - r.viewAnchor.y})
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
	r.syncViewPosToCursor(buf, Pos{buf.cursor.x - r.viewAnchor.x, buf.cursor.y - r.viewAnchor.y})
}

func (r *BufRender) updateHighlight(buf *Buffer) {
	convertBufPosToViewPos(r.hlViewStartPos, buf.hlStartPos, r.viewAnchor, r.viewStartPos, r.viewEndPos, buf.lines)
	convertBufPosToViewPos(r.hlViewEndPos, buf.hlEndPos, r.viewAnchor, r.viewStartPos, r.viewEndPos, buf.lines)
}

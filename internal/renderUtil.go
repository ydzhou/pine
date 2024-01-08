package pine

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
)

func getLinePer(buf *Buffer) int {
	if len(buf.lines) > 0 {
		return int((buf.cursor.x + 1) * 100 / len(buf.lines))
	}
	return 0
}

func drawBuffer(
	buf *Buffer,
	viewStartPos, viewEndPos, viewAnchor, viewCursor *Pos,
) {
	viewIndex := 0
	for i := 0; i < len(buf.lines); i++ {
		if i < viewAnchor.x {
			continue
		}
		// Buffer lines are out of current view point
		if viewIndex == viewEndPos.x-viewStartPos.x {
			break
		}
		viewIndex++
		drawBufferLine(buf.lines[i], i, viewIndex, viewStartPos, viewEndPos, viewAnchor, viewCursor)
	}
}

func drawBufferLine(
	line line,
	i, viewIndex int,
	viewStartPos, viewEndPos, viewAnchor, viewCursor *Pos,
) {
	renderedData := ""
	y := 0
	for _, ch := range line.txt {
		if ch == rune('\t') {
			renderedData += drawTab(runeRenderedWidth(y, ch))
		} else {
			renderedData += string(ch)
		}
		y += runeRenderedWidth(y, ch)
	}
	if y < viewAnchor.y {
		return
	}
	// TODO: it can print out of the screen. but termbox-go handles
	// this misbehavior. need to clean up this mess.
	tbprint(i-viewAnchor.x+viewStartPos.x, viewStartPos.y, tm.ColorDefault, tm.ColorDefault, renderedData[viewAnchor.y:])
	if len(renderedData[viewAnchor.y:]) >= (viewEndPos.y - viewStartPos.y) {
		tbprint(i-viewAnchor.x+viewStartPos.x, viewEndPos.y-1, tm.ColorCyan, tm.ColorDefault, ">")
	}
}

func drawTab(tablength int) string {
	tab := ""
	for i := 0; i < tablength; i++ {
		tab += " "
	}
	return tab
}

func drawCursor(viewStartPos, viewAnchor, viewCursor *Pos) {
	tm.SetCursor(viewStartPos.y+viewCursor.y-viewAnchor.y, viewStartPos.x+viewCursor.x-viewAnchor.x)
}

func offsetView(viewCursor, viewAnchor, viewStartPos, viewEndPos *Pos) {
	h := viewEndPos.x - viewStartPos.x
	w := viewEndPos.y - viewStartPos.y
	if viewAnchor.x > 0 && viewCursor.x < viewAnchor.x {
		viewAnchor.x = viewCursor.x
	}
	if viewCursor.x > viewAnchor.x+h-1 {
		viewAnchor.x = viewCursor.x - h + 1
	}
	if viewAnchor.y > 0 && viewCursor.y < viewAnchor.y {
		viewAnchor.y = viewCursor.y
	}
	if viewCursor.y > viewAnchor.y+w-1 {
		viewAnchor.y = viewCursor.y - w + 1
	}
}

func unsavedBufferMsg(bufIdxs []int) string {
	msg := "Unsaved changes at Buffer "
	for _, idx := range bufIdxs {
		msg += fmt.Sprintf("%d, ", idx)
	}
	return msg + "press ^X to discard"
}

// y width, x height
func tbprint(x, y int, fg, bg tm.Attribute, msg string) {
	for _, c := range msg {
		tm.SetCell(y, x, c, fg, bg)
		y += runewidth.RuneWidth(c)
	}
}

// Check if given coordinate is on the target area
func isOnArea(p Pos, startPos Pos, endPos Pos) bool {
	return p.x >= startPos.x && p.y >= startPos.y && p.x < endPos.x && p.y < endPos.y
}

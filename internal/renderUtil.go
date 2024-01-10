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
		renderedCh := ""
		if ch == rune('\t') {
			renderedCh = drawTab(runeRenderedWidth(y, ch))
		} else {
			renderedCh = string(ch)
		}
		y += runeRenderedWidth(y, ch)
		renderedData += renderedCh
	}
	if y < viewAnchor.y {
		return
	}
	// TODO: it can print out of the screen. but termbox-go handles
	// this misbehavior. need to clean up this mess.
	tbprint(i-viewAnchor.x+viewStartPos.x, viewStartPos.y, tm.ColorDefault, tm.ColorDefault, renderedData[viewAnchor.y:])
	if y-viewAnchor.y >= (viewEndPos.y - viewStartPos.y) {
		tbprint(i-viewAnchor.x+viewStartPos.x, viewEndPos.y-1, tm.ColorDefault, tm.ColorDefault, ">")
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

func drawHighlight(hlViewStartPos, hlViewEndPos, viewAnchor, viewStartPos, viewEndPos *Pos) {
	startX := hlViewStartPos.x + viewStartPos.x - viewAnchor.x
	startY := hlViewStartPos.y + viewStartPos.y - viewAnchor.y
	endX := hlViewEndPos.x + viewStartPos.x - viewAnchor.x
	endY := hlViewEndPos.y + viewStartPos.y - viewAnchor.y

	for i := startX; i <= endX; i++ {
		for j := 0; j < viewEndPos.y-viewStartPos.y; j++ {
			tm.SetBg(j, i, tm.ColorWhite)
			tm.SetFg(j, i, tm.ColorBlack)
		}
	}
	for j := 0; j < startY; j++ {
		tm.SetBg(j, startX, tm.ColorDefault)
		tm.SetFg(j, startX, tm.ColorDefault)
	}
	for j := endY; j < viewEndPos.y-viewStartPos.y; j++ {
		tm.SetBg(j, endX, tm.ColorDefault)
		tm.SetFg(j, endX, tm.ColorDefault)
	}
}

func convertBufPosToViewPos(
	viewPos, bufPos, viewAnchor, viewStartPos, viewEndPos *Pos,
	lines []line,
) {
	if len(lines) <= 0 || bufPos.x < 0 || bufPos.x >= len(lines) {
		return
	}
	viewPos.x = bufPos.x
	viewPos.y = 0
	currLine := &lines[bufPos.x]
	if len(currLine.txt) > 0 {
		for j := 0; j < bufPos.y; j++ {
			viewPos.y += runeRenderedWidth(viewPos.y, currLine.txt[j])
		}
	}

	offsetView(viewPos, viewAnchor, viewStartPos, viewEndPos)
}

func offsetView(viewPos, viewAnchor, viewStartPos, viewEndPos *Pos) {
	h := viewEndPos.x - viewStartPos.x
	w := viewEndPos.y - viewStartPos.y
	if viewAnchor.x > 0 && viewPos.x < viewAnchor.x {
		viewAnchor.x = viewPos.x
	}
	if viewPos.x > viewAnchor.x+h-1 {
		viewAnchor.x = viewPos.x - h + 1
	}
	if viewAnchor.y > 0 && viewPos.y < viewAnchor.y {
		viewAnchor.y = viewPos.y
	}
	if viewPos.y > viewAnchor.y+w-1 {
		viewAnchor.y = viewPos.y - w + 1
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

package ste

import (
    "fmt"
)

type Buffer struct {
    lines []line
    dirty bool
    id int
    lastModifiedCh string
    rowIndex, colIndex int
    cursor *Pos
}

type line struct {
    txt [] rune
}

func (b *Buffer) New() {
    b.lines = []line{line{}}
}

func (b *Buffer) NewLine(cursor *Pos) {
    b.lastModifiedCh = string("newline")

    x := cursor.x
    y := cursor.y

    if x > len(b.lines) - 1 || y > len(b.lines[x].txt) {
        panic(fmt.Errorf("failed to create new line at (%d,%d)", x, y))
    }

    line := line{}
    if y <= len(b.lines[x].txt) {
        line.txt = make([]rune, len(b.lines[x].txt[:y]))
        copy(line.txt, b.lines[x].txt[:y])
        nextLineTxt := make([]rune, len(b.lines[x].txt[y:]))
        copy(nextLineTxt, b.lines[x].txt[y:])
        b.lines[x].txt = nextLineTxt
    }

    b.lines = append(b.lines, line)
    copy(b.lines[x+1:], b.lines[x:])
    b.lines[x] = line

    cursor.x++
    cursor.y = 0

    return
}

func (b *Buffer) Insert(cursor *Pos, data rune) {
    x := cursor.x
    y := cursor.y
    cursor.y ++
    b.lastModifiedCh = fmt.Sprintf("+%s", string(data))

    // Append a new line if cursor is under the last line
    if x == len(b.lines) - 1 {
        b.lines = append(b.lines, line{})
    }

    if x > len(b.lines) - 1 || y > len(b.lines[x].txt) {
        panic(fmt.Errorf("failed to insert [%s] at (%d,%d)", string(data), x, y))
    }

    b.lines[x].txt = append(b.lines[x].txt, data)
    if y == len(b.lines[x].txt) - 1 {
        return
    } 

    copy(b.lines[x].txt[y+1:], b.lines[x].txt[y:])
    b.lines[x].txt[y] = data
}

func (b *Buffer) InsertTab() {
    b.Insert(b.cursor, rune('\t'))
}

func (b *Buffer) Delete(cursor *Pos) {
    x := cursor.x
    y := cursor.y
    if x == 0 && y == 0 {
        b.lastModifiedCh = string("NA")
        return
    }
    // Remove newline
    if y == 0 {
        b.lastModifiedCh = string("- newline")
        cursor.y = len(b.lines[x - 1].txt)
        b.lines[x - 1].txt = append(b.lines[x - 1].txt, b.lines[x].txt...)
        b.removeLine(x)
        cursor.x = cursor.x - 1
        b.lastModifiedCh = "-newline"
        return
    }
    b.lastModifiedCh = fmt.Sprintf("-%s", string(b.lines[x].txt[y-1]))
    cursor.y --
    b.removeRune(cursor)
}

func (b *Buffer) removeLine(x int) {
    if x < len(b.lines) - 1 {
        copy(b.lines[x:], b.lines[x+1:])
    }
    b.lines = b.lines[:len(b.lines) - 1]
}

func (b *Buffer) removeRune(cursor *Pos) {
    x := cursor.x
    y := cursor.y
    if y < len(b.lines[x].txt) - 1 {
        copy(b.lines[x].txt[y:], b.lines[x].txt[y+1:])
    }
    b.lines[x].txt = b.lines[x].txt[:len(b.lines[x].txt) - 1]
}

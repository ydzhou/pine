package pine

import (
	"bufio"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

type Buffer struct {
	lines          []line
	dirty          bool
	lastModifiedCh string
	cursor         *Pos
	filePath       string
}

type line struct {
	txt []rune
}

func (b *Buffer) New(cursor *Pos, path string) FileOpenState {
	b.init(cursor)
	b.filePath = path
	b.newEmptyBuffer()
	if path != "" {
		return b.openFile(path)
	}
	return Success
}

func (b *Buffer) init(cursor *Pos) {
	b.cursor = cursor
	b.lines = []line{}
	b.lastModifiedCh = "NA"
	b.dirty = false
}

func (b *Buffer) newEmptyBuffer() {
	b.filePath = getAbsoluteFilePath(DEFAULT_BUFFERNAME)
}

func (b *Buffer) openFile(path string) FileOpenState {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NotFound
		}
		log.Warnf("fail to open file %s", path)
		return HasError
	}
	// TODO: implement function to properly handle directory reading
	if isDirectory(f) {
		return HasError
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		b.lines = append(b.lines, line{txt: []rune(scanner.Text())})
	}
	b.filePath = path
	return Success
}

func isDirectory(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		log.Warnf("File stat failure: %s", f.Name())
		return false
	}
	if stat.IsDir() {
		return true
	}
	return false
}

func (b *Buffer) NewLine() {
	b.lastModifiedCh = string("newline")

	x := b.cursor.x
	y := b.cursor.y

	if x > len(b.lines)-1 || y > len(b.lines[x].txt) {
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

	b.cursor.x++
	b.cursor.y = 0
	b.dirty = true
}

func (b *Buffer) Insert(data rune) {
	x := b.cursor.x
	y := b.cursor.y
	b.cursor.y++
	b.lastModifiedCh = fmt.Sprintf("+%s", string(data))

	// Append a new line if cursor is under the last line
	if x == len(b.lines) {
		b.lines = append(b.lines, line{})
	}

	if x > len(b.lines)-1 || y > len(b.lines[x].txt) {
		panic(fmt.Errorf("failed to insert [%s] at (%d,%d)", string(data), x, y))
	}

	b.lines[x].txt = append(b.lines[x].txt, data)
	if y == len(b.lines[x].txt)-1 {
		return
	}

	copy(b.lines[x].txt[y+1:], b.lines[x].txt[y:])
	b.lines[x].txt[y] = data
	b.dirty = true
}

func (b *Buffer) InsertTab() {
	b.Insert(rune('\t'))
}

func (b *Buffer) DeleteLine() {
	if b.isEmpty() {
		return
	}
	b.removeLine()
	if b.cursor.x > 0 {
		b.cursor.x--
	}
	b.cursor.y = 0
	b.lastModifiedCh = "-line"
}

func (b *Buffer) Delete() {
	x := b.cursor.x
	y := b.cursor.y
	if x == 0 && y == 0 {
		b.lastModifiedCh = string("NA")
		return
	}
	// Remove newline
	if y == 0 {
		b.cursor.y = len(b.lines[x-1].txt)
		b.lines[x-1].txt = append(b.lines[x-1].txt, b.lines[x].txt...)
		b.removeLine()
		b.cursor.x = b.cursor.x - 1
		b.lastModifiedCh = "-newline"
		return
	}
	b.lastModifiedCh = fmt.Sprintf("-%s", string(b.lines[x].txt[y-1]))
	b.cursor.y--
	b.removeRune()
	b.dirty = true
}

func (b *Buffer) removeLine() {
	x := b.cursor.x
	if x == 0 {
		b.lines[0] = line{}
	}
	if x < len(b.lines)-1 {
		copy(b.lines[x:], b.lines[x+1:])
	}
	if len(b.lines) > 0 {
		b.lines = b.lines[:len(b.lines)-1]
	}
}

func (b *Buffer) removeRune() {
	x := b.cursor.x
	y := b.cursor.y
	if y < len(b.lines[x].txt)-1 {
		copy(b.lines[x].txt[y:], b.lines[x].txt[y+1:])
	}
	b.lines[x].txt = b.lines[x].txt[:len(b.lines[x].txt)-1]
}

func (b *Buffer) Save(path string) (int, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	totalbyte := 0
	for _, l := range b.lines {
		for _, d := range l.txt {
			wbyte, err := writer.WriteString(string(d))
			if err != nil {
				return 0, err
			}
			totalbyte += wbyte
		}
		wbyte, err := writer.WriteString("\n")
		if err != nil {
			return 0, err
		}
		totalbyte += wbyte
	}
	writer.Flush()

	b.filePath = path
	b.dirty = false
	return totalbyte, nil
}

func (b *Buffer) InsertString(s string) {
	for _, d := range s {
		b.Insert(d)
	}
}

func (b *Buffer) isEmpty() bool {
	return len(b.lines) == 0
}

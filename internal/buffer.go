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
	hlStartPos     *Pos
	hlEndPos       *Pos
	filePath       string
	isDir          bool
	readOnly       bool
	log            *log.Logger
}

type line struct {
	txt []rune
}

func (b *Buffer) New(path string, log *log.Logger) FileOpenState {
	b.init(log)
	b.filePath = path
	b.newEmptyBuffer()
	if path != "" {
		return b.openFile(path)
	}
	return Success
}

func (b *Buffer) ResetHightlight() {
	b.hlStartPos = &Pos{-1, -1}
	b.hlEndPos = &Pos{-1, -1}
}

func (b *Buffer) init(log *log.Logger) {
	b.log = log
	b.cursor = &Pos{x: 0, y: 0}
	b.ResetHightlight()
	b.lines = []line{}
	b.lastModifiedCh = "NA"
	b.dirty = false
}

func (b *Buffer) newEmptyBuffer() {
	b.filePath = getAbsoluteFilePath(DEFAULT_BUFFERNAME)
	b.lines = []line{}
}

func (b *Buffer) openFile(path string) FileOpenState {
	f, err := os.Open(path)
	if err != nil {
		// If file does not exist, create an empty buffer with given file path
		if os.IsNotExist(err) {
			b.filePath = path
			return NotFound
		}
		b.log.Errorf("fail to open %s: %v", path, err)
		return HasError
	}
	// TODO: implement function to properly handle directory reading
	if isDirectory(f) {
		if err = b.openDir(path); err != nil {
			b.log.Errorf("fail to list directory %s: %v", path, err)
			return HasError
		}
		return IsDir
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
		return false
	}
	if stat.IsDir() {
		return true
	}
	return false
}

func (b *Buffer) openDir(path string) error {
	fs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	b.lines = []line{
		{txt: []rune(".")},
		{txt: []rune("..")},
	}
	for _, f := range fs {
		name := f.Name()
		if f.IsDir() {
			name = f.Name() + "/"
		}
		b.lines = append(b.lines, line{txt: []rune(name)})
	}
	b.isDir = true
	b.readOnly = true
	b.filePath = path
	return nil
}

func (b *Buffer) getCurrDirPath() string {
	if !b.isDir {
		log.Errorf("Cannot get directory in non directory buffer %s", b.filePath)
		return ""
	}
	return string(b.filePath) + "/" + string(b.lines[b.cursor.x].txt)
}

func (b *Buffer) NewLine() {
	defer b.setDirty()
	b.lastModifiedCh = string("newline")

	x := b.cursor.x
	y := b.cursor.y

	if b.isEmpty() {
		b.lines = append(b.lines, line{})
		return
	}

	if x > len(b.lines)-1 || y > len(b.lines[x].txt) {
		panic(fmt.Errorf("failed to create new line at (%d,%d)", x, y))
	}

	indention := getIndention(b.lines[x].txt)

	line := line{txt: []rune{}}
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

	b.applyIndention(b.cursor.x, indention)
}

func (b *Buffer) Insert(data rune) {
	defer b.setDirty()
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
}

func (b *Buffer) InsertTab() {
	b.Insert(rune('\t'))
	b.setDirty()
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
	b.setDirty()
}

func (b *Buffer) Delete() {
	x := b.cursor.x
	y := b.cursor.y
	if x == 0 && y == 0 {
		b.lastModifiedCh = string("NA")
		return
	}
	defer b.setDirty()
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

func (b *Buffer) setDirty() {
	b.dirty = true
}

func (b *Buffer) applyIndention(idx int, indention []rune) {
	curr := b.lines[idx].txt
	existIndention := getIndention(curr)
	if len(indention) > len(existIndention) {
		b.lines[idx].txt = append(indention[:(len(indention)-len(existIndention))], b.lines[idx].txt...)
	}
	b.cursor.y += len(indention)
}

func getIndention(runes []rune) []rune {
	count := 0
	for _, r := range runes {
		if r != rune(' ') && r != rune('\t') {
			break
		}
		count++
	}
	return runes[:count]
}

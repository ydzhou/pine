package ste

import (
	"bufio"
	"fmt"
    "os"
    "path/filepath"
    "strings"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
)

type Editor struct {
    buf *Buffer
    render Render
    miscBuf *Buffer
    cursor *Pos
    miscCursor *Pos
    mode Mode
}

type Pos struct {
    x, y int
}

func (e *Editor) Init() {
    e.cursor = &Pos{0, 0}
    e.miscCursor = &Pos{0, 0}
    e.buf = &Buffer{}
    e.miscBuf = &Buffer{}
    e.buf.New(e.cursor)
    e.render = Render{}
    e.render.Init(e.buf, e.cursor)
}

func (e *Editor) Start(path string) {
    err := tm.Init()
    if err != nil {
        panic(err)
    }
    defer tm.Close()

    msg := ""
    if len(path) > 0 {
        msg = e.Open(path)
    }
    e.render.DrawScreen(e.mode, msg)
    for {
        if e.process() {
            break
        }
    }

    e.render.Clear()
    tm.Flush()
}

func (e *Editor) process() bool {
    event := tm.PollEvent()
    // TODO: mouse support
    // 1. scroll by middle wheel
    // 2. jump cursor by clicking mouse
    if event.Type != tm.EventKey {
        return false
    }
    switch e.mode {
    case EditMode:
        return e.processEditMode(event)
    case FileMode:
        e.processFileMode(event)
    case HelpMode:
        e.processHelpMode(event)
    default:
        panic("unsupported edit mode")
    }
    return false
}

func (e *Editor) processFileMode(event tm.Event) {
    msg := "Trying to open file..."
    switch event.Key {
    case tm.KeyCtrlX:
        e.mode = EditMode
        e.openFileInitEdit()
    case tm.KeyBackspace, tm.KeyBackspace2:
        e.miscBuf.Delete()
    case tm.KeyEnter:
        msg = e.Open(string(e.miscBuf.lines[0].txt))
    default:
        if runewidth.RuneWidth(event.Ch) > 0 {
            e.miscBuf.Insert(event.Ch)
        }
    }
    e.render.SyncCursorToView()
    e.render.DrawScreen(e.mode, msg)
}

func (e *Editor) processHelpMode(event tm.Event) {
    if event.Key == tm.KeyCtrlX {
        e.mode = EditMode
    }
    e.render.DrawScreen(e.mode, "")
}

func (e *Editor) processEditMode(event tm.Event) bool {
    switch event.Key {
    case tm.KeyCtrlX:
        tm.Flush()
        return true
    case tm.KeyCtrlR:
        e.mode = FileMode
        e.initFileMode()
    case tm.KeyCtrlSlash:
        e.mode = HelpMode
    case tm.KeyCtrlA:
        e.cursor.y = 0
    case tm.KeyCtrlE:
        e.moveCursorToEOL()
    case tm.KeyCtrlV:
        e.render.moveCursorToNextHalfScreen()
    case tm.KeyCtrlZ:
        e.render.moveCursorToPrevHalfScreen()
    case tm.KeyArrowUp, tm.KeyArrowDown, tm.KeyArrowLeft, tm.KeyArrowRight:
        e.moveCursor(event.Key)
        break
    case tm.KeyEnter:
        e.buf.NewLine()
        break
    case tm.KeyBackspace, tm.KeyBackspace2:
        e.buf.Delete()
        break
    case tm.KeySpace:
        e.buf.Insert(rune(' '))
    case tm.KeyTab:
        e.buf.InsertTab() 
    default:
        if runewidth.RuneWidth(event.Ch) > 0 {
            e.buf.Insert(event.Ch)
        }
    }
    e.render.SyncCursorToView()
    e.render.DrawScreen(e.mode, "")
    return false
}

func (e *Editor) moveCursor(keyType tm.Key) {
    e.render.MoveCursor(keyType)
}

func (e *Editor) moveCursorToEOL() {
    if len(e.buf.lines) == 0 {
        e.cursor.y = 0
    }
    e.cursor.y = len(e.buf.lines[e.cursor.x].txt)
}

func (e *Editor) Open(path string) string {
    err := e.openfile(path)
    if err != nil {
        msg := fmt.Sprintf("unable to open file: %s", err)
        return msg
    }
    e.mode = EditMode
    e.openFileInitEdit()
    return "@"
}

func (e *Editor) openfile(path string) (error) {
    homeDir, err := os.UserHomeDir()
    fullPath := path
    if path == "~" {
        fullPath = homeDir
    } else if strings.HasPrefix(path, "~/") {
        fullPath = filepath.Join(homeDir, path[2:])
    }
    if err != nil {
        return err
    }
    f, err := os.Open(fullPath)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    e.buf = &Buffer{}
    e.cursor = &Pos{0, 0}
    e.buf.New(e.cursor)
    for scanner.Scan() {
        e.buf.lines = append(e.buf.lines, line{txt: []rune(scanner.Text())})
    }
    return nil
}

func (e *Editor) openFileInitEdit() {
    e.render.Init(e.buf, e.cursor)
}

func (e *Editor) initFileMode() {
    resetPos(e.miscCursor)
    e.miscBuf.New(e.miscCursor)
    e.render.Init(e.miscBuf, e.miscCursor)
}

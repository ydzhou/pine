package ste

import (
	"os"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
)

type Editor struct {
    buf Buffer
    render Render
    cursor *Pos
    mode Mode
}

type Pos struct {
    x, y int
}

func (e *Editor) Init() {
    e.cursor = &Pos{0, 0}
    e.buf = Buffer{
        cursor: e.cursor,
    }
    e.buf.New()
    e.render = Render{
        buf: &e.buf,
        cursor: e.cursor,
        viewCursor: &Pos{0, 0},
        viewAnchor: &Pos{0, 0},
    }
    e.mode = EditMode
}

func (e *Editor) Start() {
    err := tm.Init()
        if err != nil {
            panic(err)
    }
    defer tm.Close()
    
    for {
        e.render.DrawScreen(e.mode)
        if e.process() {
            break
        }
    }

    e.render.Clear()
    tm.Flush()
}

func (e *Editor) process() bool {
    event := tm.PollEvent()
    if event.Type != tm.EventKey {
        return false
    }
    switch event.Key {
    case tm.KeyCtrlX:
        if e.mode == HelpMode {
            e.mode = EditMode
        } else {
            tm.Flush()
            return true
        }
    case tm.KeyCtrlSlash:
        e.mode = HelpMode
    case tm.KeyCtrlA:
        e.cursor.y = 0
    case tm.KeyCtrlE:
        e.moveCursorToEOL()
    case tm.KeyArrowUp, tm.KeyArrowDown, tm.KeyArrowLeft, tm.KeyArrowRight:
        e.moveCursor(event.Key)
        break
    case tm.KeyEnter:
        e.buf.NewLine(e.cursor)
        break
    case tm.KeyBackspace, tm.KeyBackspace2:
        e.buf.Delete(e.cursor)
        break
    case tm.KeySpace:
        e.buf.Insert(e.cursor, rune(' '))
    case tm.KeyTab:
        e.buf.InsertTab() 
    default:
        if runewidth.RuneWidth(event.Ch) > 0 {
            e.buf.lastModifiedCh = "NA"
            e.buf.Insert(e.cursor, event.Ch)
        }
    }
    e.render.SyncCursorToView()
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

func (e *Editor) Open(fileName string) {
    rd, err := os.Open(fileName)
    if err != nil {
        panic(err)
    }
    defer rd.Close()

    
}

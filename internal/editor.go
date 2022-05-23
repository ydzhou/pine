package pine

import (
	"fmt"

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
    tm.SetInputMode(tm.InputMouse)
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
    if event.Type != tm.EventKey && event.Type != tm.EventMouse {
        return false
    }
    isExit := false
    msg := ""
    switch e.mode {
    case EditMode:
        isExit = e.processEditMode(event)
    case FileOpenMode:
        msg = e.processOpenFileMode(event)
    case FileSaveMode:
        msg = e.processSaveFileMode(event)
    case HelpMode:
        e.processHelpMode(event)
    default:
        panic("unsupported edit mode")
    }
    e.render.DrawScreen(e.mode, msg)
    return isExit
}

func (e *Editor) processOpenFileMode(event tm.Event) string {
    msg := "trying to open file..."
    switch event.Key {
    case tm.KeyCtrlX:
        e.mode = EditMode
        e.fileModeToEdit()
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
    return msg
}

func (e *Editor) processSaveFileMode(event tm.Event) string {
    msg := "trying to save file..."
    switch event.Key {
    case tm.KeyCtrlX:
        e.mode = EditMode
        e.fileModeToEdit()
    case tm.KeyBackspace, tm.KeyBackspace2:
        e.miscBuf.Delete()
    case tm.KeyEnter:
        msg = e.Save(string(e.miscBuf.lines[0].txt))
    default:
        if runewidth.RuneWidth(event.Ch) > 0 {
            e.miscBuf.Insert(event.Ch)
        }
    }
    e.render.SyncCursorToView()
    return msg
}

func (e *Editor) processHelpMode(event tm.Event) {
    if event.Key == tm.KeyCtrlX {
        e.mode = EditMode
    }
}

func (e *Editor) processEditMode(event tm.Event) bool {
    isExit := false
    if event.Type == tm.EventKey {
        isExit = e.processEditModeKey(event)
    } else {
        e.processEditModeMouse(event)
    }
    e.render.SyncCursorToView()
    e.render.DrawScreen(e.mode, "")
    return isExit
}

func (e *Editor) processEditModeMouse(event tm.Event) {
    e.render.mouseCursor.x = event.MouseY
    e.render.mouseCursor.y = event.MouseX
    switch event.Key {
    case tm.MouseLeft:
        e.buf.lastModifiedCh = "+ML"
        e.moveCursorByMouse(Pos{
            x: event.MouseY - 1,
            y: event.MouseX,
        })
    case tm.MouseWheelUp:
        e.moveCursor(tm.KeyArrowUp)
    case tm.MouseWheelDown:
        e.moveCursor(tm.KeyArrowDown)
    default:
        e.buf.lastModifiedCh = "NA"
    }
}

func (e *Editor) processEditModeKey(event tm.Event) bool {
    switch event.Key {
    case tm.KeyCtrlX:
        tm.Flush()
        return true
    case tm.KeyCtrlR:
        e.mode = FileOpenMode
        e.initOpenFileMode()
    case tm.KeyCtrlO:
        e.mode = FileSaveMode
        e.initSaveFileMode()
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
    return false
}

func (e *Editor) moveCursor(keyType tm.Key) {
    e.render.MoveCursor(keyType)
}

func (e *Editor) moveCursorByMouse(tpos Pos) {
    e.render.MoveCursorByMouse(tpos)
}

func (e *Editor) moveCursorToEOL() {
    if len(e.buf.lines) == 0 {
        e.cursor.y = 0
    }
    e.cursor.y = len(e.buf.lines[e.cursor.x].txt)
}

func (e *Editor) Open(path string) string {
    fullPath, err := expandHomeDir(path)
    if err != nil {
        return fmt.Sprintf("unable to open file: %s", err)
    }
    e.buf = &Buffer{}
    e.cursor = &Pos{0, 0}
    err = e.buf.Open(e.cursor, fullPath)
    if err != nil {
        return fmt.Sprintf("unable to open file: %s", err)
    }
    e.mode = EditMode
    e.fileModeToEdit()
    return "file opened successfully"
}

func (e *Editor) Save(path string) string {
    fullPath, err := expandHomeDir(path)
    if err != nil {
        return fmt.Sprintf("unable to save file: %s", err)
    }
    wbyte, err := e.buf.Save(fullPath)
    if err != nil {
        return fmt.Sprintf("unable to save file: %s", err)
    }
    e.mode = EditMode
    e.fileModeToEdit()
    return fmt.Sprintf("file saved %d byte written", wbyte)
}

func (e *Editor) fileModeToEdit() {
    e.render.Init(e.buf, e.cursor)
}

func (e *Editor) initOpenFileMode() {
    resetPos(e.miscCursor)
    e.miscBuf.New(e.miscCursor)
    e.render.Init(e.miscBuf, e.miscCursor)
}

func (e *Editor) initSaveFileMode() {
    resetPos(e.miscCursor)
    e.miscBuf.New(e.miscCursor)
    for _, r := range e.buf.filePath {
        e.miscBuf.Insert(rune(r))
    }
    e.render.Init(e.miscBuf, e.miscCursor)
}

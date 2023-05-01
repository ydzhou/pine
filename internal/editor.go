package pine

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	tm "github.com/nsf/termbox-go"
)

type Editor struct {
	buf        *Buffer
	render     Render
	miscBuf    *Buffer
	cursor     *Pos
	miscCursor *Pos
	mode       Mode
	sett       *Setting
	key        *KeyMapper
	isExit     bool
}

type Pos struct {
	x, y int
}

func (e *Editor) Init(sett *Setting) {
	e.isExit = false
	e.cursor = &Pos{0, 0}
	e.miscCursor = &Pos{0, 0}
	e.miscBuf = &Buffer{}
	e.render = Render{sett: sett}
	e.render.Init(e.buf, e.cursor)
	e.sett = sett
	e.key = &KeyMapper{}
}

func (e *Editor) Start(path string) {
	err := tm.Init()
	tm.SetInputMode(tm.InputMouse)
	if err != nil {
		panic(err)
	}
	defer tm.Close()

	msg := e.Open(path)
	e.render.DrawScreen(e.mode, msg)
	for !e.isExit {
		if e.process() {
			break
		}
	}

	e.render.Clear()
	tm.Flush()
}

func (e *Editor) process() bool {
	event := tm.PollEvent()
	if event.Type != tm.EventKey && event.Type != tm.EventMouse {
		return false
	}
	isExit := false
	msg := ""
	e.key.Map(event)
	switch e.mode {
	case EditMode:
		e.processEditMode(event)
	case FileOpenMode:
		msg = e.processOpenFileMode()
	case FileSaveMode:
		msg = e.processSaveFileMode(e.buf.filePath)
	case HelpMode:
		e.processHelpMode()
	default:
		log.Fatal("unsupported edit mode")
	}
	e.render.DrawScreen(e.mode, msg)
	return isExit
}

func (e *Editor) processOpenFileMode() string {
	msg := "trying to open file..."
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.fileModeToEdit()
	case DeleteChOp:
		e.miscBuf.Delete()
	case InsertEnterOp:
		msg = e.Open(string(e.miscBuf.lines[0].txt))
	case InsertChOp:
		e.miscBuf.Insert(e.key.ch)
	}
	e.render.SyncCursorToView()
	return msg
}

func (e *Editor) processSaveFileMode(filepath string) string {
	msg := "trying to save file..."
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.fileModeToEdit()
	case DeleteChOp:
		e.miscBuf.Delete()
	case InsertEnterOp:
		msg = e.Save(string(e.miscBuf.lines[0].txt))
	case InsertChOp:
		e.miscBuf.Insert(e.key.ch)
	}
	e.render.SyncCursorToView()
	return msg
}

func (e *Editor) processHelpMode() {
	if e.key.op == ExitOp {
		e.mode = EditMode
	}
}

func (e *Editor) processEditMode(event tm.Event) {
	if event.Type == tm.EventKey {
		e.processEditModeKey()
	} else {
		e.processEditModeMouse(event)
	}
	e.render.SyncCursorToView()
	e.render.DrawScreen(e.mode, "")
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
		e.render.moveCursorUp()
	case tm.MouseWheelDown:
		e.render.moveCursorDown()
	default:
		e.buf.lastModifiedCh = "NA"
	}
}

func (e *Editor) processEditModeKey() {
	switch e.key.op {
	case ExitOp:
		e.setExit()
	case OpenFileOp:
		e.mode = FileOpenMode
		e.initOpenFileMode()
	case SaveFileOp:
		e.mode = FileSaveMode
		e.initSaveFileMode()
	case HelpOp:
		e.mode = HelpMode
	case GoToBOLOp:
		e.cursor.y = 0
	case GoToEOLOp:
		e.moveCursorToEOL()
	case NextHalfPageOp:
		e.render.moveCursorToNextHalfScreen()
	case PrevHalfPageOp:
		e.render.moveCursorToPrevHalfScreen()
	case MoveCursorUpOp:
		e.render.moveCursorUp()
	case MoveCursorDownOp:
		e.render.moveCursorDown()
	case MoveCursorLeftOp:
		e.render.moveCursorLeft()
	case MoveCursorRightOp:
		e.render.moveCursorRight()
	case InsertEnterOp:
		e.buf.NewLine()
	case DeleteChOp:
		e.buf.Delete()
	case DeleteLineOp:
		e.buf.DeleteLine()
	case InsertSpaceOp:
		e.buf.Insert(rune(' '))
	case InsertTabOp:
		e.buf.InsertTab()
	case InsertChOp:
		e.buf.Insert(e.key.ch)
	}
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
		return fmt.Sprintf("invalid filepath: %s", err)
	}
	if path == "" {
		fullPath = ""
	}
	e.buf = &Buffer{}
	e.cursor = &Pos{0, 0}
	state := e.buf.New(e.cursor, fullPath)
	msg := fmt.Sprintf("Open file %s", e.buf.filePath)
	if state == NotFound {
		msg = fmt.Sprintf("Create new file %s", e.buf.filePath)
	} else if state == HasError {
		msg = "Fail to open file"
	}
	e.mode = EditMode
	e.fileModeToEdit()
	return msg
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
	e.miscBuf.New(e.miscCursor, "")
	e.render.Init(e.miscBuf, e.miscCursor)
}

func (e *Editor) initSaveFileMode() {
	resetPos(e.miscCursor)
	e.miscBuf.New(e.miscCursor, "")
	e.miscBuf.InsertString(e.buf.filePath)
	e.render.Init(e.miscBuf, e.miscCursor)
}

func (e *Editor) setExit() {
	e.isExit = true
}

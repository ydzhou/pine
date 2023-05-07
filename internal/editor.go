package pine

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	tm "github.com/nsf/termbox-go"
)

type Editor struct {
	bufIdx  int
	bufs    []*Buffer
	render  Render
	miscBuf *Buffer
	mode    Mode
	sett    *Setting
	key     *KeyMapper
	isExit  bool
	msg     string
}

type Pos struct {
	x, y int
}

func (e *Editor) Init(sett *Setting) {
	e.isExit = false
	e.miscBuf = &Buffer{}
	e.render = Render{sett: sett}
	e.render.Init()
	e.sett = sett
	e.key = &KeyMapper{}
	e.bufIdx = DEFAULT_CURR_BUF_INDEX
	e.bufs = []*Buffer{}
}

func (e *Editor) Start(path string) {
	err := tm.Init()
	tm.SetInputMode(tm.InputMouse)
	if err != nil {
		panic(err)
	}
	defer tm.Close()

	e.Open(path, -1)
	e.render.DrawScreen(e.mode, e.getRenderContent())
	for !e.isExit {
		e.process()
	}

	e.render.Clear()
	tm.Flush()
}

func (e *Editor) process() {
	event := tm.PollEvent()
	if event.Type != tm.EventKey && event.Type != tm.EventMouse {
		return
	}
	e.key.Map(event)
	if e.mode == ConfirmExitOp {
		if event.Ch == rune('y') {
			e.isExit = true
		}
		e.mode = EditMode
		e.msg = "Exit cancelled"
	} else if e.mode == ConfirmCloseOp {
		if event.Ch == rune('y') {
			e.bufs = append(e.bufs[:e.bufIdx], e.bufs[e.bufIdx+1:]...)
			e.nextBuffer()
		}
		e.mode = EditMode
		e.msg = ""
	} else {
		switch e.mode {
		case EditMode:
			e.processEditMode(event)
		case FileOpenMode:
			e.processOpenFileMode()
		case FileSaveMode:
			e.processSaveFileMode(e.getBuf().filePath)
		default:
			log.Fatal("unsupported edit mode")
		}
	}
	switch e.mode {
	case FileOpenMode, FileSaveMode:
		e.render.buf = e.miscBuf
	default:
		e.render.buf = e.getBuf()
	}
	e.render.SyncCursorToView()
	e.render.DrawScreen(e.mode, e.getRenderContent())
	return
}

func (e *Editor) processOpenFileMode() {
	e.msg = "trying to open file..."
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.fileModeToEdit()
	case DeleteChOp:
		e.miscBuf.Delete()
	case InsertEnterOp:
		if len(e.miscBuf.lines) > 0 && len(e.miscBuf.lines[0].txt) > 0 {
			e.Open(string(e.miscBuf.lines[0].txt), -1)
		}
	case InsertChOp:
		e.miscBuf.Insert(e.key.ch)
	}
}

func (e *Editor) processSaveFileMode(filepath string) {
	e.msg = "trying to save file..."
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.fileModeToEdit()
	case DeleteChOp:
		e.miscBuf.Delete()
	case InsertEnterOp:
		if len(e.miscBuf.lines) > 0 && len(e.miscBuf.lines[0].txt) > 0 {
			e.Save(string(e.miscBuf.lines[0].txt))
		}
	case InsertChOp:
		e.miscBuf.Insert(e.key.ch)
	}
}

func (e *Editor) processEditMode(event tm.Event) {
	e.msg = ""
	if event.Type == tm.EventKey {
		e.processEditModeKey()
	} else {
		e.processEditModeMouse(event)
	}
}

func (e *Editor) processEditModeMouse(event tm.Event) {
	e.render.mouseCursor.x = event.MouseY
	e.render.mouseCursor.y = event.MouseX
	switch event.Key {
	case tm.MouseLeft:
		e.getBuf().lastModifiedCh = "+ML"
		e.moveCursorByMouse(Pos{
			x: event.MouseY - 1,
			y: event.MouseX,
		})
	case tm.MouseWheelUp:
		e.render.moveCursorUp()
	case tm.MouseWheelDown:
		e.render.moveCursorDown()
	default:
		e.getBuf().lastModifiedCh = "NA"
	}
}

func (e *Editor) processEditModeKey() {
	switch e.key.op {
	case ExitOp:
		e.Exit()
	case OpenFileOp:
		e.toOpenFileMode()
	case SaveFileOp:
		e.toSaveFileMode()
	case CloseFileOp:
		e.Close()
	case HelpOp:
		e.toHelpPage()
	case GoToBOLOp:
		e.getBuf().cursor.y = 0
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
	case NextBufferOp:
		e.nextBuffer()
	case PrevBufferOp:
		e.prevBuffer()
	case CmdOp:
		e.msg = "Cmd Key ^X Pressed"
	}

	if !e.getBuf().readOnly {
		switch e.key.op {
		case InsertEnterOp:
			e.getBuf().NewLine()
		case DeleteChOp:
			e.getBuf().Delete()
		case DeleteLineOp:
			e.getBuf().DeleteLine()
		case InsertSpaceOp:
			e.getBuf().Insert(rune(' '))
		case InsertTabOp:
			e.getBuf().InsertTab()
		case InsertChOp:
			e.getBuf().Insert(e.key.ch)
		}
	}
}

func (e *Editor) moveCursorByMouse(tpos Pos) {
	e.render.MoveCursorByMouse(tpos)
}

func (e *Editor) moveCursorToEOL() {
	if len(e.getBuf().lines) == 0 {
		e.getBuf().cursor.y = 0
	}
	e.getBuf().cursor.y = len(e.getBuf().lines[e.getBuf().cursor.x].txt)
}

// Try to open a given filepath in the target buffer
// If no index is given, open it in the end of buffers
// If filepath is invalid, create a new buffer
func (e *Editor) Open(path string, bufIdx int) {
	fullPath, err := expandHomeDir(path)
	if err != nil {
		e.msg = fmt.Sprintf("Buffer %d: invalid filepath: %s", e.bufIdx, err)
		fullPath = ""
	}
	if path == "" {
		fullPath = ""
	}
	if idx := e.hasFileOpened(fullPath); idx >= 0 {
		e.bufIdx = idx
		e.msg = fmt.Sprintf("Buffer %d: file %s already opened", e.bufIdx, e.getBuf().filePath)
	} else {
		buf := &Buffer{}
		if bufIdx >= 0 && bufIdx < len(e.bufs) {
			e.bufs[bufIdx] = buf
		} else {
			e.bufs = append(e.bufs, buf)
			e.bufIdx = len(e.bufs) - 1
		}
		state := buf.New(fullPath)
		e.msg = fmt.Sprintf("Buffer %d: open file %s", e.bufIdx, buf.filePath)
		if state == NotFound {
			e.msg = fmt.Sprintf("Buffer %d: create new file %s", e.bufIdx, buf.filePath)
		} else if state == HasError {
			e.msg = "Buffer %d: fail to open file"
		}
	}
	e.fileModeToEdit()
}

// Save current buffer to the given filepath
func (e *Editor) Save(path string) {
	fullPath, err := expandHomeDir(path)
	if err != nil {
		e.msg = fmt.Sprintf("unable to save file: %s", err)
	}
	wbyte, err := e.bufs[e.bufIdx].Save(fullPath)
	if err != nil {
		e.msg = fmt.Sprintf("unable to save file: %s", err)
	}
	e.mode = EditMode
	e.fileModeToEdit()
	e.msg = fmt.Sprintf("file saved %d byte written", wbyte)
}

// Close current buffer
func (e *Editor) Close() {
	e.msg = fmt.Sprintf("Closed buffer %s", e.getBuf().filePath)
	// If this is the last buffer, exit the editor
	if len(e.bufs) < 2 {
		e.Exit()
		return
	}
	// Golang automatically handle the case where bufIdx is at the last one.
	if e.getBuf().dirty && !e.getBuf().readOnly {
		e.msg = fmt.Sprintf("Unsaved changes at Buffer %d, Exit? (y) or (n)", e.bufIdx)
		e.mode = ConfirmCloseOp
	} else {
		e.bufs = append(e.bufs[:e.bufIdx], e.bufs[e.bufIdx+1:]...)
		e.nextBuffer()
	}
}

// Exit the editor
func (e *Editor) Exit() {
	modifiedBufIdx := []int{}
	for i, buf := range e.bufs {
		if buf.dirty && !buf.readOnly {
			modifiedBufIdx = append(modifiedBufIdx, i)
		}
	}
	if len(modifiedBufIdx) > 0 {
		e.msg = fmt.Sprintf("Unsaved changes at Buffer")
		for _, bufIdx := range modifiedBufIdx {
			e.msg = e.msg + fmt.Sprintf(" %d,", bufIdx)
		}
		e.msg = e.msg + " Exit? (y) or (n)"
		e.mode = ConfirmExitOp
	} else {
		e.isExit = true
	}
}

func (e *Editor) fileModeToEdit() {
	e.render.buf = e.getBuf()
	e.render.Init()
	e.mode = EditMode
}

func (e *Editor) toOpenFileMode() {
	e.miscBuf.New("")
	e.render.Init()
	e.mode = FileOpenMode
}

func (e *Editor) toSaveFileMode() {
	e.miscBuf.New("")
	e.miscBuf.InsertString(e.getBuf().filePath)
	e.render.Init()
	e.mode = FileSaveMode
}

func (e *Editor) toHelpPage() {
	e.getHelpDoc()
}

// Return the current buffer
func (e *Editor) getBuf() *Buffer {
	return e.bufs[e.bufIdx]
}

func (e *Editor) hasFileOpened(path string) int {
	for idx, buf := range e.bufs {
		if buf.filePath == path {
			return idx
		}
	}
	return -1
}

func (e *Editor) nextBuffer() {
	if e.bufIdx < len(e.bufs)-1 {
		e.bufIdx++
	} else {
		e.bufIdx = 0
	}
	e.msg = fmt.Sprintf("Switch to buffer %d", e.bufIdx)
}

func (e *Editor) prevBuffer() {
	if e.bufIdx > 0 {
		e.bufIdx--
	} else {
		e.bufIdx = len(e.bufs) - 1
	}
	e.msg = fmt.Sprintf("Switch to buffer %d", e.bufIdx)
}

func (e *Editor) getHelpDoc() {
	e.Open(HELP_DOC_PATH, -1)
	helpBuf := e.getBuf()
	if helpBuf.isEmpty() {
		helpBuf.InsertString("CTRL+\\: Exit\tCTRL+X: Exit")
		helpBuf.NewLine()
		helpBuf.InsertString("CTRL+R: Open\tCTRL+O: Save")
		helpBuf.NewLine()
		helpBuf.NewLine()
		helpBuf.InsertString("More help doc in https://github.com/ydzhou/pine.git")
		helpBuf.filePath = "help.txt"
	}
	helpBuf.readOnly = true
}

func (e *Editor) getRenderContent() RenderContent {
	return RenderContent{
		mod:      e.key.mod,
		key:      e.key.key,
		ch:       e.key.ch,
		bufIdx:   e.bufIdx,
		bufDirty: e.getBuf().dirty,
		msg:      e.msg,
	}
}

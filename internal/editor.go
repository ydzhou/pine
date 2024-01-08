package pine

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	tm "github.com/nsf/termbox-go"
)

type Editor struct {
	bufIdx  int
	bufs    []*Buffer
	miscBuf *Buffer
	render  Render
	mode    Mode
	sett    *Setting
	log     *log.Logger
	key     *KeyMapper
	isExit  bool
}

type Pos struct {
	x, y int
}

func (e *Editor) Init(sett *Setting) {
	e.sett = sett
	e.log = e.initLogger()
	e.isExit = false
	e.miscBuf = &Buffer{}
	e.render.Init(sett, e.log)
	e.key = &KeyMapper{}
	e.bufIdx = DEFAULT_CURR_BUF_INDEX
	e.bufs = []*Buffer{}
}

func (e *Editor) initLogger() *log.Logger {
	logger := log.New()
	logger.Level = log.ErrorLevel
	if e.sett.IsDebug {
		logger.Level = log.DebugLevel
	}
	logFilepath := "~/.pe.log"
	logFilepath, err := expandHomeDir(logFilepath)
	if err != nil {
		logger.Fatalf("failed to setup logger: %v", err)
	}
	f, err := os.OpenFile(logFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.SetOutput(os.Stdout)
		logger.Fatal(err)
	}
	logger.SetOutput(f)
	logger.Info("Logger setup successfully")
	return logger
}

func (e *Editor) Start(path string) {
	err := tm.Init()
	tm.SetInputMode(tm.InputMouse)
	if err != nil {
		panic(err)
	}
	defer tm.Close()

	e.Open(path, -1)
	e.renderAll()
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
		if event.Key == tm.KeyCtrlX {
			e.isExit = true
			return
		}
		e.mode = EditMode
		e.setMsg("Exit cancelled")
	} else if e.mode == ConfirmCloseOp {
		if event.Ch == rune('y') {
			e.bufs = append(e.bufs[:e.bufIdx], e.bufs[e.bufIdx+1:]...)
			e.nextBuffer()
		}
		e.mode = EditMode
		e.setMsg("")
	} else {
		e.identifyFileMode()
		switch e.mode {
		case EditMode:
			e.processEditMode(event)
		case FileOpenMode:
			e.processOpenFileMode()
		case FileSaveMode:
			e.processSaveFileMode(e.getBuf().filePath)
		case DirMode:
			e.processDirMode()
		default:
			e.log.Fatal("unsupported edit mode")
		}
	}
	if len(e.bufs) < 1 {
		e.isExit = true
		return
	}
	e.identifyFileMode()
	e.renderAll()
}

func (e *Editor) identifyFileMode() {
	if e.isExit || (e.mode != EditMode && e.mode != DirMode) {
		return
	}
	if e.getBuf().isDir {
		e.mode = DirMode
	} else {
		e.mode = EditMode
	}
}

func (e *Editor) processOpenFileMode() {
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.setMsg("Open file cancelled")
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
	switch e.key.op {
	case ExitOp:
		e.mode = EditMode
		e.setMsg("Save file cancelled")
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

func (e *Editor) processDirMode() {
	switch e.key.op {
	case ExitOp:
		e.Exit()
	case OpenFileOp:
		e.toOpenFileMode()
	case CloseFileOp:
		e.Close()
	case HelpOp:
		e.toHelpPage()
	case NextHalfPageOp:
		e.render.bufRender.moveCursorToNextHalfScreen(e.getBuf())
	case PrevHalfPageOp:
		e.render.bufRender.moveCursorToPrevHalfScreen(e.getBuf())
	case MoveCursorUpOp, MoveCursorDownOp, MoveCursorLeftOp, MoveCursorRightOp:
		e.render.MoveCursor(e.mode, e.getBuf(), e.key.op)
	case NextBufferOp:
		e.nextBuffer()
	case PrevBufferOp:
		e.prevBuffer()
	case InsertEnterOp:
		e.openDir(e.bufIdx)
	}
	switch e.key.ch {
	case 't':
		e.openDir(-1)
	case 'q':
		e.Close()
	}
}

func (e *Editor) processEditMode(event tm.Event) {
	e.setMsg("")
	if event.Type == tm.EventKey {
		e.processEditModeKey()
	} else {
		e.processEditModeMouse(event)
	}
}

func (e *Editor) processEditModeMouse(event tm.Event) {
	if e.processEditMouseBuffer(event) {
		return
	}
	switch event.Key {
	case tm.MouseLeft:
		e.getBuf().lastModifiedCh = "+ML"
		e.moveCursorByMouse(Pos{
			x: event.MouseY - 1,
			y: event.MouseX,
		})
	case tm.MouseWheelUp:
		e.render.MoveCursor(e.mode, e.getBuf(), MoveCursorUpOp)
	case tm.MouseWheelDown:
		e.render.MoveCursor(e.mode, e.getBuf(), MoveCursorDownOp)
	}
}

func (e *Editor) processEditMouseBuffer(event tm.Event) bool {
	bufStartPos, bufEndPos := e.render.getBufNamePos(e.getBuf().filePath, e.bufIdx)
	if !isOnArea(Pos{event.MouseY, event.MouseX}, bufStartPos, bufEndPos) {
		return false
	}
	switch event.Key {
	case tm.MouseLeft:
		e.nextBuffer()
	case tm.MouseWheelDown:
		e.nextBuffer()
	case tm.MouseWheelUp:
		e.prevBuffer()
	case tm.MouseRight:
		e.Close()
	}
	return true
}

func (e *Editor) processEditModeKey() {
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
		e.render.bufRender.moveCursorToNextHalfScreen(e.getBuf())
	case PrevHalfPageOp:
		e.render.bufRender.moveCursorToPrevHalfScreen(e.getBuf())
	case MoveCursorUpOp, MoveCursorDownOp, MoveCursorLeftOp, MoveCursorRightOp:
		e.render.MoveCursor(e.mode, e.getBuf(), e.key.op)
	case NextBufferOp:
		e.nextBuffer()
	case PrevBufferOp:
		e.prevBuffer()
	case CmdOp:
		e.setMsg("Cmd Mod (^X) Triggered")
	}
}

func (e *Editor) moveCursorByMouse(tpos Pos) {
	e.render.MoveCursorByMouse(e.getBuf(), tpos)
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
		e.log.Warnf(fmt.Sprintf("buffer %d: invalid filepath: %s", e.bufIdx, err))
		fullPath = ""
	}
	if path == "" {
		fullPath = ""
	}
	if idx := e.hasFileOpened(fullPath); idx >= 0 {
		e.bufIdx = idx
		e.log.Infof(fmt.Sprintf("buffer %d: file %s already opened", e.bufIdx, e.getBuf().filePath))
	} else {
		buf := &Buffer{}
		if bufIdx >= 0 && bufIdx < len(e.bufs) {
			e.bufs[bufIdx] = buf
		} else {
			e.bufs = append(e.bufs, buf)
			e.bufIdx = len(e.bufs) - 1
		}
		state := buf.New(fullPath, e.log)
		if state != Success {
			e.log.Warnf(fmt.Sprintf("buffer %d: fail to open file, %s with state %d", e.bufIdx, e.getBuf().filePath, state))
		}
		if state == IsDir {
			e.mode = DirMode
		} else {
			e.mode = EditMode
		}
	}
	e.setMsg(fmt.Sprintf("buffer %d: opened %s", e.bufIdx, e.getBuf().filePath))
}

// Save current buffer to the given filepath
func (e *Editor) Save(path string) {
	fullPath, err := expandHomeDir(path)
	if err != nil {
		log.Errorf("unable to save file %s: %v", path, err)
		e.setMsg(fmt.Sprintf("Unable to save file: %s", err))
	}
	wbyte, err := e.bufs[e.bufIdx].Save(fullPath)
	if err != nil {
		log.Errorf("Unable to save file %s: %v", path, err)
		e.setMsg(fmt.Sprintf("Unable to save file: %s", err))
	}
	e.mode = EditMode
	e.setMsg(fmt.Sprintf("File saved %d byte written", wbyte))
}

// Close current buffer
func (e *Editor) Close() {
	idx := e.bufIdx
	if !e.getBuf().readOnly && e.getBuf().dirty {
		e.setMsg(unsavedBufferMsg([]int{idx}))
		e.mode = ConfirmCloseOp
		return
	}
	e.bufs = append(e.bufs[:idx], e.bufs[idx+1:]...)
	e.nextBuffer()
}

// Exit the editor
func (e *Editor) Exit() {
	unsavedBufIdxs := []int{}
	for i, buf := range e.bufs {
		if buf.readOnly || !buf.dirty {
			continue
		}
		unsavedBufIdxs = append(unsavedBufIdxs, i)
	}
	if len(unsavedBufIdxs) > 0 {
		e.setMsg(unsavedBufferMsg(unsavedBufIdxs))
		e.mode = ConfirmExitOp
		return
	}
	e.isExit = true
}

func (e *Editor) toOpenFileMode() {
	e.miscBuf.New("", e.log)
	e.miscBuf.InsertString(e.getBuf().filePath)
	e.render.miscBufRender.Reset()
	e.mode = FileOpenMode
}

func (e *Editor) toSaveFileMode() {
	e.miscBuf.New("", e.log)
	e.miscBuf.InsertString(e.getBuf().filePath)
	e.render.miscBufRender.Reset()
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
	e.setMsg(fmt.Sprintf("Switch to buffer %d", e.bufIdx))
}

func (e *Editor) prevBuffer() {
	if e.bufIdx > 0 {
		e.bufIdx--
	} else {
		e.bufIdx = len(e.bufs) - 1
	}
	e.setMsg(fmt.Sprintf("Switch to buffer %d", e.bufIdx))
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
		buf:      e.getBuf(),
		miscBuf:  e.miscBuf,
		mode:     e.mode,
		mod:      e.key.mod,
		key:      e.key.key,
		ch:       e.key.ch,
		bufIdx:   e.bufIdx,
		bufDirty: e.getBuf().dirty,
	}
}

func (e *Editor) renderAll() {
	e.render.Draw(e.getRenderContent())
}

func (e *Editor) setMsg(msg string) {
	e.miscBuf.New("", e.log)
	e.miscBuf.InsertString(msg)
}

func (e *Editor) openDir(idx int) {
	path := e.getBuf().getCurrDirPath()
	e.Open(path, idx)
}

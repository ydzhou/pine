package pine

const (
	VERSION  = "0.2.3 alpha"
	TABWIDTH = 8
)

type Mode int64

const (
	DEFAULT_BUFFERNAME     = "untitled"
	DEFAULT_CURR_BUF_INDEX = 0
	HELP_DOC_PATH          = "/usr/share/doc/pe/help.txt"

	// Rendering offset for buffer content
	BUFFER_CONTENT_START_OFFSET = 1
	BUFFER_CONTENT_END_OFFSET   = 2
)

const (
	EditMode Mode = iota
	FileOpenMode
	FileSaveMode
	WelcomeMode
	ConfirmExitOp
	ConfirmCloseOp
)

type FileOpMode int64

const (
	OpenOp FileOpMode = iota
	SaveOp
)

type FileOpenState int64

const (
	Success FileOpenState = iota
	HasError
	NotFound
)

type KeyOps int64

const (
	NoOp KeyOps = iota
	// Editor Ops
	ExitOp
	OpenFileOp
	SaveFileOp
	CloseFileOp
	HelpOp
	NextBufferOp
	PrevBufferOp
	// Navigation Ops
	MoveCursorUpOp
	MoveCursorDownOp
	MoveCursorLeftOp
	MoveCursorRightOp
	NextHalfPageOp
	PrevHalfPageOp
	// of the line
	GoToBOLOp
	GoToEOLOp
	// of the document
	GoToBODOp
	GoToEODOp
	// Text Edit Ops
	InsertChOp
	InsertSpaceOp
	InsertTabOp
	InsertEnterOp
	DeleteChOp
	DeleteLineOp
	// Misc
	CmdOp
)

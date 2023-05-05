package pine

const (
	VERSION  = "0.2.2 alpha"
	TABWIDTH = 8
)

type Mode int64

const (
	DEFAULT_BUFFERNAME     = "untitled"
	DEFAULT_CURR_BUF_INDEX = 1
	HELP_DOC_PATH          = "/usr/share/doc/pe/help.txt"
)

const (
	EditMode Mode = iota
	FileOpenMode
	FileSaveMode
	WelcomeMode
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
)

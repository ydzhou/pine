package pine

const (
	VERSION  = "0.2.5 alpha"
	TABWIDTH = 4
)

type Mode int64

const (
	DEFAULT_BUFFERNAME     = "untitled"
	DEFAULT_CURR_BUF_INDEX = 0
	HELP_DOC_PATH          = "/usr/local/share/doc/pe/help.txt"
)

// UI
const (
	HEADLINE_OFFSET                 = 0
	STATUSLINE_OFFSET               = -1
	BUFFER_CONTENT_START_OFFSET     = 1
	BUFFER_DIR_CONTENT_START_OFFSET = 2
	BUFFER_END_OFFSET               = -2
)

const (
	EditMode Mode = iota
	FileOpenMode
	FileSaveMode
	DirMode
	SearchMode
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
	IsDir
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
	// Search
	SearchOp
	SearchNextOp
	SearchPrevOp
	// Misc
	CmdOp
)

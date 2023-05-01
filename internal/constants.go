package pine

const (
	VERSION  = "0.2.1 alpha"
	TABWIDTH = 8
)

type Mode int64

const DEFAULT_BUFFERNAME = "untitled"

const (
	EditMode Mode = iota
	HelpMode
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

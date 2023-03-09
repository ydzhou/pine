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

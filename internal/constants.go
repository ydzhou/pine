package pine

const (
	VERSION  = "0.2 alpha"
	TABWIDTH = 8
)

type Mode int64

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

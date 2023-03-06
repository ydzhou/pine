package main

import (
	"os"

	pine "github.com/ydzhou/pine/internal"
)

func main() {
	editor := pine.Editor{}
	filename := ""
	editor.Init()
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	editor.Start(filename)
}

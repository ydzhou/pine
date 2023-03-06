package main

import (
	"fmt"
	"os"

	pine "github.com/ydzhou/pine/internal"
)

func main() {
	sett, filename := parseInput(os.Args)
	editor := pine.Editor{}
	editor.Init(sett)
	editor.Start(filename)
}

func parseInput(args []string) (*pine.Setting, string) {
	sett := &pine.Setting{}
	filename := ""
	for i, arg := range args {
		if i == 0 {
			continue
		}
		if arg == "--debug" {
			sett := &pine.Setting{}
			sett.IsDebug = true
		} else if arg == "--version" || arg == "-v" {
			printVersion()
			os.Exit(0)
		} else {
			filename = arg
		}
	}
	return sett, filename
}

func printVersion() {
	fmt.Printf("Pine version %s\n\n", pine.VERSION)
}

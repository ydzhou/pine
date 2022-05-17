package main

import (
    "os"
    "github.com/ydzhou/ste/internal"
)

func main() {
    editor := ste.Editor{}
    filename := ""
    editor.Init()
    if len(os.Args) > 1 {
        filename = os.Args[1]
    }
    editor.Start(filename)
}

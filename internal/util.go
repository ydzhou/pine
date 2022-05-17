package ste

import (
    "github.com/mattn/go-runewidth"
)

func runeRenderedWidth(
    index int,
    data rune,
) int {
    if data == rune('\t') {
        return TABWIDTH - index % TABWIDTH
    }
    if data == ' ' {
        return 1
    }
    return runewidth.RuneWidth(data)
}

func resetPos(p *Pos) {
    p.x = 0
    p.y = 0
}

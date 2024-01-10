package pine

import (
	log "github.com/sirupsen/logrus"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
)

type KeyMapper struct {
	op  KeyOps
	mod tm.Modifier
	ch  rune
	key tm.Key
}

func (k *KeyMapper) Map(event tm.Event) {
	isCmd := false
	if k.op == CmdOp {
		isCmd = true
	}
	k.op = mapKey(event, isCmd)
	k.mod = event.Mod
	k.key = event.Key
	if k.op != NoOp {
		k.ch = event.Ch
	}
}

func mapKey(event tm.Event, isCmd bool) KeyOps {
	if event.Type != tm.EventKey && event.Type != tm.EventMouse {
		log.Warnf("detected non key/mouse interaction")
		return NoOp
	}
	if isCmd {
		switch event.Key {
		case tm.KeyCtrlX:
			return ExitOp
		case tm.KeyCtrlG:
			return CancelOp
		}
		switch event.Ch {
		case rune('k'):
			return CloseFileOp
		}
		return NoOp
	}
	if event.Mod == tm.ModAlt {
		switch event.Ch {
		case rune(','):
			return PrevBufferOp
		case rune('.'):
			return NextBufferOp
		}
	}
	switch event.Key {
	case tm.KeyCtrlX:
		return CmdOp
	case tm.KeyCtrlG:
		return CancelOp
	case tm.KeyCtrlR:
		return OpenFileOp
	case tm.KeyCtrlO:
		return SaveFileOp
	case tm.KeyCtrlS:
		return SearchOp
	case tm.KeyCtrlSlash:
		return HelpOp
	case tm.KeyCtrlA:
		return GoToBOLOp
	case tm.KeyCtrlE:
		return GoToEOLOp
	case tm.KeyCtrlV, tm.KeyPgdn:
		return NextHalfPageOp
	case tm.KeyCtrlZ, tm.KeyPgup:
		return PrevHalfPageOp
	case tm.KeyCtrlK:
		return DeleteLineOp
	case tm.KeyArrowUp, tm.KeyCtrlP:
		return MoveCursorUpOp
	case tm.KeyArrowDown, tm.KeyCtrlN:
		return MoveCursorDownOp
	case tm.KeyArrowLeft:
		return MoveCursorLeftOp
	case tm.KeyArrowRight:
		return MoveCursorRightOp
	case tm.KeyEnter:
		return InsertEnterOp
	case tm.KeyBackspace, tm.KeyBackspace2:
		return DeleteChOp
	case tm.KeySpace:
		return InsertSpaceOp
	case tm.KeyTab:
		return InsertTabOp
	case tm.KeyCtrlB:
		return NextBufferOp
	default:
		if runewidth.RuneWidth(event.Ch) > 0 {
			return InsertChOp
		}
	}
	return NoOp
}

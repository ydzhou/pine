package pine

import (
	log "github.com/sirupsen/logrus"

	"github.com/mattn/go-runewidth"
	tm "github.com/nsf/termbox-go"
)

type KeyMapper struct {
	op  KeyOps
	ch  rune
	key tm.Key
}

func (k *KeyMapper) Map(event tm.Event) {
	k.op = mapKey(event)
	if k.op != NoOp {
		k.ch = event.Ch
		k.key = event.Key
	}
}

func mapKey(event tm.Event) KeyOps {
	if event.Type != tm.EventKey && event.Type != tm.EventMouse {
		log.Warnf("detected non key/mouse interaction")
		return NoOp
	}
	switch event.Key {
	case tm.KeyCtrlX:
		tm.Flush()
		return ExitOp
	case tm.KeyCtrlR:
		return OpenFileOp
	case tm.KeyCtrlO:
		return SaveFileOp
	case tm.KeyCtrlSlash:
		return HelpOp
	case tm.KeyCtrlA:
		return GoToBOLOp
	case tm.KeyCtrlE:
		return GoToEOLOp
	case tm.KeyCtrlV:
		return NextHalfPageOp
	case tm.KeyCtrlZ:
		return PrevHalfPageOp
	case tm.KeyArrowUp:
		return MoveCursorUpOp
	case tm.KeyArrowDown:
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
	default:
		if runewidth.RuneWidth(event.Ch) > 0 {
			return InsertChOp
		}
	}
	return NoOp
}

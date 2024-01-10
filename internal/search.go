package pine

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

type Search struct {
	matchedStartPos  []*Pos
	matchedEndPos    []*Pos
	currCandidateIdx int
	log              *log.Logger
}

func (s *Search) Search(target string, buffer *Buffer) {
	s.matchedStartPos = []*Pos{}
	s.matchedEndPos = []*Pos{}
	s.currCandidateIdx = -1

	r, err := regexp.Compile(target)
	if err != nil {
		log.Warnf("failed to search %s: %v", target, err)
		return
	}

	for i := 0; i < len(buffer.lines); i++ {
		loc := r.FindAllStringIndex(string(buffer.lines[i].txt), -1)
		for j := 0; j < len(loc); j++ {
			s.matchedStartPos = append(s.matchedStartPos, &Pos{i, loc[j][0]})
			s.matchedEndPos = append(s.matchedEndPos, &Pos{i, loc[j][1]})
		}
	}
}

func (s *Search) SetBufferHightlight(buffer *Buffer, target string) {
	buffer.hlStartPos, buffer.hlEndPos = s.getHightlight(target)
	buffer.cursor = buffer.hlStartPos
}

func (s *Search) getHightlight(target string) (*Pos, *Pos) {
	return s.matchedStartPos[s.currCandidateIdx], s.matchedEndPos[s.currCandidateIdx]
}

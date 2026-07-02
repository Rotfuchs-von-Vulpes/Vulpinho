package util

import (
	"strings"
	"unicode"
)

type strProcessor struct {
	buff []rune
	idx  int
}

func (s *strProcessor) Mark() int {
	return s.idx
}

func (s *strProcessor) Reset(pos int) {
	s.idx = pos
}

func (s *strProcessor) TestRune(r rune) bool {
	if s.idx > len(s.buff)-1 {
		return false
	}
	rr := s.buff[s.idx]
	if rr == r {
		s.idx += 1
		return true
	}
	return false
}

func (s *strProcessor) TestString(str string) bool {
	pos := s.Mark()
	for _, r := range str {
		if !s.TestRune(r) {
			s.Reset(pos)
			return false
		}
	}
	return true
}

func (s *strProcessor) ConsumeUntil(str string) string {
	b := strings.Builder{}
	for {
		pos := s.Mark()
		if pos > len(s.buff)-1 || s.TestString(str) {
			s.Reset(pos)
			return b.String()
		}
		b.WriteRune(s.buff[pos])
		s.idx += 1
	}
}

func (s *strProcessor) RejectBlankSpace() {
	for {
		rr := s.buff[s.idx]
		if unicode.IsSpace(rr) {
			s.idx += 1
		} else {
			break
		}
	}
}

func NewStrProc(buff string) (g *strProcessor) {
	g = new(strProcessor)
	g.idx = 0
	g.buff = []rune(buff)
	return
}

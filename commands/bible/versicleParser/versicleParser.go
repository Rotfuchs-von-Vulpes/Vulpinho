package versicleParser

import (
	"vulpinho/commands/bible/versicleParser/langKit"
)

type Node struct {
	Typ      string
	Value    string
	Children []Node
}

type Versicle struct {
	scanner langKit.Scanner
}

func GetVersicleParser(text string) Versicle {
	return Versicle{langKit.GetScanner(text)}
}

type Ref struct {
	Refs []BookRef
}

type BookRef struct {
	Book string
	Refs []ChapterRef
}

type ChapterRef struct {
	Chapter string
	Spans   []Span
}

type Span struct {
	Init string
	End  string
}

func (s *Versicle) ref_1() (bool, BookRef) {
	if ok := s.scanner.String(";"); ok {
		if ok := s.__(); ok {
			if ok, bookRef := s.bookRef(); ok {
				if ok := s.__(); ok {
					return true, bookRef
				}
			}
		}
	}
	return false, BookRef{}
}

func (s *Versicle) ref() (bool, Ref) {
	node := Ref{}
	if ok, bookRef := s.bookRef(); ok {
		node.Refs = append(node.Refs, bookRef)
		pos := s.scanner.Mark()
		for {
			if ok, ref_1 := s.ref_1(); ok {
				node.Refs = append(node.Refs, ref_1)
				pos = s.scanner.Mark()
			} else {
				break
			}
		}
		s.scanner.Reset(pos)
		if ok := s.scanner.Expect(0); ok {
			return true, node
		}
	}
	return false, Ref{}
}

func (s *Versicle) bookRef_1() (bool, ChapterRef) {
	if ok := s.scanner.String(";"); ok {
		if ok := s.__(); ok {
			if ok, chapterRef := s.chapterRef(); ok {
				if ok := s.__(); ok {
					return true, chapterRef
				}
			}
		}
	}
	return false, ChapterRef{}
}

func (s *Versicle) bookRef() (bool, BookRef) {
	node := BookRef{}
	if ok, book := s.book(); ok {
		node.Book = book
		if ok := s.__(); ok {
			if ok, chapterRef := s.chapterRef(); ok {
				node.Refs = append(node.Refs, chapterRef)
				pos := s.scanner.Mark()
				for {
					if ok, bookRef_1 := s.bookRef_1(); ok {
						node.Refs = append(node.Refs, bookRef_1)
						pos = s.scanner.Mark()
					} else {
						break
					}
				}
				s.scanner.Reset(pos)
				return true, node
			}
		}
	}
	return false, BookRef{}
}

func (s *Versicle) chapterRef_1_1() (bool, Span) {
	if ok := s.scanner.String("."); ok {
		if ok, span := s.span(); ok {
			return true, span
		}
	}
	return false, Span{}
}

func (s *Versicle) chapterRef_1() (bool, []Span) {
	nodes := []Span{}
	if ok := s.scanner.String(","); ok {
		if ok, span := s.span(); ok {
			nodes = append(nodes, span)
			pos := s.scanner.Mark()
			for {
				if ok, chapterRef_1_1 := s.chapterRef_1_1(); ok {
					nodes = append(nodes, chapterRef_1_1)
					pos = s.scanner.Mark()
				} else {
					break
				}
			}
			s.scanner.Reset(pos)
			return true, nodes
		}
	}
	return false, []Span{}
}

func (s *Versicle) chapterRef() (bool, ChapterRef) {
	node := ChapterRef{}
	pos := s.scanner.Mark()
	if ok, chapter := s.chapter(); ok {
		node.Chapter = chapter
		if ok, chapterRef_1 := s.chapterRef_1(); ok {
			node.Spans = chapterRef_1
			return true, node
		}
	}
	s.scanner.Reset(pos)
	node.Chapter = "noChapter"
	if ok, span := s.span(); ok {
		node.Spans = append(node.Spans, span)
		pos := s.scanner.Mark()
		for {
			if ok, chapterRef_1_1 := s.chapterRef_1_1(); ok {
				node.Spans = append(node.Spans, chapterRef_1_1)
				pos = s.scanner.Mark()
			} else {
				break
			}
		}
		s.scanner.Reset(pos)
		return true, node
	}
	return false, ChapterRef{}
}

func (s *Versicle) span_1() (bool, string) {
	if ok := s.scanner.String("-"); ok {
		if ok, verse := s.verse(); ok {
			return true, verse
		}
	}
	return false, ""
}

func (s *Versicle) span() (bool, Span) {
	node := Span{"", ""}
	if ok, verse := s.verse(); ok {
		node.Init = verse
		node.End = verse
		pos := s.scanner.Mark()
		if ok, span_1 := s.span_1(); ok {
			node.End = span_1
		} else {
			s.scanner.Reset(pos)
		}
		return true, node
	}
	return false, Span{}
}

func (s *Versicle) book() (bool, string) {
	if ok, str := langKit.RunRegex(&s.scanner, "\\a?(-|\\w)+"); ok {
		return true, str
	}
	return false, ""
}

func (s *Versicle) chapter() (bool, string) {
	if ok, str := langKit.RunRegex(&s.scanner, "(0)!\\a\\a*"); ok {
		return true, str
	}
	return false, ""
}

func (s *Versicle) verse() (bool, string) {
	nodes := []Node{}
	if ok, str := langKit.RunRegex(&s.scanner, "(0)!\\a\\a*"); ok {
		nodes = append(nodes, Node{"string", str, []Node{}})
		return true, str
	}
	return false, ""
}

func (s *Versicle) __() bool {
	if ok, _ := langKit.RunRegex(&s.scanner, "( |\\t|\\n|\\r)*"); ok {
		return true
	}
	return false
}

func (s *Versicle) Parse() (bool, Ref) {
	return s.ref()
}

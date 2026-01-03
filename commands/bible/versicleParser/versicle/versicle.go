package versicle

import (
	"github.com/Rotfuchs-von-Vulpes/langKit"
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

func (s *Versicle) ref_1 () Node {
	nodes := []Node{}
	if ok := s.scanner.String(";"); ok {
		if bookRef := s.bookRef(); bookRef.Typ != "" {
			nodes = append(nodes, bookRef)
			return Node{"ref_1", "", nodes}
		}
	}
	return Node{}
}

func (s *Versicle) ref () Node {
	nodes := []Node{}
	if bookRef := s.bookRef(); bookRef.Typ != "" {
		nodes = append(nodes, bookRef)
		pos := s.scanner.Mark()
		for {
			if ref_1 := s.ref_1(); ref_1.Typ != "" {
				nodes = append(nodes, ref_1.Children...)
				pos = s.scanner.Mark()
			} else {
				break
			}
		}
		s.scanner.Reset(pos)
		if ok := s.scanner.Expect(0); ok {
			return Node{"ref", "", nodes}
		}
	}
	return Node{}
}

func (s *Versicle) bookRef_1 () Node {
	nodes := []Node{}
	if ok := s.scanner.String(";"); ok {
		if __ := s.__(); __.Typ != "" {
			nodes = append(nodes, __)
			if chapterRef := s.chapterRef(); chapterRef.Typ != "" {
				nodes = append(nodes, chapterRef)
				if __ := s.__(); __.Typ != "" {
					nodes = append(nodes, __)
					return Node{"bookRef_1", "", nodes}
				}
			}
		}
	}
	return Node{}
}

func (s *Versicle) bookRef () Node {
	nodes := []Node{}
	if book := s.book(); book.Typ != "" {
		nodes = append(nodes, book)
		if __ := s.__(); __.Typ != "" {
			nodes = append(nodes, __)
			if chapterRef := s.chapterRef(); chapterRef.Typ != "" {
				nodes = append(nodes, chapterRef)
				pos := s.scanner.Mark()
				for {
					if bookRef_1 := s.bookRef_1(); bookRef_1.Typ != "" {
						nodes = append(nodes, bookRef_1.Children...)
						pos = s.scanner.Mark()
					} else {
						break
					}
				}
				s.scanner.Reset(pos)
				return Node{"bookRef", "", nodes}
			}
		}
	}
	return Node{}
}

func (s *Versicle) chapterRef_1_1 () Node {
	nodes := []Node{}
	if ok := s.scanner.String("."); ok {
		if span := s.span(); span.Typ != "" {
			nodes = append(nodes, span)
			return Node{"chapterRef_1_1", "", nodes}
		}
	}
	return Node{}
}

func (s *Versicle) chapterRef_1 () Node {
	nodes := []Node{}
	if ok := s.scanner.String(","); ok {
		if span := s.span(); span.Typ != "" {
			nodes = append(nodes, span)
			pos := s.scanner.Mark()
			for {
				if chapterRef_1_1 := s.chapterRef_1_1(); chapterRef_1_1.Typ != "" {
					nodes = append(nodes, chapterRef_1_1.Children...)
					pos = s.scanner.Mark()
				} else {
					break
				}
			}
			s.scanner.Reset(pos)
			return Node{"chapterRef_1", "", nodes}
		}
	}
	return Node{}
}

func (s *Versicle) chapterRef () Node {
	nodes := []Node{}
	if chapter := s.chapter(); chapter.Typ != "" {
		nodes = append(nodes, chapter)
		pos := s.scanner.Mark()
		if chapterRef_1 := s.chapterRef_1(); chapterRef_1.Typ != "" {
			nodes = append(nodes, chapterRef_1.Children...)
		} else {
			s.scanner.Reset(pos)
		}
		return Node{"chapterRef", "", nodes}
	}
	return Node{}
}

func (s *Versicle) span_1 () Node {
	nodes := []Node{}
	if ok := s.scanner.String("-"); ok {
		if verse := s.verse(); verse.Typ != "" {
			nodes = append(nodes, verse)
			return Node{"span_1", "", nodes}
		}
	}
	return Node{}
}

func (s *Versicle) span () Node {
	nodes := []Node{}
	if verse := s.verse(); verse.Typ != "" {
		nodes = append(nodes, verse)
		pos := s.scanner.Mark()
		if span_1 := s.span_1(); span_1.Typ != "" {
			nodes = append(nodes, span_1.Children...)
		} else {
			s.scanner.Reset(pos)
		}
		return Node{"span", "", nodes}
	}
	return Node{}
}

func (s *Versicle) book () Node {
	nodes := []Node{}
	if ok, str := langKit.RunRegex(&s.scanner, "(-|\\w)+"); ok {
		nodes = append(nodes, Node{"string", str, []Node{}})
		return Node{"book", "", nodes}
	}
	return Node{}
}

func (s *Versicle) chapter () Node {
	nodes := []Node{}
	if ok, str := langKit.RunRegex(&s.scanner, "(0)!\\a\\a*"); ok {
		nodes = append(nodes, Node{"string", str, []Node{}})
		return Node{"chapter", "", nodes}
	}
	return Node{}
}

func (s *Versicle) verse () Node {
	nodes := []Node{}
	if ok, str := langKit.RunRegex(&s.scanner, "(0)!\\a\\a*"); ok {
		nodes = append(nodes, Node{"string", str, []Node{}})
		return Node{"verse", "", nodes}
	}
	return Node{}
}

func (s *Versicle) __ () Node {
	nodes := []Node{}
	if ok, str := langKit.RunRegex(&s.scanner, "( |\\t|\\n|\\r)*"); ok {
		nodes = append(nodes, Node{"string", str, []Node{}})
		return Node{"__", "", nodes}
	}
	return Node{}
}

func (s *Versicle) Parse() Node {
	return s.ref()
}
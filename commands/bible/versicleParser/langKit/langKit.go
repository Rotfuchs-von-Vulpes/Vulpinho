package langKit

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Node struct {
	Typ      string
	Value    string
	Children []Node
}

type Regex struct {
	parser Scanner
}

func GetRegexParser(text string) Regex {
	return Regex{GetScanner(text)}
}

func (s *Regex) regex() Node {
	nodes := []Node{}
	if capture := s.capture(); capture.Typ != "" {
		nodes = append(nodes, capture)
		if ok := s.parser.Expect(0); ok {
			return Node{"regex", "", nodes}
		}
	}
	return Node{}
}

func (s *Regex) capture_1() Node {
	nodes := []Node{}
	if ok := s.parser.String("|"); ok {
		if group := s.group(); group.Typ != "" {
			nodes = append(nodes, group)
			return Node{"capture_1", "", nodes}
		}
	}
	return Node{}
}

func (s *Regex) capture() Node {
	nodes := []Node{}
	if group := s.group(); group.Typ != "" {
		nodes = append(nodes, group)
		pos := s.parser.Mark()
		for {
			if capture_1 := s.capture_1(); capture_1.Typ != "" {
				nodes = append(nodes, capture_1.Children...)
				pos = s.parser.Mark()
			} else {
				break
			}
		}
		s.parser.Reset(pos)
		return Node{"capture", "", nodes}
	}
	return Node{}
}

func (s *Regex) group() Node {
	nodes := []Node{}
	if mode := s.mode(); mode.Typ != "" {
		nodes = append(nodes, mode)
		pos := s.parser.Mark()
		for {
			if mode := s.mode(); mode.Typ != "" {
				nodes = append(nodes, mode)
				pos = s.parser.Mark()
			} else {
				break
			}
		}
		s.parser.Reset(pos)
		return Node{"group", "", nodes}
	}
	return Node{}
}

func (s *Regex) mode_1() Node {
	nodes := []Node{}
	pos := s.parser.Mark()
	if ok := s.parser.String("?"); ok {
		nodes = append(nodes, Node{"string", "?", []Node{}})
		return Node{"mode_1", "", nodes}
	}
	s.parser.Reset(pos)
	if ok := s.parser.String("!"); ok {
		nodes = append(nodes, Node{"string", "!", []Node{}})
		return Node{"mode_1", "", nodes}
	}
	s.parser.Reset(pos)
	if ok := s.parser.String("*"); ok {
		nodes = append(nodes, Node{"string", "*", []Node{}})
		return Node{"mode_1", "", nodes}
	}
	s.parser.Reset(pos)
	if ok := s.parser.String("+"); ok {
		nodes = append(nodes, Node{"string", "+", []Node{}})
		return Node{"mode_1", "", nodes}
	}
	s.parser.Reset(pos)
	return Node{}
}

func (s *Regex) mode() Node {
	nodes := []Node{}
	if atom := s.atom(); atom.Typ != "" {
		nodes = append(nodes, atom)
		if mode_1 := s.mode_1(); mode_1.Typ != "" {
			nodes = append(nodes, mode_1.Children...)
			return Node{"mode", "", nodes}
		}
		return Node{"mode", "", nodes}
	}
	return Node{}
}

func (s *Regex) atom() Node {
	nodes := []Node{}
	pos := s.parser.Mark()
	if char := s.char(); char.Typ != "" {
		nodes = append(nodes, char)
		return Node{"atom", "", nodes}
	}
	s.parser.Reset(pos)
	if ok := s.parser.String("("); ok {
		if capture := s.capture(); capture.Typ != "" {
			nodes = append(nodes, capture)
			if ok := s.parser.String(")"); ok {
				return Node{"atom", "", nodes}
			}
		}
	}
	s.parser.Reset(pos)
	return Node{}
}

func (s *Regex) char() Node {
	nodes := []Node{}
	pos := s.parser.Mark()
	if ok := s.parser.String("\\"); ok {
		if ok, r := s.parser.Rune(); ok {
			if r == '(' || r == ')' || r == '[' || r == ']' || r == '+' || r == '*' || r == '?' || r == '!' || r == '|' || r == '.' {
				nodes = append(nodes, Node{"rune", string(r), []Node{}})
			} else {
				nodes = append(nodes, Node{"meta", string(r), []Node{}})
			}
			return Node{"char", "", nodes}
		}
	}
	s.parser.Reset(pos)
	if ok := s.parser.String("."); ok {
		nodes = append(nodes, Node{"meta", ".", []Node{}})
		return Node{"char", "", nodes}
	}
	s.parser.Reset(pos)
	if ok, r := s.parser.Rune(); ok {
		if r == '(' || r == ')' || r == '+' || r == '*' || r == '?' || r == '!' || r == '|' {
			return Node{}
		}
		nodes = append(nodes, Node{"rune", string(r), []Node{}})
		return Node{"char", "", nodes}
	}
	s.parser.Reset(pos)
	return Node{}
}

func (s *Regex) Parse() Node {
	return s.regex()
}

type StateIn struct {
	ID    int
	Typ   string
	Value rune
}

type State struct {
	id   int
	next []StateIn
}

type Stack struct {
	states  []State
	count   int
	dontAdd bool
}

func getUnexpectedTypeError(want string, get string) string {
	return fmt.Sprintf("This is what you want: %s, this is what you get: %s", want, get)
}

func (s *Stack) run(run Node) {
	if run.Typ != "rune" {
		panic(getUnexpectedTypeError("rune", run.Typ))
	}
	if len(run.Children) != 0 {
		panic("No rune children was unexpected")
	}
	if run.Value == "" {
		panic("Empty rune value")
	}
	literal := []rune(run.Value)[0]
	s.states[s.count].next = append(s.states[s.count].next, StateIn{s.count + 1, "rune", literal})
}

func (s *Stack) meta(meta Node) {
	if meta.Typ != "meta" {
		panic(getUnexpectedTypeError("meta", meta.Typ))
	}
	if len(meta.Children) != 0 {
		panic("No meta children was unexpected")
	}
	if meta.Value == "" {
		panic("Empty meta value")
	}
	literal := []rune(meta.Value)[0]
	s.states[s.count].next = append(s.states[s.count].next, StateIn{s.count + 1, "meta", literal})
}

func (s *Stack) char(char Node) {
	if char.Typ != "char" {
		panic(getUnexpectedTypeError("char", char.Typ))
	}
	if len(char.Children) == 0 {
		panic("No char children was unexpected")
	}
	if len(char.Children) > 1 {
		panic("Too much char children")
	}
	child := char.Children[0]
	switch child.Typ {
	case "meta":
		s.meta(child)
	case "rune":
		s.run(child)
	default:
		panic("char has illegal child")
	}
}

func (s *Stack) atom(atom Node) {
	if atom.Typ != "atom" {
		panic(getUnexpectedTypeError("atom", atom.Typ))
	}
	if len(atom.Children) == 0 {
		panic("No atom children was unexpected")
	}
	if len(atom.Children) > 2 {
		panic("Too much atom children")
	}
	child := atom.Children[0]
	switch child.Typ {
	case "char":
		s.char(child)
	case "capture":
		s.capture(child)
	default:
		panic("char has illegal child")
	}
}

func (s *Stack) mode(mode Node) {
	if mode.Typ != "mode" {
		panic(getUnexpectedTypeError("mode", mode.Typ))
	}
	if len(mode.Children) == 0 {
		panic("No mode children was unexpected")
	}
	if len(mode.Children) > 2 {
		panic("Too much mode children")
	}
	child := mode.Children[0]
	if child.Typ != "atom" {
		panic("mode has illegal child")
	}
	mark := s.count
	add_not := len(mode.Children) == 2 && mode.Children[1].Value == "!"
	if add_not {
		s.states[s.count].next = append(s.states[mark].next, StateIn{s.count + 1, "not", 0})
		s.count += 1
		s.states = append(s.states, State{s.count, []StateIn{}})
	}
	s.atom(child)
	if len(mode.Children) == 2 {
		repeat := mode.Children[1]
		if repeat.Typ != "string" {
			panic("mode has illegal child")
		}
		switch repeat.Value {
		case "?":
			s.states[mark].next = append(s.states[mark].next, StateIn{s.count + 1, "jump", 0})
		case "!":
			s.count += 1
			s.states = append(s.states, State{s.count, []StateIn{{1, "end", 0}}})
			s.states[mark].next = append(s.states[mark].next, StateIn{s.count + 1, "jump", 0})
		case "*":
			s.count += 1
			s.states = append(s.states, State{s.count, []StateIn{}})
			s.states[s.count].next = append(s.states[s.count].next, StateIn{mark, "jump", 0})
			s.states[mark].next = append(s.states[mark].next, StateIn{s.count + 1, "jump", 0})
		case "+":
			s.count += 1
			s.states = append(s.states, State{s.count, []StateIn{}})
			s.states[s.count].next = append(s.states[s.count].next, StateIn{mark, "jump", 0}, StateIn{s.count + 1, "jump", 0})
		default:
			panic("Illegal mode literal")
		}
	}
}

func (s *Stack) group(group Node) {
	if group.Typ != "group" {
		panic(getUnexpectedTypeError("group", group.Typ))
	}
	if len(group.Children) == 0 {
		panic("No group children was unexpected")
	}
	for i, mode := range group.Children {
		s.mode(mode)
		if i < len(group.Children)-1 {
			s.count += 1
			s.states = append(s.states, State{s.count, []StateIn{}})
		}
	}
}

func (s *Stack) capture(capture Node) {
	if capture.Typ != "capture" {
		panic(getUnexpectedTypeError("capture", capture.Typ))
	}
	if len(capture.Children) == 0 {
		panic("No capture children was unexpected")
	}
	if len(capture.Children) == 1 {
		s.group(capture.Children[0])
	} else {
		mark := s.count
		s.count += 1
		s.states = append(s.states, State{s.count, []StateIn{}})
		endlist := []int{}
		for _, group := range capture.Children {
			s.states[mark].next = append(s.states[mark].next, StateIn{s.count, "jump", 0})
			s.group(group)
			endlist = append(endlist, s.count)
			s.count += 1
			s.states = append(s.states, State{s.count, []StateIn{}})
		}
		for i := range capture.Children {
			end := endlist[i]
			s.states[end].next[0].ID = s.count
		}
		s.states = s.states[0 : len(s.states)-1]
		s.count -= 1
	}
}

func (s *Stack) assemble(regex Node) {
	if regex.Typ != "regex" {
		panic(getUnexpectedTypeError("regex", regex.Typ))
	}
	if len(regex.Children) == 0 {
		panic("No regex children was unexpected")
	}
	if len(regex.Children) > 1 {
		panic("Too much regex children")
	}
	s.capture(regex.Children[0])
	s.count += 1
	s.states = append(s.states, State{s.count, []StateIn{{0, "end", 0}}})
}

func GetRegexStack(regex Node) []State {
	if regex.Typ != "regex" {
		panic(getUnexpectedTypeError("regex", regex.Typ))
	}
	if len(regex.Children) == 0 {
		panic("No regex child has unexpected")
	}
	if len(regex.Children) > 1 {
		panic("Too much regex children")
	}
	final := Stack{[]State{{0, []StateIn{}}}, 0, false}
	final.assemble(regex)
	return final.states
}

func meta(r, meta rune) bool {
	switch meta {
	case '.':
		if r != 0 {
			return true
		}
	case 'w':
		if unicode.IsLetter(r) {
			return true
		}
	case 'a':
		if r >= '0' && r <= '9' {
			return true
		}
	case 'r':
		if r == '\r' {
			return true
		}
	case 'n':
		if r == '\n' {
			return true
		}
	case 't':
		if r == '\t' {
			return true
		}
	default:
		if r == meta {
			return true
		}
	}
	return false
}

func test(stack []State, runes []rune, index, pos int, inside_not bool, fromFront bool) bool {
	s := stack[pos]
	if index > len(runes)-1 {
		return false
	}
	r := runes[index]
	nullNextList := []int{}
	for _, next := range s.next {
		switch next.Typ {
		case "rune":
			if r == next.Value && test(stack, runes, index+1, next.ID, inside_not, false) {
				return true
			}
		case "meta":
			if meta(r, next.Value) && test(stack, runes, index+1, next.ID, inside_not, false) {
				return true
			}
		case "not":
			if test(stack, runes, index, next.ID, true, fromFront) {
				return false
			}
		case "end":
			if inside_not {
				return true
			} else {
				if pos == len(stack)-1 && index == len(runes)-1 {
					return true
				} else {
					return false
				}
			}
		case "jump":
			nullNextList = append(nullNextList, next.ID)
		default:
			panic(next.Typ + " não foi implementado")
		}
	}
	if len(nullNextList) != 0 {
		slices.Sort(nullNextList)
		for _, next := range nullNextList {
			if next < pos && fromFront {
				continue
			}
			if test(stack, runes, index, next, inside_not, fromFront || next < pos) {
				return true
			}
		}
	}
	return false
}

func UseStack(stack []State, str string) bool {
	runes := []rune(str)
	runes = append(runes, 0)
	return test(stack, runes, 0, 0, false, false)
}

var memo map[string][]State

func run(regex, str string) bool {
	if memo == nil {
		memo = make(map[string][]State)
	}
	if s, ok := memo[regex]; ok {
		return UseStack(s, str)
	}
	r := GetRegexParser(regex)
	n := r.Parse()
	if n.Typ == "" {
		panic("failed to parse " + regex + " regex")
	}
	s := GetRegexStack(n)
	memo[regex] = s
	return UseStack(s, str)
}

func RunRegex(s *Scanner, rule string) (bool, string) {
	buffer := strings.Builder{}
	cuttoff := false
	if run(rule, "") {
		cuttoff = true
	}
	pos := s.Mark()
	for {
		if ok, r := s.Rune(); ok {
			if run(rule, buffer.String()+string(r)) {
				cuttoff = true
			} else if cuttoff {
				s.Reset(pos)
				return true, buffer.String()
			}
			pos = s.Mark()
			buffer.WriteRune(r)
		} else {
			if cuttoff {
				s.Reset(pos)
				return true, buffer.String()
			} else {
				return false, ""
			}
		}
	}
}

type Scanner struct {
	text  string
	runes []rune
	pos   int
}

func GetScanner(text string) Scanner {
	runes := []rune{}
	for _, r := range text {
		runes = append(runes, r)
	}
	runes = append(runes, 0)
	return Scanner{text, runes, 0}
}

func (s *Scanner) next() rune {
	if s.pos >= len(s.runes) {
		return 0
	}
	r := s.runes[s.pos]
	s.pos += 1
	return r
}

func (s *Scanner) peekRune() rune {
	if s.pos == len(s.runes) {
		s.runes = append(s.runes, s.next())
	}
	return s.runes[s.pos]
}

func (s *Scanner) getRune() rune {
	r := s.peekRune()
	s.pos = s.pos + 1
	return r
}

func (s *Scanner) Mark() int {
	return s.pos
}

func (s *Scanner) Reset(p int) {
	s.pos = p
}

func (s *Scanner) Expect(arg rune) bool {
	if arg == 0 {
		for {
			if ok := s.Expect(' '); ok {
				continue
			} else if ok := s.Expect('\n'); ok {
				continue
			} else if ok := s.Expect('\r'); ok {
				continue
			}
			break
		}
	}
	r := s.peekRune()
	if r == arg {
		s.pos += 1
		return true
	}
	return false
}

func (s *Scanner) Rune() (bool, rune) {
	r := s.peekRune()
	if r == 0 {
		return false, 0
	} else {
		return true, s.getRune()
	}
}

func (s *Scanner) String(arg string) bool {
	if arg == "" {
		return false
	}
	pos := s.Mark()
	for _, r1 := range arg {
		ok, r2 := s.Rune()
		if ok {
			if r1 != r2 {
				s.Reset(pos)
				return false
			}
		} else {
			s.Reset(pos)
			return false
		}
	}
	return true
}

func (s *Scanner) LowLetter() (bool, rune) {
	r := s.peekRune()
	if r < 'a' || r > 'z' {
		return false, 0
	} else {
		return true, s.getRune()
	}
}

func (s *Scanner) HighLetter() (bool, rune) {
	r := s.peekRune()
	if r < 'A' || r > 'Z' {
		return false, 0
	} else {
		return true, s.getRune()
	}
}

func (s *Scanner) Letter() (bool, rune) {
	if ok, r := s.HighLetter(); !ok {
		if ok, r := s.LowLetter(); ok {
			return true, r
		} else {
			return false, 0
		}
	} else {
		return true, r
	}
}

func (s *Scanner) Num() (bool, rune) {
	r := s.peekRune()
	if (r < '1' || r > '9') && r != '0' {
		return false, 0
	} else {
		return true, s.getRune()
	}
}

func (s *Scanner) Number() (bool, string) {
	name := strings.Builder{}
	no_point := false
	for {
		if ok, r := s.Num(); ok {
			name.WriteRune(r)
		} else if ok := s.Expect('.'); ok && !no_point {
			no_point = true
			name.WriteRune('.')
		} else {
			break
		}
	}
	if name.Len() > 0 {
		return true, name.String()
	} else {
		return false, ""
	}
}

func (s *Scanner) Name() (bool, string) {
	name := strings.Builder{}
	for {
		if ok, r := s.Letter(); ok {
			name.WriteRune(r)
		} else if ok, r := s.Num(); ok {
			name.WriteRune(r)
		} else if ok := s.Expect('_'); ok {
			name.WriteRune('_')
		} else {
			break
		}
	}
	if name.Len() > 0 {
		return true, name.String()
	} else {
		return false, ""
	}
}

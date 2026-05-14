package svg

import (
	"bytes"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

func SvgToPng(code string) (out io.Reader, ok bool, errStr string) {
	process := exec.Command("python", ".\\svgToPng.py", code)
	process.Dir = "resources/svgToPng"
	stdin, err := process.StdinPipe()
	if err != nil {
		slog.Error("Não foi possivel preparar comando ao script Python", "error", err.Error())
		return
	}
	defer stdin.Close()
	png := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	process.Stdout = png
	process.Stderr = errBuff

	if err = process.Run(); err != nil {
		if err.Error() == "exit status 2" {
			final := "Erro de xml:\n```\n" + errBuff.String() + "```"
			return nil, true, final
		}
		if err.Error() == "exit status 2" {
			final := "Erro de svg:\n```\n" + errBuff.String() + "```"
			return nil, true, final
		}
		slog.Error("Erro ao executar script Python", "error", err.Error())
		if errBuff.Len() > 0 {
			slog.Error("Erro interno do script Python", "error", errBuff.String())
		}
		return nil, false, ""
	}

	out = bytes.NewReader(png.Bytes())
	ok = true
	return
}

type grabber struct {
	buff []rune
	idx  int
}

func (s *grabber) mark() int {
	return s.idx
}

func (s *grabber) reset(pos int) {
	s.idx = pos
}

func (s *grabber) testRune(r rune) bool {
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

func (s *grabber) testString(str string) bool {
	pos := s.mark()
	for _, r := range str {
		if !s.testRune(r) {
			s.reset(pos)
			return false
		}
	}
	return true
}

func (s *grabber) consumeUntil(str string) string {
	b := strings.Builder{}
	for {
		pos := s.mark()
		if pos > len(s.buff)-1 || s.testString(str) {
			s.reset(pos)
			return b.String()
		}
		b.WriteRune(s.buff[pos])
		s.idx += 1
	}
}

func newGrabber(buff string) (g *grabber) {
	g = new(grabber)
	g.idx = 0
	g.buff = []rune(buff)
	return
}

func ExtractSvgCodeFromMsg(msg string) string {
	g := newGrabber(msg)
	g.consumeUntil("```")
	if g.testString("```svg") || g.testString("```") {
		if text := g.consumeUntil("```"); len(text) > 0 {
			if g.testString("```") {
				return text
			}
		}
	}
	return ""
}

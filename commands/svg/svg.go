package svg

import (
	"bytes"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"vulpinho/util"
)

func SvgToPng(code string) (out io.Reader, ok bool, errStr string) {
	process := exec.Command("python", ".\\svgToPng.py")
	process.Dir = "resources/svgToPng"
	process.Stdin = strings.NewReader(code)
	png := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	process.Stdout = png
	process.Stderr = errBuff

	if err := process.Run(); err != nil {
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

	out = png
	ok = true
	return
}

func ExtractSvgCodeFromMsg(msg string) string {
	g := util.NewStrProc(msg)
	pos := g.Mark()
	g.ConsumeUntil("```")
	if g.TestString("```svg") || g.TestString("```") {
		g.RejectBlankSpace()
		if text := g.ConsumeUntil("```"); len(text) > 0 {
			if g.TestString("```") {
				return text
			}
		}
	}
	g.Reset(pos)
	return msg
}

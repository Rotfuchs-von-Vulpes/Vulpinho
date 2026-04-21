package chemistry

import (
	"bytes"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

func Init() {

}

func convertSvgToPng(in *bytes.Buffer) (out bytes.Buffer, ok bool) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("C:/Program Files/Inkscape/bin/inkscape.exe", "--pipe", "--export-type=png", "--export-filename=-", "-w", "512", "-h", "512")
	cmd.Stdin = in
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Não foi possivel executar o programa Inkscape", "error", err.Error())
		return
	}

	if stdout.Len() == 0 {
		slog.Error("Não foi coletado dados do programa Inkscape")
		return
	}

	ok = true
	out = stdout

	return
}

func errFilter(input string) string {
	reading := false
	space := false
	newLine := true
	b := strings.Builder{}
	for _, r := range input {
		if r == '[' && newLine {
			reading = false
			newLine = false
		}
		if r == '\n' || r == '\r' {
			newLine = true
		}
		if reading {
			if space {
				space = false
			} else {
				b.WriteRune(r)
			}
		}
		if r == ']' {
			reading = true
			space = true
		}
	}
	return b.String()
}

func Smiles(code string) (out io.Reader, ok bool, errStr string) {
	process := exec.Command("python", ".\\smiles2img.py", code)
	process.Dir = "resources/chemistry"
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
			final := "```\n" + errFilter(errBuff.String()) + "```"
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

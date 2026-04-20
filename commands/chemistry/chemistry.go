package chemistry

import (
	"bytes"
	"io"
	"log/slog"
	"os/exec"
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

func Smiles(code string) (out io.Reader, ok bool) {
	process := exec.Command("smiles2img", code, "-f", "PNG", "-s", "512", "512", "--stdout")
	stdin, err := process.StdinPipe()
	if err != nil {
		slog.Error("Não foi possivel preparar comando ao programa Open Babel", "error", err.Error())
		return
	}
	defer stdin.Close()
	png := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	process.Stdout = png
	process.Stderr = errBuff

	if err = process.Run(); err != nil {
		slog.Error("Erro ao executar programa Open Babel", "error", err.Error())
		if errBuff.Len() > 0 {
			slog.Error("Erro interno do Open Babel", "error", errBuff.String())
		}
		return
	}

	ok = true
	out = bytes.NewReader(png.Bytes())

	return
}

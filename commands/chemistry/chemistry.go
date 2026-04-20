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
	args := []string{"-:" + code, "-osvg", "-xb", "none", "-xB", "white"}
	process := exec.Command("obabel", args...)
	stdin, err := process.StdinPipe()
	if err != nil {
		slog.Error("Erro ao preparar programa Open Babel", "error", err.Error())
		return
	}
	defer stdin.Close()
	svg := new(bytes.Buffer)
	process.Stdout = svg

	if err = process.Run(); err != nil {
		slog.Error("Erro ao executar programa Open Babel", "error", err.Error())
		return
	}

	png, converted := convertSvgToPng(svg)
	if !converted {
		return
	}
	ok = true
	out = bytes.NewReader(png.Bytes())
	return
}

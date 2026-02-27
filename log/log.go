package log

import (
	"log/slog"
	"os"
)

var logFile *os.File

func Init() bool {
	logFile, err := os.Create("log/log.txt")
	if err != nil {
		slog.Error("Inicio do sistema de log falhou.", "erro", err.Error())
		return true
	}
	logger := slog.New(slog.NewMultiHandler(slog.NewTextHandler(os.Stdout, nil), slog.NewJSONHandler(logFile, nil)))
	slog.SetDefault(logger)
	return false
}

func End() {
	logFile.Close()
}

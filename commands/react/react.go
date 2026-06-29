package react

import (
	"encoding/csv"
	"log/slog"
	"os"
	"slices"
	"strings"
	"unicode"
)

var reactMap map[string]int
var allEmojis []string
var AllPrefix []string

func Init() (ok bool) {
	ok = false

	reactMap = make(map[string]int)

	f, err := os.Open("commands/react/react.csv")
	if err != nil {
		slog.Error("Não foi possivel abrir react.csv.", "error", err.Error())
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	data, err := r.ReadAll()
	if err != nil {
		slog.Error("Não foi possivel parsear react.csv.", "error", err.Error())
	}

	for idx, line := range data {
		if idx == 0 {
			continue
		}
		if idx := slices.Index(allEmojis, line[1]); idx >= 0 {
			reactMap[line[0]] = idx
		} else {
			reactMap[line[0]] = len(allEmojis)
			allEmojis = append(allEmojis, line[1])
		}
		if line[2] == "true" {
			if !slices.Contains(AllPrefix, line[1]) {
				AllPrefix = append(AllPrefix, line[1])
			}
			AllPrefix = append(AllPrefix, line[0])
		}
	}

	if len(AllPrefix) == 0 {
		slog.Error("Prefixo não especificado.")
		return
	}

	slog.Info("Dados de palavras para reagir foi lido com sucesso.", "Palavras", len(reactMap), "Prefixos", len(AllPrefix), "Emojis", len(allEmojis))

	ok = true
	return
}

func Detect(text string) (reactions []string) {
	words := strings.Split(text, " ")
	for idx, word := range words {
		if slices.Contains(allEmojis, word) {
			continue
		}
		b := strings.Builder{}
		for _, r := range word {
			if unicode.IsLetter(r) || r == '\'' {
				b.WriteRune(r)
			}
		}
		words[idx] = b.String()
	}
	for _, word := range words {
		if idx, ok := reactMap[word]; ok && !slices.Contains(reactions, allEmojis[idx]) {
			ok = true
			reactions = append(reactions, allEmojis[idx])
		}
	}
	return
}

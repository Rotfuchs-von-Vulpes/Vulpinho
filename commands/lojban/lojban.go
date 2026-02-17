package lojban

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"vulpinho/log"
)

type Dictionary struct {
	XMLName   xml.Name `xml:"dictionary"`
	Text      string   `xml:",chardata"`
	Direction []struct {
		Text  string `xml:",chardata"`
		From  string `xml:"from,attr"`
		To    string `xml:"to,attr"`
		Valsi []struct {
			Text       string `xml:",chardata"`
			Word       string `xml:"word,attr"`
			Type       string `xml:"type,attr"`
			Unofficial string `xml:"unofficial,attr"`
			Selmaho    string `xml:"selmaho"`
			User       struct {
				Text     string `xml:",chardata"`
				Username string `xml:"username"`
				Realname string `xml:"realname"`
			} `xml:"user"`
			Definition   string `xml:"definition"`
			Definitionid string `xml:"definitionid"`
			Score        string `xml:"score"`
			Glossword    []struct {
				Text  string `xml:",chardata"`
				Word  string `xml:"word,attr"`
				Sense string `xml:"sense,attr"`
			} `xml:"glossword"`
			Notes   string `xml:"notes"`
			Keyword []struct {
				Text  string `xml:",chardata"`
				Word  string `xml:"word,attr"`
				Place string `xml:"place,attr"`
				Sense string `xml:"sense,attr"`
			} `xml:"keyword"`
			Rafsi []string `xml:"rafsi"`
		} `xml:"valsi"`
		Nlword []struct {
			Text  string `xml:",chardata"`
			Word  string `xml:"word,attr"`
			Sense string `xml:"sense,attr"`
			Valsi string `xml:"valsi,attr"`
			Place string `xml:"place,attr"`
		} `xml:"nlword"`
	} `xml:"direction"`
}

//go:embed jbovlaste-en.xml
var lojbanDictionaryRaw []byte

var lojbanDictionary Dictionary
var ph_lu map[string]string

var logger *slog.Logger
var dictReadyToRead bool

func LojbanInit() {
	logger = log.Logger

	if err := xml.Unmarshal(lojbanDictionaryRaw, &lojbanDictionary); err != nil {
		logger.Error("Incapaz de parsear XML", "error", err)
		dictReadyToRead = false
		return
	}

	ph_lu = make(map[string]string)

	ph_lu["«"] = ""
	ph_lu["-"] = "."
	ph_lu["»"] = ""
	ph_lu["?"] = ""
	ph_lu[","] = ""
	ph_lu["."] = "ʔ"
	ph_lu[" "] = " "
	ph_lu["ˈ"] = "ˈ"
	ph_lu["a"] = "ɑ"
	ph_lu["e"] = "ɛ"
	ph_lu["i"] = "i"
	ph_lu["o"] = "o"
	ph_lu["u"] = "u"
	ph_lu["y"] = "ə"
	ph_lu["ai"] = "ɑj"
	ph_lu["ei"] = "ɛj"
	ph_lu["oi"] = "oj"
	ph_lu["au"] = "aw"
	ph_lu["ia"] = "jɑ"
	ph_lu["ie"] = "jɛ"
	ph_lu["ii"] = "ji"
	ph_lu["io"] = "jo"
	ph_lu["iu"] = "ju"
	ph_lu["ua"] = "wa"
	ph_lu["ue"] = "wɛ"
	ph_lu["ui"] = "wi"
	ph_lu["uo"] = "wo"
	ph_lu["uu"] = "wu"
	ph_lu["iy"] = "jə"
	ph_lu["uy"] = "wə"
	ph_lu["c"] = "ʃ"
	ph_lu["j"] = "ʒ"
	ph_lu["s"] = "s"
	ph_lu["z"] = "z"
	ph_lu["f"] = "f"
	ph_lu["v"] = "v"
	ph_lu["x"] = "x"
	ph_lu["'"] = "h"
	ph_lu["dj"] = "dʒ"
	ph_lu["tc"] = "tʃ"
	ph_lu["dz"] = "ʣ"
	ph_lu["ts"] = "ʦ"
	ph_lu["r"] = "r"
	ph_lu["n"] = "n"
	ph_lu["m"] = "m"
	ph_lu["l"] = "l"
	ph_lu["b"] = "b"
	ph_lu["d"] = "d"
	ph_lu["g"] = "g"
	ph_lu["k"] = "k"
	ph_lu["p"] = "p"
	ph_lu["t"] = "t"
	dictReadyToRead = true
	logger.Info("Todos os dados sobre Lojban foram carregados com sucesso.")
}

func toIPA(text string) string {
	final := strings.Builder{}
	ignoreNext := false
	for i, r := range text {
		if ignoreNext {
			ignoreNext = false
			continue
		}

		char := string(r)
		pair := ""

		if i < len(text)-1 {
			rr := []rune(text)[i+1]
			pair = string(r) + string(rr)
		}

		symbol, ok := ph_lu[pair]
		if !ok {
			symbol, ok = ph_lu[char]
			if !ok {
				logger.Error("Character indefinido", "char", char)
			}
		} else {
			ignoreNext = true
		}

		if symbol != "" {
			final.WriteString(symbol)
		}
	}
	return final.String()
}

func parseNotes(note string) string {
	note = parseDefinition(note)
	noteBuilder := strings.Builder{}
	for _, r := range note {
		if r != '{' && r != '}' {
			noteBuilder.WriteRune(r)
		}
	}
	return noteBuilder.String()
}

func parseDefinition(def string) string {
	readingArgument := false
	readingSubscript := false
	defBuilder := strings.Builder{}
	for _, r := range def {
		if r == '$' {
			if readingArgument {
				readingArgument = false
				readingSubscript = false
				defBuilder.WriteRune('_')
			} else {
				readingArgument = true
				defBuilder.WriteRune('_')
			}
		} else {
			if readingArgument {
				if r == '_' {
					readingSubscript = true
				}
			}
			if readingSubscript {
				if r == '{' || r == '}' {
					continue
				}
				switch r {
				case '0':
					defBuilder.WriteRune('₀')
				case '1':
					defBuilder.WriteRune('₁')
				case '2':
					defBuilder.WriteRune('₂')
				case '3':
					defBuilder.WriteRune('₃')
				case '4':
					defBuilder.WriteRune('₄')
				case '5':
					defBuilder.WriteRune('₅')
				case '6':
					defBuilder.WriteRune('₆')
				case '7':
					defBuilder.WriteRune('₇')
				case '8':
					defBuilder.WriteRune('₈')
				case '9':
					defBuilder.WriteRune('₉')
				}
			} else if r != '_' && r != '{' && r != '}' {
				defBuilder.WriteRune(r)
			}
		}
	}

	return defBuilder.String()
}

func Sisku(word string) string {
	if dictReadyToRead {
		for _, valsi := range lojbanDictionary.Direction[0].Valsi {
			if valsi.Word == word {
				response := "**" + valsi.Word + "** /" + toIPA(valsi.Word) + "/\n-# " + valsi.Type + "\n" + parseDefinition(valsi.Definition)
				if valsi.Notes != "" {
					response += "\n-# " + parseNotes(valsi.Notes)
				}
				return response
			}
		}
		return "Such lojban word does not occur in my database."
	}
	return "Estou sem o dicionário hoje!"
}

func Facki(text string) []string {
	if dictReadyToRead {
		final := []string{}
		for _, nlword := range lojbanDictionary.Direction[1].Nlword {
			if nlword.Word == text {
				var response string
				var valsi string
				if nlword.Place == "" {
					valsi = nlword.Valsi
				} else {
					valsi = nlword.Valsi + " (" + nlword.Place + ")"
				}
				if nlword.Sense == "" {
					response = "**" + valsi + "**"
				} else {
					response = "**" + valsi + "** [" + nlword.Sense + "]"
				}
				final = append(final, response)
			}
		}
		if len(final) == 0 {
			return []string{"No literal lojban translation word has been found in the database."}
		}
		return final
	}
	return []string{"Estou sem o dicionário hoje!"}
}

func Gerna(text string) string {
	words := strings.Split(text, " ")
	hasArgs := false
	if words[0] == "-m" {
		hasArgs = true
	}

	var args []string
	if hasArgs {
		command := words[1]
		text := strings.Join(words[2:], " ")
		args = []string{"resources/lojban/ilmentufa/run_camxes", "-ckt", "-m", command, text}
	} else {
		args = []string{"resources/lojban/ilmentufa/run_camxes", "-ckt", text}
	}
	process := exec.Command("node", args...)
	stdin, err := process.StdinPipe()
	if err != nil {
		logger.Error("Erro ao preparar programa", "error", err.Error())
	}
	defer stdin.Close()
	buf := new(bytes.Buffer)
	process.Stdout = buf
	process.Stderr = os.Stderr

	if err = process.Start(); err != nil {
		logger.Error("Erro ao executar programa", "error", err.Error())
	}
	process.Wait()

	return buf.String()
}

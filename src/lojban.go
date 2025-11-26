package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
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

var lojbanDictionary Dictionary
var ph_lu map[string]string

func lojbanInit() {
	filePath := "resources/lojban/dictionary/jbovlaste-en.xml"
	f, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("unable to read input file "+filePath, "error", err)
	}

	if err := xml.Unmarshal(f, &lojbanDictionary); err != nil {
		logger.Error("unable to parse xml file", "error", err)
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
				logger.Error("Undefined character: ", "char", char)
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
	defBuinder := strings.Builder{}
	for _, r := range def {
		if r == '$' {
			if readingArgument {
				readingArgument = false
				readingSubscript = false
				defBuinder.WriteRune('_')
			} else {
				readingArgument = true
				defBuinder.WriteRune('_')
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
					defBuinder.WriteRune('₀')
				case '1':
					defBuinder.WriteRune('₁')
				case '2':
					defBuinder.WriteRune('₂')
				case '3':
					defBuinder.WriteRune('₃')
				case '4':
					defBuinder.WriteRune('₄')
				case '5':
					defBuinder.WriteRune('₅')
				case '6':
					defBuinder.WriteRune('₆')
				case '7':
					defBuinder.WriteRune('₇')
				case '8':
					defBuinder.WriteRune('₈')
				case '9':
					defBuinder.WriteRune('₉')
				}
			} else if r != '_' && r != '{' && r != '}' {
				defBuinder.WriteRune(r)
			}
		}
	}

	return defBuinder.String()
}

func sisku(word string) string {
	for _, valsi := range lojbanDictionary.Direction[0].Valsi {
		if valsi.Word == word {
			response := "**" + valsi.Word + "** /" + toIPA(valsi.Word) + "/\n-# " + valsi.Type + "\n" + parseDefinition(valsi.Definition)
			if valsi.Notes != "" {
				response += "\n-# " + parseNotes(valsi.Notes)
			}
			return response
		}
	}
	return "Such lojban word does not occurr in my database."
}

func facki(text string) []string {
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
		return []string{"No literal lojban translation word has founded in the database."}
	}
	return final
}

func gerna(text string) string {
	commands := []rune{'J', 'I', 'M', 'S', 'T', 'C', 'R', 'N', 'G'}
	words := strings.Split(text, " ")
	invalid := false
	for _, r := range words[0] {
		found := slices.Contains(commands, r)
		if !found {
			invalid = true
			break
		}
	}

	var args []string
	if invalid {
		args = []string{"resources/lojban/ilmentufa/run_camxes", text}
	} else {
		command := words[0]
		text, _ = strings.CutPrefix(text, command)
		args = []string{"resources/lojban/ilmentufa/run_camxes", "-m", command, text}
	}
	process := exec.Command("node", args...)
	stdin, err := process.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	defer stdin.Close()
	buf := new(bytes.Buffer)
	process.Stdout = buf
	process.Stderr = os.Stderr

	if err = process.Start(); err != nil {
		fmt.Println("An error occured: ", err)
	}
	process.Wait()

	return buf.String()
}

package lojban

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"log/slog"
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

//go:embed jbovlaste-en.xml
var lojbanDictionaryRaw []byte

var lojbanDictionary Dictionary
var ph_lu map[string]string

var dictReadyToRead bool

func LojbanInit() {
	if err := xml.Unmarshal(lojbanDictionaryRaw, &lojbanDictionary); err != nil {
		slog.Error("Incapaz de parsear XML", "error", err)
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
	ph_lu["a"] = "a"
	ph_lu["e"] = "ɛ"
	ph_lu["i"] = "i"
	ph_lu["o"] = "o"
	ph_lu["u"] = "u"
	ph_lu["y"] = "ə"
	ph_lu["ai"] = "aj"
	ph_lu["ei"] = "ɛj"
	ph_lu["oi"] = "oj"
	ph_lu["au"] = "aw"
	ph_lu["ia"] = "ja"
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
	slog.Info("Todos os dados sobre Lojban foram carregados com sucesso.")
}

func runProcess(args []string) string {
	constantArgs := []string{"resources/lojban/ilmentufa/run_camxes", "-ckt"}
	args = append(constantArgs, args...)
	process := exec.Command("node", args...)
	stdin, err := process.StdinPipe()
	if err != nil {
		slog.Error("Erro ao preparar o programa NodeJS", "error", err.Error())
	}
	defer stdin.Close()
	buf := new(bytes.Buffer)
	process.Stdout = buf
	process.Stderr = os.Stderr

	if err = process.Start(); err != nil {
		slog.Error("Erro ao executar o programa NodeJS", "error", err.Error())
	}
	process.Wait()

	return buf.String()
}

type word struct {
	idx int
	buf string
}

func newWord(text string) (w *word) {
	w = new(word)
	w.idx = 0
	w.buf = text
	return
}

func (s *word) get(count int) (r string) {
	idx := s.idx + count
	if idx > len(s.buf)-1 {
		return
	}
	return string(s.buf[idx])
}

func (s *word) add(count int) {
	s.idx += count
}

func (s *word) remain(count int) *word {
	b := strings.Builder{}
	i := count
	for {
		r := s.get(i)
		if r == "" {
			break
		}
		b.WriteString(r)
		i += 1
	}
	return newWord(b.String())
}

var vowels = []string{"a", "e", "i", "o", "u", "y", "ai", "ei", "au", "oi", "ia", "ie", "ii", "io", "iu", "ua", "ue", "ui", "uo", "uu", "iy", "uy"}
var consonant = []string{"l", "m", "n", "r", "c", "j", "s", "z", "b", "d", "g", "v", "j", "z", "p", "t", "k", "f", "c", "s", "x", "'"}
var sylConsonant = []string{"l", "m", "n", "r"}
var forbidden = []string{"c", "j", "s", "z"}
var preferCoda = []string{"l", "m", "n", "r", "c", "j", "s", "z"}
var forbiddenPair = []string{"cx", "kx", "xc", "xk"}
var voiced = []string{"b", "d", "g", "v", "j", "z"}
var unvoiced = []string{"p", "t", "k", "f", "c", "s", "x"}

func identifyOnset(text *word) string {
	first := func() string {
		r1 := text.get(0)
		r2 := text.get(1)
		if slices.Contains(vowels, r1) {
			return ""
		} else {
			if slices.Contains(vowels, r2) {
				text.add(1)
				return r1
			}
			if slices.Contains(forbiddenPair, r1+r2) {
				text.add(1)
				return r1
			} else if slices.Contains(voiced, r1) && slices.Contains(unvoiced, r2) {
				text.add(1)
				return r1
			} else if slices.Contains(unvoiced, r1) && slices.Contains(voiced, r2) {
				text.add(1)
				return r1
			} else if slices.Contains(sylConsonant, r1) {
				text.add(1)
				return r1
			} else if r1 == r2 {
				text.add(1)
				return r1
			} else if slices.Contains(forbidden, r1) && slices.Contains(forbidden, r2) {
				text.add(1)
				return r1
			}
			text.add(2)
			return r1 + r2
		}
	}
	m := text.idx
	ro := first()
	if ro != "" {
		w := text.remain(0)
		rn := identifyNucleus(w)
		if rn == "" {
			text.idx = m
			return ""
		}
	}
	return ro
}

func identifyNucleus(text *word) string {
	r1 := text.get(0)
	r2 := text.get(1)

	if slices.Contains(vowels, r1+r2) {
		text.add(2)
		return r1 + r2
	} else if slices.Contains(vowels, r1) {
		text.add(1)
		return r1
	} else if slices.Contains(sylConsonant, r1) && !slices.Contains(vowels, r2) {
		text.add(1)
		return r1
	}
	return ""
}

func identifyCoda(text *word) string {
	if slices.Contains(vowels, text.get(0)) {
		return ""
	} else {
		r1 := text.get(0)
		r2 := text.get(1)
		if slices.Contains(vowels, r2) {
			return ""
		}
		if slices.Contains(preferCoda, r1) {
			text.add(1)
			return r1
		}
		w := text.remain(0)
		r := identifyOnset(w)
		if r != "" {
			rr := identifyNucleus(w)
			if !slices.Contains(sylConsonant, rr) {
				return ""
			}
		}
		text.add(1)
		return r1
	}
}

func toSyllables(text string) (final []string) {
	w := newWord(text)
	for {
		onset := identifyOnset(w)
		nucleus := identifyNucleus(w)
		coda := identifyCoda(w)
		if onset+nucleus+coda == "" {
			break
		}
		final = append(final, onset+nucleus+coda)
	}
	return
}

func toIPA(text string) string {
	final := strings.Builder{}
	syllables := toSyllables(text)
	for i, syllable := range syllables {
		ignoreNext := false
		if i == len(syllables)-2 {
			final.WriteRune('ˈ')
		}
		for i, r := range syllable {
			if ignoreNext {
				ignoreNext = false
				continue
			}

			char := string(r)
			pair := ""

			if i < len(syllable)-1 {
				rr := []rune(syllable)[i+1]
				pair = string(r) + string(rr)
			}

			symbol, ok := ph_lu[pair]
			if !ok {
				symbol, ok = ph_lu[char]
				if !ok {
					slog.Error("Character indefinido", "char", char)
				}
			} else {
				ignoreNext = true
			}

			if symbol != "" {
				final.WriteString(symbol)
			}
		}
		if i < len(syllables)-1 && i != len(syllables)-3 {
			final.WriteRune('.')
		}
	}
	return final.String()
}

func formatNotes(note string) string {
	note = formatDefinition(note)
	noteBuilder := strings.Builder{}
	for _, r := range note {
		if r != '{' && r != '}' {
			noteBuilder.WriteRune(r)
		}
	}
	return noteBuilder.String()
}

func formatDefinition(def string) string {
	readingArgument := false
	readingSubscript := false

	type reading int
	const (
		argument reading = iota
		subscript
	)

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
				if r == '=' {
					defBuilder.WriteRune('_')
					defBuilder.WriteRune('=')
					defBuilder.WriteRune('_')
					readingArgument = true
					readingSubscript = false
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
				response := "**" + valsi.Word + "** /" + toIPA(valsi.Word) + "/\n-# " + valsi.Type + "\n" + formatDefinition(valsi.Definition)
				if valsi.Notes != "" {
					response += "\n-# " + formatNotes(valsi.Notes)
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
	var args []string
	parts := strings.Split(text, " ")
	if parts[0] == "-m" {
		args = append(args, "-m")
		args = append(args, parts[1])
		args = append(args, strings.Join(parts[2:], " "))
	} else {
		args = append(args, text)
	}
	return runProcess(args)
}

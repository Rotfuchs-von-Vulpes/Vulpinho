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

func readLojbanDict() {
	filePath := "resources/lojban/dictionary/jbovlaste-en.xml"
	f, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("unable to read input file "+filePath, "error", err)
	}

	if err := xml.Unmarshal(f, &lojbanDictionary); err != nil {
		logger.Error("unable to parse xml file", "error", err)
	}
}

func sisku(word string) string {
	for _, valsi := range lojbanDictionary.Direction[0].Valsi {
		if valsi.Word == word {
			response := "**" + valsi.Word + "** [" + valsi.Type + "]: " + parseDefinition(valsi.Definition) + " " + parseNotes(valsi.Notes)
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
	buf := new(bytes.Buffer) // THIS STORES THE NODEJS OUTPUT
	process.Stdout = buf
	process.Stderr = os.Stderr

	if err = process.Start(); err != nil {
		fmt.Println("An error occured: ", err)
	}
	process.Wait()

	return buf.String()
}

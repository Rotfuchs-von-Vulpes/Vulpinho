package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"
)

func openWh(filePath string) [][]string {
	f1, err := os.Open("resources/warhammer/" + filePath)
	if err != nil {
		logger.Error("Unable to read input file "+filePath, "error", err.Error())
		return [][]string{}
	}
	defer f1.Close()

	csvReader := csv.NewReader(f1)
	csvReader.Comma = '|'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("Unable to parse file as CSV for "+filePath, "error", err.Error())
		return [][]string{}
	}

	records[0][0], _ = strings.CutPrefix(records[0][0], string(rune(65279)))

	return records
}

var abilities [][]string
var data_abilities [][]string
var detachment_abilities [][]string
var data_detachment_abilities [][]string
var detachments [][]string
var data_enhancements [][]string
var enhancements [][]string
var keywords [][]string
var leader [][]string
var models_cost [][]string
var models [][]string
var options [][]string
var data_stratagems [][]string
var stratagems [][]string
var unit_composition [][]string
var data_wargear [][]string
var datasheets [][]string
var factions [][]string

func readWh() {
	abilities = openWh("Abilities.csv")
	data_abilities = openWh("Datasheets_abilities.csv")
	data_detachment_abilities = openWh("Datasheets_detachment_abilities.csv")
	data_enhancements = openWh("Datasheets_enhancements.csv")
	keywords = openWh("Datasheets_keywords.csv")
	leader = openWh("Datasheets_leader.csv")
	models_cost = openWh("Datasheets_models_cost.csv")
	models = openWh("Datasheets_models.csv")
	options = openWh("Datasheets_options.csv")
	data_stratagems = openWh("Datasheets_stratagems.csv")
	unit_composition = openWh("Datasheets_unit_composition.csv")
	data_wargear = openWh("Datasheets_wargear.csv")
	datasheets = openWh("Datasheets.csv")
	detachments = openWh("Detachments.csv")
	detachment_abilities = openWh("Detachment_abilities.csv")
	factions = openWh("Factions.csv")
	stratagems = openWh("Stratagems.csv")
	enhancements = openWh("Enhancements.csv")
	// 	all_data := [][][]string{
	// 		abilities,
	// 		data_abilities,
	// 		data_detachment_abilities,
	// 		data_enhancements,
	// 		keywords,
	// 		leader,
	// 		models_cost,
	// 		models,
	// 		options,
	// 		data_stratagems,
	// 		unit_composition,
	// 		data_wargear,
	// 		datasheets,
	// 		detachments,
	// 		detachment_abilities,
	// 		factions,
	// 		stratagems,
	// 		enhancements,
	// 	}
	// main_loop:
	// 	for i, data := range all_data {
	// 		for j, line := range data {
	// 			for _, str := range line {
	// 				_, err := parseTag(str)
	// 				if !err {
	// 					fmt.Printf("%d° banco %d° linha tem tag inesperada\n", i, j)
	// 					continue main_loop
	// 				}
	// 			}
	// 		}
	// 	}
}

func expected(target string, idx int, r rune) bool {
	runes := []rune(target)
	if idx > len(runes)-1 {
		return false
	}
	return runes[idx] == r
}

func parseTags(str string) string {
	final := strings.Builder{}
	link := strings.Builder{}
	type State int
	const (
		Text State = iota
		Tag
		Between
	)
	type LinkState int
	const (
		None LinkState = iota
		Reading
	)
	s := Text
	ss := None

	idx := 0
	for _, r := range str {
		switch s {
		case Text:
			if r == '<' {
				s = Tag
			} else {
				final.WriteRune(r)
			}
		case Tag:
			if r == '>' {
				idx = 0
				if link.Len() != 0 {
					s = Between
					final.WriteRune('[')
				} else {
					s = Text
				}
			} else {
				if expected("href=\"", idx, r) {
					idx += 1
				} else {
					idx = 0
				}
				if ss == Reading {
					if r == '"' {
						ss = None
					} else {
						link.WriteRune(r)
					}
				}
				if idx == len("href=\"") {
					idx = 0
					ss = Reading
				}
			}
		case Between:
			if r == '<' {
				s = Tag
				if link.Len() != 0 {
					fmt.Fprintf(&final, "](<https://wahapedia.ru/%s>)", link.String())
					link.Reset()
				}
			} else {
				final.WriteRune(r)
			}
		}
	}
	return final.String()
}

func decode(id string) string {
	final := strings.Builder{}
	prefix := true
	for _, r := range id {
		if prefix {
			if r != '0' {
				prefix = false
			}
		}
		if !prefix {
			final.WriteRune(r)
		}
	}
	return final.String()
}

func encode(id string) string {
	count := 9 - len(id)
	if count < 0 {
		return id
	}
	final := strings.Builder{}
	for range count {
		final.WriteRune('0')
	}
	final.WriteString(id)
	return final.String()
}

func searchLineWh(data [][]string, id string) []string {
	fmt.Println("\"" + id + "\"")
	final := []string{}
	first := true
	for _, line := range data {
		if first {
			fmt.Println([]rune(line[0]))
			fmt.Println([]rune(id))
			fmt.Println(line[0] == id)
			first = false
		}
		if line[0] == id {
			list := []string{}
			for _, str := range line {
				list = append(list, parseTags(str))
			}
			final = append(final, strings.Join(list[1:], "; "))
		}
	}
	return final
}

func getKeywords(id string) []string {
	result := []string{}
	for _, line := range keywords {
		if line[0] == id {
			result = append(result, line[1])
		}
	}
	return result
}

func searchDatasheetByKeyword(key string) []string {
	id_list := []string{}
	unitBlackList := []string{}
	for _, line := range keywords {
		if strings.ToLower(line[1]) == key {
			if !slices.Contains(unitBlackList, line[0]) {
				unitBlackList = append(unitBlackList, line[0])
				id_list = append(id_list, line[0])
			}
		}
	}
	return id_list
}

func getUnitNameAndURL(id string) string {
	for _, line := range datasheets {
		if line[0] == id {
			return fmt.Sprintf("[%s](<%s>) (id: %s)", line[1], line[13], decode(id))
		}
	}
	return ""
}

func KeySearch(key string) []string {
	unitsIDs := searchDatasheetByKeyword(strings.ToLower(key))
	final := []string{}
	for _, unit := range unitsIDs {
		final = append(final, getUnitNameAndURL(unit))
	}
	if len(final) == 0 {
		return []string{"unit with that keyword dont exist"}
	}
	return final
}

func getWh(data string, id string) []string {
	id = encode(id)
	switch data {
	case "abilities":
		return searchLineWh(abilities, id)
	case "datasheets_abilities":
		return searchLineWh(data_abilities, id)
	case "datasheets_detachment_abilities":
		return searchLineWh(data_detachment_abilities, id)
	case "datasheets_enhancements":
		return searchLineWh(data_enhancements, id)
	case "datasheets_keywords":
		return getKeywords(id)
	case "datasheets_leader":
		return searchLineWh(leader, id)
	case "datasheets_models_cost":
		return searchLineWh(models_cost, id)
	case "datasheets_models":
		return searchLineWh(models, id)
	case "datasheets_options":
		return searchLineWh(options, id)
	case "datasheets_stratagems":
		return searchLineWh(data_stratagems, id)
	case "datasheets_unit_composition":
		return searchLineWh(unit_composition, id)
	case "datasheets_wargear":
		return searchLineWh(data_wargear, id)
	case "datasheets":
		return searchLineWh(datasheets, id)
	case "detachments":
		return searchLineWh(detachments, id)
	case "detachment_abilities":
		return searchLineWh(detachment_abilities, id)
	case "factions":
		return searchLineWh(factions, id)
	case "stratagems":
		return searchLineWh(stratagems, id)
	case "enhancements":
		return searchLineWh(enhancements, id)
	}
	return []string{}
}

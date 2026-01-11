package wh40k

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"unicode"
	"vulpinho/log"
)

var logger *slog.Logger

func openWh(filePath string) [][]string {
	f1, err := os.Open("resources/warhammer/" + filePath)
	if err != nil {
		logger.Error("Incapaz de ler o arquivo.", "error", err.Error())
		return [][]string{}
	}
	defer f1.Close()

	csvReader := csv.NewReader(f1)
	csvReader.Comma = '|'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("Incapaz de parsear CSV.", "error", err.Error())
		return [][]string{}
	}

	records[0][0], _ = strings.CutPrefix(records[0][0], string(rune(65279)))

	return records
}

var abilities [][]string
var datasheetsAbilities [][]string
var detachmentAbilities [][]string
var datasheetsDetachmentAbilities [][]string
var detachments [][]string
var datasheetsEnhancements [][]string
var enhancements [][]string
var datasheetsKeywords [][]string
var datasheetsLeader [][]string
var datasheetsModelsCost [][]string
var datasheetsModels [][]string
var datasheetsOptions [][]string
var datasheetsStratagems [][]string
var stratagems [][]string
var datasheetsUnitComposition [][]string
var datasheetsWargear [][]string
var datasheets [][]string
var factions [][]string

func ReadWh() {
	logger = log.Logger

	abilities = openWh("Abilities.csv")
	datasheetsAbilities = openWh("Datasheets_abilities.csv")
	datasheetsDetachmentAbilities = openWh("Datasheets_detachment_abilities.csv")
	datasheetsEnhancements = openWh("Datasheets_enhancements.csv")
	datasheetsKeywords = openWh("Datasheets_keywords.csv")
	datasheetsLeader = openWh("Datasheets_leader.csv")
	datasheetsModelsCost = openWh("Datasheets_models_cost.csv")
	datasheetsModels = openWh("Datasheets_models.csv")
	datasheetsOptions = openWh("Datasheets_options.csv")
	datasheetsStratagems = openWh("Datasheets_stratagems.csv")
	datasheetsUnitComposition = openWh("Datasheets_unit_composition.csv")
	datasheetsWargear = openWh("Datasheets_wargear.csv")
	datasheets = openWh("Datasheets.csv")
	detachments = openWh("Detachments.csv")
	detachmentAbilities = openWh("Detachment_abilities.csv")
	factions = openWh("Factions.csv")
	stratagems = openWh("Stratagems.csv")
	enhancements = openWh("Enhancements.csv")
	allData := [][][]string{
		abilities,
		datasheetsAbilities,
		datasheetsDetachmentAbilities,
		datasheetsEnhancements,
		datasheetsKeywords,
		datasheetsLeader,
		datasheetsModelsCost,
		datasheetsModels,
		datasheetsOptions,
		datasheetsStratagems,
		datasheetsUnitComposition,
		datasheetsWargear,
		datasheets,
		detachments,
		detachmentAbilities,
		factions,
		stratagems,
		enhancements,
	}
	allRight := true
	for _, data := range allData {
		if len(data) == 0 {
			allRight = false
			break
		}
	}
	if allRight {
		logger.Info("Todos os dados sobre Warhammer 40k foram carregados com sucesso.")
	}
	// for _, data := range all_data {
	// 	for _, line := range data {
	// 		for _, str := range line {
	// 			parseTags(str)
	// 		}
	// 	}
	// }
}

func expected(target string, idx int, r rune) bool {
	runes := []rune(target)
	if idx > len(runes)-1 {
		return false
	}
	return runes[idx] == r
}

var blacklist2 []string

func push(s []string, e string) []string {
	return append(s, e)
}

func pop(s []string) []string {
	return s[:len(s)-1]
}

func parseTags(str string) (string, bool) {
	debug := false
	tagStack := []string{}
	final := strings.Builder{}
	link := strings.Builder{}
	tag := strings.Builder{}
	type State int
	const (
		Text State = iota
		Tag
	)
	type LinkState int
	const (
		None LinkState = iota
		Reading
	)
	type TagType int
	const (
		Notag TagType = iota
		Head
		Tail
		Both
	)
	mainState := Text
	linkState := None
	readingState := None
	tagState := Notag

	idx := 0
	for _, r := range str {
		switch mainState {
		case Text:
			if r == '<' {
				mainState = Tag
				readingState = Reading
				tagState = Notag
			} else {
				final.WriteRune(r)
			}
		case Tag:
			if tagState == Notag {
				if r == '/' {
					tagState = Tail
				} else if unicode.IsLetter(r) {
					tagState = Head
				}
			}
			if readingState == Reading {
				if unicode.IsLetter(r) {
					tag.WriteRune(r)
				} else if tag.Len() > 0 {
					switch tagState {
					case Head:
						tagStack = push(tagStack, tag.String())
					case Tail:
						if len(tagStack) > 1 {
							last := tagStack[len(tagStack)-1]
							if last != tag.String() {
								if !slices.Contains(blacklist2, last) {
									// fmt.Println(last)
									blacklist2 = append(blacklist2, last)
									debug = true
								}
								tagStack = pop(tagStack)
							}
							tagStack = pop(tagStack)
						}
					}
					tag.Reset()
					tagState = Notag
					readingState = None
				}
			}
			if r == '>' {
				idx = 0
				mainState = Text
				readingState = None
				if link.Len() > 0 {
					final.WriteRune('[')
				}
				if tagState == Tail && link.Len() > 0 {
					fmt.Fprintf(&final, "](<https://wahapedia.ru/%s>)", link.String())
					link.Reset()
				}
			} else if tagState == Head {
				if expected("href=\"", idx, r) {
					idx += 1
				} else {
					idx = 0
				}
				if linkState == Reading {
					if r == '"' {
						linkState = None
					} else {
						link.WriteRune(r)
					}
				}
				if idx == len("href=\"") {
					idx = 0
					linkState = Reading
				}
			}
		}
	}
	return final.String(), debug
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
	final := []string{}
	for _, line := range data {
		if line[0] == id {
			list := []string{}
			for _, str := range line {
				text, _ := parseTags(str)
				list = append(list, text)
			}
			final = append(final, strings.Join(list[1:], "; "))
		}
	}
	return final
}

func getKeywords(id string) []string {
	result := []string{}
	for _, line := range datasheetsKeywords {
		if line[0] == id {
			result = append(result, line[1])
		}
	}
	return result
}

func searchDatasheetByKeyword(key string) []string {
	id_list := []string{}
	unitBlackList := []string{}
	for _, line := range datasheetsKeywords {
		if strings.ToLower(line[1]) == key {
			if !slices.Contains(unitBlackList, line[0]) {
				unitBlackList = append(unitBlackList, line[0])
				id_list = append(id_list, line[0])
			}
		}
	}
	return id_list
}

func filterByKeyword(idList []string, keyword string) []string {
	final := []string{}
loop:
	for _, id := range idList {
		for _, line := range datasheetsKeywords {
			if line[0] == id && line[1] == keyword {
				final = append(final, id)
				continue loop
			}
		}
	}
	return final
}

func getUnitNameAndURL(id string) string {
	for _, line := range datasheets {
		if line[0] == id {
			return fmt.Sprintf("[%s](<%s>) (id: %s)", line[1], line[13], decode(id))
		}
	}
	return ""
}

func KeySearch(keys []string) []string {
	unitsIDs := searchDatasheetByKeyword(strings.ToLower(keys[0]))
	for i, key := range keys {
		if i == 0 {
			continue
		}
		unitsIDs = filterByKeyword(unitsIDs, key)
	}
	final := []string{}
	for _, unit := range unitsIDs {
		final = append(final, getUnitNameAndURL(unit))
	}
	if len(final) == 0 {
		return []string{"Unit with that keyword doesn't exist."}
	}
	return final
}

func GetWh(data string, id string) []string {
	id = encode(id)
	switch data {
	case "abilities":
		return searchLineWh(abilities, id)
	case "datasheets_abilities":
		return searchLineWh(datasheetsAbilities, id)
	case "datasheets_detachment_abilities":
		return searchLineWh(datasheetsDetachmentAbilities, id)
	case "datasheets_enhancements":
		return searchLineWh(datasheetsEnhancements, id)
	case "datasheets_keywords":
		return searchLineWh(datasheetsKeywords, id)
	case "datasheets_leader":
		return searchLineWh(datasheetsLeader, id)
	case "datasheets_models_cost":
		return searchLineWh(datasheetsModelsCost, id)
	case "datasheets_models":
		return searchLineWh(datasheetsModels, id)
	case "datasheets_options":
		return searchLineWh(datasheetsOptions, id)
	case "datasheets_stratagems":
		return searchLineWh(datasheetsStratagems, id)
	case "datasheets_unit_composition":
		return searchLineWh(datasheetsUnitComposition, id)
	case "datasheets_wargear":
		return searchLineWh(datasheetsWargear, id)
	case "datasheets":
		return searchLineWh(datasheets, id)
	case "detachments":
		return searchLineWh(detachments, id)
	case "detachment_abilities":
		return searchLineWh(detachmentAbilities, id)
	case "factions":
		return searchLineWh(factions, id)
	case "stratagems":
		return searchLineWh(stratagems, id)
	case "enhancements":
		return searchLineWh(enhancements, id)
	}
	return []string{}
}

package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

var bible [][]string

func readBible() {
	filePath := "resources/bible/bible.csv"
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath+": ", err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath+": ", err)
	}

	bible = records
}

func versicle(say func(string), raw string) {
	words := strings.Split(raw, " ")

	var versicle_temp []string
	if len(words) == 2 {
		pair_1 := strings.Split(words[1], ",")
		pair_2 := strings.Split(words[1], ":")
		if len(pair_1) == 2 {
			versicle_temp = pair_1
		} else if len(pair_2) == 2 {
			versicle_temp = pair_2
		}

		if len(versicle_temp) == 2 {
			words[1] = versicle_temp[0]
			words = append(words, versicle_temp[1])
		}
	}

	if len(words) == 3 {
		words[1] = strings.Map(func(r rune) rune {
			if r == ',' {
				return -1
			}
			return r
		}, words[1])

		rang := strings.Split(words[2], "-")

		if len(rang) == 1 {
			for _, line := range bible {
				if line[1] == words[0] && line[2] == words[1] && line[3] == words[2] {
					say("**" + line[3] + "**. " + line[4])
					break
				}
			}
		} else if len(rang) == 2 {
			_, ok1 := SnowflakeToUint64(rang[0])
			_, ok2 := SnowflakeToUint64(rang[1])
			if ok1 && ok2 {
				reading := false
				chapter := ""
				text := ""
				for _, line := range bible {
					if line[1] == words[0] && line[2] == words[1] && line[3] == rang[0] {
						chapter = line[2]
						reading = true
					}
					if reading {
						text += "**" + line[3] + "**. " + line[4]
						if line[3] == rang[1] || line[2] != chapter {
							break
						}
						text += "\n"
					}
				}
				say(text)
			}
		}
	}
}

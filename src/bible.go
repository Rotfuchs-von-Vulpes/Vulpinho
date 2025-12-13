package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
)

var bible [][]string

func readBible() {
	filePath := "resources/bible/bible.csv"
	f1, err := os.Open(filePath)
	if err != nil {
		logger.Error("Unable to read input file "+filePath, "error", err.Error())
	}
	defer f1.Close()

	csvReader := csv.NewReader(f1)
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("Unable to parse file as CSV for "+filePath, "error", err.Error())
	}

	bible = records

	f2, err := os.Create("resources/bible/missing.txt")
	if err != nil {
		logger.Error("Can't create missing list file")
	}

	var previous int64 = 0
	for _, line := range bible {
		num, err := strconv.ParseInt(line[3], 10, 32)
		if err == nil {
			if num == 1 {
				previous = 0
			}
			if num-previous > 1 {
				f2.WriteString(line[1] + " " + line[2] + " " + strconv.FormatInt(previous+1, 10) + " não existe\n")
			}
			previous = num
		}
	}
	f2.Close()
}

func versicle(raw string) []string {
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
					return []string{"**" + line[3] + "**. " + line[4]}
				}
			}
		} else if len(rang) == 2 {
			_, ok1 := SnowflakeToUint64(rang[0])
			_, ok2 := SnowflakeToUint64(rang[1])
			if ok1 && ok2 {
				reading := false
				chapter := ""
				text := []string{}
				for _, line := range bible {
					if line[1] == words[0] && line[2] == words[1] && line[3] == rang[0] {
						chapter = line[2]
						reading = true
					}
					if reading {
						if line[3] == rang[1] || line[2] != chapter {
							break
						}
						text = append(text, "**"+line[3]+"**. "+line[4])
					}
				}
				return text
			}
		}
	}
	return []string{}
}

type ChapterCount struct {
	chapter string
	count   int
}

func search() string {
	chapters := []ChapterCount{}
	chapters = append(chapters, ChapterCount{bible[0][2], 0})
	for _, line := range bible {
		last := len(chapters) - 1
		chapter := line[2]
		if chapter != chapters[last].chapter {
			chapters = append(chapters, ChapterCount{chapter, 0})
		}
		last = len(chapters) - 1
		chapters[last].count += 1
	}
	biggest := chapters[0]
	for _, chapter := range chapters {
		if chapter.count > biggest.count {
			biggest = chapter
		}
	}
	return "O maior é o " + biggest.chapter + " com " + strconv.Itoa(biggest.count) + " versículos."
}

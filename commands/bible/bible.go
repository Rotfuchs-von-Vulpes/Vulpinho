package bible

import (
	"encoding/csv"
	"log/slog"
	"os"
	"strconv"
	"vulpinho/commands/bible/versicleParser"
	"vulpinho/log"
)

var logger *slog.Logger

var bible [][]string

func ReadBible() {
	logger = log.Logger
	filePath := "resources/bible/bible.csv"
	f1, err := os.Open(filePath)
	if err != nil {
		log.Logger.Error("Unable to read input file "+filePath, "error", err.Error())
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

func Versicle(raw string) []string {
	p := versicleParser.GetVersicleParser(raw)
	text := []string{}
	if ok, ref := p.Parse(); ok {
		for _, BookRef := range ref.Refs {
			for _, ref := range BookRef.Refs {
				text = append(text, "**"+BookRef.Book+" "+ref.Chapter+"**")
				for _, span := range ref.Spans {
					reading := false
					for _, line := range bible {
						if line[1] == BookRef.Book && line[2] == ref.Chapter && line[3] == span.Init {
							reading = true
						}
						if reading {
							if line[2] != ref.Chapter {
								break
							}
							text = append(text, "**"+line[3]+"**. "+line[4])
							if line[3] == span.End {
								break
							}
						}
					}
				}
			}
		}
	}
	return text
}

type ChapterCount struct {
	chapter string
	count   int
}

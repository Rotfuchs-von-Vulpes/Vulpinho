package bible

import (
	"encoding/csv"
	"log/slog"
	"os"
	"strconv"
	"vulpinho/commands/bible/versicleParser"
	"vulpinho/log"
)

var abbrTab map[string]string
var bible [][]string
var logger *slog.Logger
var bibleIsReady bool

func ReadBible() {
	logger = log.Logger
	filePath := "resources/bible/bible.csv"
	abbrTab = map[string]string{
		"gn":   "genesis",
		"ex":   "exodo",
		"lv":   "levitico",
		"nr":   "numeros",
		"dr":   "deuteronomio",
		"js":   "josue",
		"jz":   "juizes",
		"rt":   "rute",
		"1sm":  "i-samuel",
		"2sm":  "ii-samuel",
		"1rs":  "i-reis",
		"2rs":  "ii-reis",
		"1cr":  "i-cronicas",
		"2cr":  "ii-cronicas",
		"esd":  "esdras",
		"ne":   "neemias",
		"tb":   "tobias",
		"jdt":  "judite",
		"est":  "ester",
		"jó":   "jo",
		"sl":   "salmos",
		"pr":   "proverbios",
		"ecl":  "eclesiastes",
		"ct":   "cantico-dos-canticos",
		"sb":   "sabedoria",
		"eclo": "eclesiastico",
		"is":   "isaias",
		"jr":   "jeremias",
		"lm":   "lamentacoes",
		"br":   "baruc",
		"ez":   "ezequiel",
		"dn":   "daniel",
		"os":   "oseias",
		"jl":   "joel",
		"am":   "amos",
		"ab":   "abdias",
		"jn":   "jonas",
		"mq":   "miqueias",
		"na":   "naum",
		"hab":  "habacuc",
		"sf":   "sofonias",
		"ag":   "ageu",
		"zc":   "zacarias",
		"ml":   "malaquias",
		"1mac": "i-macabeus",
		"2mac": "ii-macabeus",
		"mt":   "sao-mateus",
		"mc":   "sao-marcos",
		"kc":   "sao-lucas",
		"jo":   "sao-joao",
		"at":   "atos-dos-apostolos",
		"rm":   "romanos",
		"1cor": "i-corintios",
		"2cor": "ii-corintios",
		"gl":   "galatas",
		"ef":   "efesios",
		"fl":   "filipenses",
		"cl":   "colossenses",
		"1ts":  "i-tessalonicenses",
		"2ts":  "ii-tessalonicenses",
		"1tm":  "i-timoteo",
		"2tm":  "ii-timoteo",
		"tt":   "tito",
		"fm":   "filemon",
		"hb":   "hebreus",
		"tg":   "sao-tiago",
		"1pd":  "i-sao-pedro",
		"2pd":  "ii-sao-pedro",
		"1jo":  "i-sao-joao",
		"2jo":  "ii-sao-joao",
		"3jo":  "iii-sao-joao",
		"jd":   "sao-judas",
		"ap":   "apocalipse",
	}
	f1, err := os.Open(filePath)
	if err != nil {
		logger.Error("Incapaz de ler o arquivo.", "error", err.Error())
		bibleIsReady = false
		return
	}
	defer f1.Close()

	csvReader := csv.NewReader(f1)
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("Incapaz de parsear CSV no arquivo.", "error", err.Error())
		bibleIsReady = false
		return
	}

	bible = records
	bibleIsReady = true
	logger.Info("A biblia foi carregada com sucesso.")

	f2, missingError := os.Create("resources/bible/missing.txt")
	if missingError != nil {
		logger.Error("Incapaz de criar arquivo com a lista de versiculos faltantes.", "error", missingError.Error())
	} else {
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
}

func Versicle(raw string) []string {
	p := versicleParser.GetVersicleParser(raw)
	text := []string{}
	if ok, ref := p.Parse(); ok {
		if bibleIsReady {
			for _, BookRef := range ref.Refs {
				if v, ok := abbrTab[BookRef.Book]; ok {
					BookRef.Book = v
				}
				for _, ref := range BookRef.Refs {
					first := true
					for _, span := range ref.Spans {
						reading := false
						for _, line := range bible {
							if line[1] == BookRef.Book && line[2] == ref.Chapter && line[3] == span.Init {
								reading = true
								if first {
									text = append(text, "**"+BookRef.Book+" "+ref.Chapter+"**")
									first = false
								}
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
			return text
		} else {
			return []string{"Estou sem a biblia hoje!"}
		}
	}
	return []string{}
}

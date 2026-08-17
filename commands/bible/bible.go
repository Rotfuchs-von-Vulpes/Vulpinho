package bible

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"vulpinho/commands/bible/versicleParser"
)

//go:embed catholicism/bible.csv
var bibleRaw string

//go:embed catholicism/ccc.csv
var cccRaw string

var abbrTab map[string]string
var bible [][]string
var ccc [][]string
var bibleIsReady bool
var cccIsReady bool

func ReadBible() {
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
		"lc":   "sao-lucas",
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

	bibleReader := strings.NewReader(bibleRaw)
	bibleCsvReader := csv.NewReader(bibleReader)

	records1, err := bibleCsvReader.ReadAll()
	if err != nil {
		slog.Error("Incapaz de parsear CSV no arquivo da biblia.", "error", err.Error())
		bibleIsReady = false
		return
	}

	bible = records1
	bibleIsReady = true
	slog.Info("A biblia foi carregada com sucesso.")

	f2, missingError := os.Create("resources/bible/missing.txt")
	if missingError != nil {
		slog.Error("Incapaz de criar arquivo com a lista de versiculos faltantes.", "error", missingError.Error())
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

	cccReader := strings.NewReader(cccRaw)
	cccCsvReader := csv.NewReader(cccReader)
	records2, err := cccCsvReader.ReadAll()
	if err != nil {
		slog.Error("Incapaz de parsear CSV no arquivo do catecismo.", "error", err.Error())
		cccIsReady = false
		return
	}

	ccc = records2
	cccIsReady = true
	slog.Info("O catecismo foi carregado com sucesso.")
}

type bookToRead int

const (
	t_bible bookToRead = iota
	t_ccc
)

func Versicle(raw string) []string {
	p := versicleParser.GetVersicleParser(raw)
	text := []string{}
	if ok, ref := p.Parse(); ok {
		for _, BookRef := range ref.Refs {
			var wich bookToRead
			var book [][]string
			if v, ok := abbrTab[strings.ToLower(BookRef.Book)]; ok {
				BookRef.Book = v
			}
			if BookRef.Book == "ccc" || BookRef.Book == "catechism" {
				if !cccIsReady {
					return []string{"Estou sem o catecismo hoje!"}
				}
				BookRef.Book = "catechism"
				wich = t_ccc
				book = ccc
			} else {
				if !bibleIsReady {
					return []string{"Estou sem a biblia hoje!"}
				}
				wich = t_bible
				book = bible
			}
			for _, ref := range BookRef.Refs {
				if ref.Chapter == "noChapter" {
					if wich == t_ccc {
						ref.Chapter = "1"
					} else {
						continue
					}
				}
				first := true
				for _, span := range ref.Spans {
					reading := false
					for _, line := range book {
						if line[1] == BookRef.Book && line[2] == ref.Chapter && line[3] == span.Init {
							reading = true
							if first {
								var title string
								if wich == t_bible {
									title = fmt.Sprint("**" + BookRef.Book + " " + ref.Chapter + "**")
								} else {
									title = "**" + BookRef.Book + "**"
								}
								text = append(text, title)
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
	}
	return []string{}
}

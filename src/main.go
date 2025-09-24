package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/badgerodon/peg"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func readCsvFile(filePath string) [][]string {
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

	return records
}

func SnowflakeToUint64(snowflake string) (uint64, bool) {
	result, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		// log.Fatalf("error: %s", err.Error())
		return 0, false
	}
	return result, true
}

type Value int

const (
	Number = iota
	operator
)

type (
	OP struct {
		val  float64
		op   int
		next *OP
	}
)

var (
	prec = []int{'*', '/', '+', '-'}
	ops  = map[int]func(float64, float64) (bool, float64){
		'*': func(a, b float64) (bool, float64) {
			return true, a * b
		},
		'/': func(a, b float64) (bool, float64) {
			if b == 0 {
				return false, 0
			}
			return true, a / b
		},
		'+': func(a, b float64) (bool, float64) {
			return true, a + b
		},
		'-': func(a, b float64) (bool, float64) {
			return true, a - b
		},
	}
)

func reduce(tree *peg.ExpressionTree) (bool, float64) {
	// If we're at a number just parse it
	if tree.Name == "Number" {
		str := ""
		for _, c := range tree.Children {
			str += string(rune(c.Value))
		}
		i, _ := strconv.ParseFloat(str, 64)
		return true, i
	}

	// We have to collapse all sub expressions into a flattened linked list
	//   of expressions each of which has an operator. We will then execute
	//   each of the operators in order of precedence.
	fst := &OP{0, '+', nil}
	lst := fst
	var visit func(*peg.ExpressionTree)
	visit = func(t *peg.ExpressionTree) {
		switch t.Name {
		case "Expression":
			if len(t.Children) > 1 {
				_, reduced := reduce(t.Children[0])
				nxt := &OP{reduced, t.Children[1].Value, nil}
				lst.next = nxt
				lst = nxt
				visit(t.Children[2])
				return
			}
		case "Parentheses":
			_, reduced := reduce(t.Children[1])
			nxt := &OP{reduced, 0, nil}
			lst.next = nxt
			lst = nxt
			return
		}

		if len(t.Children) > 0 {
			_, reduced := reduce(t.Children[0])
			nxt := &OP{reduced, 0, nil}
			lst.next = nxt
			lst = nxt
		}
	}
	visit(tree)

	// Foreach operator in order of precedence
	for _, o := range prec {
		cur := fst
		for cur.next != nil {
			if cur.op == o {
				ok := true
				ok, cur.val = ops[o](cur.val, cur.next.val)
				if !ok {
					return false, 0
				}
				cur.op = cur.next.op
				cur.next = cur.next.next
			} else {
				cur = cur.next
			}
		}
	}

	return true, fst.val
}

func (op *OP) String() string {
	str := ""
	if op.op == 0 {
		str = "(" + fmt.Sprint(op.val) + ") "
	} else {
		str = "(" + fmt.Sprint(op.val) + " " + string(rune(op.op)) + ") "
	}
	if op.next != nil {
		str += op.next.String()
	}
	return str
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		return
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Error when connecting to discord: %s", err.Error())
		return
	}

	bible := readCsvFile("resources/bible/bible.csv")

	f, err := os.Create("resources/bible/missing.txt")
	if err != nil {
		log.Fatalf("Can't create missing list file")
	}

	var previous int64 = 0
	for _, line := range bible {
		num, err := strconv.ParseInt(line[3], 10, 32)
		if err == nil {
			if num == 1 {
				previous = 0
			}
			if num-previous > 1 {
				f.WriteString(line[1] + " " + line[2] + " " + strconv.FormatInt(previous+1, 10) + " não existe\n")
			}
			previous = num
		}
	}
	f.Close()

	parser := peg.NewParser()

	start := parser.NonTerminal("Start")
	expr := parser.NonTerminal("Expression")
	paren := parser.NonTerminal("Parentheses")
	number := parser.NonTerminal("Number")

	start.Expression = expr
	expr.Expression = parser.Sequence(
		parser.OrderedChoice(
			paren,
			number,
		),
		parser.Optional(
			parser.Sequence(
				parser.OrderedChoice(
					parser.Terminal('-'),
					parser.Terminal('+'),
					parser.Terminal('*'),
					parser.Terminal('/'),
				),
				expr,
			),
		),
	)
	paren.Expression = parser.Sequence(
		parser.Terminal('('),
		expr,
		parser.Terminal(')'),
	)
	number.Expression = parser.Sequence(
		parser.Sequence(
			parser.OneOrMore(
				parser.Range('0', '9'),
			),
		),
		parser.Optional(
			parser.Terminal('.'),
		),
		parser.Sequence(
			parser.ZeroOrMore(
				parser.Range('0', '9'),
			),
		),
	)

	// tree := parser.Parse("(0.5123651*3.14159+15)/2")
	// fmt.Println(tree)
	// fmt.Println(reduce(tree))

	lastMsg := map[string]string{}
	bannedPeople := map[string][]string{}
	bannedMsg := map[string]string{}
	repeatedMsgCount := map[string]int{}
	minimum := map[string]int{}

	waitingPong := false
	last_time := int64(0)

	discord.AddHandler(func(_ *discordgo.Session, message *discordgo.MessageCreate) {
		channelID := message.ChannelID
		serverID := message.GuildID

		if message.Author.ID == discord.State.User.ID {
			if waitingPong {
				if message.Content == "Pong!" {
					latency := message.Timestamp.UnixMilli() - last_time
					discord.ChannelMessageEdit(channelID, message.ID, fmt.Sprintf("Pong! Latência é %dms", latency))
					waitingPong = false
				}
			}
			return
		}
		if len(message.Content) <= 0 {
			return
		}

		_, ok := minimum[serverID]

		if !ok {
			minimum[serverID] = 2
		}

		messageContent := strings.ToLower(message.Content)

		switch messageContent {
		case "fox!":
			discord.ChannelMessageSend(channelID, ":fox::+1:")
		case "fox! ping":
			last_time = message.Timestamp.UnixMilli()
			waitingPong = true
			discord.ChannelMessageSend(channelID, "Pong!")
		default:
			msgText := message.Content
			if msgText == lastMsg[channelID] {
				banned := slices.Contains(bannedPeople[channelID], message.Author.ID)
				if !banned {
					bannedPeople[channelID] = append(bannedPeople[channelID], message.Author.ID)
					repeatedMsgCount[channelID] += 1
				}
			} else if msgText != bannedMsg[channelID] {
				bannedPeople[channelID] = nil
				bannedPeople[channelID] = append(bannedPeople[channelID], message.Author.ID)
				lastMsg[channelID] = msgText
				repeatedMsgCount[channelID] = 0
			}

			if repeatedMsgCount[channelID] == minimum[serverID] {
				discord.ChannelMessageSend(channelID, lastMsg[channelID])

				bannedMsg[channelID] = lastMsg[channelID]
				bannedPeople[channelID] = nil
				repeatedMsgCount[channelID] = 0
				lastMsg[channelID] = ""
				minimum[serverID] += 1
			}

			words := strings.Split(strings.ToLower(message.Content), " ")

			if len(words) >= 3 {
				if words[0] == "fox!" && words[1] == "calc" {
					final_str := ""

					for i, word := range words {
						if i < 2 {
							continue
						}
						final_str = fmt.Sprintf("%s%s", final_str, word)
					}
					tree := parser.Parse(final_str)
					ok, result := reduce(tree)
					if ok {
						discord.ChannelMessageSend(channelID, fmt.Sprint(result))
					} else {
						discord.ChannelMessageSend(channelID, "Uma ideterminação foi encontrada")
					}
				}
			}

			fops_list := []string{"raposa", "raposo", "raposinha", "raposinhas", "raposas", "raposos", "fops", "fox", "poposa", "poposas", "foxes", "fxoe"}
		loop:
			for _, word := range words {
				for _, fops := range fops_list {
					if word == fops {
						discord.MessageReactionAdd(channelID, message.ID, "🦊")
						break loop
					}
				}
			}

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
							discord.ChannelMessageSend(channelID, "**"+line[3]+"**. "+line[4])
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
						discord.ChannelMessageSend(channelID, text)
					}
				}
			}

			if message.MentionEveryone {
				discord.ChannelMessageSend(channelID, "<:memojo_really:1411209850213498890>")
			} else {
				if len(message.Mentions) == 1 && message.Mentions[0].ID == discord.State.User.ID {
					text, found := strings.CutPrefix(message.Content, fmt.Sprintf("<@%s> ", discord.State.User.ID))
					if found {
						commands := []rune{'J', 'I', 'M', 'S', 'T', 'C', 'R', 'N', 'G'}
						words := strings.Split(text, " ")
						invalid := false
						for _, r := range words[0] {
							found := false
							for _, m := range commands {
								if m == r {
									found = true
									break
								}
							}
							if !found {
								invalid = true
								break
							}
						}

						cmd := "node"
						var args []string
						if invalid {
							args = []string{"resources/javascript/run_camxes", text}
						} else {
							command := words[0]
							text, _ = strings.CutPrefix(text, command)
							args = []string{"resources/javascript/run_camxes", "-m", command, text}
						}
						process := exec.Command(cmd, args...)
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
						discord.ChannelMessageSend(channelID, buf.String())
					}
				}
			}
		}
	})

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT)

	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening discord session: %s", err.Error())
		return
	}
	defer func() {
		log.Printf("closing discord session...")
		if err := discord.Close(); err != nil {
			log.Fatalf("error closing discord session: %s", err.Error())
		}
	}()

	log.Printf("Online")

	<-signalChannel

	log.Print("Shutting down")
}

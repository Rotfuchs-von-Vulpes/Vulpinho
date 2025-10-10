package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/badgerodon/peg"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func parseNotes(note string) string {
	note = parseDefinition(note)
	noteBuilder := strings.Builder{}
	for _, r := range note {
		if r != '{' && r != '}' {
			noteBuilder.WriteRune(r)
		}
	}
	return noteBuilder.String()
}

func parseDefinition(def string) string {
	readingArgument := false
	readingSubscript := false
	defBuinder := strings.Builder{}
	for _, r := range def {
		if r == '$' {
			if readingArgument {
				readingArgument = false
				readingSubscript = false
				defBuinder.WriteRune('_')
			} else {
				readingArgument = true
				defBuinder.WriteRune('_')
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
				switch r {
				case '0':
					defBuinder.WriteRune('₀')
				case '1':
					defBuinder.WriteRune('₁')
				case '2':
					defBuinder.WriteRune('₂')
				case '3':
					defBuinder.WriteRune('₃')
				case '4':
					defBuinder.WriteRune('₄')
				case '5':
					defBuinder.WriteRune('₅')
				case '6':
					defBuinder.WriteRune('₆')
				case '7':
					defBuinder.WriteRune('₇')
				case '8':
					defBuinder.WriteRune('₈')
				case '9':
					defBuinder.WriteRune('₉')
				}
			} else if r != '_' && r != '{' && r != '}' {
				defBuinder.WriteRune(r)
			}
		}
	}

	return defBuinder.String()
}

func removePonctuation(text string) string {
	respondeBuilder := strings.Builder{}
	for _, r := range text {
		if unicode.IsLetter(r) {
			respondeBuilder.WriteRune(r)
		}
	}

	return respondeBuilder.String()
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

	readBible()
	readLojbanDict()

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
		say := func(text string) {
			discord.ChannelMessageSend(message.ChannelID, text)
		}
		sayList := func(list []string) {
			for _, text := range list {
				discord.ChannelMessageSend(message.ChannelID, text)
			}
		}

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
			say(lastMsg[channelID])

			bannedMsg[channelID] = lastMsg[channelID]
			bannedPeople[channelID] = nil
			repeatedMsgCount[channelID] = 0
			lastMsg[channelID] = ""
			minimum[serverID] += 1
		}

		words := strings.Split(strings.ToLower(message.Content), " ")
		fops_list := []string{"vulpinho", "raposa", "raposo", "raposinha", "raposinhas", "raposas", "raposos", "fop", "fops", "fox", "poposa", "poposas", "poposo", "poposos", "foxes", "fxoe"}

		if len(words) >= 1 {
			word_minus, found := strings.CutSuffix(words[0], "!")
			if found && slices.Contains(fops_list, word_minus) {
				if len(words) == 1 {
					say(":fox::+1:")
				} else if len(words) == 2 {
					if words[1] == "ping" {
						last_time = message.Timestamp.UnixMilli()
						waitingPong = true
						say("Pong!")
					}
				} else if len(words) >= 3 {
					if words[1] == "calc" {
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
							say(fmt.Sprint(result))
						} else {
							say("Uma ideterminação foi encontrada")
						}
					} else if words[1] == "gerna" {
						text, found := strings.CutPrefix(message.Content, "fox! gerna ")
						if found {
							say(gerna(text))
						}
					} else if words[1] == "sisku" {
						word := strings.Split(message.Content, " ")[2]
						say(sisku(word))
					} else if words[1] == "facki" {
						word := strings.Split(message.Content, " ")[2]
						sayList(facki(word))
					}
				}
			}
		}

	loop:
		for _, word := range words {
			word = removePonctuation(word)
			for _, fops := range fops_list {
				if word == fops {
					discord.MessageReactionAdd(channelID, message.ID, "🦊")
					break loop
				}
			}
		}

		versicle(say, message.Content)

		if message.MentionEveryone {
			say("<:memojo_really:1411209850213498890>")
		} else {
			msgRef := message.ReferencedMessage
			to_emote := true
			if msgRef != nil {
				if msgRef.Author.ID == discord.State.User.ID {
					to_emote = false
				}
			}
			if to_emote {
				if len(message.Mentions) == 1 && message.Mentions[0].ID == discord.State.User.ID {
					say("<a:foxexcite:1421359331361816678>")
				} else if len(message.Mentions) > 1 {
					for _, user := range message.Mentions {
						if user.ID == discord.State.User.ID {
							say("<:pepe_think:1421357826051407962>")
							break
						}
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

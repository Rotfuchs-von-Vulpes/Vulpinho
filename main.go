package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
	"vulpinho/commands/bible"
	"vulpinho/commands/chemistry"
	"vulpinho/commands/lojban"
	"vulpinho/commands/remind"
	"vulpinho/commands/svg"
	"vulpinho/commands/wh40k"
	"vulpinho/log"
	"vulpinho/update"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type promptProc struct {
	buff           []rune
	attachs        []bytes.Buffer
	idx            int
	attachmentsURL []string
}

func newCommand(msg string, files []string) (c *promptProc) {
	c = new(promptProc)
	c.buff = []rune(msg)
	c.idx = 0
	c.attachmentsURL = files
	return
}

func (s *promptProc) mark() int {
	return s.idx
}

func (s *promptProc) reset(pos int) {
	s.idx = pos
}

func (s *promptProc) space() bool {
	if s.idx > len(s.buff)-1 {
		return false
	}
	r := s.buff[s.idx]
	if unicode.IsSpace(r) {
		s.idx += 1
		return true
	}
	return false
}

func (s *promptProc) consumeAllSpace() {
	for {
		if !s.space() {
			return
		}
	}
}

func (s *promptProc) testRune(r rune) bool {
	if s.idx > len(s.buff)-1 {
		return false
	}
	rr := s.buff[s.idx]
	if rr == r {
		s.idx += 1
		return true
	}
	return false
}

func (s *promptProc) testString(str string) bool {
	s.consumeAllSpace()
	pos := s.mark()
	for _, r := range str {
		if !s.testRune(r) {
			s.reset(pos)
			return false
		}
	}
	return true
}

func (s *promptProc) testVarious(list []string) bool {
	for _, str := range list {
		if s.testString(str) {
			return true
		}
	}
	return false
}

func getFileData(url string) (string, error) {
	// Perform the GET request to Discord's CDN
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Cant request data", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the body into a byte slice
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Cant read data", "error", err)
		return "", err
	}

	return string(data), nil
}

func (s *promptProc) getRemains() string {
	s.consumeAllSpace()
	b := strings.Builder{}
	for {
		pos := s.mark()
		if pos > len(s.buff)-1 {
			break
		}
		s.idx += 1
		b.WriteRune(s.buff[pos])
	}
	for _, url := range s.attachmentsURL {
		if fileText, err := getFileData(url); err == nil {
			b.WriteString(fileText)
		}
	}
	return b.String()
}

func removePonctuation(text string) string {
	respondeBuilder := strings.Builder{}
	for _, r := range text {
		if unicode.IsLetter(r) || r == '🦊' {
			respondeBuilder.WriteRune(r)
		}
	}

	return respondeBuilder.String()
}

func SnowflakeToUint64(snowflake string) (uint64, bool) {
	result, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		return 0, false
	}
	return result, true
}

type Author struct {
	ID          string
	AvatarURL   string
	DisplayName string
}

type Message struct {
	ID          string
	Author      *discordgo.User
	ChannelID   string
	GuildID     string
	Content     string
	Time        int64
	Attachments []*discordgo.MessageAttachment
}

func wrapMessageCreate(message *discordgo.MessageCreate) Message {
	return Message{message.ID, message.Author, message.ChannelID, message.GuildID, message.Content, message.Timestamp.UnixMilli(), message.Attachments}
}

func wrapMessageUpdate(message *discordgo.MessageUpdate) Message {
	return Message{message.ID, message.Author, message.ChannelID, message.GuildID, message.Content, message.Timestamp.UnixMilli(), message.Attachments}
}

func main() {
	if log.Init() {
		return
	}

	args := os.Args[1:]
	isTest := false

	if len(args) > 0 {
		switch args[0] {
		case "update":
			slog.Info("Atualizando dados sobre Warhammer 40k")
			update.UpdateWarhammer()
			return
		case "test":
			isTest = true
		}
	}

	update.GetLastEdit()
	bible.ReadBible()
	lojban.LojbanInit()
	wh40k.ReadWh()
	chemistry.Init()

	isRemberOk, missedReminds := remind.Init()
	fmt.Println("iniciando...")

	if isTest {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT)

		<-signalChannel

		slog.Info("Desligando...")
		log.End()

		return
	}

	err := godotenv.Load(".env")
	if err != nil {
		slog.Error("Erro ao ler arquivo .env.", "error", err.Error())
		return
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		slog.Error("Erro ao se conectar com o discord.", "error", err.Error())
		return
	}

	if isRemberOk {
		for _, msg := range missedReminds {
			if _, err := discord.ChannelMessageSend(msg.ChannelID, msg.Message); err != nil {
				slog.Error("Lembrete falhou", "error", err.Error())
			}
		}
		go func() {
			msg := <-remind.RemberChan
			if _, err := discord.ChannelMessageSend(msg.ChannelID, msg.Message); err != nil {
				slog.Error("Lembrete falhou", "error", err.Error())
			}
		}()
	}

	history := make(map[string][]string)

	lastMsg := map[string]string{}
	bannedPeople := map[string][]string{}
	bannedMsg := map[string]string{}
	repeatedMsgCount := map[string]int{}
	minimum := map[string]int{}

	waitingPong := false
	last_time := int64(0)

	reactReceiverMessages := []string{}

	runCommands := func(message Message) {
		say := func(text string) {
			if m, err := discord.ChannelMessageSend(message.ChannelID, text); err == nil {
				history[message.ID] = []string{m.ID}
			} else {
				r := strings.NewReader(text)
				if m, err2 := discord.ChannelFileSend(message.ChannelID, "response.txt", r); err2 == nil {
					history[message.ID] = []string{m.ID}
				} else {
					slog.Error("Não foi possivel enviar mensagem", "error", err.Error())
				}
			}
		}
		sayList := func(list []string) {
			if len(list) != 0 {
				allMsgs := []string{}
				final := []string{""}
				for _, text := range list {
					last := len(final) - 1
					if len(final[last])+len(text)+1 < 2000 {
						final[last] += text + "\n"
					} else {
						final = append(final, text+"\n")
					}
				}
				for _, msg := range final {
					if m, err := discord.ChannelMessageSend(message.ChannelID, msg); err == nil {
						allMsgs = append(allMsgs, m.ID)
					} else {
						slog.Error("Não foi possivel enviar mensagem", "error", err.Error())
					}
				}
				history[message.ID] = allMsgs
			}
		}
		sayEmbed := func() {
			embed := discordgo.MessageEmbed{}
			embed.Title = "Título do Teste"
			embed.Description = "Descrição do teste"
			embed.Type = discordgo.EmbedTypeArticle
			embed.Color = 0xff6000
			author := discordgo.MessageEmbedAuthor{}
			author.IconURL = message.Author.AvatarURL("")
			author.Name = message.Author.DisplayName()
			embed.Author = &author
			if msg, err := discord.ChannelMessageSendEmbed(message.ChannelID, &embed); err == nil {
				history[message.ID] = []string{msg.ID}
				discord.MessageReactionAdd(message.ChannelID, msg.ID, "🍇")
				reactReceiverMessages = append(reactReceiverMessages, msg.ID)
			} else {
				slog.Error("Não foi possivel enviar embed", "error", err.Error())
			}
		}
		sayImage := func(imageName string, imageBuffer io.Reader) {
			if msg, err := discord.ChannelFileSend(message.ChannelID, imageName+".png", imageBuffer); err == nil {
				history[message.ID] = []string{msg.ID}
			} else {
				slog.Error("Não foi possivel enviar imagem", "error", err.Error())
			}
		}

		channelID := message.ChannelID
		serverID := message.GuildID

		if message.Author.ID == discord.State.User.ID {
			if waitingPong {
				if message.Content == "Pong!" {
					latency := message.Time - last_time
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

		fops_list := []string{
			// Special words and symbols
			"vulpinho", "vulpinha", "vulpinhos", "vulpinhas", "🦊",

			// Fox species (without "chama" and "cana", common words in portuguese)
			"vulpini", "lagopus", "velox", "macrotis", "corsac", "pallida", "bengalensis", "ferrilata", "rueppellii", "zerda",

			// Conlang
			"semimi", "semi", "lorxu", "vulpino", "vulpo", "vulpinoj", "vulpoj", "oram", "kıtta",

			// Portuguese
			"raposa", "raposo", "raposas", "raposos", "raposinha", "raposinho", "raposinhas", "raposinhos", "poposa", "poposo", "poposas", "poposos", "posa", "poso", "poposinha", "poposinho", "poposinhas", "poposinhos",

			// German
			"fuchs", "füchse", "fuchse", "füchses", "fuchses",

			// Spanish
			"zorro", "zorra", "zorros", "zorras", "zorrita", "zorrito", "zorritas", "zorritos",

			// English
			"fox", "foxe", "foxes", "foxy", "foxys", "foxis", "fxoe", "fxoes", "foex", "foexes", "fux", "fuxes", "fop", "fops", "fopse", "fopses", "vix", "vixes", "vixen", "vixens",

			// French
			"renard", "renards", "renarde", "renardes", "renardeau", "renardeaux",

			// Italian
			"volpe", "volpi", "volpes", "volpeses",

			// Russian
			"лиса", "лисица", "лис", "лисы", "лисички", "лисичка", "лисичкина",
			"lisa", "lisitsa", "lis", "lisy", "lisichki", "lisichka", "lisichkina",

			// Other languages
			"狐", "kitsune", "キツネ", "여우", "ثعلب", "الثعالب", "लोमड़ी", "लोमड़ियों", "שׁוּעָל", "שועלים", "αλεπού", "vulpiculus",
		}

		attachmentURLs := []string{}
		for _, a := range message.Attachments {
			attachmentURLs = append(attachmentURLs, a.URL)
		}
		proc := newCommand(message.Content, attachmentURLs)

		if proc.testVarious(fops_list) && proc.testRune('!') {
			if proc.testString("ping") {
				last_time = message.Time
				waitingPong = true
				say("Pong!")
			} else if proc.testString("teste") {
				sayEmbed()
			} else if proc.testString("gerna") {
				say(lojban.Gerna(proc.getRemains()))
			} else if proc.testString("sisku") {
				say(lojban.Sisku(proc.getRemains()))
			} else if proc.testString("facki") {
				sayList(lojban.Facki(proc.getRemains()))
			} else if proc.testString("wh40k") {
				if proc.testString("keyword") || proc.testString("keywords") {
					text := proc.getRemains()
					if text != "" {
						keys := strings.Split(text, ",")
						for i, key := range keys {
							keys[i] = strings.TrimSpace(key)
						}
						sayList(wh40k.KeySearch(keys))
					}
				} else {
					text := proc.getRemains()
					words := strings.Split(text, " ")
					if len(words) >= 1 {
						data := words[0]
						id := words[1]
						sayList(wh40k.GetWh(data, id))
					}
				}
			} else if proc.testString("smiles") {
				if buff, ok, errStr := chemistry.Smiles(proc.getRemains()); ok {
					if errStr == "" {
						sayImage("smilesImg", buff)
					} else {
						say(errStr)
					}
				}
			} else if proc.testString("svg") {
				if code := svg.ExtractSvgCodeFromMsg(proc.getRemains()); len(code) > 0 {
					if buff, ok, errStr := svg.SvgToPng(code); ok {
						if errStr == "" {
							sayImage("svgImg", buff)
						} else {
							say(errStr)
						}
					}
				}
			} else {
				say("<a:fox_wave:1426439130253885440>")
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

		text := bible.Versicle(message.Content)
		if len(text) > 0 {
			if len(text) == 1 {
				say(text[0])
			} else {
				sayList(text)
			}
		}
	}

	discord.AddHandler(func(_ *discordgo.Session, reaction *discordgo.MessageReactionRemove) {

	})

	discord.AddHandler(func(_ *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
		say := func(text string) {
			discord.ChannelMessageSend(reaction.ChannelID, text)
		}
		if slices.Contains(reactReceiverMessages, reaction.MessageID) {
			if reaction.UserID != discord.State.User.ID {
				if reaction.Emoji.ID == "" {
					say(reaction.Member.DisplayName() + " reagiu com " + reaction.Emoji.Name)
				} else {
					if reaction.Emoji.Animated {
						say(reaction.Member.DisplayName() + " reagiu com <a:" + reaction.Emoji.Name + ":" + reaction.Emoji.ID + ">")
					} else {
						say(reaction.Member.DisplayName() + " reagiu com <:" + reaction.Emoji.Name + ":" + reaction.Emoji.ID + ">")
					}
				}
			}
		}
	})

	discord.AddHandler(func(_ *discordgo.Session, message *discordgo.MessageUpdate) {
		if list, ok := history[message.ID]; ok {
			history[message.ID] = nil
			discord.ChannelMessagesBulkDelete(message.ChannelID, list)
		}
		runCommands(wrapMessageUpdate(message))
	})

	discord.AddHandler(func(_ *discordgo.Session, message *discordgo.MessageCreate) {
		if message.MentionEveryone {
			discord.ChannelMessageSend(message.ChannelID, "<:memojo_really:1411209850213498890>")
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
					discord.ChannelMessageSend(message.ChannelID, "<a:foxexcite:1421359331361816678>")
				} else if len(message.Mentions) > 1 {
					for _, user := range message.Mentions {
						if user.ID == discord.State.User.ID {
							discord.ChannelMessageSend(message.ChannelID, "<:pepe_think:1421357826051407962>")
							break
						}
					}
				}
			}
		}

		runCommands(wrapMessageCreate(message))
	})

	discord.AddHandler(func(_ *discordgo.Session, message *discordgo.MessageDelete) {
		if list, ok := history[message.ID]; ok {
			history[message.ID] = nil
			discord.ChannelMessagesBulkDelete(message.ChannelID, list)
		}
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT)

	go func() {
		seconds := 2
		limit := 60 * 60
		var lastErr string
		for {
			select {
			case <-signalChan:
				return
			default:
				err = discord.Open()
				if err != nil {
					if err.Error() != lastErr {
						lastErr = err.Error()
						slog.Error("Erro ao abrir uma sessão no Discord.", "error", err.Error())
					}
					slog.Info("Reconectando...")
					time.Sleep(time.Duration(seconds) * time.Second)
					if seconds < limit {
						seconds *= seconds
					}
				} else {
					return
				}
			}
		}
	}()

	defer func() {
		slog.Info("Fechando a sessão...")
		if err := discord.Close(); err != nil {
			slog.Error("Erro ao fechar a sessão.", "error", err.Error())
		}
	}()

	slog.Info("Online.")

	<-signalChan

	slog.Info("Desligando...")
	log.End()
}

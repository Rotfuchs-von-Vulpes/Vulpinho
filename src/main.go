package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var logger *slog.Logger

func removeFirstTwo(wordList []string) []string {
	final := []string{}
	for i, word := range wordList {
		if i >= 2 {
			final = append(final, word)
		}
	}
	return final
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
		// logger.Error("error: %s", err.Error())
		return 0, false
	}
	return result, true
}

func fileExist(path string) (bool, bool) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, false
	} else if err == nil {
		return true, false
	} else {
		slog.Error("error viewing stat", "err", err)
		return false, true
	}
}

func removeAndCreate(path string) (*os.File, bool) {
	exist, fatal := fileExist(path)
	if fatal {
		return nil, true
	}
	if exist {
		if err := os.Remove(path); err != nil {
			slog.Error("error removing file", "err", err.Error())
			return nil, true
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE, os.ModeDevice)
	if err != nil {
		slog.Error("error opening file", "err", err.Error())
		return nil, true
	}
	return f, false
}

func main() {
	folderLog, fatal := fileExist("log/")
	if fatal {
		return
	}
	if !folderLog {
		os.Mkdir("log", 0)
	}
	var f *os.File
	f, fatal = removeAndCreate("log/log.txt")
	if fatal {
		return
	}
	defer f.Close()
	logger = slog.New(NewCopyHandler(slog.NewTextHandler(os.Stdout, nil), slog.NewTextHandler(f, nil)))

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("Error loading .env file")
		return
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		logger.Error("Error when connecting to discord", "error", err.Error())
		return
	}

	readBible()
	lojbanInit()
	readWh()

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
		fops_list := []string{"vulpinho", "🦊", "raposa", "raposo", "raposinha", "raposinhas", "raposas", "raposos", "fop", "fops", "fox", "poposa", "poposas", "poposo", "poposos", "foxes", "fxoe"}

		if len(words) >= 1 {
			word_minus, found := strings.CutSuffix(words[0], "!")
			if found && slices.Contains(fops_list, word_minus) {
				if len(words) == 1 {
					say("<a:fox_wave:1426439130253885440>")
				} else if len(words) == 2 {
					if words[1] == "ping" {
						last_time = message.Timestamp.UnixMilli()
						waitingPong = true
						say("Pong!")
					}
				} else if len(words) >= 3 {
					switch words[1] {
					case "gerna":
						text := removeFirstTwo(strings.Split(message.Content, " "))
						say(gerna(strings.Join(text, " ")))
					case "sisku":
						word := strings.Split(message.Content, " ")[2]
						say(sisku(word))
					case "facki":
						text := removeFirstTwo(strings.Split(message.Content, " "))
						sayList(facki(strings.Join(text, " ")))
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

	go func() {
		seconds := 2
		limit := 60 * 60
		var lastErr string
		for {
			select {
			case <-signalChannel:
				return
			default:
				err = discord.Open()
				if err != nil {
					if err.Error() != lastErr {
						lastErr = err.Error()
						logger.Error("error opening discord session", "error", err.Error())
					}
					logger.Info("reconnect...")
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

	// err = discord.Open()
	// if err != nil {
	// 	logger.Error("error opening discord session", "error", err.Error())
	// 	return
	// }
	defer func() {
		logger.Info("closing discord session...")
		if err := discord.Close(); err != nil {
			logger.Error("error closing discord session", "error", err.Error())
		}
	}()

	logger.Info("Online")

	<-signalChannel

	logger.Info("Shutting down")
}

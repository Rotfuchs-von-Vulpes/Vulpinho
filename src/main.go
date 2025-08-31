package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func readCsvFile(filePath string) [][]string {
	f, err := os.Open("resources/" + filePath)
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

func SnowflakeToUint64(snowflake string) uint64 {
	result, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return 0
	}
	return result
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		return
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("error when connecting to discord: %s", err.Error())
		return
	}

	bible := readCsvFile("bible.csv")

	log.Print(bible[0][0])

	discord.AddHandler(func(_ *discordgo.Session, message *discordgo.MessageCreate) {
		if message.Author.Bot {
			return
		}
		if len(message.Content) <= 0 {
			return
		}

		channelID := message.ChannelID
		messageContent := strings.ToLower(message.Content)

		switch messageContent {
		case "fox!":
			discord.ChannelMessageSend(channelID, ":fox::+1:")
		case "fox! ping":
			latency := time.Now().UTC().UnixMilli() - message.Timestamp.UnixMilli()
			discord.ChannelMessageSend(channelID, fmt.Sprintf("Pong! Latência é %dms", latency))
		default:
			words := strings.Split(strings.ToLower(message.Content), " ")

			if len(words) == 3 {
				for _, line := range bible {
					if line[1] == words[0] && line[2] == words[1] && line[3] == words[2] {
						discord.ChannelMessageSend(channelID, line[4])
						break
					}
				}
			}

			if message.MentionEveryone {
				discord.ChannelMessageSend(channelID, "<:memojo_really:1411209850213498890>")
			} else {
				for _, user := range message.Mentions {
					if user.ID == discord.State.User.ID {
						discord.ChannelMessageSend(channelID, "<:pepe_ping:954135254329852014>")
						break
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

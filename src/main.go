package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
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

func SnowflakeToUint64(snowflake string) (uint64, bool) {
	result, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		// log.Fatalf("error: %s", err.Error())
		return 0, false
	}
	return result, true
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

	bible := readCsvFile("bible.csv")

	f, err := os.Create("resources/missing.txt")
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

	lastMsg := map[string]string{}
	bannedPeople := map[string][]string{}
	bannedMsg := map[string]string{}
	repeatedMsgCount := map[string]int{}
	minimum := 2

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

			if repeatedMsgCount[channelID] == minimum {
				discord.ChannelMessageSend(channelID, lastMsg[channelID])
				bannedPeople[channelID] = nil

				bannedMsg[channelID] = lastMsg[channelID]
				lastMsg[channelID] = ""
				minimum += 1
			}

			words := strings.Split(strings.ToLower(message.Content), " ")

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

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	Token string
)

func main() {
	fmt.Println("main")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Token = os.Getenv("DISCORD_TOKEN")
	fmt.Println(Token)
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == "!bootstrap" {
		s.ChannelMessageSend(m.ChannelID, "bootstrapping server")
		err := bootstrapServer(s, m.ChannelID, m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error bootstrapping server")
		}
	}

	if m.Content == "!nuke" {
		s.ChannelMessageSend(m.ChannelID, "nuking server")
		err := nukeServer(s, m.ChannelID, m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error nuking server")
		}
	}

	if m.Content == "!clear" {
		s.ChannelMessageSend(m.ChannelID, "clearing messages")
		err := clearMessages(s, m.ChannelID, m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error clearing messages")
		}
	}

	if strings.Contains(m.Content, "!kick") {
		s.ChannelMessageSend(m.ChannelID, "kicking user")
		userToKick := strings.Split(m.Content, " ")[1]
		err := s.GuildMemberDelete(m.GuildID, userToKick)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error kicking user")
		}
	}
}

func clearMessages(s *discordgo.Session, channelID string, guildID string) error {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return err
	}

	for _, c := range channels {
		if c.ID == channelID {
			if c.Messages == nil {
				fmt.Println("No Messages Found")
			}
			for _, message := range c.Messages {
				fmt.Println(message.Content)
				err := s.ChannelMessageDelete(c.ID, message.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func bootstrapServer(s *discordgo.Session, channelID string, guildID string) error {
	var channelNames = []string{"general", "🤖 bot-commands", "📣 - Announcements", "🆕 New Members", "📜 Rules"}
	for _, channelName := range channelNames {
		if !isChannelNameUsed(s, channelName) {
			_, err := s.GuildChannelCreate(guildID, channelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				return err
			}
			s.ChannelMessageSend(channelID, fmt.Sprintf("-> %s channel created", channelName))
		} else {
			s.ChannelMessageSend(channelID, fmt.Sprintf("-> %s channel already exists ... skipping", channelName))
		}
	}
	return nil
}

func nukeServer(s *discordgo.Session, channelID string, guildID string) error {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return err
	}

	for _, c := range channels {
		if c.Name != "general" {
			_, err = s.ChannelDelete(c.ID)
			if err != nil {
				return err
			}
			s.ChannelMessageSend(channelID, fmt.Sprintf("-> %s channel deleted", c.Name))
		}
	}
	return nil
}

func isChannelNameUsed(s *discordgo.Session, channelName string) bool {
	for _, guild := range s.State.Guilds {
		channels, _ := s.GuildChannels(guild.ID)

		for _, c := range channels {
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			if channelName == c.Name {
				return true
			}
		}
	}
	return false
}

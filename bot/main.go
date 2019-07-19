package main

import (
	//"context"
	"errors"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	//"time"

	//pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/bot/commands"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

var (
	// Token is the bot's Token
	Token string
	// prefix is the bot's prefix, temporarily here until I work out a good configuration set-up
	prefix = "!"
)

func parseCommand(content string) (args []string, err error) {
	if !strings.HasPrefix(content, prefix) {
		err = errors.New("string does not contain prefix")
		return
	}
	content = strings.TrimPrefix(content, prefix)
	args = strings.Split(content, " ")
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, conn *grpc.ClientConn) {
	if len(m.Content) <= 0 || !strings.HasPrefix(m.Content, prefix) ||
		m.Author.ID == s.State.User.ID {
		return
	}

	args, err := parseCommand(m.Content)
	if err != nil {
		log.Println(err)
		return
	}

	/* this shouldn't theoeretically be possible */
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "help":
		help, err := commands.GetHelp()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, help)
		return
	case "prefix":
		help, err := commands.Prefix(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Prefix changed to "+args[1]+", was "+prefix)
		prefix = args[1]
		return
	case "gettable":
		out, err := commands.GetTable(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "```"+out+"```")
		return
	case "adduser":
		help, err := commands.AddUser(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "User added!")
		return
	case "deluser":
		help, err := commands.DelUser(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "User deleted!")
		return
	case "addcol":
		help, err := commands.AddCol(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Column added!")
		return
	case "delcol":
		help, err := commands.DelCol(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Column deleted!")
		return
	case "incact":
		help, err := commands.IncAct(conn, args)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		if help != "" {
			s.ChannelMessageSend(m.ChannelID, help)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Activity Incremented!")
		return
	}
}

func main() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("error creating Discord session, %v", err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { messageCreate(s, m, conn) })

	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection, %v", err)
		return
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

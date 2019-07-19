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

// Check for quoted terms in a given byte slice, also sorts byte slice into a slice of strings
// Does not currently support escaping quotes. \" is still treated as a quote
// This is a very poor implementation of this, currently leading characters that are coupled with a quote, like
// aaaaaa"cats are cute" are ignored, and trailing characters ("cats are cute"aaaa) are added into ret
// There's a much better way to do this, fix later
func splitQuotes(in []byte) (ret []string, err error) {
	var inquote bool
	var n int
	for i := 0; i != len(in); i++ {
		if in[i] == '"' {
			if inquote {
				inquote = false
			} else {
				n = i + 1
				inquote = true
			}
		} else if !inquote && in[i] == ' ' {
			if i != 0 && in[i-1] == '"' {
				ret = append(ret, string(in[n:i-1]))
			} else {
				ret = append(ret, string(in[n:i]))
			}
			n = i + 1
		}
	}
	if inquote {
		return nil, errors.New("error: unterminated quote")
	}

	if in[len(in)-1] == '"' {
		ret = append(ret, string(in[n:len(in)-1]))
	} else {
		ret = append(ret, string(in[n:]))
	}

	return
}

func parseCommand(content string) (args []string, err error) {
	if !strings.HasPrefix(content, prefix) {
		err = errors.New("string does not contain prefix")
		return
	}

	// removes prefix from command
	content = strings.TrimPrefix(content, prefix)
	// removes multiple instances of whitespace
	content = strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	args, err = splitQuotes([]byte(content))
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

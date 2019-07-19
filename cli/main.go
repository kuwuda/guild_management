package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kuwuda/guild_management/cli/commands"
	"google.golang.org/grpc"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile     = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
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
	// removes multiple instances of whitespace
	content = strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	args, err = splitQuotes([]byte(content))
	return
}

func processCommand(conn *grpc.ClientConn, in string) error {
	if len(in) <= 0 {
		return errors.New("empty slice")
	}

	input, err := parseCommand(in)
	if err != nil {
		return err
	}

	switch input[0] {
	case "help":
		err := commands.GetHelp()
		if err != nil {
			return err
		}
		return nil
	case "gettable":
		err := commands.GetTable(conn, input)
		if err != nil {
			return err
		}
		return nil
	case "adduser":
		err := commands.AddUser(conn, input)
		if err != nil {
			return err
		}
		return nil
	case "deluser":
		err := commands.DelUser(conn, input)
		if err != nil {
			return err
		}
		return nil
	case "addcol":
		err := commands.AddCol(conn, input)
		if err != nil {
			return err
		}
		return nil
	case "delcol":
		err := commands.DelCol(conn, input)
		if err != nil {
			return err
		}
		return nil
	case "incact":
		err := commands.IncAct(conn, input)
		if err != nil {
			return err
		}
		return nil
	}

	return errors.New("command not found")
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	fmt.Println("Enter your commands:")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ucl := scanner.Text()
		err := processCommand(conn, ucl)
		if err != nil {
			log.Printf("error processing command: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("error: %v", err)
		return
	}
}

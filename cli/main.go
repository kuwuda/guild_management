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

func processCommand(conn *grpc.ClientConn, input []string) error {
	if len(input) <= 0 {
		return errors.New("empty slice")
	}

	if input[0] == "gettable" {
		err := commands.GetTable(conn, input)
		if err != nil {
			return err
		}
		return nil
	}
	if input[0] == "adduser" {
		err := commands.AddUser(conn, input)
		if err != nil {
			return err
		}
		return nil
	}
	if input[0] == "addcol" {
		err := commands.AddCol(conn, input)
		if err != nil {
			return err
		}
		return nil
	}
	if input[0] == "delcol" {
		err := commands.DelCol(conn, input)
		if err != nil {
			return err
		}
		return nil
	}
	if input[0] == "incact" {
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
		err := processCommand(conn, strings.Split(ucl, " "))
		if err != nil {
			log.Printf("error processing command: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("error: %v", err)
		return
	}
}

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"git.sr.ht/~sircmpwn/getopt"
	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile     = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
)

func digits(n int) (digits int) {
	for n != 0 {
		n /= 10
		digits++
	}
	return
}

// Returns the greater int of the pair
// necessary since Golang's standard library has no function for ints, only floats
func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

func sendToServer(conn *grpc.ClientConn, input []string) error {
	if len(input) <= 0 {
		return errors.New("empty slice")
	}

	if input[0] == "getActivities" {
		opts, optind, err := getopt.Getopts(input, "n:u:h")
		if err != nil {
			log.Fatal(err)
		}
		// get specified arguments
		// there has to be a more eloquent way to do this
		var request pb.ActivityRequest
		if optind > 1 {
			for _, opt := range opts {
				switch opt.Option {
				case 'n':
					// amount
					t, err := strconv.Atoi(opt.Value)
					if err != nil {
						log.Print(err)
					}
					request.Amount = uint32(t)
				case 'u':
					//user
					request.User = opt.Value
				case 'h':
					fmt.Println("usage: getActivities [options]")
					fmt.Println("Options:")
					fmt.Println("-n Specifies the maximum amount of results to return\n" +
						"-u Specifies the specific user to get\n" +
						"-h Prints this dialog\n" +
						"If no options are input, will default to returning every user.")
					return nil
				}
			}

			// get remaining arguments
			// commented out because it's currently unused
			//for _, arg := range os.Args[optind:] {
			//	println(arg)
			//}
		}

		activities, err := client.GetActivities(conn, &request)
		if err != nil {
			log.Println("error: ", err)
			return err
		}

		if len(activities) <= 0 {
			fmt.Println("No activities found!")
			return nil
		}

		// get the keys of db (doesn't use client.GetKeys since that would incur an unnecessary  request to server)
		var keys []string
		keys = append(keys, "Name")
		for k := range activities[0].Activities {
			keys = append(keys, k)
		}

		// since the way ranges iterate on maps is non-deterministic, store each row in a slice of string slices instead
		var rows [][]string
		// appends keys to rows, for convenience purposes
		rows = append(rows, keys)
		for _, v := range activities {
			var row []string
			row = append(row, v.Name)
			for _, k := range keys {
				// ints are converted to strings since this is just for display purposes anyway
				if k != "Name" {
					row = append(row, strconv.Itoa(int(v.Activities[k])))
				}
			}
			rows = append(rows, row)
		}

		// get the maximum length of every entry for the purpose of printing
		var maxlens []int
		for _, r := range rows {
			for n, v := range r {
				if n > len(maxlens)-1 {
					maxlens = append(maxlens, len(v))
				} else {
					maxlens[n] = max(maxlens[n], len(v))
				}
			}
		}

		var sb strings.Builder
		for _, r := range rows {
			for n, v := range r {
				sb.WriteString(v)
				i := maxlens[n]
				for i > len(v) {
					sb.WriteRune(' ')
					i--
				}
				sb.WriteRune(' ')
			}
			fmt.Println(sb.String())
			sb.Reset()
		}
	}
	if input[0] == "newActivityMember" {
		opts, optind, err := getopt.Getopts(input, "h")
		if err != nil {
			log.Fatal(err)
		}

		keys, err := client.GetKeys(conn)
		if err != nil {
			log.Fatal(err)
		}

		// Not really a fan of initializing this when it isn't used
		// But even less of a fan of doing it twice
		// Probably a better solution
		var help strings.Builder
		help.WriteString("usage: newActivityMember [options] name ")
		for _, v := range keys {
			help.WriteRune('"')
			help.WriteString(v)
			help.WriteRune('"')
			help.WriteRune(' ')
		}
		help.WriteString("\nWhere name etc. is replaced with the value for the input\n" +
			"EX:\n" +
			"newActivityMember Apples 4 8\n")
		help.WriteString("Any non-specified values will default to 0\n Extraneous values are ignored")
		help.WriteString("Options:\n " +
			"-h Prints this message")

		var parsed []string
		// get specified arguments
		if len(input) > 1 {
			for _, opt := range opts {
				switch opt.Option {
				case 'h':
					fmt.Println(help.String())
					return nil
				}
			}

			// get remaining arguments
			for _, arg := range input[optind:] {
				parsed = append(parsed, arg)
			}
		} else {
			fmt.Println(help.String())
			return nil
		}
		activityItem := &pb.ActivityItem{Name: parsed[0]}
		activityItem.Activities = make(map[string]uint32)

		for len(parsed)-1 < len(keys) {
			parsed = append(parsed, "0")
		}

		for i, v := range keys {
			parsedAsInt, err := strconv.Atoi(parsed[i+1])
			if err != nil {
				return err
			}
			activityItem.Activities[v] = uint32(parsedAsInt)
		}

		var activities []*pb.ActivityItem
		activities = append(activities, activityItem)

		resp, err := client.WriteMembers(conn, activities)
		fmt.Println(resp)
		fmt.Println(err)
	}
	if input[0] == "newColumn" {
		opts, _, err := getopt.Getopts(input, "h")
		if err != nil {
			log.Fatal(err)
		}

		var help strings.Builder
		help.WriteString("usage: newColumn [options] key")
		help.WriteString("\nWhere key is the name of the column being created\n" +
			"EX:\n" +
			"newColumn Cats\n")
		help.WriteString("Options:\n " +
			"-h Prints this message")

		// get specified arguments
		if len(input) > 1 {
			for _, opt := range opts {
				switch opt.Option {
				case 'h':
					fmt.Println(help.String())
					return nil
				}
			}
		} else {
			fmt.Println(help.String())
			return nil
		}
		key := input[1]
		var keys []string
		keys = append(keys, key)

		resp, err := client.AddColumns(conn, keys)
		fmt.Println(resp)
		fmt.Println(err)
	}

	return nil
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
		err := sendToServer(conn, strings.Split(ucl, " "))
		if err != nil {
			log.Printf("could not send to server: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("error: %v", err)
		return
	}
}

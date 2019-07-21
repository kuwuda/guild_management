package commands

import (
	"strings"
)

// GetHelp returns a help string which lists commands
func GetHelp() (string, error) {
	s, err := getHelp("help", nil)
	if err != nil {
		return "", err
	}
	return s, nil
}

// I really am not sure what a good way to do all this stuff is
func getHelp(command string, keys []string) (string, error) {
	var help strings.Builder

	switch command {
	case "help":
		help.WriteString("Currently supported commands:\n" +
			"prefix: sets the prefix to a given value \n" +
			"gettable: gets the current table from the database\n" +
			"adduser: adds a user to the database\n" +
			"deluser: deletes a user from the database\n" +
			"addcol: adds a key/column to the database\n" +
			"delcol: deletes a column from the database\n" +
			"incact: increments an activity for the provided users\n")
		help.WriteString("Add -h as an option to any command to get help with it! (EX: " +
			"gettable -h)")
		return help.String(), nil
	case "prefix":
		help.WriteString("usage: " + command + " [options] prefix\n")
		help.WriteString("Options:\n")
		help.WriteString("-h Prints this dialog\n")
		help.WriteString("EX:\n" +
			command + " @")
		return help.String(), nil
	case "gettable":
		help.WriteString("usage: " + command + " [options]\n")
		help.WriteString("Options:\n")
		help.WriteString("-n Specifies the maximum amount of results to return\n" +
			"-u Specifies the specific user to get\n" +
			"-h Prints this dialog\n" +
			"If no options are input, will default to returning every user.")
		return help.String(), nil
	case "adduser":
		help.WriteString("usage: " + command + " [options] name ")
		for _, v := range keys {
			help.WriteRune('"')
			help.WriteString(v)
			help.WriteRune('"')
			help.WriteRune(' ')
		}
		help.WriteString("\nWhere name etc. is replaced with the value for the input\n" +
			"EX:\n" +
			command + " Apples 4 8\n")
		help.WriteString("Any non-specified values will default to 0\n Extraneous values are ignored")
		help.WriteString("Options:\n " +
			"-h Prints this message")
		return help.String(), nil
	case "addcol":
		help.WriteString("usage: " + command + " [options] key")
		help.WriteString("\nWhere key is the name of the column being created\n" +
			"EX:\n" +
			command + " Cats\n")
		help.WriteString("Options:\n " +
			"-h Prints this message")
		return help.String(), nil
	case "delcol":
		help.WriteString("usage: " + command + " [options] key")
		help.WriteString("\nWhere key is the name of the column being deleted\n" +
			"EX:\n" +
			command + " Cats\n")
		help.WriteString("Options:\n " +
			"-h Prints this message")
		return help.String(), nil
	case "incact":
		help.WriteString("usage: incact [options] activity members")
		help.WriteString("\nWhere activity is the name of the activity being incremented, and" +
			"members is the total list of users who are to be incremented\n" +
			"EX:\n" +
			command + " -n 4 kraken kusu apples tyny\n" +
			"Would increase the activity \"kraken\" for all users by 4")
		help.WriteString("Options:\n " +
			"-h Prints this message\n" +
			"-n Specify the amount to increment by (defaults to 1)")
		return help.String(), nil
	}
	return "", nil
}

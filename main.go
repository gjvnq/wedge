package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mgutz/str"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const UNSET_STR = "\tUNSET\n"

var DB *sql.DB

var completer = readline.NewPrefixCompleter(
	readline.PcItem("exit"),
	readline.PcItem("account",
		readline.PcItem("show"),
		readline.PcItem("add"),
		readline.PcItem("edit"),
		readline.PcItem("del"),
	),
	readline.PcItem("asset",
		readline.PcItem("value",
			readline.PcItem("show"),
			readline.PcItem("add"),
			readline.PcItem("edit"),
			readline.PcItem("del")),
		readline.PcItem("kind",
			readline.PcItem("show"),
			readline.PcItem("add"),
			readline.PcItem("edit"),
			readline.PcItem("del")),
	),
)

func getHistoryFile() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(usr.HomeDir, ".wedge_history")
}

func jsonFromFile(filename string, v interface{}) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(dat, v)
}

func jsonToFile(filename string, v interface{}) error {
	dat, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, dat, os.FileMode(int(0555)))
}

func set_str(source string, destination *string) {
	if source != UNSET_STR {
		*destination = source
	}
}

func main() {
	// Preapre readline
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
		HistoryFile:       getHistoryFile(),
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Open database
	fmt.Println("Opening database...")
	DB, err = sql.Open("sqlite3", "wedge.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()
	EnsureTables(DB)
	fmt.Println("Database ready")

	// Prepare stuff
	account_prep()

	for {
		raw_line, err := l.Readline()
		// Basic parsing
		line := str.ToArgv(raw_line)
		err_str := ""
		if err != nil {
			err_str = err.Error()
		}
		// Avoid out of range errors
		for len(line) < 10 {
			line = append(line, "")
		}
		// Interpret
		switch {
		case line[0] == "exit" || err_str == "EOF":
			os.Exit(0)
		case line[0] == "account" && line[1] == "show":
			account_show(line[2:])
		case line[0] == "account" && line[1] == "add":
			account_add(line[2:])
		case line[0] == "account" && line[1] == "edit":
			account_edit(line[2:])
		case line[0] == "account" && line[1] == "del":
			account_del(line[2:])
		default:
			fmt.Printf("Unknown command: %+v Additional error: %+v\n", line, err)
		}
	}
}

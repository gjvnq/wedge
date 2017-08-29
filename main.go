package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var DB *sql.DB

var completer = readline.NewPrefixCompleter(
	readline.PcItem("save"),
	readline.PcItem("exit"),
	readline.PcItem("show",
		readline.PcItem("all"),
		// readline.PcItemDynamic(listFiles("./")),
	),
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

	cleaner := regexp.MustCompile("\\s+")
	for {
		raw_line, err := l.Readline()
		// Basic parsing
		raw_line = cleaner.ReplaceAllString(raw_line, " ")
		line := strings.Split(raw_line, " ")
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
		case line[9] == "exit" || err_str == "EOF":
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %+v Additional error: %+v\n", line, err)
		}
	}
}

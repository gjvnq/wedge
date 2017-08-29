package main

import (
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	"io/ioutil"
	"log"
	"os"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os/user"
	"path/filepath"
)

var DB *sql.DB

var completer = readline.NewPrefixCompleter(
	readline.PcItem("save"),
	readline.PcItem("exit"),
	readline.PcItem("show",
		readline.PcItem("all"),
		// readline.PcItemDynamic(listFiles("./")),
	),
	readline.PcItem("set",
		readline.PcItem("account"),
		readline.PcItem("transaction"),
		readline.PcItem("currency"),
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

	for {
		line, err := l.Readline()
		err_str := ""
		if err != nil {
			err_str = err.Error()
		}
		switch {
		case line == "exit" || err_str == "EOF":
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %+v Additional error: %+v\n", line, err)
		}
	}
}

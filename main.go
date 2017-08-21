package main

import (
	"database/sql"
	"fmt"
	"github.com/chzyer/readline"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os/user"
	"path/filepath"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("save"),
	readline.PcItem("exit"),
	readline.PcItem("show",
		readline.PcItem("all"),
		// readline.PcItemDynamic(listFiles("./")),
	),
	readline.PcItem("add"),
)

func getHistoryFile() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(usr.HomeDir, ".wedge_history")
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
		panic(err)
	}
	defer l.Close()

	// Open database
	db, err := sql.Open("sqlite3", "wedge.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	schema_handler(db)

	line, err := l.Readline()
	fmt.Println(line, err)
}

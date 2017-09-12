package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/chzyer/readline"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mgutz/str"
)

const UNSET_STR = "\tUNSET\n"

var DB *sql.DB
var GlobalLine *readline.Instance
var LocalLine *readline.Instance
var NotImplementedErr = errors.New("Not Implemented")

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
	var err error
	// Preapre readline
	GlobalLine, err = readline.NewEx(&readline.Config{
		Prompt:            "» ",
		HistoryFile:       getHistoryFile(),
		AutoComplete:      Completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	// Preapre readline
	LocalLine, err = readline.NewEx(&readline.Config{
		Prompt:          "» ",
		HistoryLimit:    -1,
		InterruptPrompt: "^C",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer LocalLine.Close()

	// Open database
	filename := "wedge.db"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	fmt.Println("Opening database...")
	fmt.Println("  Filename: " + filename)
	DB, err = sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()
	EnsureTables(DB)
	fmt.Println("Database ready")

	for {
		GlobalLine.SetPrompt("\033[31m»\033[0m ")
		raw_line, err := GlobalLine.Readline()
		// Basic parsing
		line := str.ToArgv(raw_line)
		err_str := ""
		if err != nil {
			err_str = err.Error()
		}
		// Interpret
		switch {
		case len(line) == 0 && err_str != "EOF":
			continue
		case len(line) == 0 && err_str == "EOF":
			os.Exit(0)
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
		case line[0] == "asset" && line[1] == "kind" && line[2] == "show":
			asset_kind_show(line[3:])
		case line[0] == "asset" && line[1] == "kind" && line[2] == "add":
			asset_kind_add(line[3:])
		case line[0] == "asset" && line[1] == "kind" && line[2] == "edit":
			asset_kind_edit(line[3:])
		case line[0] == "asset" && line[1] == "kind" && line[2] == "del":
			asset_kind_del(line[3:])
		case line[0] == "asset" && line[1] == "value" && line[2] == "show":
			asset_value_show(line[3:])
		case line[0] == "asset" && line[1] == "value" && line[2] == "add":
			asset_value_add(line[3:])
		case line[0] == "asset" && line[1] == "value" && line[2] == "edit":
			asset_value_edit(line[3:])
		case line[0] == "asset" && line[1] == "value" && line[2] == "del":
			asset_value_del(line[3:])
		case line[0] == "transaction" && line[1] == "show":
			transaction_show(line[2:])
		case line[0] == "transaction" && line[1] == "add":
			transaction_add(line[2:])
		case line[0] == "transaction" && line[1] == "edit":
			transaction_edit(line[2:])
		case line[0] == "transaction" && line[1] == "del":
			transaction_del(line[2:])
		case line[0] == "transaction" && line[1] == "part" && line[2] == "show":
			transaction_part_show(line[3:])
		case line[0] == "transaction" && line[1] == "part" && line[2] == "add":
			transaction_part_add(line[3:])
		case line[0] == "transaction" && line[1] == "part" && line[2] == "edit":
			transaction_part_edit(line[3:])
		case line[0] == "transaction" && line[1] == "part" && line[2] == "del":
			transaction_part_del(line[3:])
		case line[0] == "transaction" && line[1] == "item" && line[2] == "show":
			transaction_item_show(line[3:])
		case line[0] == "transaction" && line[1] == "item" && line[2] == "add":
			transaction_item_add(line[3:])
		case line[0] == "transaction" && line[1] == "item" && line[2] == "edit":
			transaction_item_edit(line[3:])
		case line[0] == "transaction" && line[1] == "item" && line[2] == "del":
			asset_value_del(line[3:])
		default:
			fmt.Printf("Unknown command: %+v Additional error: %+v\n", line, err)
		}
	}
}

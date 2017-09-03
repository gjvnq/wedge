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
	"strings"
)

const UNSET_STR = "\tUNSET\n"

var DB *sql.DB
var GlobalLine *readline.Instance
var LocalLine *readline.Instance

var PcItemAccount = readline.PcItemDynamic(CompleteAccountFunc)
var PcItemAssetValue = readline.PcItemDynamic(CompleteAssetValueFunc)
var PcItemAssetKind = readline.PcItemDynamic(CompleteAssetKindFunc)
var CompleterAccount = readline.NewPrefixCompleter(PcItemAccount)
var CompleterAssetValue = readline.NewPrefixCompleter(PcItemAssetValue)
var CompleterAssetKind = readline.NewPrefixCompleter(PcItemAssetKind)
var Completer = readline.NewPrefixCompleter(
	readline.PcItem("exit"),
	readline.PcItem("account",
		readline.PcItem("show", PcItemAccount),
		readline.PcItem("add"),
		readline.PcItem("edit", PcItemAccount),
		readline.PcItem("del", PcItemAccount)),
	readline.PcItem("asset",
		readline.PcItem("value",
			readline.PcItem("show", PcItemAssetValue),
			readline.PcItem("add"),
			readline.PcItem("edit", PcItemAssetValue),
			readline.PcItem("del", PcItemAssetValue)),
		readline.PcItem("kind",
			readline.PcItem("show", PcItemAssetKind),
			readline.PcItem("add"),
			readline.PcItem("edit", PcItemAssetKind),
			readline.PcItem("del", PcItemAssetKind))))

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

func ask_user(line *readline.Instance, prompt string, what string) (string, error) {
	line.SetPrompt(prompt)
	return line.ReadlineWithDefault(what)
}

func must_ask_user(line *readline.Instance, prompt string, what string) string {
	line.SetPrompt(prompt)
	s, err := line.ReadlineWithDefault(what)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(s)
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
	fmt.Println("Opening database...")
	DB, err = sql.Open("sqlite3", "wedge.db")
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
		// case line[0] == "asset" && line[1] == "value" && line[2] == "edit":
		// 	asset_value_edit(line[3:])
		// case line[0] == "asset" && line[1] == "value" && line[2] == "del":
		// 	asset_value_del(line[3:])
		default:
			fmt.Printf("Unknown command: %+v Additional error: %+v\n", line, err)
		}
	}
}

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

var PcItemAccount = readline.PcItemDynamic(CompleteAccountFunc)
var PcItemAssetValue = readline.PcItemDynamic(CompleteAssetValueFunc)
var PcItemAssetKind = readline.PcItemDynamic(CompleteAssetKindFunc)
var PcItemTransaction = readline.PcItemDynamic(CompleteTransactionFunc)
var PcItemTransactionPart = readline.PcItemDynamic(CompleteTransactionPartFunc)
var PcItemTransactionItem = readline.PcItemDynamic(CompleteTransactionItemFunc)
var PcItemTransactionStatus = readline.PcItemDynamic(CompleteTransactionStatusFunc)
var CompleterAccount = readline.NewPrefixCompleter(PcItemAccount)
var CompleterAssetValue = readline.NewPrefixCompleter(PcItemAssetValue)
var CompleterAssetKind = readline.NewPrefixCompleter(PcItemAssetKind)
var CompleterTransaction = readline.NewPrefixCompleter(PcItemTransaction)
var CompleterTransactionPart = readline.NewPrefixCompleter(PcItemTransactionPart)
var CompleterTransactionItem = readline.NewPrefixCompleter(PcItemTransactionItem)
var CompleterTransactionStatus = readline.NewPrefixCompleter(PcItemTransactionStatus)
var CompleterEmpty = readline.NewPrefixCompleter()
var Completer = readline.NewPrefixCompleter(
	readline.PcItem("exit"),
	readline.PcItem("timeline",
		readline.PcItem("summary"),
		readline.PcItem("plot")),
	readline.PcItem("account",
		readline.PcItem("show", PcItemAccount),
		readline.PcItem("add"),
		readline.PcItem("edit", PcItemAccount),
		readline.PcItem("del", PcItemAccount),
		readline.PcItem("balance", PcItemAccount)),
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
			readline.PcItem("del", PcItemAssetKind))),
	readline.PcItem("transaction",
		readline.PcItem("show", PcItemTransaction),
		readline.PcItem("add"),
		readline.PcItem("del", PcItemTransaction),
		readline.PcItem("edit", PcItemTransaction),
		readline.PcItem("part",
			readline.PcItem("show", PcItemTransactionPart),
			readline.PcItem("add"),
			readline.PcItem("del", PcItemTransactionPart),
			readline.PcItem("edit", PcItemTransactionPart),
			readline.PcItem("show-by-account", PcItemAccount)),
		readline.PcItem("item",
			readline.PcItem("show", PcItemTransactionItem),
			readline.PcItem("add"),
			readline.PcItem("del", PcItemTransactionItem),
			readline.PcItem("edit", PcItemTransactionItem))))

const DATE_FMT = "2006-01-02-15:04:05-MST"
const DATE_FMT_SPACES = "2006-01-02 15:04:05 MST"
const DAY_FMT = "2006-01-02"
const MONTH_FMT = "2006-01"
const YEAR_FMT = "2006"

func EndOfDay(input time.Time) time.Time {
	y, m, d := input.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, time.Local)
}

func EndOfMonth(input time.Time) time.Time {
	ans := input.AddDate(0, 1, 0)
	for ans.Day() < 28 {
		ans = ans.AddDate(0, 0, -1)
	}
	y, m, d := ans.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, time.Local)
}

func EndOfYear(input time.Time) time.Time {
	ans := input.AddDate(1, 0, 0)
	for ans.Day() < 28 {
		ans = ans.AddDate(0, 0, -1)
	}
	y, m, d := ans.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, time.Local)
}

func True(s string) bool { return true }

func IsFloat(s string) bool {
	val := 3.14
	n, err := fmt.Sscanf(s, "%f", &val)
	return n == 1 && err == nil
}

func IsInt(s string) bool {
	val := 3
	n, err := fmt.Sscanf(s, "%d", &val)
	return n == 1 && err == nil
}

func IsDay(s string) bool {
	_, err := time.Parse(DAY_FMT, s)
	return err == nil
}

func IsAccount(s string) bool {
	n := 0
	err := DB.QueryRow("SELECT COUNT() FROM `Account` WHERE `Id` = ?", s).
		Scan(&n)
	return n > 0 && err == nil
}

func IsAccountOrEmpty(s string) bool {
	if s == "" {
		return true
	}
	return IsAccount(s)
}

func IsTransaction(s string) bool {
	n := 0
	err := DB.QueryRow("SELECT COUNT() FROM `Transaction` WHERE `Id` = ?", s).
		Scan(&n)
	return n > 0 && err == nil
}

func IsAssetKind(s string) bool {
	n := 0
	err := DB.QueryRow("SELECT COUNT() FROM `AssetKind` WHERE `Id` = ?", s).
		Scan(&n)
	return n > 0 && err == nil
}

func IsAssetKindOrEmpty(s string) bool {
	if s == "" {
		return true
	}
	return IsAssetKind(s)
}

func ToBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "yes" || s == "on" || s == "1" || s == "y"
}

func IsBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "yes" || s == "on" || s == "1" || s == "y" || s == "false" || s == "no" || s == "off" || s == "0" || s == "n"
}

func ask_user(line *readline.Instance, prompt string, what string, completer *readline.PrefixCompleter, validator func(string) bool) string {
	for {
		line.SetPrompt(prompt)
		set_completer(line, completer)
		s, err := line.ReadlineWithDefault(what)
		if err != nil {
			log.Fatal(err)
		}
		s = strings.TrimSpace(s)
		if validator(s) {
			return s
		}
		fmt.Println("Invalid information, try again.")
	}
}

func set_completer(line *readline.Instance, completer *readline.PrefixCompleter) {
	if completer == nil {
		completer = CompleterEmpty
	}
	line.Config.AutoComplete = completer
}

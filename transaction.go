package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	. "github.com/logrusorgru/aurora"
	"github.com/mgutz/str"
	"github.com/satori/go.uuid"
)

const (
	TS_UNSET     = ""
	TS_PLANNED   = "P"
	TS_SCHEDULED = "S" // E.g.: transações futuras já programadas
	TS_ON_GOING  = "G"
	TS_FINISHED  = "F"
	TS_CANCELED  = "C"
)

type Transaction struct {
	Id          string
	Name        string
	Desc        string
	RefTimeSpan TimePeriod
	Parts       []TransactionPart
	Items       []TransactionItem
	Tags        map[string]bool
}

func NewTransaction() *Transaction {
	tr := Transaction{}
	tr.Init()
	return &tr
}

func (tr Transaction) MultilineString() string {
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("    Id:"), tr.Id)
	s += fmt.Sprintf("%s %s\n", Bold("  Name:"), tr.Name)
	s += fmt.Sprintf("%s %s\n", Bold("  Desc:"), tr.Desc)
	s += fmt.Sprintf("%s %s\n", Bold("Period:"), tr.RefTimeSpan.String())
	s += fmt.Sprintf("------------------------------ %s -------------------------------\n", Bold("Transaction Parts"))
	for _, tp := range tr.Parts {
		s += tp.ANSIString() + "\n"
	}
	s += fmt.Sprintf("------------------------------ %s -------------------------------\n", Bold("Transaction Items"))
	for _, ti := range tr.Items {
		s += ti.ANSIString() + "\n"
	}
	return s
}

func (tr *Transaction) Init() {
	if tr.Id == "" {
		tr.Id = uuid.NewV4().String()
	}
	if tr.Parts == nil {
		tr.Parts = make([]TransactionPart, 0)
	}
	if tr.Items == nil {
		tr.Items = make([]TransactionItem, 0)
	}
	if tr.Tags == nil {
		tr.Tags = make(map[string]bool)
	}
}

func (tr *Transaction) Load(id string) error {
	var start, end int64
	// Load basic info
	tr.Init()
	err := DB.QueryRow("SELECT `Id`, `Name`, `Desc`, `RefStart`, `RefEnd` FROM `Transaction` WHERE `Id` = ?", id).
		Scan(&tr.Id, &tr.Name, &tr.Desc, &start, &end)
	tr.RefTimeSpan.Start = time.Unix(start, 0)
	tr.RefTimeSpan.End = time.Unix(end, 0)
	if err != nil {
		return err
	}
	err = tr.load_parts()
	if err != nil {
		return err
	}
	err = tr.load_items()
	if err != nil {
		return err
	}
	return nil
}

func (tr *Transaction) load_parts() error {
	rows, err := DB.Query("SELECT `Id` FROM `TransactionPart` WHERE `TransactionId` = ?", tr.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		tp := TransactionPart{}
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		err = tp.Load(id)
		if err != nil {
			log.Fatal(err)
		}
		tr.Parts = append(tr.Parts, tp)
	}
	return nil
}

func (tr *Transaction) load_items() error {
	rows, err := DB.Query("SELECT `Id` FROM `TransactionItem` WHERE `TransactionId` = ?", tr.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		ti := TransactionItem{}
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		err = ti.Load(id)
		if err != nil {
			log.Fatal(err)
		}
		tr.Items = append(tr.Items, ti)
	}
	return nil
}

func (tr Transaction) Del(id string) error {
	tr.Init()
	_, err := DB.Exec("DELETE FROM `Transaction` WHERE `Id` = ?", id)
	if err != nil {
		return err
	}
	_, err = DB.Exec("DELETE FROM `TransactionPart` WHERE `TransactionId` = ?", id)
	if err != nil {
		return err
	}
	_, err = DB.Exec("DELETE FROM `TransactionItem` WHERE `TransactionId` = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func (tr *Transaction) Save() error {
	tr.Init()
	_, err := DB.Exec("INSERT INTO `Transaction` (`Id`, `Name`, `Desc`, `RefStart`, `RefEnd`) VALUES (?, ?, ?, ?, ?)",
		tr.Id,
		tr.Name,
		tr.Desc,
		tr.RefTimeSpan.Start.Unix(),
		tr.RefTimeSpan.End.Unix())
	if err != nil {
		return err
	}
	return tr.Update()
}

func (tr *Transaction) Update() error {
	tr.Init()
	_, err := DB.Exec("UPDATE `Transaction` SET `Name` = ?, `Desc` = ?, `RefStart` = ?, `RefEnd` = ? WHERE `Id` = ?",
		tr.Name,
		tr.Desc,
		tr.RefTimeSpan.Start.Unix(),
		tr.RefTimeSpan.End.Unix(),
		tr.Id)
	if err != nil {
		return err
	}
	err = tr.UpdateParts()
	if err != nil {
		return err
	}
	err = tr.UpdateItems()
	if err != nil {
		return err
	}
	return nil
}

func (tr *Transaction) UpdateParts() error {
	tr.Init()
	// First, delete all
	_, err := DB.Exec("DELETE FROM `TransactionPart` WHERE `TransactionId` = ?", tr.Id)
	if err != nil {
		return err
	}
	// Now let us add them back
	for _, tp := range tr.Parts {
		err = tp.Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tr *Transaction) UpdateItems() error {
	tr.Init()
	// First, delete all
	_, err := DB.Exec("DELETE FROM `TransactionItem` WHERE `TransactionId` = ?", tr.Id)
	if err != nil {
		return err
	}
	// Now let us add them back
	for _, ti := range tr.Items {
		err = ti.Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func transaction_add(line []string) {
	var err error
	tr := NewTransaction()
	// Ask user for basic info
	tr.Name = ask_user(
		LocalLine,
		Sprintf(Bold("  Name: ")),
		"",
		nil,
		True)
	tr.Desc = ask_user(
		LocalLine,
		Sprintf(Bold("  Desc: ")),
		"",
		nil,
		True)
	period := ask_user(
		LocalLine,
		Sprintf(Bold("Period: ")),
		"",
		nil,
		func(s string) bool {
			_, err := ParseTimePeriod(s)
			return err == nil
		})

	// Parse stuff
	tr.RefTimeSpan, err = ParseTimePeriod(period)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Ask user for transaction parts
	for {
		flag := ToBool(ask_user(
			LocalLine,
			Sprintf(Bold("Add transaction part? [y/n] ")),
			"",
			nil,
			IsBool))
		if flag != true {
			break
		}
		// Ask transaction part details
		tp := NewTransactionPart()
		tp.TransactionId = tr.Id
		tp.AccountId = ask_user(
			LocalLine,
			Sprintf(Bold("AccountId: ")),
			"",
			CompleterAccount,
			IsAccount)
		tp.AssetKindId = ask_user(
			LocalLine,
			Sprintf(Bold("Asset: ")),
			"",
			CompleterAssetKind,
			IsAssetKind)
		val_str := ask_user(
			LocalLine,
			Sprintf(Bold("Value: ")),
			"",
			nil,
			IsFloat)
		schdul := ask_user(
			LocalLine,
			Sprintf(Bold("Scheduled for: ")),
			"",
			nil,
			IsDay)
		actual := ask_user(
			LocalLine,
			Sprintf(Bold("Actual date: ")),
			schdul,
			nil,
			IsDay)
		status := ask_user(
			LocalLine,
			Sprintf(Bold("Status: ")),
			"",
			CompleterTransactionStatus,
			func(s string) bool {
				tp := NewTransactionPart()
				return tp.SetStatus(s) == nil
			})
		tp.SetValue(val_str)
		tp.SetDates(schdul, actual)
		tp.SetStatus(status)
		tr.Parts = append(tr.Parts, *tp)
	}
	// Ask user for transaction items
	for {
		flag := ToBool(ask_user(
			LocalLine,
			Sprintf(Bold("Add transaction item? [y/n] ")),
			"",
			nil,
			IsBool))
		if flag != true {
			break
		}
		// Ask transaction part details
		ti := NewTransactionItem()
		ti.TransactionId = tr.Id
		ti.Name = ask_user(
			LocalLine,
			Sprintf(Bold("Name: ")),
			"",
			nil,
			True)
		ti.AssetKindId = ask_user(
			LocalLine,
			Sprintf(Bold("AssetId: ")),
			"",
			CompleterAssetKind,
			IsAssetKind)
		tot_str := ask_user(
			LocalLine,
			Sprintf(Bold("TotalCost: ")),
			"",
			nil,
			IsFloat)
		ti.Quantity = str.ToFloatOr(ask_user(
			LocalLine,
			Sprintf(Bold("Quantity: ")),
			"",
			nil,
			IsFloat), 0)
		guess := fmt.Sprintf("%f", str.ToFloatOr(tot_str, 0)/ti.Quantity)
		uni_str := ask_user(
			LocalLine,
			Sprintf(Bold("UnitCost: ")),
			guess,
			nil,
			IsFloat)
		ti.SetTotalCost(tot_str)
		ti.SetUnitCost(uni_str)
		tr.Items = append(tr.Items, *ti)
	}
	// Save
	err = tr.Save()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func transaction_edit(line []string) {
	var err error
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	// Load transaction
	tr := NewTransaction()
	err = tr.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Ask user for basic info
	tr.Name = ask_user(
		LocalLine,
		Sprintf(Bold("  Name: ")),
		tr.Name,
		nil,
		True)
	tr.Desc = ask_user(
		LocalLine,
		Sprintf(Bold("  Desc: ")),
		tr.Desc,
		nil,
		True)
	period := ask_user(
		LocalLine,
		Sprintf(Bold("Period: ")),
		tr.RefTimeSpan.StringDay(),
		nil,
		func(s string) bool {
			_, err := ParseTimePeriod(s)
			return err == nil
		})

	// Parse stuff
	tr.RefTimeSpan, err = ParseTimePeriod(period)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Save
	err = tr.Update()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
func transaction_show(line []string) {
	spec := ""
	if len(line) > 0 {
		spec = line[0]
	}

	tr := NewTransaction()
	err := tr.Load(spec)
	if err == nil {
		fmt.Printf(tr.MultilineString())
		return
	}

	rows, err := DB.Query("SELECT `Id`, `Name`, `RefStart`, `RefEnd` FROM `Transaction` WHERE `Name` LIKE '%%"+spec+"%%' OR ? = '' LIMIT 64", spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read Stuff
	for rows.Next() {
		var id, name string
		var start_int, end_int int64
		err := rows.Scan(&id, &name, &start_int, &end_int)
		if err != nil {
			log.Fatal(err)
		}
		start := time.Unix(start_int, 0)
		end := time.Unix(end_int, 0)
		tmp_id := Sprintf(Gray(id))
		tmp_name := Sprintf(Bold(fmt.Sprintf("%-19.19s", name)))
		fmt.Printf("%s %s %10s - %10s\n", tmp_id, tmp_name, start.Format(DAY_FMT), end.Format(DAY_FMT))
	}
}

func transaction_del(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	id := line[len(line)-1]
	deleter(id, NewTransaction())
}

func CompleteTransactionFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `Transaction` WHERE `Id` LIKE '"+spec+"%%' OR ? = '' LIMIT 64", spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read accounts
	found := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		s := ""
		err := rows.Scan(&s)
		if err != nil {
			log.Fatal(err)
		}
		found = append(found, s)
	}
	return found
}

func CompleteTransactionStatusFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	statuses := []string{
		"TS_UNSET",
		"TS_PLANNED",
		"TS_SCHEDULED",
		"TS_ON_GOING",
		"TS_FINISHED",
		"TS_CANCELED"}
	ret := make([]string, 0)
	for _, status := range statuses {
		if strings.HasPrefix(status, spec) {
			ret = append(ret, status)
		}
	}
	return ret
}

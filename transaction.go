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
	Itens       []TransactionItem
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
	return s
}

func (tr *Transaction) Init() {
	if tr.Id == "" {
		tr.Id = uuid.NewV4().String()
	}
	if tr.Parts == nil {
		tr.Parts = make([]TransactionPart, 0)
	}
	if tr.Itens == nil {
		tr.Itens = make([]TransactionItem, 0)
	}
	if tr.Tags == nil {
		tr.Tags = make(map[string]bool)
	}
}

func (tr *Transaction) Load(id string) error {
	var start, end int64

	tr.Init()
	err := DB.QueryRow("SELECT `Id`, `Name`, `Desc`, `RefStart`, `RefEnd` FROM `Transaction` WHERE `Id` = ?", id).
		Scan(&tr.Id, &tr.Name, &tr.Desc, &start, &end)
	tr.RefTimeSpan.Start = time.Unix(start, 0)
	tr.RefTimeSpan.End = time.Unix(end, 0)
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
	err = tr.UpdateItens()
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
	if len(tr.Parts) == 0 {
		// Bit let us be lazy first :)
		return nil
	}
	for _, tp := range tr.Parts {
		err = tp.Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tr *Transaction) UpdateItens() error {
	tr.Init()
	// First, delete all
	_, err := DB.Exec("DELETE FROM `TransactionItem` WHERE `TransactionId` = ?", tr.Id)
	if err != nil {
		return err
	}
	// Now let us add them back
	if len(tr.Itens) == 0 {
		// Bit let us be lazy first :)
		return nil
	}
	for _, ti := range tr.Itens {
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
	// Ask user for transaction itens
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
		tr.Itens = append(tr.Itens, *ti)
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
	rows, err := DB.Query("SELECT `Id` FROM `Transaction` WHERE `Id` = ? OR ? = '' LIMIT 64", spec, spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read raw data
	tmp := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		tmp = append(tmp, id)
	}
	rows.Close()
	// Load data
	trs := make([]Transaction, 0)
	for _, id := range tmp {
		tr := Transaction{}
		err := tr.Load(id)
		if err != nil {
			log.Fatal(err)
		}
		trs = append(trs, tr)
	}
	if len(trs) == 1 {
		fmt.Printf(trs[0].MultilineString())
		return
	}
	fmt.Printf("%36s | %15.15s | %10s | %10s\n", "Id", "Name", "Start", "End")
	fmt.Println("-------------------------------------+-----------------+------------+-----------")
	for _, tr := range trs {
		fmt.Printf("%36s | %-15.15s | %10s | %10s\n", tr.Id, tr.Name, tr.RefTimeSpan.Start.Format(DAY_FMT), tr.RefTimeSpan.End.Format(DAY_FMT))
	}
}

func transaction_del(line []string) {
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

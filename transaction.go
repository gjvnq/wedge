package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	. "github.com/logrusorgru/aurora"
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

type TimePeriod struct {
	Start time.Time
	End   time.Time
}

func (p TimePeriod) IsOneDay() bool {
	return p.Start.Equal(p.End) && p.Start.IsZero() == false
}

// Valid formats: '{DAY_FMT}' and '{DAY_FMT} - {DAY_FMT}'
func ParseTimePeriod(input string) (TimePeriod, error) {
	date, err := time.Parse(DAY_FMT, input)
	if err == nil {
		return TimePeriod{Start: date, End: date}, nil
	}
	parts := strings.Split(input, " - ")
	if len(parts) != 2 {
		log.Println("Failed to parse time period: " + input)
		log.Println("Too many or too few parts:", parts)
		return TimePeriod{}, errors.New("failed to parse: " + input)
	}
	start, err := time.Parse(DAY_FMT, parts[0])
	if err != nil {
		log.Println("Failed to parse time period: " + input)
		log.Println("Failed to parse time " + parts[0] + ": " + err.Error())
		return TimePeriod{}, errors.New("failed to parse: " + parts[0])
	}
	end, err := time.Parse(DAY_FMT, parts[1])
	if err != nil {
		log.Println("Failed to parse time period: " + input)
		log.Println("Failed to parse time " + parts[1] + ": " + err.Error())
		return TimePeriod{}, errors.New("failed to parse: " + parts[1])
	}
	return TimePeriod{Start: start, End: end}, nil
}

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

type TransactionPart struct {
	Id            string
	TransactionId string
	AccountId     string
	Status        string
	ScheduledFor  time.Time
	ActualDate    time.Time
	Value         int
	AssetKindId   int
	Tags          map[string]bool
}

func (tp *TransactionPart) Init() {
	if tp.Id == "" {
		tp.Id = uuid.NewV4().String()
	}
	if tp.Tags == nil {
		tp.Tags = make(map[string]bool)
	}
}

func (tp *TransactionPart) Save() error {
	tp.Init()
	_, err := DB.Exec("INSERT INTO `TransactionPart` (`Id`, `TransactionId`, `AccountId`, `Status`, `ScheduledFor`, `ActualDate`, `Value`, `AssetKindId`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		tp.Id,
		tp.TransactionId,
		tp.AccountId,
		tp.Status,
		tp.ScheduledFor.Unix(),
		tp.ActualDate.Unix(),
		tp.Value,
		tp.AssetKindId)
	return err
}

type TransactionItem struct {
	Id            string
	TransactionId string
	Name          string
	UnitCost      int
	KindId        string
	Quantity      float64
	TotalCost     int
	Tags          map[string]bool
}

func (ti *TransactionItem) Init() {
	if ti.Id == "" {
		ti.Id = uuid.NewV4().String()
	}
	if ti.Tags == nil {
		ti.Tags = make(map[string]bool)
	}
}

func (ti *TransactionItem) Save() error {
	ti.Init()
	_, err := DB.Exec("INSERT INTO `TransactionItem` (`Id`, `TransactionId`, `Name`, `UnitCost`, `KindId`, `Quantity`, `TotalCost`) VALUES (?, ?, ?, ?, ?, ?, ?)",
		ti.Id,
		ti.TransactionId,
		ti.Name,
		ti.UnitCost,
		ti.KindId,
		ti.Quantity,
		ti.TotalCost)
	return err
}

func transaction_add(line []string) {
	var err error
	tr := NewTransaction()
	// Ask user for basic info
	tr.Name = must_ask_user(LocalLine, Sprintf(Bold("  Name: ")), "")
	tr.Desc = must_ask_user(LocalLine, Sprintf(Bold("  Desc: ")), "")
	period := must_ask_user(LocalLine, Sprintf(Bold("Period: ")), "")
	// Parse stuff
	tr.RefTimeSpan, err = ParseTimePeriod(period)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Save
	err = tr.Save()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
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

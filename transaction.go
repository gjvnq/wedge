package main

import (
	"errors"
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

type TimePeriod struct {
	Start time.Time
	End   time.Time
}

func (p TimePeriod) IsOneDay() bool {
	return p.Start.Equal(p.End) && p.Start.IsZero() == false
}

// Valid formats: '2006-01-02', '2006-01-02 2006-01-02', '2006-01', '2006-01 2006-01', '2006' and '2006 2006'
// Note that the end date will be at the end of the smallest time unit specified.
// Ex: '2017' -> '2017-01-01 00:00:00' to '2017-12-31 23:59:59'
// Ex: '2017 2018' -> '2017-01-01 00:00:00' to '2018-12-31 23:59:59'
func ParseTimePeriod(input string) (TimePeriod, error) {
	var err, err1, err2 error
	var start, end, date time.Time

	// Try '2006-01-02'
	date, err = time.Parse(DAY_FMT, input)
	if err == nil {
		return TimePeriod{Start: date, End: date}, nil
	}
	// Try '2006-01'
	start, err = time.Parse(MONTH_FMT, input)
	if err == nil {
		end = EndOfMonth(start)
		return TimePeriod{Start: start, End: date}, nil
	}
	// Try '2006'
	start, err = time.Parse(YEAR_FMT, input)
	if err == nil {
		end = EndOfYear(start)
		return TimePeriod{Start: start, End: date}, nil
	}

	// Try multipart
	parts := strings.Split(input, " ")
	if len(parts) != 2 {
		log.Println("Failed to parse time period: " + input)
		log.Println("Too many or too few parts:", parts)
		return TimePeriod{}, errors.New("failed to parse: " + input)
	}
	// Try '2006-01-02 2006-01-02'
	start, err1 = time.Parse(DAY_FMT, parts[0])
	end, err2 = time.Parse(DAY_FMT, parts[1])
	if err1 == nil && err2 == nil {
		return TimePeriod{Start: start, End: end}, nil
	}
	// Try '2006-01 2006-01'
	start, err1 = time.Parse(MONTH_FMT, parts[0])
	end, err2 = time.Parse(MONTH_FMT, parts[1])
	if err1 == nil && err2 == nil {
		end = EndOfMonth(start)
		return TimePeriod{Start: start, End: end}, nil
	}
	// Try '2006 2006'
	start, err1 = time.Parse(YEAR_FMT, parts[0])
	end, err2 = time.Parse(YEAR_FMT, parts[1])
	if err1 == nil && err2 == nil {
		end = EndOfYear(start)
		return TimePeriod{Start: start, End: end}, nil
	}

	// Everything failed
	return TimePeriod{}, errors.New("failed to parse: " + input)
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
	AssetKindId   string
	Tags          map[string]bool
}

func NewTransactionPart() *TransactionPart {
	tp := TransactionPart{}
	tp.Init()
	return &tp
}

func (tp *TransactionPart) Init() {
	if tp.Id == "" {
		tp.Id = uuid.NewV4().String()
	}
	if tp.Tags == nil {
		tp.Tags = make(map[string]bool)
	}
}

func (tp *TransactionPart) SetValue(input string) error {
	var err error
	tp.Value, err = full_decimal_parse(input, tp.AssetKindId)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (tp *TransactionPart) SetDates(scheduled_str, actual_str string) error {
	var err error
	tp.ScheduledFor, err = time.Parse(DAY_FMT, scheduled_str)
	if err != nil {
		return err
	}
	tp.ActualDate, err = time.Parse(DAY_FMT, actual_str)
	if err != nil {
		return err
	}
	return nil
}

func (tp *TransactionPart) SetStatus(input string) error {
	switch input {
	case TS_UNSET:
		tp.Status = TS_UNSET
	case "TS_UNSET":
		tp.Status = TS_UNSET
	case TS_CANCELED:
		tp.Status = TS_CANCELED
	case "TS_CANCELED":
		tp.Status = TS_CANCELED
	case TS_FINISHED:
		tp.Status = TS_FINISHED
	case "TS_FINISHED":
		tp.Status = TS_FINISHED
	case TS_SCHEDULED:
		tp.Status = TS_SCHEDULED
	case "TS_SCHEDULED":
		tp.Status = TS_SCHEDULED
	case TS_PLANNED:
		tp.Status = TS_PLANNED
	case "TS_PLANNED":
		tp.Status = TS_PLANNED
	case TS_ON_GOING:
		tp.Status = TS_ON_GOING
	case "TS_ON_GOING":
		tp.Status = TS_ON_GOING
	default:
		return errors.New("failed to parse")
	}
	return nil
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
	AssetKindId   string
	Quantity      float64
	TotalCost     int
	Tags          map[string]bool
}

func NewTransactionItem() *TransactionItem {
	ti := TransactionItem{}
	ti.Init()
	return &ti
}

func (ti *TransactionItem) Init() {
	if ti.Id == "" {
		ti.Id = uuid.NewV4().String()
	}
	if ti.Tags == nil {
		ti.Tags = make(map[string]bool)
	}
}

func (ti *TransactionItem) SetTotalCost(input string) error {
	var err error
	ti.TotalCost, err = full_decimal_parse(input, ti.AssetKindId)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (ti *TransactionItem) SetUnitCost(input string) error {
	var err error
	ti.UnitCost, err = full_decimal_parse(input, ti.AssetKindId)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (ti *TransactionItem) Save() error {
	ti.Init()
	_, err := DB.Exec("INSERT INTO `TransactionItem` (`Id`, `TransactionId`, `Name`, `UnitCost`, `AssetKindId`, `Quantity`, `TotalCost`) VALUES (?, ?, ?, ?, ?, ?, ?)",
		ti.Id,
		ti.TransactionId,
		ti.Name,
		ti.UnitCost,
		ti.AssetKindId,
		ti.Quantity,
		ti.TotalCost)
	return err
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

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

func (tp *TransactionPart) Load(id string) error {
	var schdul, actual int64

	tp.Init()
	err := DB.QueryRow("SELECT `Id`, `TransactionId`, `AccountId`, `Status`, `ScheduledFor`, `ActualDate`, `Value`, `AssetKindId` FROM `TransactionPart` WHERE `Id` = ?", id).
		Scan(&tp.Id, &tp.TransactionId, &tp.AccountId, &tp.Status, &schdul, &actual, &tp.Value, &tp.AssetKindId)
	tp.ScheduledFor = time.Unix(schdul, 0)
	tp.ActualDate = time.Unix(actual, 0)
	if err != nil {
		return err
	}
	return nil
}

func (tp TransactionPart) Date() string {
	date := tp.ScheduledFor.Format(DAY_FMT)
	if tp.Status == TS_FINISHED {
		date = tp.ActualDate.Format(DAY_FMT)
	}
	return date
}

func (tp TransactionPart) ANSIString() string {
	tmp_num := fmt.Sprintf("%11.11s", tp.ValueToStr())
	tmp_id := Bold(fmt.Sprintf("%3.3s", tp.AssetKindId))
	if tp.Value > 0 {
		tmp_num = Sprintf(Cyan(tmp_num))
	} else {
		tmp_num = Sprintf(Red(tmp_num))
	}
	return fmt.Sprintf("%s %-14.14s %s %10s %s %s", Sprintf(Gray(tp.Id)), tp.AccountId, tp.Status, tp.Date(), tmp_num, tmp_id)
}

func (tp TransactionPart) String() string {
	return fmt.Sprintf("[%s] %s %s %s (%s) %s", tp.Id, tp.AccountId, tp.ValueToStr(), tp.AssetKindId, tp.Status, tp.Date())
}

func (tp TransactionPart) MultilineString() string {
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("           Id:"), tp.Id)
	s += fmt.Sprintf("%s %s\n", Bold("TransactionId:"), tp.TransactionId)
	s += fmt.Sprintf("%s %s\n", Bold("    AccountId:"), tp.AccountId)
	s += fmt.Sprintf("%s %s\n", Bold("        Value:"), tp.ValueToStr())
	s += fmt.Sprintf("%s %s\n", Bold("  AssetKindId:"), tp.AssetKindId)
	s += fmt.Sprintf("%s %s\n", Bold("       Status:"), tp.Status)
	s += fmt.Sprintf("%s %s\n", Bold(" ScheduledFor:"), tp.ScheduledFor.Format(DATE_FMT_SPACES))
	s += fmt.Sprintf("%s %s\n", Bold("   ActualDate:"), tp.ActualDate.Format(DATE_FMT_SPACES))
	return s
}

func (tp *TransactionPart) SetValue(input string) error {
	var err error
	tp.Value, err = full_decimal_parse(input, tp.AssetKindId)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (tp *TransactionPart) ValueToStr() string {
	s, err := full_decimal_fmt(tp.Value, tp.AssetKindId)
	if err != nil {
		log.Println(err)
	}
	return s
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

func (tp *TransactionPart) Update() error {
	tp.Init()
	_, err := DB.Exec("UPDATE `TransactionPart` SET `AccountId` = ?, `Status` = ?, `ScheduledFor` = ?, `ActualDate` = ?, `Value` = ?, `AssetKindId` = ? WHERE `Id` = ?",
		tp.AccountId,
		tp.Status,
		tp.ScheduledFor.Unix(),
		tp.ActualDate.Unix(),
		tp.Value,
		tp.AssetKindId,
		tp.Id)
	return err
}

func (tp TransactionPart) Del(id string) error {
	tp.Init()
	_, err := DB.Exec("DELETE FROM `TransactionPart` WHERE `Id` = ?", id)
	return err
}

func transaction_part_add(line []string) {
	var err error
	tp := NewTransactionPart()
	tp.TransactionId = ask_user(
		LocalLine,
		Sprintf(Bold("TransactionId: ")),
		"",
		CompleterTransaction,
		IsTransaction)
	tp.AccountId = ask_user(
		LocalLine,
		Sprintf(Bold("    AccountId: ")),
		"",
		CompleterAccount,
		IsAccount)
	tp.AssetKindId = ask_user(
		LocalLine,
		Sprintf(Bold("        Asset: ")),
		"",
		CompleterAssetKind,
		IsAssetKind)
	val_str := ask_user(
		LocalLine,
		Sprintf(Bold("        Value: ")),
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
		Sprintf(Bold("  Actual date: ")),
		schdul,
		nil,
		IsDay)
	status := ask_user(
		LocalLine,
		Sprintf(Bold("       Status: ")),
		"",
		CompleterTransactionStatus,
		func(s string) bool {
			tp := NewTransactionPart()
			return tp.SetStatus(s) == nil
		})
	tp.SetValue(val_str)
	tp.SetDates(schdul, actual)
	tp.SetStatus(status)
	// Save
	err = tp.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func transaction_part_edit(line []string) {
	var err error
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	// Load transaction part
	tp := NewTransactionPart()
	err = tp.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tp.TransactionId = ask_user(
		LocalLine,
		Sprintf(Bold("TransactionId: ")),
		tp.TransactionId,
		CompleterTransaction,
		IsTransaction)
	tp.AccountId = ask_user(
		LocalLine,
		Sprintf(Bold("    AccountId: ")),
		tp.AccountId,
		CompleterAccount,
		IsAccount)
	tp.AssetKindId = ask_user(
		LocalLine,
		Sprintf(Bold("        Asset: ")),
		tp.AssetKindId,
		CompleterAssetKind,
		IsAssetKind)
	val_str := ask_user(
		LocalLine,
		Sprintf(Bold("        Value: ")),
		tp.ValueToStr(),
		nil,
		IsFloat)
	schdul := ask_user(
		LocalLine,
		Sprintf(Bold("Scheduled for: ")),
		tp.ScheduledFor.Format(DAY_FMT),
		nil,
		IsDay)
	actual := ask_user(
		LocalLine,
		Sprintf(Bold("  Actual date: ")),
		tp.ActualDate.Format(DAY_FMT),
		nil,
		IsDay)
	status := ask_user(
		LocalLine,
		Sprintf(Bold("       Status: ")),
		"",
		CompleterTransactionStatus,
		func(s string) bool {
			tp := NewTransactionPart()
			return tp.SetStatus(s) == nil
		})
	tp.SetValue(val_str)
	tp.SetDates(schdul, actual)
	tp.SetStatus(status)
	// Save
	err = tp.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func transaction_part_show(line []string) {
	spec := ""
	if len(line) > 0 {
		spec = line[0]
	}

	tp := NewTransactionPart()
	err := tp.Load(spec)
	if err == nil {
		fmt.Printf(tp.MultilineString())
		return
	}

	rows, err := DB.Query("SELECT `Id` FROM `TransactionPart` WHERE `AccountId` = ? OR `Status` = ? OR ? = '' LIMIT 64", spec, spec, spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read Stuff
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		err = tp.Load(id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(tp.ANSIString())
	}
}

func transaction_part_del(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	id := line[len(line)-1]
	deleter(id, NewTransactionPart())
}

func CompleteTransactionPartFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `TransactionPart` WHERE `Id` LIKE '"+spec+"%%' OR ? = '' LIMIT 64", spec)
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

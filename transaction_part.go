package main

import (
	"errors"
	"fmt"
	"log"
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

func (tp TransactionPart) String() string {
	return fmt.Sprintf("[%s] %s %s %s (%s) %s\n", tp.Id, tp.AccountId, tp.ValueToStr(), tp.AssetKindId, tp.Status, tp.Date())
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

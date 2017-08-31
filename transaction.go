package main

import (
	"time"
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

type Transaction struct {
	Id          string
	Name        string
	Desc        string
	Tags        map[string]bool
	RefTimeSpan TimePeriod
	Parts       []TransactionPart
	Itens       []TransactionItem
}

type TransactionPart struct {
	Id            string
	TransactionId string
	AccountId     string
	Status        string
	ScheledFor    time.Time
	ActualDate    time.Time
	Tags          map[string]bool
	Value         int
	KindId        int
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

func (t Transaction) TypeName() string {
	return "Transaction"
}

func (tp TransactionPart) TypeName() string {
	return "TransactionPart"
}

func (ti TransactionItem) TypeName() string {
	return "TransactionItem"
}

func CompleteTransaction(prefix string) []string {
	return []string{"A", "B"}
}

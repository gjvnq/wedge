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
	Movements   []GMov
}

type TransactionPart struct {
	Account    string
	Value      int
	Status     string
	ScheledFor time.Time
	ActualDate time.Time
	Tags       map[string]bool
}

type TransactionItem struct {
	Name    string
	Cost    int
	CostCur string
	Tags    map[string]bool
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

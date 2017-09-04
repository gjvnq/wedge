package main

import (
	"fmt"
	"log"

	. "github.com/logrusorgru/aurora"
	"github.com/satori/go.uuid"
)

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

func (ti *TransactionItem) Load(id string) error {
	ti.Init()
	err := DB.QueryRow("SELECT `Id`, `TransactionId`, `Name`, `UnitCost`, `Quantity`, `TotalCost`, `AssetKindId` FROM `TransactionItem` WHERE `Id` = ?", id).
		Scan(&ti.Id, &ti.TransactionId, &ti.Name, &ti.UnitCost, &ti.Quantity, &ti.TotalCost, &ti.AssetKindId)
	if err != nil {
		return err
	}
	return nil
}

func (ti TransactionItem) String() string {
	return fmt.Sprintf("[%s] %s %s * %f = %s\n", ti.Id, ti.Name, ti.UnitCostToStr(), ti.Quantity, ti.TotalCostToStr())
}

func (ti TransactionItem) MultilineString() string {
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("           Id:"), ti.Id)
	s += fmt.Sprintf("%s %s\n", Bold("TransactionId:"), ti.TransactionId)
	s += fmt.Sprintf("%s %s\n", Bold("         Name:"), ti.Name)
	s += fmt.Sprintf("%s %s\n", Bold("     UnitCost:"), ti.UnitCostToStr())
	s += fmt.Sprintf("%s %f\n", Bold("     Quantity:"), ti.Quantity)
	s += fmt.Sprintf("%s %s\n", Bold("    TotalCost:"), ti.TotalCostToStr())
	return s
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

func (ti *TransactionItem) UnitCostToStr() string {
	s, err := full_decimal_fmt(ti.UnitCost, ti.AssetKindId)
	if err != nil {
		log.Println(err)
	}
	return s
}

func (ti *TransactionItem) TotalCostToStr() string {
	s, err := full_decimal_fmt(ti.TotalCost, ti.AssetKindId)
	if err != nil {
		log.Println(err)
	}
	return s
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

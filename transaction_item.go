package main

import (
	"fmt"
	"log"
	"strings"

	. "github.com/logrusorgru/aurora"
	"github.com/mgutz/str"
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
	return fmt.Sprintf("[%s] %s %s * %f = %s", ti.Id, ti.Name, ti.UnitCostToStr(), ti.Quantity, ti.TotalCostToStr())
}

func (ti TransactionItem) ANSIString() string {
	tmp_id := Bold(fmt.Sprintf("%3.3s", ti.AssetKindId))
	tmp_num := fmt.Sprintf("%11.11s", ti.TotalCostToStr())
	tmp_num = Sprintf(Cyan(tmp_num))
	return fmt.Sprintf("%s %-22.22s %4.1f %s %s", Sprintf(Gray(ti.Id)), ti.Name, ti.Quantity, tmp_num, tmp_id)
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

func (ti *TransactionItem) Update() error {
	ti.Init()
	_, err := DB.Exec("UPDATE `TransactionItem` SET `Name` = ?, `UnitCost` = ?, `AssetKindId` = ?, `Quantity` = ?, `TotalCost` = ? WHERE `Id` = ?",
		ti.Name,
		ti.UnitCost,
		ti.AssetKindId,
		ti.Quantity,
		ti.TotalCost,
		ti.Id)
	return err
}

func (ti TransactionItem) Del(id string) error {
	ti.Init()
	_, err := DB.Exec("DELETE FROM `TransactionItem` WHERE `Id` = ?", id)
	return err
}

func transaction_item_add(line []string) {
	var err error
	// Ask transaction part details
	ti := NewTransactionItem()
	ti.TransactionId = ask_user(
		LocalLine,
		Sprintf(Bold("TransactionId: ")),
		"",
		CompleterTransaction,
		IsTransaction)
	ti.Name = ask_user(
		LocalLine,
		Sprintf(Bold("         Name: ")),
		"",
		nil,
		True)
	ti.AssetKindId = ask_user(
		LocalLine,
		Sprintf(Bold("      AssetId: ")),
		"",
		CompleterAssetKind,
		IsAssetKind)
	tot_str := ask_user(
		LocalLine,
		Sprintf(Bold("    TotalCost: ")),
		"",
		nil,
		IsFloat)
	ti.Quantity = str.ToFloatOr(ask_user(
		LocalLine,
		Sprintf(Bold("     Quantity: ")),
		"",
		nil,
		IsFloat), 0)
	guess := fmt.Sprintf("%f", str.ToFloatOr(tot_str, 0)/ti.Quantity)
	uni_str := ask_user(
		LocalLine,
		Sprintf(Bold("     UnitCost: ")),
		guess,
		nil,
		IsFloat)
	ti.SetTotalCost(tot_str)
	ti.SetUnitCost(uni_str)
	// Save
	err = ti.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func transaction_item_edit(line []string) {
	var err error
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	// Load transaction part
	ti := NewTransactionItem()
	err = ti.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Ask transaction part details
	ti.TransactionId = ask_user(
		LocalLine,
		Sprintf(Bold("TransactionId: ")),
		ti.TransactionId,
		CompleterTransaction,
		IsTransaction)
	ti.Name = ask_user(
		LocalLine,
		Sprintf(Bold("         Name: ")),
		ti.Name,
		nil,
		True)
	ti.AssetKindId = ask_user(
		LocalLine,
		Sprintf(Bold("      AssetId: ")),
		ti.AssetKindId,
		CompleterAssetKind,
		IsAssetKind)
	tot_str := ask_user(
		LocalLine,
		Sprintf(Bold("    TotalCost: ")),
		ti.TotalCostToStr(),
		nil,
		IsFloat)
	ti.Quantity = str.ToFloatOr(ask_user(
		LocalLine,
		Sprintf(Bold("     Quantity: ")),
		fmt.Sprintf("%f", ti.Quantity),
		nil,
		IsFloat), 0)
	uni_str := ask_user(
		LocalLine,
		Sprintf(Bold("     UnitCost: ")),
		ti.UnitCostToStr(),
		nil,
		IsFloat)
	ti.SetTotalCost(tot_str)
	ti.SetUnitCost(uni_str)
	// Save
	err = ti.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func transaction_item_show(line []string) {
}

func transaction_item_del(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	id := line[len(line)-1]
	deleter(id, NewTransactionItem())
}

func CompleteTransactionItemFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `TransactionItem` WHERE `Id` LIKE '"+spec+"%%' OR ? = '' LIMIT 64", spec)
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

package main

import (
	"errors"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"log"
	"strings"
	"time"
)

const DATE_FMT = "2006-01-02-15:04:05-MST"

type AssetValue struct {
	Id      string
	AssetId string
	RefId   string
	Value   int // Value of AssetId in terms of RefId
	Date    time.Time
	Notes   string
}

func (av AssetValue) TypeName() string {
	return "AssetValue"
}

func (av *AssetValue) GenId() {
	av.Id = fmt.Sprintf("%s-in-%s-at-%s",
		av.AssetId,
		av.RefId,
		av.Date.Format(DATE_FMT),
	)
}

func (av AssetValue) Save() error {
	if len(av.Id) <= 0 {
		return errors.New("All asset values must have a non empty id")
	}
	if len(av.AssetId) <= 0 {
		return errors.New("All asset values must have a non empty AssetId")
	}
	if len(av.RefId) <= 0 {
		return errors.New("All asset values must have a non empty RefId")
	}
	_, err := DB.Exec("INSERT INTO `AssetValue` (`Id`, `AssetId`, `RefId`, `Value`, `Date`, `Notes`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", av.Id, av.AssetId, av.RefId, av.Value, av.Date.Unix(), av.Notes)
	return err
}

func (av AssetValue) Update() error {
	_, err := DB.Exec("UPDATE `AssetValue` SET `Value` = ?, `Notes` = ? WHERE `Id` = ?", av.Value, av.Notes, av.Id)
	return err
}

func (av AssetValue) Del(id string) error {
	_, err := DB.Exec("DELETE FROM `AssetValue` WHERE `Id` = ?", id)
	return err
}

func (av *AssetValue) Load(id string) error {
	var tmp int64
	err := DB.QueryRow("SELECT `Id`, `AssetId`, `RefId`, `Value`, `Date`, `Notes` FROM `AssetValue` WHERE `Id` = ?", id).
		Scan(&av.Id, &av.AssetId, &av.RefId, &av.Value, &tmp, &av.Notes)
	av.Date = time.Unix(tmp, 0)
	return err
}

func (av AssetValue) MultilineString() string {
	val_str := fmt.Sprintf("%d (raw/faield to load AssetKind)", av.Value)
	ak := AssetKind{}
	err := ak.Load(av.RefId)
	if err == nil {
		val_str = fmt.Sprintf("1 %s = %s %s", Bold(av.AssetId), fmt_decimal(av.Value, ak.DecimalPlaces), Bold(av.RefId))
	}
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("     Id:"), av.Id)
	s += fmt.Sprintf("%s %s\n", Bold("AssetId:"), av.AssetId)
	s += fmt.Sprintf("%s %s\n", Bold("  RefId:"), av.RefId)
	s += fmt.Sprintf("%s %s\n", Bold("  Value:"), val_str)
	s += fmt.Sprintf("%s %s\n", Bold("   Date:"), av.Date.Format(DATE_FMT))
	s += fmt.Sprintf("%s %s\n", Bold("  Notes:"), av.Notes)
	return s
}

func CompleteAssetValue(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `AssetValue` WHERE `Id` LIKE '"+spec+"%%' OR ? = ''", spec)
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

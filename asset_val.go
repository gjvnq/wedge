package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	. "github.com/logrusorgru/aurora"
)

const DATE_FMT = "2006-01-02-15:04:05-MST"
const DATE_FMT_SPACES = "2006-01-02 15:04:05 MST"
const DAY_FMT = "2006-01-02"

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

func (av *AssetValue) Save() error {
	if len(av.Id) <= 0 {
		av.GenId()
	}
	if len(av.AssetId) <= 0 {
		return errors.New("All asset values must have a non empty AssetId")
	}
	if len(av.RefId) <= 0 {
		return errors.New("All asset values must have a non empty RefId")
	}
	_, err := DB.Exec("INSERT INTO `AssetValue` (`Id`, `AssetId`, `RefId`, `Value`, `Date`, `Notes`) VALUES (?, ?, ?, ?, ?, ?)", av.Id, av.AssetId, av.RefId, av.Value, av.Date.Unix(), av.Notes)
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

func (av AssetValue) ValueToStr() string {
	ak := AssetKind{}
	err := ak.Load(av.RefId)
	if err != nil {
		return "ERR"
	}
	return fmt_decimal(av.Value, ak.DecimalPlaces)
}

func (av AssetValue) StrToValue(val_str string) {
	// Load Ref AssetKind
	ak := AssetKind{}
	err := ak.Load(av.RefId)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// Parse stuff
	av.Value = parse_decimal(val_str, ak.DecimalPlaces)
}

func (av AssetValue) MultilineString() string {
	val_str := fmt.Sprintf("%d (raw/faield to load AssetKind)", av.Value)
	ak := AssetKind{}
	err := ak.Load(av.RefId)
	if err == nil {
		val_str = fmt.Sprintf("%s %s = %s %s", Cyan("1"), Bold(av.AssetId), Cyan(fmt_decimal(av.Value, ak.DecimalPlaces)), Bold(av.RefId))
	}
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("     Id:"), av.Id)
	s += fmt.Sprintf("%s %s\n", Bold("AssetId:"), av.AssetId)
	s += fmt.Sprintf("%s %s\n", Bold("  RefId:"), av.RefId)
	s += fmt.Sprintf("%s %s\n", Bold("  Value:"), val_str)
	s += fmt.Sprintf("%s %s\n", Bold("   Date:"), av.Date.Format(DATE_FMT_SPACES))
	s += fmt.Sprintf("%s %s\n", Bold("  Notes:"), av.Notes)
	return s
}

func asset_value_show(line []string) {
	spec := ""
	if len(line) > 0 {
		spec = line[0]
	}
	rows, err := DB.Query("SELECT `Id` FROM `AssetValue` WHERE `Id` = ? OR `AssetId` = ? OR ? = ''", spec, spec, spec, spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read accounts
	avs := make([]AssetValue, 0)
	defer rows.Close()
	for rows.Next() {
		av := AssetValue{}
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		err = av.Load(id)
		if err != nil {
			log.Fatal(err)
		}
		avs = append(avs, av)
	}
	if len(avs) == 1 {
		fmt.Printf(avs[0].MultilineString())
		return
	}
	fmt.Printf("%10s | %10s | %10s | %s\n", "AssetId", "RefId", "Value", "Date")
	fmt.Println("-----------+------------+------------+------------------------------------------")
	for _, av := range avs {
		fmt.Printf("%10s | %10s | %10s | %s\n", av.AssetId, av.RefId, av.ValueToStr(), av.Date.Format(DAY_FMT))
	}
}

func asset_value_add(line []string) {
	var err error
	av := AssetValue{}
	// Ask user
	LocalLine.Config.AutoComplete = CompleterAssetKind
	av.AssetId = must_ask_user(LocalLine, Sprintf(Bold("AssetId: ")), "")
	av.RefId = must_ask_user(LocalLine, Sprintf(Bold("RefId: ")), "")
	LocalLine.Config.AutoComplete = nil
	val_str := must_ask_user(LocalLine, Sprintf(Bold("Value: ")), "")
	date_str := must_ask_user(LocalLine, Sprintf(Bold("Date: ")), "")
	// Parse stuff
	av.StrToValue(val_str)
	av.Date, err = time.Parse(DAY_FMT, date_str)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Save
	err = av.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_value_edit(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	av := AssetValue{}
	err := av.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(Bold("     Id:"), av.Id, Gray(" (non editable)"))
	fmt.Println(Bold("AssetId:"), av.AssetId, Gray(" (non editable)"))
	fmt.Println(Bold("  RefId:"), av.RefId, Gray(" (non editable)"))
	fmt.Println(Bold("   Date:"), av.Date, Gray(" (non editable)"))
	tmp := must_ask_user(LocalLine, Sprintf(Bold("  Value: ")), av.ValueToStr())
	av.Notes = must_ask_user(LocalLine, Sprintf(Bold("  Notes: ")), av.Notes)
	av.StrToValue(tmp)
	err = av.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_value_del(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	id := line[len(line)-1]
	av := AssetValue{}
	conf := "DEL-" + id
	input := ""
	fmt.Printf("Type '%s' to confirm deletion: ", Bold(Red(conf)))
	fmt.Scanln(&input)
	if input != conf {
		fmt.Println(Bold("Deletion avoided"))
		return
	}

	err := av.Del(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Bold("Deletion done"))
}

func CompleteAssetValueFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `AssetValue` WHERE `Id` LIKE '"+spec+"%%' OR ? = '' LIMIT 64", spec)
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

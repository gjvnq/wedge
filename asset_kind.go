package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	. "github.com/logrusorgru/aurora"
	"github.com/mgutz/str"
)

type AssetKind struct {
	Id            string
	Name          string
	Desc          string
	DecimalPlaces int
	Tags          map[string]bool
}

func NewAssetKind() *AssetKind {
	ak := AssetKind{}
	ak.Init()
	return &ak
}

func (ak *AssetKind) Init() {
	if ak.Tags == nil {
		ak.Tags = make(map[string]bool)
	}
}

func (ak AssetKind) Save() error {
	if len(ak.Id) <= 0 {
		return errors.New("All asset kinds must have a non empty id")
	}
	if len(ak.Name) <= 0 {
		return errors.New("All asset kinds must have a non empty name")
	}
	_, err := DB.Exec("INSERT INTO `AssetKind` (`Id`, `Name`, `Desc`, `DecimalPlaces`) VALUES (?, ?, ?, ?)", ak.Id, ak.Name, ak.Desc, ak.DecimalPlaces)
	return err
}

func (ak AssetKind) Update() error {
	_, err := DB.Exec("UPDATE `AssetKind` SET `Name` = ?, `Desc` = ?, `DecimalPlaces` = ? WHERE `Id` = ?", ak.Name, ak.Desc, ak.DecimalPlaces, ak.Id)
	return err
}

func (ak AssetKind) Del(id string) error {
	_, err := DB.Exec("DELETE FROM `AssetKind` WHERE `Id` = ?", id)
	return err
}

func (ak *AssetKind) Load(id string) error {
	err := DB.QueryRow("SELECT `Id`, `Name`, `Desc`, `DecimalPlaces` FROM `AssetKind` WHERE `Id` = ?", id).
		Scan(&ak.Id, &ak.Name, &ak.Desc, &ak.DecimalPlaces)
	return err
}

func (ak AssetKind) MultilineString() string {
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("            Id:"), ak.Id)
	s += fmt.Sprintf("%s %s\n", Bold("          Name:"), ak.Name)
	s += fmt.Sprintf("%s %s\n", Bold("          Desc:"), ak.Desc)
	s += fmt.Sprintf("%s %d\n", Bold("Decimal places:"), ak.DecimalPlaces)
	return s
}

func parse_decimal(input string, decimal_places int) int {
	input = strings.TrimSpace(input)
	tmp := ""
	mul := 1
	places := 0
	after_dot := false
	for _, cur_rune := range input {
		char := string(cur_rune)
		switch {
		case char == "-":
			mul = -1
		case char == ".":
			after_dot = true
		case after_dot == false:
			tmp += char
		case after_dot == true && places < decimal_places:
			tmp += char
			places++
		default:
			continue
		}
	}
	for places < decimal_places {
		tmp += "0"
		places++
	}
	return str.ToIntOr(tmp, 0) * mul
}

func full_decimal_parse(val_str string, asset_kind_id string) (int, error) {
	// Load AssetKind
	ak := AssetKind{}
	err := ak.Load(asset_kind_id)
	if err != nil {
		return 0, err
	}
	// Parse stuff
	return parse_decimal(val_str, ak.DecimalPlaces), nil
}

func full_decimal_fmt(val_raw int, asset_kind_id string) (string, error) {
	// Load AssetKind
	ak := AssetKind{}
	err := ak.Load(asset_kind_id)
	if err != nil {
		return "", err
	}
	// Parse stuff
	return fmt_decimal(val_raw, ak.DecimalPlaces), nil
}

func full_decimal_fmt_pad(val_raw int, pad int, asset_kind_id string) (string, error) {
	// Load AssetKind
	ak := AssetKind{}
	err := ak.Load(asset_kind_id)
	if err != nil {
		return "", err
	}
	// Parse stuff
	return fmt_decimal_pad(val_raw, ak.DecimalPlaces, pad), nil
}

func fmt_decimal(raw, decimal_places int) string {
	div := int(math.Pow10(decimal_places))
	quotient := raw / div
	remainder := raw - div*quotient
	return fmt.Sprintf("%d.%d", quotient, remainder)
}

func fmt_decimal_pad(raw, decimal_places, pad int) string {
	div := int(math.Pow10(decimal_places))
	quotient := raw / div
	remainder := raw - div*quotient
	pad_str := fmt.Sprintf("%d", pad)
	return fmt.Sprintf("%%"+pad_str+"d.%d", quotient, remainder)
}

func asset_kind_show(line []string) {
	spec := ""
	if len(line) > 0 {
		spec = line[0]
	}
	rows, err := DB.Query("SELECT `Id`, `Name`, `Desc`, `DecimalPlaces` FROM `AssetKind` WHERE `Id` = ? OR `Name` LIKE '%%"+spec+"%%' OR ? = ''", spec, spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read accounts
	aks := make([]AssetKind, 0)
	defer rows.Close()
	for rows.Next() {
		ak := AssetKind{}
		err := rows.Scan(&ak.Id, &ak.Name, &ak.Desc, &ak.DecimalPlaces)
		if err != nil {
			log.Fatal(err)
		}
		aks = append(aks, ak)
	}
	if len(aks) == 1 {
		fmt.Printf(aks[0].MultilineString())
		return
	}
	fmt.Printf("%10s | %8s | %s\n", Bold("Id"), "Min Frac", "Name")
	fmt.Println("-----------+----------+---------------------------------------------------------")
	for _, ak := range aks {
		places := fmt.Sprintf("1/10^%d", ak.DecimalPlaces)
		fmt.Printf("%10s | %-8s | %s\n", Bold(ak.Id), places, ak.Name)
	}
}

func asset_kind_add(line []string) {
	ak := AssetKind{}
	// Ask user
	ak.Id = ask_user(
		LocalLine,
		Sprintf(Bold("  Id: ")),
		"",
		nil,
		True)
	ak.Name = ask_user(
		LocalLine,
		Sprintf(Bold("Name: ")),
		"",
		nil,
		True)
	ak.Desc = ask_user(
		LocalLine,
		Sprintf(Bold("Desc: ")),
		"",
		nil,
		True)
	ak.DecimalPlaces = str.ToIntOr(ask_user(
		LocalLine,
		Sprintf(Bold("DecimalPlaces: ")),
		"",
		nil,
		IsInt), 0)

	err := ak.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_kind_edit(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	ak := AssetKind{}
	err := ak.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(Bold("   Id:"), ak.Id, Gray(" (non editable)"))
	ak.Name = ask_user(
		LocalLine,
		Sprintf(Bold("Name: ")),
		ak.Name,
		nil,
		True)
	ak.Desc = ask_user(
		LocalLine,
		Sprintf(Bold("Desc: ")),
		ak.Desc,
		nil,
		True)
	ak.DecimalPlaces = str.ToIntOr(ask_user(
		LocalLine,
		Sprintf(Bold("DecimalPlaces: ")),
		Sprintf(ak.DecimalPlaces),
		nil,
		IsInt), 0)

	err = ak.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_kind_del(line []string) {
	if len(line) == 0 {
		fmt.Println(Red("No id specified"))
		return
	}
	id := line[len(line)-1]
	deleter(id, NewAssetKind())
}

func CompleteAssetKindFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `AssetKind` WHERE `Id` LIKE '"+spec+"%%' OR ? = ''", spec)
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

package main

import (
	"errors"
	"fmt"
	"github.com/chzyer/readline"
	. "github.com/logrusorgru/aurora"
	"github.com/mgutz/str"
	"log"
	"math"
	"strings"
)

var AssetKindLine *readline.Instance

type AssetKind struct {
	Id            string
	Name          string
	Desc          string
	DecimalPlaces int
	Tags          map[string]bool
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

func fmt_decimal(raw, decimal_places int) string {
	div := int(math.Pow10(decimal_places))
	quotient := raw / div
	remainder := raw - div
	return fmt.Sprintf("%d.%d", quotient, remainder)
}

func fmt_decimal_pad(raw, decimal_places, pad int) string {
	div := int(math.Pow10(decimal_places))
	quotient := raw / div
	remainder := raw - div
	pad_str := fmt.Sprintf("%d", pad)
	return fmt.Sprintf("%%"+pad_str+"d.%d", quotient, remainder)
}

func asset_kind_add(line []string) {
	ak := AssetKind{}
	// Ask user
	ak.Id = must_ask_user(AssetKindLine, Sprintf(Bold("Id: ")), "")
	ak.Name = must_ask_user(AssetKindLine, Sprintf(Bold("Name: ")), "")
	ak.Desc = must_ask_user(AssetKindLine, Sprintf(Bold("Desc: ")), "")
	ak.DecimalPlaces = str.ToIntOr(must_ask_user(AssetKindLine, Sprintf(Bold("DecimalPlaces: ")), ""), 0)
	err := ak.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_kind_show(line []string) {
	spec := line[0]
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

func asset_kind_edit(line []string) {
	ak := AssetKind{}
	err := ak.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ak.Name = must_ask_user(AccountLine, Sprintf(Bold("Name: ")), ak.Name)
	ak.Desc = must_ask_user(AccountLine, Sprintf(Bold("Desc: ")), ak.Desc)
	tmp := fmt.Sprintf("%d", ak.DecimalPlaces)
	ak.DecimalPlaces = str.ToIntOr(must_ask_user(AssetKindLine, Sprintf(Bold("DecimalPlaces: ")), tmp), ak.DecimalPlaces)
	err = ak.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func asset_kind_del(line []string) {
	id := line[len(line)-1]
	ak := AssetKind{}
	conf := "DEL-" + id
	input := ""
	fmt.Printf("Type '%s' to confirm deletion: ", Bold(Red(conf)))
	fmt.Scanln(&input)
	if input != conf {
		fmt.Println(Bold("Deletion avoided"))
		return
	}

	err := ak.Del(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Bold("Deletion done"))
}

func asset_kind_prep() {
	var err error
	// Preapre readline
	AssetKindLine, err = readline.NewEx(&readline.Config{
		Prompt:          "» ",
		HistoryLimit:    -1,
		InterruptPrompt: "^C",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func CompleteAssetKind(prefix string) []string {
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

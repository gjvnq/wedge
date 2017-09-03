package main

import (
	"errors"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"log"
	"strings"
)

type Account struct {
	Id       string
	ParentId string
	Name     string
	Desc     string
	Tags     map[string]bool
}

func (acc Account) Save() error {
	if len(acc.Id) <= 0 {
		return errors.New("All accounts must have a non empty id")
	}
	if len(acc.Name) <= 0 {
		return errors.New("All accounts must have a non empty name")
	}
	_, err := DB.Exec("INSERT INTO `Account` (`Id`, `ParentId`, `Name`, `Desc`) VALUES (?, ?, ?, ?)", acc.Id, acc.ParentId, acc.Name, acc.Desc)
	return err
}

func (acc Account) Update() error {
	_, err := DB.Exec("UPDATE `Account` SET `ParentId` = ?, `Name` = ?, `Desc` = ? WHERE `Id` = ?", acc.ParentId, acc.Name, acc.Desc, acc.Id)
	return err
}

func (acc Account) Del(id string) error {
	_, err := DB.Exec("DELETE FROM `Account` WHERE `Id` = ?", id)
	return err
}

func (acc *Account) Load(id string) error {
	err := DB.QueryRow("SELECT `Id`, `ParentId`, `Name`, `Desc` FROM `Account` WHERE `Id` = ?", id).
		Scan(&acc.Id, &acc.ParentId, &acc.Name, &acc.Desc)
	return err
}

func (acc Account) MultilineString() string {
	s := ""
	s += fmt.Sprintf("%s %s\n", Bold("      Id:"), acc.Id)
	s += fmt.Sprintf("%s %s\n", Bold("ParentId:"), acc.ParentId)
	s += fmt.Sprintf("%s %s\n", Bold("    Name:"), acc.Name)
	s += fmt.Sprintf("%s %s\n", Bold("    Desc:"), acc.Desc)
	return s
}

func account_show(line []string) {
	spec := ""
	if len(line) > 0 {
		spec = line[0]
	}
	rows, err := DB.Query("SELECT `Id`, `ParentId`, `Name`, `Desc` FROM `Account` WHERE `Id` = ? OR `Name` LIKE '%%"+spec+"%%' OR ? = ''", spec, spec)
	if err != nil {
		log.Fatal(err)
	}
	// Read accounts
	accs := make([]Account, 0)
	defer rows.Close()
	for rows.Next() {
		acc := Account{}
		err := rows.Scan(&acc.Id, &acc.ParentId, &acc.Name, &acc.Desc)
		if err != nil {
			log.Fatal(err)
		}
		accs = append(accs, acc)
	}
	if len(accs) == 1 {
		fmt.Printf(accs[0].MultilineString())
		return
	}
	printed := make(map[string]bool)
	account_show_print_children(-1, Account{}, printed, accs)
	for _, acc := range accs {
		if !printed[acc.Id] {
			fmt.Printf("├─ %s (child of %s) %s\n", Bold(acc.Id), acc.ParentId, acc.Name)
		}
	}
}

func account_show_print_children(level int, parent Account, printed map[string]bool, accs []Account) {
	print_line := func(has_children bool) {
		if printed[parent.Id] == false && parent.Id != "" {
			for i := 0; i < level; i++ {
				fmt.Printf("┆")
			}
			if has_children {
				fmt.Printf("├┬ %s %s\n", Bold(parent.Id), parent.Name)
			} else {
				fmt.Printf("├─ %s %s\n", Bold(parent.Id), parent.Name)
			}
			printed[parent.Id] = true
		}
	}

	for _, acc := range accs {
		if acc.ParentId == parent.Id {
			// Print this account
			print_line(true)
			// Print children
			account_show_print_children(level+1, acc, printed, accs)
		}
	}
	print_line(false)
}

func account_add(line []string) {
	acc := Account{}
	acc.Id = must_ask_user(LocalLine, Sprintf(Bold("Id: ")), "")
	LocalLine.Config.AutoComplete = CompleterAccount
	acc.ParentId = must_ask_user(LocalLine, Sprintf(Bold("ParentId: ")), "")
	LocalLine.Config.AutoComplete = nil
	acc.Name = must_ask_user(LocalLine, Sprintf(Bold("Name: ")), "")
	acc.Desc = must_ask_user(LocalLine, Sprintf(Bold("Desc: ")), "")
	err := acc.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func account_edit(line []string) {
	acc := Account{}
	err := acc.Load(line[len(line)-1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	LocalLine.Config.AutoComplete = CompleterAccount
	acc.ParentId = must_ask_user(LocalLine, Sprintf(Bold("ParentId: ")), acc.ParentId)
	LocalLine.Config.AutoComplete = nil
	acc.Name = must_ask_user(LocalLine, Sprintf(Bold("Name: ")), acc.Name)
	acc.Desc = must_ask_user(LocalLine, Sprintf(Bold("Desc: ")), acc.Desc)
	err = acc.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func account_del(line []string) {
	id := line[len(line)-1]
	acc := Account{}
	conf := "DEL-" + id
	input := ""
	fmt.Printf("Type '%s' to confirm deletion: ", Bold(Red(conf)))
	fmt.Scanln(&input)
	if input != conf {
		fmt.Println(Bold("Deletion avoided"))
		return
	}

	err := acc.Del(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Bold("Deletion done"))
}

func CompleteAccountFunc(prefix string) []string {
	tmp := strings.Split(prefix, " ")
	spec := tmp[len(tmp)-1]
	rows, err := DB.Query("SELECT `Id` FROM `Account` WHERE `Id` LIKE '"+spec+"%%' OR ? = ''", spec)
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

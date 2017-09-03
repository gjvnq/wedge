package main

import (
	"errors"
	"flag"
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

var AccountFlagSet *flag.FlagSet
var AccountFlagSetId string
var AccountFlagSetName string
var AccountFlagSetDesc string
var AccountFlagSetParentId string

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
	spec := line[0]
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
			fmt.Printf("├─ %s (child of %s) - %s\n", acc.Id, acc.ParentId, acc.Name)
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
				fmt.Printf("├┬ %s - %s\n", parent.Id, parent.Name)
			} else {
				fmt.Printf("├─ %s - %s\n", parent.Id, parent.Name)
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
	AccountFlagSet.Parse(line)
	set_str(AccountFlagSetId, &acc.Id)
	set_str(AccountFlagSetName, &acc.Name)
	set_str(AccountFlagSetDesc, &acc.Desc)
	set_str(AccountFlagSetParentId, &acc.ParentId)
	err := acc.Save()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func account_edit(line []string) {
	acc := Account{}
	AccountFlagSet.Parse(line)
	err := acc.Load(AccountFlagSetId)
	if err != nil {
		fmt.Println(err.Error())
	}
	set_str(AccountFlagSetName, &acc.Name)
	set_str(AccountFlagSetDesc, &acc.Desc)
	set_str(AccountFlagSetParentId, &acc.ParentId)
	err = acc.Update()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func account_del(line []string) {
	id := ""
	acc := Account{}
	AccountFlagSet.Parse(line)
	set_str(AccountFlagSetId, &id)

	conf := "DEL-" + id
	input := ""
	fmt.Printf("Type '%s' to confirm deletion: ", conf)
	fmt.Scanln(&input)
	if input != conf {
		fmt.Println("Deletion avoided")
		fmt.Printf("You typed '%s' insted of '%s'\n", input, conf)
		return
	}

	err := acc.Del(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Deletion done")
}

func account_prep() {
	AccountFlagSet = flag.NewFlagSet("account", flag.ContinueOnError)
	AccountFlagSet.StringVar(&AccountFlagSetId, "id", UNSET_STR, "")
	AccountFlagSet.StringVar(&AccountFlagSetName, "name", UNSET_STR, "")
	AccountFlagSet.StringVar(&AccountFlagSetDesc, "desc", UNSET_STR, "")
	AccountFlagSet.StringVar(&AccountFlagSetParentId, "parent-id", UNSET_STR, "")
}

func CompleteAccount(prefix string) []string {
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

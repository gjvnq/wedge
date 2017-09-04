package main

import (
	"fmt"

	. "github.com/logrusorgru/aurora"
)

type ILoadable interface {
	Load(id string) error
}

type IDeletable interface {
	Del(id string) error
}

func deleter(id string, obj IDeletable) {
	conf := "DEL-" + id
	fmt.Printf("Type '%s' to confirm deletion: ", Bold(Red(conf)))
	input := ask_user(
		LocalLine,
		fmt.Sprintf("Type '%s' to confirm deletion: ", Bold(Red(conf))),
		"",
		nil,
		True)
	if input != conf {
		fmt.Println(Bold("Deletion avoided"))
		return
	}
	err := obj.Del(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Bold("Deletion done"))
}

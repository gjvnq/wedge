package main

type Account struct {
	Id       string
	ParentId string
	Name     string
	Desc     string
	Tags     map[string]bool
}

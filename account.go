package main

type Account struct {
	Id       string
	Name     string
	Type     string
	Desc     string
	Currency string
	Tags     map[string]bool
}

func (a Account) TypeName() string {
	return "Account"
}

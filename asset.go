package main

import (
	"time"
)

type Asset struct {
	Id   string
	Name string
	Desc string
	Tags map[string]bool
}

type AssetValue struct {
	Id       string
	AssetId  string
	Value    int
	ValueCur string
	Date     time.Time
	Notes    string
}

func (a Asset) TypeName() string {
	return "Asset"
}

func (av AssetValue) TypeName() string {
	return "AssetValue"
}

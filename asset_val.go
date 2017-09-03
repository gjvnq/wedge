package main

import (
	"time"
)

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

func CompleteAssetValue(prefix string) []string {
	return []string{"A", "B"}
}

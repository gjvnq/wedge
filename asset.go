package main

import (
	"time"
)

type AssetKind struct {
	Id            string
	Name          string
	Desc          string
	DecimalPlaces int
	Tags          map[string]bool
}

type AssetValue struct {
	Id      string
	AssetId string
	RefId   string
	Value   int // Value of AssetId in terms of RefId
	Date    time.Time
	Notes   string
}

func (a AssetKind) TypeName() string {
	return "Asset"
}

func (av AssetValue) TypeName() string {
	return "AssetValue"
}

func CompleteAssetKind(prefix string) []string {
	return []string{"A", "B"}
}

func CompleteAssetValue(prefix string) []string {
	return []string{"A", "B"}
}

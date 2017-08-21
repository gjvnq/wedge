package main

type World struct {
	Accounts     map[string]Account
	Assets       map[string]Asset
	AssetValues  map[string]AssetValue
	AssetValues  map[string]AssetValue
	Groups       map[string]Group
	Transactions map[string]Transaction
	// The fields below are generated when the DB is loaded. The are all maps Id to Id
	GMovByAccount        map[string]string `json:"-"`
	AssetValueByAsset    map[string]string `json:"-"`
	TransactionByAccount map[string]string `json:"-"`
	IdsByTag             map[string]string `json:"-"`
}

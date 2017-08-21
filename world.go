package main

type World struct {
	Accounts             map[string]Account
	Assets               map[string]Asset
	AssetValues          map[string]AssetValue
	AssetValues          map[string]AssetValue
	Transactions         map[string]Transaction
	Groups               map[string]Group
	GMovs                map[string]GMov
	GMovByAccount        map[string]string
	AssetValueByAsset    map[string]string
	TransactionByAccount map[string]string
	IdsByTag             map[string]string
}

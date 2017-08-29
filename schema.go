package main

import (
	"database/sql"
	"log"
)

func EnsureTables(db *sql.DB) {
	text := ""
	text += "CREATE TABLE IF NOT EXISTS `Account` ( `Id` TEXT NOT NULL UNIQUE, `ParentId` TEXT NOT NULL, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, PRIMARY KEY(`Id`));\n"
	text += "CREATE TABLE IF NOT EXISTS `AssetKind` ( `Id` TEXT NOT NULL UNIQUE, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, `DecimalPlaces` INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(`Id`));\n"
	text += "CREATE TABLE IF NOT EXISTS `AssetValues` ( `Id` TEXT NOT NULL UNIQUE, `AssetId` TEXT NOT NULL, `RefId` TEXT NOT NULL, `Value` INTEGER NOT NULL DEFAULT 0, `Date` INTEGER NOT NULL DEFAULT 0, `Notes` TEXT NOT NULL, PRIMARY KEY(`Id`));\n"
	text += "CREATE TABLE IF NOT EXISTS `Tags` ( `ObjectId` TEXT NOT NULL, `Tag` TEXT NOT NULL );\n"
	text += "CREATE TABLE IF NOT EXISTS `Transaction` ( `Id` TEXT NOT NULL UNIQUE, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, `RefStart` INTEGER NOT NULL DEFAULT 0, `RefEnd` INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(`Id`));\n"
	text += "CREATE TABLE IF NOT EXISTS `TransactionItem` ( `Id` TEXT NOT NULL UNIQUE, `TransactionId` TEXT NOT NULL, `Name` TEXT NOT NULL, `CostAmount` INTEGER NOT NULL DEFAULT 0, `CostAssetKindId` TEXT NOT NULL, `UnitCost` INTEGER NOT NULL DEFAULT 0, `Quantity` REAL NOT NULL, PRIMARY KEY(`Id`));\n"
	text += "CREATE TABLE IF NOT EXISTS `TransactionParts` ( `Id` TEXT NOT NULL UNIQUE, `TransactionId` TEXT NOT NULL, `AccountId` TEXT NOT NULL, `Status` TEXT NOT NULL, `ScheduledFor` INTEGER NOT NULL DEFAULT 0, `ActualDate` INTEGER NOT NULL DEFAULT 0, `CostAmount` INTEGER NOT NULL DEFAULT 0, `CostAssetKindId` TEXT NOT NULL, PRIMARY KEY(`Id`));\n"
	text += "CREATE UNIQUE INDEX IF NOT EXISTS `IndexUniTag` ON `Tags` (`ObjectId` ASC,`Tag` ASC);\n"
	_, err := db.Exec(text)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"database/sql"
	"log"
)

func EnsureTables(db *sql.DB) {
	codes := make([]string, 0)
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `Account` ( `Id` TEXT NOT NULL UNIQUE, `ParentId` TEXT NOT NULL, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, PRIMARY KEY(`Id`));")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `AssetKind` ( `Id` TEXT NOT NULL UNIQUE, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, `DecimalPlaces` INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(`Id`));")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `AssetValue` ( `Id` TEXT NOT NULL UNIQUE, `AssetId` TEXT NOT NULL, `RefId` TEXT NOT NULL, `Value` INTEGER NOT NULL DEFAULT 0, `Date` INTEGER NOT NULL DEFAULT 0, `Notes` TEXT NOT NULL, PRIMARY KEY(`Id`));")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `Tags` ( `ObjectId` TEXT NOT NULL, `Tag` TEXT NOT NULL );")
	codes = append(codes, "CREATE UNIQUE INDEX IF NOT EXISTS `IndexUniTag` ON `Tags` (`ObjectId` ASC,`Tag` ASC);")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `Transaction` ( `Id` TEXT NOT NULL UNIQUE, `Name` TEXT NOT NULL, `Desc` TEXT NOT NULL, `RefStart` INTEGER NOT NULL DEFAULT 0, `RefEnd` INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(`Id`));")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `TransactionPart` ( `Id` TEXT NOT NULL UNIQUE, `TransactionId` TEXT NOT NULL, `AccountId` TEXT NOT NULL, `Status` TEXT NOT NULL, `ScheduledFor` INTEGER NOT NULL DEFAULT 0, `ActualDate` INTEGER NOT NULL DEFAULT 0, `Value` INTEGER NOT NULL DEFAULT 0, `AssetKindId` TEXT NOT NULL, PRIMARY KEY(`Id`));")
	codes = append(codes, "CREATE TABLE IF NOT EXISTS `TransactionItem` ( `Id` TEXT NOT NULL UNIQUE, `TransactionId` TEXT NOT NULL, `Name` TEXT NOT NULL, `UnitCost` INTEGER NOT NULL DEFAULT 0, `AssetKindId` TEXT NOT NULL, `Quantity` REAL NOT NULL,  `TotalCost` INTEGER NOT NULL, PRIMARY KEY(`Id`));")
	for _, code := range codes {
		_, err := db.Exec(code)
		if err != nil {
			log.Fatal(err)
		}
	}
}

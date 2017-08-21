package main

import (
"database/sql"
"log"
)

func schema_handler(db *sql.DB) {
	cmd := "CREATE TABLE IF NOT EXISTS `data` ("
	cmd += "`id` TEXT NOT NULL PRIMARY KEY,"
	cmd += "`type` TEXT NOT NULL,"
	cmd += "`data` TEXT NOT NULL"
	cmd += ")"
	_, err := db.Exec(cmd)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	cmd = "CREATE INDEX IF NOT EXISTS `idx_data_type` ON `data` (`type`)"
	_, err = db.Exec(cmd)
	if err != nil {
		log.Fatal(err)
	}
	// Used to store names, descs and similar stuff (dates and times are always stored as an unix timestamp)
	cmd := "CREATE TABLE IF NOT EXISTS `meta` ("
	cmd += "`id` TEXT NOT NULL PRIMARY KEY,"
	cmd += "`key` TEXT NOT NULL,"
	cmd += "`value` TEXT NOT NULL"
	cmd += ")"
	_, err := db.Exec(cmd)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	cmd = "CREATE UNIQUE INDEX IF NOT EXISTS `idx_meta_key` ON `meta` (`id`, `key`)"
	_, err = db.Exec(cmd)
	if err != nil {
		log.Fatal(err)
	}
	// Used to store tags and similar stuff
	cmd := "CREATE TABLE IF NOT EXISTS `poly_meta` ("
	cmd += "`id` TEXT NOT NULL PRIMARY KEY,"
	cmd += "`key` TEXT NOT NULL,"
	cmd += "`value` TEXT NOT NULL"
	cmd += ")"
	_, err := db.Exec(cmd)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	cmd = "CREATE INDEX IF NOT EXISTS `idx_poly_meta_key` ON `poly_meta` (`key`)"
	_, err = db.Exec(cmd)
	if err != nil {
		log.Fatal(err)
	}
}
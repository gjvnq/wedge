package main

import (
	"time"
)

type Group struct {
	Id   string
	Name string
	Desc string
	Tags map[string]bool
}

// Group movement
type GMov struct {
	Id            string
	ObjectId      string //Can be an account, an asset or another group
	DestinationId string
	Status        string
	ScheledFor    time.Time
	ActualDate    time.Time
}

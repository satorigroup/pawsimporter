package main

import (
	"fmt"
	"satori/pawsimporter/config"
	"satori/pawsimporter/db"
	. "satori/pawsimporter/xlsximporter"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		fmt.Printf("%s \n", err)
		return
	}

	db, err := db.NewService(conf.ConnString, "development")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	importer := Importer{}
	importer.DB = db
	importer.DataFile = conf.SourceData
	importer.Begin()
}

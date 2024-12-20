package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "os"
	_ "time"

	"github.com/mikolajgs/prototyping"
	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"

	_ "github.com/lib/pq"
)

const dbDSN = "host=localhost user=protouser password=protopass port=54320 dbname=protodb sslmode=disable"

func main() {
	p, err := prototyping.NewPrototype(
		prototyping.Config{
			DatabaseDSN:     dbDSN,
			UserConstructor: func() interface{} { return &User{} },
		},
		func() interface{} { return &Item{} },
		func() interface{} { return &ItemGroup{} },
	)
	if err != nil {
		log.Fatalf("error creating new prototype: %s", err.Error())
	}

	err = p.CreateDB()
	if err != nil {
		log.Fatalf("error creating database: %s", err.Error())
	}

	// creating dummy objects in the database
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatal("Error connecting to db")
	}
	// proto_ is the default db table prefix (will be configurable later)
	s2db := stdb.NewController(db, "proto_", nil)
	item := &Item{}
	itemGroup := &ItemGroup{}
	for i := 0; i < 301; i++ {
		item.ID = 0
		item.Flags = int64(i)
		item.Title = fmt.Sprintf("Item %d", i)
		item.Text = fmt.Sprintf("Description %d", i)
		s2db.Save(item, stdb.SaveOptions{})
	}
	for i := 0; i < 73; i++ {
		itemGroup.ID = 0
		itemGroup.Flags = int64(i)
		itemGroup.Name = fmt.Sprintf("Name %d", i)
		itemGroup.Description = fmt.Sprintf("Description %d", i)
		s2db.Save(itemGroup, stdb.SaveOptions{})
	}
	db.Close()
	// end of creating dummy objects

	err = p.Run()
	if err != nil {
		log.Fatalf("error running prototype: %s", err.Error())
	}
}

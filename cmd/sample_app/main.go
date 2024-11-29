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

const dbUser = "uiuser"
const dbPass = "uipass"
const dbName = "uidb"
const dbPort = "54320"

func main() {
	p, err := prototyping.NewPrototype(prototyping.DbConfig{
		Host: "localhost",
		Port: "54320",
		User: "uiuser",
		Pass: "uipass",
		Name: "uidb",
		TablePrefix: "ui_",
	}, []func() interface{}{
		func() interface{} { return &Item{} },
		func() interface{} { return &ItemGroup{} },
	},
		prototyping.HttpConfig{
		Port: "9001",
		ApiUri: "/api/v1/",
		UiUri: "/ui/v1/",
	})
	if err != nil {
		log.Fatalf("error creating new prototype: %s", err.Error())
	}

	err = p.CreateDB()
	if err != nil {
		log.Fatalf("error creating database: %s", err.Error())
	}

	// creating dummy objects in the database
	db, err := sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dbPort, dbName))
	if err != nil {
		log.Fatal("Error connecting to db")
	}
	s2db := stdb.NewController(db, "ui_", nil)
	item := &Item{}
	itemGroup := &ItemGroup{}
	for i:=0; i<301; i++ {
		item.ID = 0;
		item.Flags = int64(i);
		item.Title = fmt.Sprintf("Item %d", i)
		item.Text = fmt.Sprintf("Description %d", i)
		s2db.Save(item, stdb.SaveOptions{})
	}
	for i:=0; i<73; i++ {
		itemGroup.ID = 0;
		itemGroup.Flags = int64(i);
		itemGroup.Name = fmt.Sprintf("Name %d", i)
		itemGroup.Description = fmt.Sprintf("Description %d", i)
		s2db.Save(itemGroup, stdb.SaveOptions{})
	}
	db.Close()

	err = p.Run()
	if err != nil {
		log.Fatalf("error running prototype: %s", err.Error())
	}
}

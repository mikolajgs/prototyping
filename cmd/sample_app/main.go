package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "os"
	_ "time"

	restapi "github.com/mikolajgs/prototyping/pkg/rest-api"
	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	"github.com/mikolajgs/prototyping/pkg/ui"

	_ "github.com/lib/pq"
)

const dbUser = "uiuser"
const dbPass = "uipass"
const dbName = "uidb"
const dbPort = "54320"

func main() {
	db, err := sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dbPort, dbName))
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	uiCtl := ui.NewController(db, "ui_")
	apiCtl := restapi.NewController(db, "ui_", nil)
	s2db := stdb.NewController(db, "ui_", nil)

	item := &Item{}
	itemGroup := &ItemGroup{}

	err2 := s2db.DropTables(item, itemGroup)
	if err2 != nil {
		log.Printf("Error with dropping tables: %s", err2.Error())
	}

	err2 = s2db.CreateTables(item, itemGroup)
	if err2 != nil {
		log.Fatalf("Error with creating tables: %s", err.Error())
	}
	
	http.Handle("/ui/v1/", uiCtl.GetHTTPHandler(
		"/ui/v1/",
		func() interface{}{ return &Item{} },
		func() interface{}{ return &ItemGroup{} },
	))
	http.Handle("/api/v1/items/", apiCtl.Handler(
		"/api/v1/items/",
		func() interface{}{ return &Item{} },
		restapi.HandlerOptions{},
	))
	http.Handle("/api/v1/item_groups/", apiCtl.Handler(
		"/api/v1/item_groups/",
		func() interface{}{ return &ItemGroup{} },
		restapi.HandlerOptions{},
	))
	log.Fatal(http.ListenAndServe(":9001", nil))
}

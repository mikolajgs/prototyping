package ui

import (
	"database/sql"

	"github.com/mikolajgs/crud/pkg/struct2db"
)

// Controller is the main component that gets and saves objects in the database and generates HTTP handler
// that can be attached to an HTTP server.
type Controller struct {
	struct2db *struct2db.Controller
	uriStructNameFunc map[string]map[string]func() interface{}
	uriStructArgsName map[string]map[int]string
	structNames map[string][]string
}

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string) *Controller {
	c := &Controller{}
	c.struct2db = struct2db.NewController(dbConn, tblPrefix, &struct2db.ControllerConfig{
		TagName: "ui",
	})
	return c
}

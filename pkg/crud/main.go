package crud

import (
	"database/sql"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

// Controller is the main component that gets and saves objects in the database and generates CRUD HTTP handler
// that can be attached to an HTTP server.
type Controller struct {
	dbConn       *sql.DB
	dbTblPrefix  string
	modelHelpers map[string]*struct2sql.Struct2sql
}

// Values for CRUD operations
const OpRead = 2
const OpUpdate = 4
const OpCreate = 8
const OpDelete = 16
const OpList = 32

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string) *Controller {
	c := &Controller{
		dbConn:      dbConn,
		dbTblPrefix: tblPrefix,
	}
	c.modelHelpers = make(map[string]*struct2sql.Struct2sql)
	return c
}

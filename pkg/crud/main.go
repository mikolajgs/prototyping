package crud

import (
	"database/sql"

	"github.com/mikolajgs/crud/pkg/struct2db"
)

// Controller is the main component that gets and saves objects in the database and generates CRUD HTTP handler
// that can be attached to an HTTP server.
type Controller struct {
	struct2db *struct2db.Controller
}

// Values for CRUD operations
const OpRead = 2
const OpUpdate = 4
const OpCreate = 8
const OpDelete = 16
const OpList = 32

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string) *Controller {
	c := &Controller{}
	c.struct2db = struct2db.NewController(dbConn, tblPrefix, &struct2db.ControllerConfig{
		TagName: "crud",
	})
	return c
}

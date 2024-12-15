package restapi

import (
	"database/sql"

	struct2db "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

// Controller is the main component that gets and saves objects in the database and generates CRUD HTTP handler
// that can be attached to an HTTP server.
type Controller struct {
	struct2db *struct2db.Controller
}

type ControllerConfig struct {
	TagName           string
	PasswordGenerator func(string) string
}

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string, cfg *ControllerConfig) *Controller {
	c := &Controller{}

	tagName := "restapi"
	if cfg != nil && cfg.TagName != "" {
		tagName = cfg.TagName
	}

	c.struct2db = struct2db.NewController(dbConn, tblPrefix, &struct2db.ControllerConfig{
		TagName: tagName,
	})

	return c
}

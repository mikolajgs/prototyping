package struct2db

import (
	"database/sql"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

// Controller is the main component that gets and saves objects in the database.
type Controller struct {
	dbConn        *sql.DB
	dbTblPrefix   string
	sqlGenerators map[string]*struct2sql.Struct2sql
	tagName string
}

type ControllerConfig struct {
	TagName string
}

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string, cfg *ControllerConfig) *Controller {
	c := &Controller{
		dbConn:      dbConn,
		dbTblPrefix: tblPrefix,
	}

	if cfg != nil && cfg.TagName != "" {
		c.tagName = cfg.TagName
	}

	if c.tagName == "" {
		c.tagName = "2db"
	}

	c.sqlGenerators = make(map[string]*struct2sql.Struct2sql)
	return c
}

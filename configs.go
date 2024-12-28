package prototyping

import (
	ui "github.com/go-phings/crud-ui"
)

type Config struct {
	DatabaseDSN       string
	UserConstructor   func() interface{}
	IntFieldValues    map[string]ui.IntFieldValues
	StringFieldValues map[string]ui.StringFieldValues
	ORM               ORM
}

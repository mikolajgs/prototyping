package prototyping

import "github.com/mikolajgs/prototyping/pkg/ui"

type Config struct {
	DatabaseDSN     string
	UserConstructor func() interface{}
	IntFieldValues  map[string]ui.FieldValues
}

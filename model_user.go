package prototyping

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type DefaultUser struct {
	ctl         *sdb.Controller
	user        UserInterface
	constructor func() UserInterface
}

func (g *DefaultUser) CreateDBTable() error {
	err := g.ctl.CreateTable(g.user)
	if err != nil {
		return err
	}
	return nil
}
func (g *DefaultUser) GetID() int64 {
	return g.user.GetID()
}
func (g *DefaultUser) GetEmail() string {
	return g.user.GetEmail()
}
func (g *DefaultUser) GetPassword() string {
	return g.user.GetPassword()
}
func (g *DefaultUser) GetEmailActivationKey() string {
	return g.user.GetEmailActivationKey()
}
func (g *DefaultUser) GetFlags() int64 {
	return g.user.GetFlags()
}
func (g *DefaultUser) GetExtraField(n string) string {
	if n == "name" {
		return g.user.GetName()
	}
	return ""
}
func (g *DefaultUser) SetEmail(e string) {
	g.user.SetEmail(e)
}
func (g *DefaultUser) SetPassword(p string) {
	g.user.SetPassword(p)
}
func (g *DefaultUser) SetEmailActivationKey(k string) {
	g.user.SetEmailActivationKey(k)
}
func (g *DefaultUser) SetFlags(flags int64) {
	g.user.SetFlags(flags)
}
func (g *DefaultUser) SetExtraField(n string, v string) {
	if n == "name" {
		g.user.SetName(v)
	}
}
func (g *DefaultUser) Save() error {
	errCrud := g.ctl.Save(g.user, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("error in sdb.SaveToDB: %w", errCrud)
	}
	return nil
}
func (g *DefaultUser) GetByID(id int64) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return g.constructor() }, sdb.GetOptions{
		Order:   []string{"ID", "asc"},
		Limit:   1,
		Offset:  0,
		Filters: map[string]interface{}{"ID": id},
	})
	if errCrud != nil {
		return false, fmt.Errorf("error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(UserInterface)
	return true, nil
}
func (g *DefaultUser) GetByEmail(email string) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return g.constructor() }, sdb.GetOptions{
		Order:   []string{"ID", "asc"},
		Limit:   1,
		Offset:  0,
		Filters: map[string]interface{}{g.user.GetEmailFieldName(): email},
	})
	if errCrud != nil {
		return false, fmt.Errorf("error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(UserInterface)
	return true, nil
}
func (g *DefaultUser) GetByEmailActivationKey(key string) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return g.constructor() }, sdb.GetOptions{
		Order:   []string{"id", "asc"},
		Limit:   1,
		Offset:  0,
		Filters: map[string]interface{}{g.user.GetEmailActivationKeyFieldName(): key},
	})
	if errCrud != nil {
		return false, fmt.Errorf("error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(UserInterface)
	return true, nil
}

func (g *DefaultUser) GetUser() interface{} {
	return &g.user
}

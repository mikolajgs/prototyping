package prototyping

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type defaultUser struct {
	ctl         *sdb.Controller
	user        userInterface
	constructor func() userInterface
}

func (g *defaultUser) CreateDBTable() error {
	err := g.ctl.CreateTable(g.user)
	if err != nil {
		return err
	}
	return nil
}
func (g *defaultUser) GetID() int64 {
	return g.user.GetID()
}
func (g *defaultUser) GetEmail() string {
	return g.user.GetEmail()
}
func (g *defaultUser) GetPassword() string {
	return g.user.GetPassword()
}
func (g *defaultUser) GetEmailActivationKey() string {
	return g.user.GetEmailActivationKey()
}
func (g *defaultUser) GetFlags() int64 {
	return g.user.GetFlags()
}
func (g *defaultUser) GetExtraField(n string) string {
	if n == "name" {
		return g.user.GetName()
	}
	return ""
}
func (g *defaultUser) SetEmail(e string) {
	g.user.SetEmail(e)
}
func (g *defaultUser) SetPassword(p string) {
	g.user.SetPassword(p)
}
func (g *defaultUser) SetEmailActivationKey(k string) {
	g.user.SetEmailActivationKey(k)
}
func (g *defaultUser) SetFlags(flags int64) {
	g.user.SetFlags(flags)
}
func (g *defaultUser) SetExtraField(n string, v string) {
	if n == "name" {
		g.user.SetName(v)
	}
}
func (g *defaultUser) Save() error {
	errCrud := g.ctl.Save(g.user, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("error in sdb.SaveToDB: %w", errCrud)
	}
	return nil
}
func (g *defaultUser) GetByID(id int64) (bool, error) {
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

	g.user = users[0].(userInterface)
	return true, nil
}
func (g *defaultUser) GetByEmail(email string) (bool, error) {
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

	g.user = users[0].(userInterface)
	return true, nil
}
func (g *defaultUser) GetByEmailActivationKey(key string) (bool, error) {
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

	g.user = users[0].(userInterface)
	return true, nil
}

func (g *defaultUser) GetUser() interface{} {
	return &g.user
}

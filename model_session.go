package prototyping

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type defaultSession struct {
	ctl         *sdb.Controller
	session     SessionInterface
	constructor func() SessionInterface
}

func (g *defaultSession) CreateDBTable() error {
	err := g.ctl.CreateTable(g.session)
	if err != nil {
		return err
	}
	return nil
}
func (g *defaultSession) GetFlags() int64 {
	return g.session.GetFlags()
}
func (g *defaultSession) GetKey() string {
	return g.session.GetKey()
}
func (g *defaultSession) GetExpiresAt() int64 {
	return g.session.GetExpiresAt()
}
func (g *defaultSession) GetUserID() int64 {
	return g.session.GetUserID()
}
func (g *defaultSession) SetFlags(flags int64) {
	g.session.SetFlags(flags)
}
func (g *defaultSession) SetKey(k string) {
	g.session.SetKey(k)
}
func (g *defaultSession) SetExpiresAt(exp int64) {
	g.session.SetExpiresAt(exp)
}
func (g *defaultSession) SetUserID(i int64) {
	g.session.SetUserID(i)
}
func (g *defaultSession) Save() error {
	errCrud := g.ctl.Save(g.session, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("error in sdb.SaveDB: %w", errCrud)
	}
	return nil
}
func (g *defaultSession) GetByKey(key string) (bool, error) {
	sessions, errCrud := g.ctl.Get(func() interface{} { return g.constructor() }, sdb.GetOptions{
		Order:   []string{"ID", "asc"},
		Limit:   1,
		Offset:  0,
		Filters: map[string]interface{}{"Key": key},
	})

	if errCrud != nil {
		return false, fmt.Errorf("error in sdb.Get: %w", errCrud)
	}
	if len(sessions) == 0 {
		return false, nil
	}

	g.session = sessions[0].(SessionInterface)
	return true, nil
}
func (g *defaultSession) GetSession() interface{} {
	return &g.session
}

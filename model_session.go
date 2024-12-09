package prototyping

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type DefaultSession struct {
	ctl         *sdb.Controller
	session     SessionInterface
	constructor func() SessionInterface
}

func (g *DefaultSession) CreateDBTable() error {
	err := g.ctl.CreateTable(g.session)
	if err != nil {
		return err
	}
	return nil
}
func (g *DefaultSession) GetFlags() int64 {
	return g.session.GetFlags()
}
func (g *DefaultSession) GetKey() string {
	return g.session.GetKey()
}
func (g *DefaultSession) GetExpiresAt() int64 {
	return g.session.GetExpiresAt()
}
func (g *DefaultSession) GetUserID() int64 {
	return g.session.GetUserID()
}
func (g *DefaultSession) SetFlags(flags int64) {
	g.session.SetFlags(flags)
}
func (g *DefaultSession) SetKey(k string) {
	g.session.SetKey(k)
}
func (g *DefaultSession) SetExpiresAt(exp int64) {
	g.session.SetExpiresAt(exp)
}
func (g *DefaultSession) SetUserID(i int64) {
	g.session.SetUserID(i)
}
func (g *DefaultSession) Save() error {
	errCrud := g.ctl.Save(g.session, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("error in sdb.SaveDB: %w", errCrud)
	}
	return nil
}
func (g *DefaultSession) GetByKey(key string) (bool, error) {
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
func (g *DefaultSession) GetSession() interface{} {
	return &g.session
}

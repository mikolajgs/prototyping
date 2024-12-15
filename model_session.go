package prototyping

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type Session struct {
	ID          int64  `json:"session_id"`
	Flags       int64  `json:"flags"`
	Key         string `json:"key" ui:"uniq lenmin:32 lenmax:2000"`
	ExpiresAt   int64  `json:"expires_at"`
	UserID      int64  `json:"user_id" ui:"req"`
	Description string `json:"description"`
}

type defaultSession struct {
	ctl         *sdb.Controller
	session     *Session
	constructor func() *Session
}

func (g *defaultSession) CreateDBTable() error {
	err := g.ctl.CreateTable(g.session)
	if err != nil {
		return err
	}
	return nil
}
func (g *defaultSession) GetFlags() int64 {
	return g.session.Flags
}
func (g *defaultSession) GetKey() string {
	return g.session.Key
}
func (g *defaultSession) GetExpiresAt() int64 {
	return g.session.ExpiresAt
}
func (g *defaultSession) GetUserID() int64 {
	return g.session.UserID
}
func (g *defaultSession) SetFlags(flags int64) {
	g.session.Flags = flags
}
func (g *defaultSession) SetKey(k string) {
	g.session.Key = k
}
func (g *defaultSession) SetExpiresAt(exp int64) {
	g.session.ExpiresAt = exp
}
func (g *defaultSession) SetUserID(i int64) {
	g.session.UserID = i
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

	g.session = sessions[0].(*Session)
	return true, nil
}
func (g *defaultSession) GetSession() interface{} {
	return &g.session
}

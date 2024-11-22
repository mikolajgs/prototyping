package umbrella

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type Session struct {
	ID        int    `json:"session_id"`
	Flags     int    `json:"flags"`
	Key       string `json:"key" 2db:"uniq lenmin:32 lenmax:2000"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    int    `json:"user_id" 2db:"req"`
}

// DefaultSession is default implementation of SessionInterface using struct-db-postgres
type DefaultSession struct {
	ctl *sdb.Controller
	session          *Session
}

func (g *DefaultSession) CreateDBTable() error {
	session := &Session{}
	err := g.ctl.CreateTable(session)
	if err != nil {
		return err
	}
	return nil
}
func (g *DefaultSession) GetFlags() int {
	return g.session.Flags
}
func (g *DefaultSession) GetKey() string {
	return g.session.Key
}
func (g *DefaultSession) GetExpiresAt() int64 {
	return g.session.ExpiresAt
}
func (g *DefaultSession) GetUserID() int {
	return g.session.UserID
}
func (g *DefaultSession) SetFlags(flags int) {
	g.session.Flags = flags
}
func (g *DefaultSession) SetKey(k string) {
	g.session.Key = k
}
func (g *DefaultSession) SetExpiresAt(exp int64) {
	g.session.ExpiresAt = exp
}
func (g *DefaultSession) SetUserID(i int) {
	g.session.UserID = i
}
func (g *DefaultSession) Save() error {
	errCrud := g.ctl.Save(g.session, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("Error in sdb.SaveDB: %w", errCrud)
	}
	return nil
}
func (g *DefaultSession) GetByKey(key string) (bool, error) {
	sessions, errCrud := g.ctl.Get(func() interface{} { return &Session{} }, sdb.GetOptions{
		Order: []string{"ID", "asc"},
		Limit: 1,
		Offset: 0,
		Filters: map[string]interface{}{"Key": key},
	})

	if errCrud != nil {
		return false, fmt.Errorf("Error in sdb.Get: %w", errCrud)
	}
	if len(sessions) == 0 {
		return false, nil
	}

	g.session = sessions[0].(*Session)
	return true, nil
}

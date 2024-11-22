package umbrella

import (
	"fmt"

	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

type User struct {
	ID                 int    `json:"user_id"`
	Flags              int    `json:"flags"`
	Name               string `json:"name" 2db:"lenmin:0 lenmax:50"`
	Email              string `json:"email" 2db:"req"`
	Password           string `json:"password"`
	EmailActivationKey string `json:"email_activation_key" 2db:""`
	CreatedAt          int    `json:"created_at"`
	CreatedByUserID    int    `json:"created_by_user_id"`
	LastModifiedAt          int    `json:"last_modified_at"`
	LastModifiedByUserID    int    `json:"last_modified_by_user_id"`
}

// DefaultUserModel is default implementation of UserInterface using struct-db-postgres package
type DefaultUser struct {
	ctl *sdb.Controller
	user *User
}

func (g *DefaultUser) CreateDBTable() error {
	user := &User{}
	err := g.ctl.CreateTable(user)
	if err != nil {
		return err
	}
	return nil
}
func (g *DefaultUser) GetID() int {
	return g.user.ID
}
func (g *DefaultUser) GetEmail() string {
	return g.user.Email
}
func (g *DefaultUser) GetPassword() string {
	return g.user.Password
}
func (g *DefaultUser) GetEmailActivationKey() string {
	return g.user.EmailActivationKey
}
func (g *DefaultUser) GetFlags() int {
	return g.user.Flags
}
func (g *DefaultUser) GetExtraField(n string) string {
	if n == "name" {
		return g.user.Name
	}
	return ""
}
func (g *DefaultUser) SetEmail(e string) {
	g.user.Email = e
}
func (g *DefaultUser) SetPassword(p string) {
	g.user.Password = p
}
func (g *DefaultUser) SetEmailActivationKey(k string) {
	g.user.EmailActivationKey = k
}
func (g *DefaultUser) SetFlags(flags int) {
	g.user.Flags = flags
}
func (g *DefaultUser) SetExtraField(n string, v string) {
	if n == "name" {
		g.user.Name = v
	}
}
func (g *DefaultUser) Save() error {
	errCrud := g.ctl.Save(g.user, sdb.SaveOptions{})
	if errCrud != nil {
		return fmt.Errorf("Error in sdb.SaveToDB: %w", errCrud)
	}
	return nil
}
func (g *DefaultUser) GetByID(id int) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, sdb.GetOptions{
		Order: []string{"ID", "asc"},
		Limit: 1,
		Offset: 0,
		Filters: map[string]interface{}{"ID": id},
	})
	if errCrud != nil {
		return false, fmt.Errorf("Error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(*User)
	return true, nil
}
func (g *DefaultUser) GetByEmail(email string) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, sdb.GetOptions{
		Order: []string{"ID", "asc"},
		Limit: 1,
		Offset: 0,
		Filters: map[string]interface{}{"Email": email},
	})
	if errCrud != nil {
		return false, fmt.Errorf("Error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(*User)
	return true, nil
}
func (g *DefaultUser) GetByEmailActivationKey(key string) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, sdb.GetOptions{
		Order: []string{"id", "asc"},
		Limit: 1,
		Offset: 0,
		Filters: map[string]interface{}{"EmailActivationKey": key},
	})
	if errCrud != nil {
		return false, fmt.Errorf("Error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(*User)
	return true, nil
}

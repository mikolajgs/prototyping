package umbrella

import (
	"fmt"
)

type User struct {
	ID                 int64  `json:"user_id"`
	Flags              int64  `json:"flags"`
	Name               string `json:"name" 2db:"lenmin:0 lenmax:50"`
	Email              string `json:"email" 2db:"req"`
	Password           string `json:"password"`
	EmailActivationKey string `json:"email_activation_key"`
	CreatedAt          int64  `json:"created_at"`
	CreatedBy          int64  `json:"created_by"`
	LastModifiedAt     int64  `json:"last_modified_at"`
	LastModifiedBy     int64  `json:"last_modified_by"`
}

const FlagUserActive = 1
const FlagUserEmailConfirmed = 2
const FlagUserAllowLogin = 4

// DefaultUserModel is default implementation of UserInterface using struct-db-postgres package
type DefaultUser struct {
	ctl  ORM
	user *User
}

func (g *DefaultUser) CreateDBTable() error {
	user := &User{}
	err := g.ctl.CreateTables(user)
	if err != nil {
		return err
	}
	return nil
}
func (g *DefaultUser) GetID() int64 {
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
func (g *DefaultUser) GetFlags() int64 {
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
func (g *DefaultUser) SetFlags(flags int64) {
	g.user.Flags = flags
}
func (g *DefaultUser) SetExtraField(n string, v string) {
	if n == "name" {
		g.user.Name = v
	}
}
func (g *DefaultUser) Save() error {
	errCrud := g.ctl.Save(g.user)
	if errCrud != nil {
		return fmt.Errorf("Error in sdb.SaveToDB: %w", errCrud)
	}
	return nil
}
func (g *DefaultUser) GetByID(id int64) (bool, error) {
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, []string{"ID", "asc"}, 1, 0, map[string]interface{}{"ID": id}, nil)
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
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, []string{"ID", "asc"}, 1, 0, map[string]interface{}{"Email": email}, nil)
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
	users, errCrud := g.ctl.Get(func() interface{} { return &User{} }, []string{"id", "asc"}, 1, 0, map[string]interface{}{"EmailActivationKey": key}, nil)
	if errCrud != nil {
		return false, fmt.Errorf("Error in sdb.GetFromDB: %w", errCrud)
	}
	if len(users) == 0 {
		return false, nil
	}

	g.user = users[0].(*User)
	return true, nil
}

func (g *DefaultUser) GetUser() interface{} {
	return g.user
}

package umbrella

type UserInterface interface {
	CreateDBTable() error

	GetID() int64
	GetEmail() string
	SetEmail(string)
	GetPassword() string
	SetPassword(string)
	GetEmailActivationKey() string
	SetEmailActivationKey(string)
	GetFlags() int64
	SetFlags(int64)
	GetExtraField(n string) string
	SetExtraField(n string, v string)

	Save() error
	GetByID(int64) (bool, error)
	GetByEmail(string) (bool, error)
	GetByEmailActivationKey(string) (bool, error)

	GetUser() interface{}
}

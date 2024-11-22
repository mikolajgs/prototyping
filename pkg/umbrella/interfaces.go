package umbrella

type UserInterface interface {
	CreateDBTable() error

	GetID() int
	GetEmail() string
	SetEmail(string)
	GetPassword() string
	SetPassword(string)
	GetEmailActivationKey() string
	SetEmailActivationKey(string)
	GetFlags() int
	SetFlags(int)
	GetExtraField(n string) string
	SetExtraField(n string, v string)

	Save() error
	GetByID(int) (bool, error)
	GetByEmail(string) (bool, error)
	GetByEmailActivationKey(string) (bool, error)
}

type SessionInterface interface {
	CreateDBTable() error

	GetKey() string
	SetKey(string)
	GetExpiresAt() int64
	SetExpiresAt(int64)
	GetUserID() int
	SetUserID(int)
	GetFlags() int
	SetFlags(int)

	Save() error
	GetByKey(string) (bool, error)
}

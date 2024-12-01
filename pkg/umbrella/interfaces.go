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
}

type SessionInterface interface {
	CreateDBTable() error

	GetKey() string
	SetKey(string)
	GetExpiresAt() int64
	SetExpiresAt(int64)
	GetUserID() int64
	SetUserID(int64)
	GetFlags() int64
	SetFlags(int64)

	Save() error
	GetByKey(string) (bool, error)
}

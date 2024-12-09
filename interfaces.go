package prototyping

type UserInterface interface {
	GetID() int64
	GetEmail() string
	GetPassword() string
	GetEmailActivationKey() string
	GetFlags() int64
	GetName() string
	SetEmail(string)
	SetPassword(string)
	SetEmailActivationKey(string)
	SetFlags(int64)
	SetName(string)
	GetEmailFieldName() string
	GetEmailActivationKeyFieldName() string
}

type SessionInterface interface {
	GetFlags() int64
	GetKey() string
	GetExpiresAt() int64
	GetUserID() int64
	SetFlags(int64)
	SetKey(string)
	SetExpiresAt(int64)
	SetUserID(int64)
	GetKeyFieldName() string
}

package prototyping

type userInterface interface {
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

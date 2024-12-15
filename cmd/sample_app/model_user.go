package main

type User struct {
	ID                 int64  `json:"user_id"`
	Flags              int64  `json:"flags"`
	Name               string `json:"name" ui:"lenmin:0 lenmax:50"`
	Email              string `json:"email" ui:"req"`
	Password           string `json:"password" ui:"hidden password uipassword dblentry"`
	EmailActivationKey string `json:"email_activation_key" ui:"hidden"`
	CreatedAt          int64  `json:"created_at"`
	CreatedBy          int64  `json:"created_by"`
	LastModifiedAt     int64  `json:"last_modified_at"`
	LastModifiedBy     int64  `json:"last_modified_by"`
}

func (u *User) GetID() int64 {
	return u.ID
}
func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetPassword() string {
	return u.Password
}
func (u *User) GetEmailActivationKey() string {
	return u.EmailActivationKey
}
func (u *User) GetFlags() int64 {
	return u.Flags
}
func (u *User) GetName() string {
	return u.Name
}
func (u *User) SetEmail(e string) {
	u.Email = e
}
func (u *User) SetPassword(p string) {
	u.Password = p
}
func (u *User) SetEmailActivationKey(k string) {
	u.EmailActivationKey = k
}
func (u *User) SetFlags(i int64) {
	u.Flags = i
}
func (u *User) SetName(n string) {
	u.Name = n
}
func (u User) GetEmailFieldName() string {
	return "Email"
}
func (u User) GetEmailActivationKeyFieldName() string {
	return "EmailActivationKey"
}

package main

type Item struct {
	ID             int64  `json:"item_id"`
	Flags          int64  `json:"item_flags"`
	Title          string `ui:"req lenmin:5 lenmax:200" json:"title"`
	Text           string `ui:"lenmax:5000 db_type:VARCHAR(5000)" json:"text"`
	CreatedAt      int64  `json:"created_at"`
	CreatedBy      int64  `json:"created_by"`
	LastModifiedAt int64  `json:"last_modified_at"`
	LastModifiedBy int64  `json:"last_modified_by"`
}

type ItemGroup struct {
	ID             int64  `json:"item_group_id"`
	Flags          int64  `json:"item_group_flags"`
	Name           string `ui:"req lenmin:3 lenmax:30" json:"name"`
	Description    string `ui:"lenmax:255 db_type:VARCHAR(255)" json:"description"`
	CreatedAt      int64  `json:"created_at"`
	CreatedBy      int64  `json:"created_by"`
	LastModifiedAt int64  `json:"last_modified_at"`
	LastModifiedBy int64  `json:"last_modified_by"`
}

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

type Session struct {
	ID        int64  `json:"session_id"`
	Flags     int64  `json:"flags"`
	Key       string `json:"key" ui:"uniq lenmin:32 lenmax:2000"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    int64  `json:"user_id" ui:"req"`
}

func (s *Session) GetFlags() int64 {
	return s.Flags
}
func (s *Session) GetKey() string {
	return s.Key
}
func (s *Session) GetExpiresAt() int64 {
	return s.ExpiresAt
}
func (s *Session) GetUserID() int64 {
	return s.UserID
}
func (s *Session) SetFlags(i int64) {
	s.Flags = i
}
func (s *Session) SetKey(k string) {
	s.Key = k
}
func (s *Session) SetExpiresAt(i int64) {
	s.ExpiresAt = i
}
func (s *Session) SetUserID(i int64) {
	s.UserID = i
}
func (s Session) GetKeyFieldName() string {
	return "Key"
}

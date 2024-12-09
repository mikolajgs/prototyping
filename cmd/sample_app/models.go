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
	Password           string `json:"password" ui:"hidden"`
	EmailActivationKey string `json:"email_activation_key" ui:"hidden"`
	CreatedAt          int64  `json:"created_at"`
	CreatedBy          int64  `json:"created_by"`
	LastModifiedAt     int64  `json:"last_modified_at"`
	LastModifiedBy     int64  `json:"last_modified_by"`
}

type Session struct {
	ID        int64  `json:"session_id"`
	Flags     int64  `json:"flags"`
	Key       string `json:"key" ui:"uniq lenmin:32 lenmax:2000"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    int64  `json:"user_id" ui:"req"`
}

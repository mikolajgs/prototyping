package umbrella

type Session struct {
	ID          int64  `json:"session_id"`
	Flags       int64  `json:"flags"`
	Key         string `json:"key" 2db:"uniq lenmin:32 lenmax:2000"`
	ExpiresAt   int64  `json:"expires_at"`
	UserID      int64  `json:"user_id" 2db:"req"`
	Description string `json:"string"`
}

const FlagSessionActive = 1
const FlagSessionLoggedOut = 2

func GetSessionFlagsSingleChoice() map[int]string {
	return map[int]string{
		1: "Active",
		2: "LoggedOut",
	}
}

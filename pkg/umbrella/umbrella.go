package umbrella

import (
	"context"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgrijalva/jwt-go"
	sdb "github.com/go-phings/struct-db-postgres"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const DisableRegister = 1
const DisableConfirm = 2
const DisableLogin = 4
const DisableCheck = 8
const RegisterConfirmed = 16
const RegisterAllowedToLogin = 32

type Umbrella struct {
	dbConn          *sql.DB
	dbTblPrefix     string
	orm             ORM
	jwtConfig       *JWTConfig
	Hooks           *Hooks
	Interfaces      *Interfaces
	Flags           int
	UserExtraFields []UserExtraField
	tagName         string
}

type JWTConfig struct {
	Key               string
	ExpirationMinutes int
	Issuer            string
}

type Hooks struct {
	PostRegisterSuccess func(http.ResponseWriter, string) bool
	PostConfirmSuccess  func(http.ResponseWriter) bool
	PostLoginSuccess    func(http.ResponseWriter, string, string, int64) bool
	PostCheckSuccess    func(http.ResponseWriter, string, int64, bool) bool
	PostLogoutSuccess   func(http.ResponseWriter, string) bool
	// More to come
}

type Interfaces struct {
	User func() UserInterface
}

type UserExtraField struct {
	Name         string
	RegExp       *regexp.Regexp
	DefaultValue string
}

type HandlerConfig struct {
	UseCookie          string
	CookiePath         string
	SuccessRedirectURL string
	FailureRedirectURL string
}

type UmbrellaConfig struct {
	TagName           string
	NoUserConstructor bool
	ORM               ORM
}

type UmbrellaContextValue string

type customClaims struct {
	jwt.StandardClaims
	SID string
}

func NewUmbrella(dbConn *sql.DB, tblPrefix string, jwtConfig *JWTConfig, cfg *UmbrellaConfig) *Umbrella {
	u := &Umbrella{
		dbConn:      dbConn,
		dbTblPrefix: tblPrefix,
		jwtConfig:   jwtConfig,
	}

	if dbConn == nil {
		log.Fatalf("Umbrella requires DB Connection")
	}

	tagName := "2db"
	if cfg != nil && cfg.TagName != "" {
		tagName = cfg.TagName
	}
	u.tagName = tagName

	if cfg != nil && cfg.ORM != nil {
		u.orm = cfg.ORM
	} else {
		u.orm = newWrappedStruct2db(dbConn, tblPrefix, u.tagName)
	}

	if cfg != nil && cfg.NoUserConstructor {
		return u
	}

	u.Interfaces = &Interfaces{
		User: func() UserInterface {
			user := &User{}
			return &DefaultUser{
				ctl:  u.orm,
				user: user,
			}
		},
	}

	return u
}

func (u Umbrella) CreateDBTables() *ErrUmbrella {
	user := u.Interfaces.User()
	err := user.CreateDBTable()
	if err != nil {
		return &ErrUmbrella{
			Op:  "CreateDBTables",
			Err: err,
		}
	}

	err2 := u.orm.CreateTables(&Session{})
	if err2 != nil {
		return &ErrUmbrella{
			Op:  "CreateDBTables",
			Err: err2,
		}
	}

	err2 = u.orm.CreateTables(&Permission{})
	if err2 != nil {
		return &ErrUmbrella{
			Op:  "CreateDBTables",
			Err: err2,
		}
	}

	return nil
}

func (u Umbrella) GetHTTPHandler(uri string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := u.getURIFromRequest(r, uri)

		switch uri {
		case "register":
			if u.Flags&DisableRegister > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleRegister(w, r)
			}
		case "confirm":
			if u.Flags&DisableConfirm > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleConfirm(w, r)
			}
		case "login":
			u.handleLogin(w, r, nil, "", "")
		case "check":
			if u.Flags&DisableCheck > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleCheck(w, r)
			}
		case "logout":
			u.handleLogout(w, r, nil, "", "")
		default:
			u.writeErrText(w, http.StatusNotFound, "invalid_uri")
		}
	})
}

func (u Umbrella) GetLoginHTTPHandler(config HandlerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.handleLogin(w, r, &http.Cookie{
			Name:     config.UseCookie,
			Path:     config.CookiePath,
			Value:    "ReplaceMe",
			HttpOnly: false,
		}, config.SuccessRedirectURL, config.FailureRedirectURL)
	})
}

func (u Umbrella) GetLogoutHTTPHandler(config HandlerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.handleLogout(w, r, &http.Cookie{
			Name:     config.UseCookie,
			Path:     config.CookiePath,
			Value:    "ReplaceMe",
			HttpOnly: false,
		}, config.SuccessRedirectURL, config.FailureRedirectURL)
	})
}

func (u Umbrella) GetHTTPHandlerWrapper(next http.Handler, config HandlerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		if config.UseCookie != "" {
			cookie, err := r.Cookie(config.UseCookie)
			if err != nil {
				token = ""
			} else {
				token = cookie.Value
			}
		} else {
			token = GetAuthorizationBearerToken(r)
		}
		_, _, userID, _ := u.check(token, false)
		ctx := context.WithValue(r.Context(), UmbrellaContextValue("UmbrellaUserID"), int64(userID))
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func GetAuthorizationBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.Split(h, "Bearer")
	if len(parts) != 2 {
		return ""
	}
	token := strings.TrimSpace(parts[1])
	if len(token) < 1 {
		return ""
	}
	return token
}

func GetUserIDFromRequest(r *http.Request) int64 {
	v := r.Context().Value(UmbrellaContextValue("UmbrellaUserID")).(int64)
	return v
}

func (u Umbrella) getURIFromRequest(r *http.Request, uri string) string {
	uriPart := r.RequestURI[len(uri):]
	xs := strings.SplitN(uriPart, "?", 2)
	return xs[0]
}

func (u Umbrella) GeneratePassword(pass string) (string, *ErrUmbrella) {
	passEncrypted, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", &ErrUmbrella{
			Op:  "GeneratePassword",
			Err: err,
		}
	}

	return base64.StdEncoding.EncodeToString(passEncrypted), nil
}

func (u Umbrella) CreateUser(email string, pass string, extraFields map[string]string) (string, *ErrUmbrella) {
	pass, err := u.GeneratePassword(pass)
	if err != nil {
		return "", err
	}

	key := uuid.New().String()

	user := u.Interfaces.User()
	user.SetEmail(email)
	user.SetPassword(pass)
	for k, v := range extraFields {
		user.SetExtraField(k, v)
	}
	user.SetEmailActivationKey(key)

	var flags int64
	flags = FlagUserActive
	if u.Flags&RegisterConfirmed > 0 {
		flags += FlagUserEmailConfirmed
	}
	if u.Flags&RegisterAllowedToLogin > 0 {
		flags += FlagUserAllowLogin
	}
	user.SetFlags(flags)

	err2 := user.Save()
	if err2 != nil {
		return "", &ErrUmbrella{
			Op:  "SaveToDB",
			Err: err2,
		}
	}

	return key, nil
}

func (u Umbrella) ConfirmEmail(key string) *ErrUmbrella {
	user := u.Interfaces.User()
	got, err := user.GetByEmailActivationKey(key)

	if !got {
		if err == nil {
			return &ErrUmbrella{
				Op:  "NoRow",
				Err: nil,
			}
		}
		if err != nil {
			return &ErrUmbrella{
				Op:  "GetFromDB",
				Err: err,
			}
		}
	}
	if user.GetFlags()&FlagUserActive == 0 {
		return &ErrUmbrella{
			Op:  "UserInactive",
			Err: err,
		}
	}

	user.SetFlags(user.GetFlags() | FlagUserEmailConfirmed | FlagUserAllowLogin)
	user.SetEmailActivationKey("")
	err = user.Save()
	if err != nil {
		return &ErrUmbrella{
			Op:  "SaveToDB",
			Err: err,
		}
	}

	return nil
}

func (u Umbrella) GetUserOperationAllowedTypes(i int64, o int) (map[string]bool, error) {
	perms, err := u.orm.Get(func() interface{} { return &Permission{} }, []string{"Flags", "ASC"}, 30, 0, map[string]interface{}{
			"_raw": []interface{}{
				"(.ForType=? OR (.ForType=? AND .ForItem=?)) AND .Ops&?>0",
				ForTypeEveryone,
				ForTypeUser,
				i,
				o,
			},
			"_rawConjuction": sdb.RawConjuctionOR,
		}, nil)
	if err != nil {
		return map[string]bool{}, err
	}

	types := map[string]bool{}
	for _, v := range perms {
		p := v.(*Permission)
		// only allow flags for now
		if p.Flags&FlagTypeAllow == 0 {
			continue
		}
		// everyone or particular user
		if !(p.ForType == ForTypeEveryone || (p.ForType == ForTypeUser && p.ForItem == i)) {
			continue
		}

		types[p.ToType] = true
	}

	return types, nil
}

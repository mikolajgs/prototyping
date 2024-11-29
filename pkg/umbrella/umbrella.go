package umbrella

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	"golang.org/x/crypto/bcrypt"
)

const FlagUserActive = 1
const FlagUserEmailConfirmed = 2
const FlagUserAllowLogin = 4

const FlagSessionActive = 1
const FlagSessionLoggedOut = 2

const DisableRegister = 1
const DisableConfirm = 2
const DisableLogin = 4
const DisableCheck = 8
const RegisterConfirmed = 16
const RegisterAllowedToLogin = 32

type Umbrella struct {
	dbConn           *sql.DB
	dbTblPrefix      string
	goCRUDController *sdb.Controller
	jwtConfig        *JWTConfig
	Hooks            *Hooks
	Interfaces       *Interfaces
	Flags            int
	UserExtraFields  []UserExtraField
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
	User    func() UserInterface
	Session func() SessionInterface
}

type UserExtraField struct {
	Name         string
	RegExp       *regexp.Regexp
	DefaultValue string
}

type customClaims struct {
	jwt.StandardClaims
	SID string
}

func NewUmbrella(dbConn *sql.DB, tblPrefix string, jwtConfig *JWTConfig) *Umbrella {
	u := &Umbrella{
		dbConn:      dbConn,
		dbTblPrefix: tblPrefix,
		jwtConfig:   jwtConfig,
	}

	if dbConn == nil {
		log.Fatalf("Umbrella requires DB Connection")
	}

	u.goCRUDController = sdb.NewController(dbConn, tblPrefix, &sdb.ControllerConfig{
		TagName: "2db",
	})

	u.Interfaces = &Interfaces{
		User: func() UserInterface {
			user := &User{}
			return &DefaultUser{
				ctl: u.goCRUDController,
				user:             user,
			}
		},
		Session: func() SessionInterface {
			session := &Session{}
			return &DefaultSession{
				ctl: u.goCRUDController,
				session:          session,
			}
		},
	}

	return u
}

func (u Umbrella) CreateDBTables() *ErrUmbrella {
	user := u.Interfaces.User()
	session := u.Interfaces.Session()

	err := user.CreateDBTable()
	if err != nil {
		return &ErrUmbrella{
			Op:  "CreateDBTables",
			Err: err,
		}
	}

	err = session.CreateDBTable()
	if err != nil {
		return &ErrUmbrella{
			Op:  "CreateDBTables",
			Err: err,
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
			if u.Flags&DisableLogin > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleLogin(w, r)
			}
		case "check":
			if u.Flags&DisableCheck > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleCheck(w, r)
			}
		case "logout":
			if u.Flags&DisableLogin > 0 {
				u.writeErrText(w, http.StatusNotFound, "invalid_uri")
			} else {
				u.handleLogout(w, r)
			}
		default:
			u.writeErrText(w, http.StatusNotFound, "invalid_uri")
		}
	})
}

func (u Umbrella) GetHTTPHandlerWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetAuthorizationBearerToken(r)
		_, _, userID, _ := u.check(token, false)
		ctx := context.WithValue(r.Context(), "UmbrellaUserID", int64(userID))
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
	v := r.Context().Value("UmbrellaUserID").(int64)
	return v
}

func (u Umbrella) getURIFromRequest(r *http.Request, uri string) string {
	uriPart := r.RequestURI[len(uri):]
	xs := strings.SplitN(uriPart, "?", 2)
	return xs[0]
}

func (u Umbrella) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	if !u.isValidEmail(email) {
		u.writeErrText(w, http.StatusBadRequest, "invalid_email")
		return
	}
	if !u.isValidPassword(password) {
		u.writeErrText(w, http.StatusBadRequest, "invalid_or_weak_password")
		return
	}

	ok, err := u.isEmailExists(email)
	if err != nil {
		u.writeErrText(w, http.StatusInternalServerError, "database_error")
		return
	}
	if ok {
		u.writeErrText(w, http.StatusOK, "email_registered")
		return
	}

	extraFields := map[string]string{}
	if u.UserExtraFields != nil && len(u.UserExtraFields) > 0 {
		for _, v := range u.UserExtraFields {
			postVal := r.FormValue(v.Name)
			if v.RegExp != nil {
				if !v.RegExp.MatchString(postVal) {
					u.writeErrText(w, http.StatusBadRequest, "invalid_"+v.Name)
					return
				}
			}
			if postVal != "" {
				postVal = v.DefaultValue
			}
			extraFields[v.Name] = postVal
		}
	}

	_, err2 := u.createUser(email, password, extraFields)
	if err2 != nil {
		u.writeErrText(w, http.StatusInternalServerError, "create_error")
		return
	}

	if u.Hooks != nil && u.Hooks.PostRegisterSuccess != nil {
		if !u.Hooks.PostRegisterSuccess(w, email) {
			return
		}
	}

	u.writeOK(w, http.StatusCreated, map[string]interface{}{})
}

func (u Umbrella) handleConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}
	key := r.FormValue("key")
	if !u.isValidActivationKey(key) {
		u.writeErrText(w, http.StatusBadRequest, "invalid_key")
		return
	}

	err2 := u.confirmEmail(key)
	if err2 != nil {
		var errUmb *ErrUmbrella
		if errors.As(err2, &errUmb) {
			if errUmb.Op == "NoRow" || errUmb.Op == "UserInactive" {
				u.writeErrText(w, http.StatusNotFound, "invalid_key")
			} else if errUmb.Op == "GetFromDB" {
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				u.writeErrText(w, http.StatusInternalServerError, "confirm_error")
			}
		}
		return
	}

	if u.Hooks != nil && u.Hooks.PostConfirmSuccess != nil {
		if !u.Hooks.PostConfirmSuccess(w) {
			return
		}
	}

	u.writeOK(w, http.StatusOK, map[string]interface{}{})
}

func (u Umbrella) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	if !u.isValidEmail(email) {
		u.writeErrText(w, http.StatusBadRequest, "invalid_credentials")
		return
	}
	if password == "" {
		u.writeErrText(w, http.StatusBadRequest, "invalid_credentials")
		return
	}

	token, expiresAt, err := u.login(email, password)
	if err != nil {
		var errUmb *ErrUmbrella
		if errors.As(err, &errUmb) {
			if errUmb.Op == "NoRow" || errUmb.Op == "UserInactive" || errUmb.Op == "InvalidPassword" {
				u.writeErrText(w, http.StatusNotFound, "invalid_credentials")
			} else if errUmb.Op == "GetFromDB" {
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				u.writeErrText(w, http.StatusInternalServerError, "login_error")
			}
		}
		return
	}

	if u.Hooks != nil && u.Hooks.PostLoginSuccess != nil {
		if !u.Hooks.PostLoginSuccess(w, email, token, expiresAt) {
			return
		}
	}

	u.writeOK(w, http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (u Umbrella) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}

	token := r.FormValue("token")
	if !u.isValidToken(token) {
		u.writeErrText(w, http.StatusBadRequest, "invalid_token")
		return
	}

	refresh := false
	if r.FormValue("refresh") == "1" {
		refresh = true
	}

	token2, expiresAt, _, err := u.check(token, refresh)
	if err != nil {
		var errUmb *ErrUmbrella
		if errors.As(err, &errUmb) {
			if errUmb.Op == "InvalidToken" || errUmb.Op == "UserInactive" || errUmb.Op == "Expired" || errUmb.Op == "InvalidSession" || errUmb.Op == "InvalidUser" || errUmb.Op == "ParseToken" {
				u.writeErrText(w, http.StatusNotFound, "invalid_credentials")
			} else if errUmb.Op == "GetFromDB" {
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				u.writeErrText(w, http.StatusInternalServerError, "check_error")
			}
		}
		return
	}

	if u.Hooks != nil && u.Hooks.PostCheckSuccess != nil {
		if !u.Hooks.PostCheckSuccess(w, token, expiresAt, refresh) {
			return
		}
	}

	u.writeOK(w, http.StatusOK, map[string]interface{}{
		"token":      token2,
		"expires_at": expiresAt,
		"refreshed":  refresh,
	})
}

func (u Umbrella) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}
	token := r.FormValue("token")
	if token == "" {
		u.writeErrText(w, http.StatusBadRequest, "invalid_token")
		return
	}

	err := u.logout(token)
	if err != nil {
		var errUmb *ErrUmbrella
		if errors.As(err, &errUmb) {
			if errUmb.Op == "InvalidToken" || errUmb.Op == "Expired" || errUmb.Op == "ParseToken" || errUmb.Op == "InvalidSession" {
				u.writeErrText(w, http.StatusNotFound, "invalid_credentials")
			} else if errUmb.Op == "GetFromDB" {
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				u.writeErrText(w, http.StatusInternalServerError, "login_error")
			}
		}
		return
	}

	if u.Hooks != nil && u.Hooks.PostLogoutSuccess != nil {
		if !u.Hooks.PostLogoutSuccess(w, token) {
			return
		}
	}

	u.writeOK(w, http.StatusOK, map[string]interface{}{})
}

func (u Umbrella) writeErrText(w http.ResponseWriter, status int, errText string) {
	r := NewHTTPResponse(0, errText)
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}

func (u Umbrella) writeOK(w http.ResponseWriter, status int, data map[string]interface{}) {
	r := NewHTTPResponse(1, "")
	r.Data = data
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}

func (u Umbrella) createUser(email string, pass string, extraFields map[string]string) (string, *ErrUmbrella) {
	passEncrypted, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", &ErrUmbrella{
			Op:  "GeneratePassword",
			Err: err,
		}
	}

	key := uuid.New().String()

	user := u.Interfaces.User()
	user.SetEmail(email)
	user.SetPassword(base64.StdEncoding.EncodeToString(passEncrypted))
	for k, v := range extraFields {
		user.SetExtraField(k, v)
	}
	user.SetEmailActivationKey(key)

	flags := FlagUserActive
	if u.Flags&RegisterConfirmed > 0 {
		flags += FlagUserEmailConfirmed
	}
	if u.Flags&RegisterAllowedToLogin > 0 {
		flags += FlagUserAllowLogin
	}
	user.SetFlags(flags)

	err = user.Save()
	if err != nil {
		return "", &ErrUmbrella{
			Op:  "SaveToDB",
			Err: err,
		}
	}

	return key, nil
}

func (u Umbrella) confirmEmail(key string) *ErrUmbrella {
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

func (u Umbrella) login(email string, password string) (string, int64, *ErrUmbrella) {
	user := u.Interfaces.User()
	got, err := user.GetByEmail(email)

	if !got {
		if err == nil {
			return "", 0, &ErrUmbrella{
				Op:  "NoRow",
				Err: nil,
			}
		}
		if err != nil {
			return "", 0, &ErrUmbrella{
				Op:  "GetFromDB",
				Err: err,
			}
		}
	}
	if user.GetFlags()&FlagUserActive == 0 || user.GetFlags()&FlagUserAllowLogin == 0 {
		return "", 0, &ErrUmbrella{
			Op:  "UserInactive",
			Err: err,
		}
	}

	passwordInDBDecoded, err := base64.StdEncoding.DecodeString(user.GetPassword())
	if err != nil {
		return "", 0, &ErrUmbrella{
			Op:  "InvalidPassword",
			Err: err,
		}
	}
	err = bcrypt.CompareHashAndPassword(passwordInDBDecoded, []byte(password))
	if err != nil {
		return "", 0, &ErrUmbrella{
			Op:  "InvalidPassword",
			Err: err,
		}
	}

	sUUID := uuid.New().String()
	token, expiresAt, err := u.createToken(sUUID)
	if err != nil {
		return "", 0, &ErrUmbrella{
			Op:  "CreateToken",
			Err: err,
		}
	}

	userID := user.GetID()

	sess := u.Interfaces.Session()
	sess.SetKey(sUUID)
	sess.SetExpiresAt(expiresAt)
	sess.SetUserID(userID)
	sess.SetFlags(FlagSessionActive)
	err = sess.Save()
	if err != nil {
		return "", 0, &ErrUmbrella{
			Op:  "SaveToDB",
			Err: err,
		}
	}

	return token, expiresAt, nil
}

func (u Umbrella) logout(token string) *ErrUmbrella {
	sID, errUmbrella := u.parseTokenWithCheck(token)
	if errUmbrella != nil {
		return errUmbrella
	}

	session := u.Interfaces.Session()
	got, err := session.GetByKey(sID)

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

	if session.GetFlags()&FlagSessionActive == 0 || session.GetFlags()&FlagSessionLoggedOut > 0 {
		return &ErrUmbrella{
			Op:  "InvalidSession",
			Err: err,
		}
	}

	session.SetFlags(session.GetFlags() | FlagSessionLoggedOut)
	if session.GetFlags()&FlagSessionActive > 0 {
		session.SetFlags(session.GetFlags() - FlagSessionActive)
	}
	err = session.Save()
	if err != nil {
		return &ErrUmbrella{
			Op:  "SaveToDB",
			Err: err,
		}
	}

	return nil
}

func (u Umbrella) check(token string, refresh bool) (string, int64, int, *ErrUmbrella) {
	sID, errUmbrella := u.parseTokenWithCheck(token)
	if errUmbrella != nil {
		return "", 0, 0, errUmbrella
	}

	session := u.Interfaces.Session()
	got, err := session.GetByKey(sID)
	if !got {
		if err == nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "InvalidSession",
				Err: nil,
			}
		}
		if err != nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "GetFromDB",
				Err: err,
			}
		}
	}

	if session.GetFlags()&FlagSessionActive == 0 || session.GetFlags()&FlagSessionLoggedOut > 0 {
		return "", 0, 0, &ErrUmbrella{
			Op:  "InvalidSession",
			Err: err,
		}
	}

	user := u.Interfaces.User()
	got, err = user.GetByID(session.GetUserID())
	if !got {
		if err == nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "InvalidUser",
				Err: err,
			}
		}
		if err != nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "GetFromDB",
				Err: err,
			}
		}
	}
	if user.GetFlags()&FlagUserActive == 0 || user.GetFlags()&FlagUserAllowLogin == 0 {
		return "", 0, 0, &ErrUmbrella{
			Op:  "UserInactive",
			Err: nil,
		}
	}

	if refresh {
		token2, expiresAt, err := u.createToken(sID)
		if err != nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "CreateToken",
				Err: err,
			}
		}

		session.SetExpiresAt(expiresAt)
		err = session.Save()
		if err != nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "SaveToDB",
				Err: err,
			}
		}
		return token2, expiresAt, session.GetUserID(), nil
	}

	return token, 0, session.GetUserID(), nil
}

func (u Umbrella) parseTokenWithCheck(token string) (string, *ErrUmbrella) {
	sID, expired, err := u.parseToken(token)
	if err != nil {
		return "", &ErrUmbrella{
			Op:  "ParseToken",
			Err: err,
		}
	}

	if expired {
		return "", &ErrUmbrella{
			Op:  "Expired",
			Err: err,
		}
	}

	if !u.isValidSessionID(sID) {
		return "", &ErrUmbrella{
			Op:  "InvalidSession",
			Err: err,
		}
	}

	return sID, nil
}

func (u Umbrella) createToken(sid string) (string, int64, error) {
	cc := customClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(u.jwtConfig.ExpirationMinutes) * time.Minute).Unix(),
			Issuer:    u.jwtConfig.Issuer,
		},
		SID: sid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cc)
	st, err := token.SignedString([]byte(u.jwtConfig.Key))
	if err != nil {
		return "", 0, fmt.Errorf("couldn't sign token in createToken %w", err)
	}
	return st, cc.StandardClaims.ExpiresAt, nil
}

func (u Umbrella) parseToken(st string) (string, bool, error) {
	token, err := jwt.ParseWithClaims(st, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("parseWithClaims different algorithms used")
		}
		return []byte(u.jwtConfig.Key), nil
	})

	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return token.Claims.(*customClaims).SID, true, nil
		}
	}

	if err != nil {
		return "", false, fmt.Errorf("couldn't ParseWithClaims in parseToken %w", err)
	}

	if token.Valid {
		return token.Claims.(*customClaims).SID, false, nil
	}

	return "", false, fmt.Errorf("token not valid in parseToken")
}

func (u Umbrella) isEmailExists(e string) (bool, error) {
	user := u.Interfaces.User()
	got, err := user.GetByEmail(e)
	if !got {
		if err == nil {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("Error with GetByEmail: %w", err)
		}
	}

	return true, nil
}

func (u Umbrella) isValidEmail(s string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegex.MatchString(s)
}

func (u Umbrella) isValidPassword(s string) bool {
	if len(s) < 12 {
		return false
	}
	return true
}

func (u Umbrella) isValidActivationKey(s string) bool {
	var keyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,255}$`)
	return keyRegex.MatchString(s)
}

func (u Umbrella) isValidSessionID(s string) bool {
	var keyRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{1,255}$`)
	return keyRegex.MatchString(s)
}

func (u Umbrella) isValidToken(s string) bool {
	var keyRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.\$]+$`)
	return keyRegex.MatchString(s)
}

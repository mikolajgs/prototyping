package umbrella

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

	_, err2 := u.CreateUser(email, password, extraFields)
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

	err2 := u.ConfirmEmail(key)
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

func (u Umbrella) handleLogin(w http.ResponseWriter, r *http.Request, setCookie *http.Cookie, successURI string, failureURI string) {
	if u.Flags&DisableLogin > 0 {
		if failureURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?err=disabled", failureURI))
			return
		}
		u.writeErrText(w, http.StatusNotFound, "invalid_uri")
		return
	}

	if r.Method != http.MethodPost {
		if failureURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?err=unknown", failureURI))
			return
		}
		u.writeErrText(w, http.StatusBadRequest, "invalid_request")
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	if !u.isValidEmail(email) {
		if failureURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?err=invalid_credentials", failureURI))
			return
		}
		u.writeErrText(w, http.StatusBadRequest, "invalid_credentials")
		return
	}
	if password == "" {
		if failureURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?err=invalid_credentials", failureURI))
			return
		}
		u.writeErrText(w, http.StatusBadRequest, "invalid_credentials")
		return
	}

	token, expiresAt, err := u.login(email, password)
	if err != nil {
		var errUmb *ErrUmbrella
		if errors.As(err, &errUmb) {
			if errUmb.Op == "NoRow" || errUmb.Op == "UserInactive" || errUmb.Op == "InvalidPassword" {
				if failureURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?err=invalid_credentials", failureURI))
					return
				}
				u.writeErrText(w, http.StatusNotFound, "invalid_credentials")
			} else if errUmb.Op == "GetFromDB" {
				if failureURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?err=database_error", failureURI))
					return
				}
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				if failureURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?err=login_error", failureURI))
					return
				}
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

	if successURI != "" {
		if setCookie != nil {
			cookieClone := http.Cookie{
				Name:     setCookie.Name,
				Path:     setCookie.Path,
				Value:    token,
				HttpOnly: setCookie.HttpOnly,
			}
			http.SetCookie(w, &cookieClone)
		}
		u.writeRedirect(w, successURI)
		return
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

func (u Umbrella) handleLogout(w http.ResponseWriter, r *http.Request, useCookie *http.Cookie, successURI string, failureURI string) {
	if u.Flags&DisableLogin > 0 {
		if failureURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?err=disabled", failureURI))
			return
		}
		u.writeErrText(w, http.StatusNotFound, "invalid_uri")
		return
	}

	var token string
	if useCookie != nil {
		if r.Method != http.MethodGet {
			if failureURI != "" {
				u.writeRedirect(w, fmt.Sprintf("%s?err=invalid_request", failureURI))
				return
			}
			u.writeErrText(w, http.StatusBadRequest, "invalid_request")
			return
		}

		cookie, err := r.Cookie(useCookie.Name)
		if err != nil {
			if failureURI != "" {
				u.writeRedirect(w, fmt.Sprintf("%s?err=no_cookie", failureURI))
				return
			}
			u.writeErrText(w, http.StatusNotFound, "no_cookie")
			return
		}
		token = cookie.Value
	} else {
		if r.Method != http.MethodPost {
			if failureURI != "" {
				u.writeRedirect(w, fmt.Sprintf("%s?err=invalid_request", failureURI))
				return
			}
			u.writeErrText(w, http.StatusBadRequest, "invalid_request")
			return
		}
		token = r.FormValue("token")
	}

	if token == "" {
		// If token is empty then user is logged out
		if successURI != "" {
			u.writeRedirect(w, fmt.Sprintf("%s?suc=no_cookie", successURI))
			return
		}
		u.writeErrText(w, http.StatusBadRequest, "invalid_token")
		return
	}

	err := u.logout(token)
	if err != nil {
		var errUmb *ErrUmbrella
		if errors.As(err, &errUmb) {
			if errUmb.Op == "InvalidToken" || errUmb.Op == "Expired" || errUmb.Op == "ParseToken" || errUmb.Op == "InvalidSession" {
				if successURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?suc=logged_out", successURI))
					return
				}
				u.writeErrText(w, http.StatusNotFound, "invalid_credentials")
			} else if errUmb.Op == "GetFromDB" {
				if failureURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?err=database_error", failureURI))
					return
				}
				u.writeErrText(w, http.StatusInternalServerError, "database_error")
			} else {
				if failureURI != "" {
					u.writeRedirect(w, fmt.Sprintf("%s?err=login_error", failureURI))
					return
				}
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

	if successURI != "" {
		if useCookie != nil {
			cookieClone := http.Cookie{
				Name:     useCookie.Name,
				Path:     useCookie.Path,
				Value:    "",
				HttpOnly: useCookie.HttpOnly,
			}
			http.SetCookie(w, &cookieClone)
		}
		u.writeRedirect(w, successURI)
		return
	}

	u.writeOK(w, http.StatusOK, map[string]interface{}{})
}

func (u Umbrella) writeRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusSeeOther)
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

	sess := &Session{
		Key:       sUUID,
		ExpiresAt: expiresAt,
		UserID:    userID,
		Flags:     FlagSessionActive,
	}

	errCrud := u.orm.Save(sess)
	if errCrud != nil {
		return "", 0, &ErrUmbrella{
			Op:  "SaveToDB",
			Err: errCrud,
		}
	}

	return token, expiresAt, nil
}

func (u *Umbrella) getSession(key string) (*Session, error) {
	sessions, errCrud := u.orm.Get(func() interface{} { return &Session{} }, []string{"ID", "asc"}, 1, 0, map[string]interface{}{"Key": key}, nil)

	if errCrud != nil {
		return nil, fmt.Errorf("Error in sdb.Get: %w", errCrud)
	}
	if len(sessions) == 0 {
		return nil, nil
	}

	return sessions[0].(*Session), nil
}

func (u Umbrella) logout(token string) *ErrUmbrella {
	sID, errUmbrella := u.parseTokenWithCheck(token)
	if errUmbrella != nil {
		return errUmbrella
	}

	session, err := u.getSession(sID)
	if session == nil {
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

	if session.Flags&FlagSessionActive == 0 || session.Flags&FlagSessionLoggedOut > 0 {
		return &ErrUmbrella{
			Op:  "InvalidSession",
			Err: err,
		}
	}

	session.Flags = session.Flags | FlagSessionLoggedOut
	if session.Flags&FlagSessionActive > 0 {
		session.Flags = session.Flags - FlagSessionActive
	}
	errCrud := u.orm.Save(session)
	if errCrud != nil {
		return &ErrUmbrella{
			Op:  "SaveToDB",
			Err: errCrud,
		}
	}

	return nil
}

func (u Umbrella) check(token string, refresh bool) (string, int64, int64, *ErrUmbrella) {
	sID, errUmbrella := u.parseTokenWithCheck(token)
	if errUmbrella != nil {
		return "", 0, 0, errUmbrella
	}

	session, err := u.getSession(sID)
	if session == nil {
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

	if session.Flags&FlagSessionActive == 0 || session.Flags&FlagSessionLoggedOut > 0 {
		return "", 0, 0, &ErrUmbrella{
			Op:  "InvalidSession",
			Err: err,
		}
	}

	user := u.Interfaces.User()
	got, err := user.GetByID(session.UserID)
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

		session.ExpiresAt = expiresAt
		errCrud := u.orm.Save(session)
		if errCrud != nil {
			return "", 0, 0, &ErrUmbrella{
				Op:  "SaveToDB",
				Err: errCrud,
			}
		}
		return token2, expiresAt, session.UserID, nil
	}

	return token, 0, session.UserID, nil
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
	return len(s) > 11
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

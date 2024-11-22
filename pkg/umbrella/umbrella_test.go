package umbrella

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHTTPHandlerWithInvalidInput(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("invalidfield1", "somevalue")
	data.Set("invalidfield2", "somevalue2")
	b := makeRequest("POST", false, "register", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on register endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_email" {
		t.Fatalf("POST method on register did not return invalid_email")
	}

	data = url.Values{}
	data.Set("email", testEmail)
	b = makeRequest("POST", false, "register", data.Encode(), http.StatusBadRequest, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on register endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_or_weak_password" {
		t.Fatalf("POST method on register did not return invalid_or_weak_password")
	}
}

func TestRegisterHTTPHandlerWithInvalidPassword(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", testEmail)
	data.Set("password", "weak")
	b := makeRequest("POST", false, "register", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on register endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_or_weak_password" {
		t.Fatalf("POST method on register did not return invalid_or_weak_password")
	}
}

func TestRegisterHTTPHandlerWithNonExistingEmail(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", testEmail)
	data.Set("password", testPassword)
	b := makeRequest("POST", false, "register", data.Encode(), http.StatusCreated, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on register endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "" {
		t.Fatalf("POST method on register returned error text")
	}
	if r.OK != 1 {
		t.Fatalf("POST method on register did not return ok for valid input and non-existing email")
	}

	id, email, password, _, _, err := getUserByEmail(testEmail)
	if err != nil {
		t.Fatalf("POST method on register - failed to check if record added in the database")
	}
	if id == 0 || email != testEmail {
		t.Fatalf("POST method on register failed to add record")
	}
	passwordInDBDecoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		t.Fatalf("POST method on register - failed to decode the password from the database")
	}
	err = bcrypt.CompareHashAndPassword(passwordInDBDecoded, []byte(testPassword))
	if err != nil {
		t.Fatalf("POST method on register failed to insert password to the database properly")
	}

	if testPostRegisterSuccessVariable != true {
		t.Fatalf("POST method on register failed - post register success hook was not executed")
	}
}

func TestRegisterHTTPHandlerWithExistingEmail(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", testEmail)
	data.Set("password", testPassword)
	b := makeRequest("POST", false, "register", data.Encode(), http.StatusOK, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on register endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "email_registered" {
		t.Fatalf("POST method on register with existing email did not return email_registered error text")
	}
	if r.OK != 0 {
		t.Fatalf("POST method on register returned ok for valid input and existing email")
	}
}

func TestConfirmHTTPHandlerWithInvalidInput(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("invalidfield1", "somevalue")
	data.Set("invalidfield2", "somevalue2")
	b := makeRequest("POST", false, "confirm", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on confirm endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_key" {
		t.Fatalf("POST method on register did not return invalid_key")
	}

	data = url.Values{}
	data.Set("key", `%%%(((%%%))))`)
	b = makeRequest("POST", false, "confirm", data.Encode(), http.StatusBadRequest, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on confirm endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_key" {
		t.Fatalf("POST method on register did not return invalid_key")
	}
}

func TestConfirmHTTPHandlerWithValidKey(t *testing.T) {
	id, _, _, activationKey, _, err := getUserByEmail(testEmail)
	if err != nil {
		t.Fatalf("POST method on confirm - failed to check if record added in the database")
	}
	if id == 0 {
		t.Fatalf("POST method on confirm - failed to get any record matching email address")
	}

	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("key", activationKey)
	b := makeRequest("POST", false, "confirm", data.Encode(), http.StatusOK, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on confirm endpoint returned wrong json output, error marshaling: %s", err.Error())
	}

	id, _, _, _, flags, err := getUserByEmail(testEmail)
	if err != nil {
		t.Fatalf("POST method on confirm - failed to check if record added in the database")
	}
	if id == 0 {
		t.Fatalf("POST method on confirm - failed to get any record matching email address")
	}
	if flags&FlagUserEmailConfirmed == 0 {
		t.Fatalf("POST method on confirm - failed to change flag in the database")
	}
	if flags&FlagUserAllowLogin == 0 {
		t.Fatalf("POST method on confirm - failed to change flag in the database")
	}
}

func TestConfirmHTTPHandlerWithInvalidKey(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("key", "nonexistingkey")
	b := makeRequest("POST", false, "confirm", data.Encode(), http.StatusNotFound, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on confirm endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_key" {
		t.Fatalf("POST method on register did not return invalid_key")
	}
}

func TestLoginHTTPHandlerWithInvalidInput(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("invalidfield1", "somevalue")
	data.Set("invalidfield2", "somevalue2")
	b := makeRequest("POST", false, "login", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on login did not return invalid_credentials")
	}

	data = url.Values{}
	data.Set("email", "somevalue")
	data.Set("password", "somevalue2")
	b = makeRequest("POST", false, "login", data.Encode(), http.StatusBadRequest, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on login did not return invalid_credentials when invalid email")
	}

	data = url.Values{}
	data.Set("email", testEmail)
	b = makeRequest("POST", false, "login", data.Encode(), http.StatusBadRequest, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on login did not return invalid_credentials when password empty")
	}
}

func TestLoginHTTPHandlerWithValidEmailAndPassword(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", testEmail)
	data.Set("password", testPassword)
	b := makeRequest("POST", false, "login", data.Encode(), http.StatusOK, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "" {
		t.Fatalf("POST method on login returned non-empty err_text")
	}
	if r.Data["token"].(string) == "" {
		t.Fatalf("POST method on login returned empty token")
	}
	if r.Data["expires_at"].(float64) == 0 {
		t.Fatalf("POST method on login return zero expires_at")
	}

	flags, key, expiresAt, userID, err := getSessionByID(1)
	if err != nil {
		t.Fatalf("POST method on login - failed to check if session record added in the database")
	}
	if key == "" {
		t.Fatalf("POST method on login failed to add session key to database")
	}
	if userID != 1 {
		t.Fatalf("POST method on login failed to add user ID to session in the database")
	}
	if int64(r.Data["expires_at"].(float64)) != expiresAt {
		t.Fatalf("POST method on login failed to add session expiration in the database")
	}
	if flags&FlagSessionActive == 0 {
		t.Fatalf("POST method on login failed to set session flags")
	}

	// Used in later tests
	sessionToken = r.Data["token"].(string)
}

func TestHTTPHandlerWrapperWithValidToken(t *testing.T) {
	b := makeRequest("GET", true, "", "", http.StatusOK, sessionToken, t)
	if string(b) != "RestrictedAreaContent" {
		t.Fatalf("Invalid output from HTTP request to wrapped endpoint when valid token")
	}
}

func TestHTTPHandlerWrapperWithInvalidToken(t *testing.T) {
	b := makeRequest("GET", true, "", "", http.StatusUnauthorized, "invalidToken", t)
	if string(b) != "NoAccess" {
		t.Fatalf("Invalid output from HTTP request to wrapped endpoint when invalid token")
	}
}

func TestLoginHTTPHandlerWithNonExistingEmail(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", "nonexisting@example.com")
	data.Set("password", testPassword)
	b := makeRequest("POST", false, "login", data.Encode(), http.StatusNotFound, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on login did not return invalid_credentials")
	}
}

func TestLoginHTTPHandlerWithInvalidPassword(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("email", testEmail)
	data.Set("password", "invalidPassword!")
	b := makeRequest("POST", false, "login", data.Encode(), http.StatusNotFound, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on login endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on login did not return invalid_credentials")
	}
}

func TestCheckHTTPHandlerWithInvalidInput(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("invalidfield1", "somevalue")
	data.Set("invalidfield2", "somevalue2")
	b := makeRequest("POST", false, "check", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on check endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_token" {
		t.Fatalf("POST method on check did not return invalid_token")
	}
}

func TestCheckHTTPHandlerWithInvalidToken(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("token", "12345678900123456789012345678901234567890123")
	b := makeRequest("POST", false, "check", data.Encode(), http.StatusNotFound, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on check endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on check did not return invalid_token when invalid token")
	}
}

func TestCheckHTTPHandlerWithValidToken(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("token", sessionToken)
	b := makeRequest("POST", false, "check", data.Encode(), http.StatusOK, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on check endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "" {
		t.Fatalf("POST method on check returned non-empty err_text")
	}
	if r.Data["token"].(string) == "" {
		t.Fatalf("POST method on check returned empty token")
	}
	if r.Data["expires_at"].(float64) != 0 {
		t.Fatalf("POST method on check returned non-zero expires_at")
	}
	if r.Data["refreshed"].(bool) != false {
		t.Fatalf("POST method on check returned invalid refreshed value")
	}
}

func TestCheckHTTPHandlerWithValidTokenWithRefresh(t *testing.T) {
	time.Sleep(3000 * time.Millisecond)

	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("token", sessionToken)
	data.Set("refresh", "1")
	b := makeRequest("POST", false, "check", data.Encode(), http.StatusOK, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on check endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "" {
		t.Fatalf("POST method on check returned non-empty err_text")
	}
	if r.Data["token"].(string) == "" {
		t.Fatalf("POST method on check returned empty token")
	}
	if r.Data["token"].(string) == sessionToken {
		t.Fatalf("POST method on check with refresh returned same token")
	}
	if r.Data["expires_at"].(float64) == 0 {
		t.Fatalf("POST method on check returned expires_at value of 0")
	}
	if r.Data["refreshed"].(bool) != true {
		t.Fatalf("POST method on check returned invalid refreshed value")
	}
}
func TestLogoutHTTPHandlerWithInvalidInput(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("invalidfield1", "somevalue")
	data.Set("invalidfield2", "somevalue2")
	b := makeRequest("POST", false, "logout", data.Encode(), http.StatusBadRequest, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on logout endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_token" {
		t.Fatalf("POST method on logout did not return invalid_token")
	}
}

func TestLogoutHTTPHandlerWithValidToken(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("token", sessionToken)
	b := makeRequest("POST", false, "logout", data.Encode(), http.StatusOK, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on logout endpoint returned wrong json output, error marshaling: %s", err.Error())
	}

	flags, _, _, _, err := getSessionByID(1)
	if err != nil {
		t.Fatalf("POST method on logout - failed to check if session record added in the database")
	}
	if flags&FlagSessionActive != 0 || flags&FlagSessionLoggedOut == 0 {
		t.Fatalf("POST method on logout failed to set session flags")
	}
}

func TestLogoutHTTPHandlerWithInvalidToken(t *testing.T) {
	r := NewHTTPResponse(0, "")

	data := url.Values{}
	data.Set("token", sessionToken)
	b := makeRequest("POST", false, "logout", data.Encode(), http.StatusNotFound, "", t)
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on logout endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on logout returned wrong err_text")
	}

	data = url.Values{}
	data.Set("token", "invalidtoken")
	b = makeRequest("POST", false, "logout", data.Encode(), http.StatusNotFound, "", t)
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("POST method on logout endpoint returned wrong json output, error marshaling: %s", err.Error())
	}
	if r.ErrText != "invalid_credentials" {
		t.Fatalf("POST method on logout returned wrong err_text")
	}
}

func TestHTTPHandlerWrapperAfterLogout(t *testing.T) {
	b := makeRequest("GET", true, "", "", http.StatusUnauthorized, sessionToken, t)
	if string(b) != "NoAccess" {
		t.Fatalf("Invalid output from HTTP request to wrapped endpoint after logout")
	}
}

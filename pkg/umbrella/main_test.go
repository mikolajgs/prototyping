package umbrella

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
)

// Global vars used across all the tests
var dbUser = "goumbrellatest"
var dbPass = "secret"
var dbName = "goumbrella"
var dbConn *sql.DB

var dockerPool *dockertest.Pool
var dockerResource *dockertest.Resource

var httpPort = "32777"
var httpCancelCtx context.CancelFunc
var httpURI = "/v1/umbrella/"
var httpURI2 = "/v1/restricted_stuff/"

var testEmail = "code@forthcoming.pl"
var testPassword = "T0ugh3rPassw0rd444!"

var sessionToken = ""

var testUmbrella *Umbrella

var testPostRegisterSuccessVariable = false

func TestMain(m *testing.M) {
	createDocker()
	createUmbrella()
	createHTTPServer()

	code := m.Run()
	removeDocker()
	os.Exit(code)
}

func createDocker() {
	var err error
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	dockerResource, err = dockerPool.Run("postgres", "13", []string{"POSTGRES_PASSWORD=" + dbPass, "POSTGRES_USER=" + dbUser, "POSTGRES_DB=" + dbName})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	if err = dockerPool.Retry(func() error {
		var err error
		dbConn, err = sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dockerResource.GetPort("5432/tcp"), dbName))
		if err != nil {
			return err
		}
		return dbConn.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func createUmbrella() {
	testUmbrella = NewUmbrella(dbConn, "gasiordev_", &JWTConfig{
		Key:               "someSecretKey--.",
		Issuer:            "prototyping.gasior.dev",
		ExpirationMinutes: 1,
	}, nil)
	testUmbrella.Hooks = &Hooks{
		PostRegisterSuccess: func(w http.ResponseWriter, email string) bool {
			testPostRegisterSuccessVariable = true
			return true
		},
	}
	testUmbrella.Flags = 0
	testUmbrella.UserExtraFields = []UserExtraField{
		UserExtraField{
			Name:         "Name",
			RegExp:       nil,
			DefaultValue: "Unknown",
		},
	}
	err := testUmbrella.CreateDBTables()
	if err != nil {
		log.Fatalf("Failed to create DB tables")
	}
}

func getRestrictedStuffHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromRequest(r)
		if userID != 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("RestrictedAreaContent"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("NoAccess"))
		}
	})
}

func createHTTPServer() {
	var ctx context.Context
	ctx, httpCancelCtx = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		go func() {
			http.Handle(httpURI, testUmbrella.GetHTTPHandler(httpURI))
			http.Handle(httpURI2, testUmbrella.GetHTTPHandlerWrapper(getRestrictedStuffHTTPHandler(), HandlerConfig{}))
			http.ListenAndServe(":"+httpPort, nil)
		}()
	}(ctx)
	time.Sleep(2 * time.Second)
}

func removeDocker() {
	dockerPool.Purge(dockerResource)
}

func makeRequest(method string, wrapped bool, additionalURI string, data string, status int, bearerToken string, t *testing.T) []byte {
	uri := httpURI
	if wrapped {
		uri = httpURI2
	}

	req, err := http.NewRequest(method, "http://localhost:"+httpPort+uri+additionalURI, strings.NewReader(data))
	if err != nil {
		t.Fatalf("failed to make a request")
	}
	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if bearerToken != "" {
		req.Header.Add("Authorization", "Bearer "+bearerToken)
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("failed to make a request")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body")
	}
	if resp.StatusCode != status {
		log.Print(string(b))
		t.Fatalf("request returned wrong status code, wanted %d, got %d", status, resp.StatusCode)
	}

	return b
}

func getUserByEmail(email string) (int64, string, string, string, int64, error) {
	var id, flags int64
	var email2, password, activationKey string
	err := dbConn.QueryRow(fmt.Sprintf("SELECT user_id, email, password, email_activation_key, user_flags FROM gasiordev_users WHERE email = '%s'", email)).Scan(&id, &email2, &password, &activationKey, &flags)
	return id, email2, password, activationKey, flags, err
}

func getSessionByID(id int64) (int64, string, int64, int64, error) {
	var flags, expiresAt, userID int64
	var key2 string
	err := dbConn.QueryRow(fmt.Sprintf("SELECT session_flags, key, expires_at, user_id FROM gasiordev_sessions WHERE session_id = %d", id)).Scan(&flags, &key2, &expiresAt, &userID)
	return flags, key2, expiresAt, userID, err
}

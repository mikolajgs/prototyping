package restapi

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
)

// Global vars used across all the tests
var dbUser = "gocrudtest"
var dbPass = "secret"
var dbName = "gocrud"
var dbConn *sql.DB

var dockerPool *dockertest.Pool
var dockerResource *dockertest.Resource

var httpPort = "32777"
var httpCancelCtx context.CancelFunc
var httpURI = "/v1/testobjects/"
var httpURI2 = "/v1/testobjects/price/"
var httpURIPassFunc = "/v1/testobjects/password/"
var httpURIJoined = "/v1/joined/"

var ctl *Controller

var testStructNewFunc func() interface{}
var testStructCreateNewFunc func() interface{}
var testStructReadNewFunc func() interface{}
var testStructUpdateNewFunc func() interface{}
var testStructListNewFunc func() interface{}
var testStructUpdatePriceNewFunc func() interface{}
var testStructUpdatePasswordWithFuncNewFunc func() interface{}
var testStructObj *TestStruct

// Test struct for all the tests
type TestStruct struct {
	ID    int64 `json:"test_struct_id"`
	Flags int64 `json:"test_struct_flags"`

	// Test email validation
	PrimaryEmail   string `json:"email" crud:"req"`
	EmailSecondary string `json:"email2" crud:"req email"`

	// Test length validation
	FirstName string `json:"first_name" crud:"req lenmin:2 lenmax:30"`
	LastName  string `json:"last_name" crud:"req lenmin:0 lenmax:255"`

	// Test int value validation
	Age   int `json:"age" crud:"valmin:1 valmax:120"`
	Price int `json:"price" crud:"valmin:0 valmax:999"`

	// Test regular expression
	PostCode  string `json:"post_code" crud:"req lenmin:6 regexp:^[0-9]{2}\\-[0-9]{3}$"`
	PostCode2 string `json:"post_code2" crud:"lenmin:6" crud_regexp:"^[0-9]{2}\\-[0-9]{3}$"`

	// Test HTTP endpoint tags
	Password        string `json:"password"`
	CreatedByUserID int64  `json:"created_by_user_id" crud_val:"55"`

	// Test unique tag
	Key string `json:"key" crud:"req uniq lenmin:30 lenmax:255"`
}

// Test structs for HTTP endpoints
// Create
type TestStruct_Create struct {
	ID           int64  `json:"test_struct_id"`
	PrimaryEmail string `json:"email" crud:"req"`
	FirstName    string `json:"first_name" crud:"req lenmin:2 lenmax:30"`
	LastName     string `json:"last_name" crud:"req lenmin:0 lenmax:255"`
	Key          string `json:"key" crud:"req uniq lenmin:30 lenmax:255"`
}

type TestStruct_Update struct {
	ID        int64  `json:"test_struct_id"`
	FirstName string `json:"first_name" crud:"req lenmin:2 lenmax:30"`
	LastName  string `json:"last_name" crud:"req lenmin:0 lenmax:255"`
}

type TestStruct_UpdatePrice struct {
	ID    int64 `json:"test_struct_id"`
	Price int   `json:"price" crud:"valmin:0 valmax:999"`
}

type TestStruct_UpdatePasswordWithFunc struct {
	ID       int64  `json:"test_struct_id"`
	Password string `json:"password" crud:"valmin:0 valmax:999 password"`
}

type TestStruct_Read struct {
	ID           int64  `json:"test_struct_id"`
	PrimaryEmail string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	Price        int    `json:"price"`
	PostCode     string `json:"post_code"`
	Password     string `json:"password" crud:"hidden"`
}

type TestStruct_List struct {
	ID           int64  `json:"test_struct_id"`
	Price        int    `json:"price"`
	PrimaryEmail string `json:"email" crud:"req email"`
	FirstName    string `json:"first_name"`
	Age          int    `json:"age"`
	Password     string `json:"password" crud:"hidden"`
}

func TestMain(m *testing.M) {
	createDocker()
	createController()
	createDBStructure()
	createHTTPServer()

	code := m.Run()
	//removeDocker()
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

func createController() {
	ctl = NewController(dbConn, "crud_", &ControllerConfig{
		TagName: "crud",
		PasswordGenerator: func(p string) string {
			return p + p
		},
	})
	testStructNewFunc = func() interface{} {
		return &TestStruct{}
	}
	testStructCreateNewFunc = func() interface{} {
		return &TestStruct_Create{}
	}
	testStructUpdateNewFunc = func() interface{} {
		return &TestStruct_Update{}
	}
	testStructReadNewFunc = func() interface{} {
		return &TestStruct_Read{}
	}
	testStructListNewFunc = func() interface{} {
		return &TestStruct_List{}
	}
	testStructUpdatePriceNewFunc = func() interface{} {
		return &TestStruct_UpdatePrice{}
	}
	testStructUpdatePasswordWithFuncNewFunc = func() interface{} {
		return &TestStruct_UpdatePasswordWithFunc{}
	}
	testStructObj = testStructNewFunc().(*TestStruct)
}

func createDBStructure() {
	ctl.struct2db.CreateTables(testStructObj)
}

func getWrappedHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "UserID", 123)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func createHTTPServer() {
	var ctx context.Context
	ctx, httpCancelCtx = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		go func() {
			http.Handle(httpURI, getWrappedHTTPHandler(ctl.Handler(httpURI, testStructNewFunc, HandlerOptions{
				CreateConstructor: testStructCreateNewFunc,
				ReadConstructor:   testStructReadNewFunc,
				UpdateConstructor: testStructUpdateNewFunc,
				ListConstructor:   testStructListNewFunc,
			})))
			http.Handle(httpURI2, ctl.Handler(httpURI2, testStructNewFunc, HandlerOptions{
				Operations:        OpUpdate,
				UpdateConstructor: testStructUpdatePriceNewFunc,
			}))
			http.Handle(httpURIPassFunc, ctl.Handler(httpURIPassFunc, testStructNewFunc, HandlerOptions{
				Operations:        OpUpdate,
				UpdateConstructor: testStructUpdatePasswordWithFuncNewFunc,
			}))
			http.Handle(httpURIJoined, ctl.Handler(httpURIJoined, func() interface{} { return &Product_WithDetails{} }, HandlerOptions{
				Operations: OpRead | OpList,
				ForceName:  "Product",
			}))
			http.ListenAndServe(":"+httpPort, nil)
		}()
	}(ctx)
	time.Sleep(2 * time.Second)
}

func removeDocker() {
	dockerPool.Purge(dockerResource)
}

func getTableNameCnt(tblName string) (int64, error) {
	var cnt int64
	err := dbConn.QueryRow("SELECT COUNT(table_name) AS c FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1", tblName).Scan(&cnt)
	return cnt, err
}

func getRow() (int64, int64, string, string, string, string, int, int, string, string, string, int64, string, error) {
	var id, flags, createdByUserID int64
	var primaryEmail, emailSecondary, firstName, lastName, postCode, postCode2, password, key string
	var age, price int
	err := dbConn.QueryRow("SELECT * FROM crud_test_structs ORDER BY test_struct_id DESC LIMIT 1").Scan(&id, &flags, &primaryEmail, &emailSecondary, &firstName, &lastName, &age, &price, &postCode, &postCode2, &password, &createdByUserID, &key)
	return id, flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err
}

func getRowById(id int64) (int64, string, string, string, string, int, int, string, string, string, int64, string, error) {
	var id2, flags, createdByUserID int64
	var primaryEmail, emailSecondary, firstName, lastName, postCode, postCode2, password, key string
	var age, price int
	err := dbConn.QueryRow(fmt.Sprintf("SELECT * FROM crud_test_structs WHERE test_struct_id = %d", id)).Scan(&id2, &flags, &primaryEmail, &emailSecondary, &firstName, &lastName, &age, &price, &postCode, &postCode2, &password, &createdByUserID, &key)
	return flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err
}

func getRowCntById(id int64) (int64, error) {
	var cnt int64
	err := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) AS c FROM crud_test_structs WHERE test_struct_id = %d", id)).Scan(&cnt)
	return cnt, err
}

func truncateTable() error {
	_, err := dbConn.Exec("TRUNCATE TABLE crud_test_structs")
	return err
}

func getTestStructWithData() *TestStruct {
	ts := testStructNewFunc().(*TestStruct)
	ts.Flags = 4
	ts.PrimaryEmail = "primary@example.com"
	ts.EmailSecondary = "secondary@example.com"
	ts.FirstName = "John"
	ts.LastName = "Smith"
	ts.Age = 37
	ts.Price = 444
	ts.PostCode = "00-000"
	ts.PostCode2 = "11-111"
	ts.Password = "yyy"
	ts.CreatedByUserID = 4
	ts.Key = fmt.Sprintf("12345679012345678901234567890%d", time.Now().UnixNano())
	return ts
}

func areTestStructObjectSame(ts1 *TestStruct, ts2 *TestStruct) bool {
	if ts1.Flags != ts2.Flags {
		return false
	}
	if ts1.PrimaryEmail != ts2.PrimaryEmail {
		return false
	}
	if ts1.EmailSecondary != ts2.EmailSecondary {
		return false
	}
	if ts1.FirstName != ts2.FirstName {
		return false
	}
	if ts1.LastName != ts2.LastName {
		return false
	}
	if ts1.Age != ts2.Age {
		return false
	}
	if ts1.Price != ts2.Price {
		return false
	}
	if ts1.PostCode != ts2.PostCode {
		return false
	}
	if ts1.PostCode2 != ts2.PostCode2 {
		return false
	}
	if ts1.Password != ts2.Password {
		return false
	}
	if ts1.CreatedByUserID != ts2.CreatedByUserID {
		return false
	}
	if ts1.Key != ts2.Key {
		return false
	}
	return true
}

func makePUTInsertRequest(j string, status int, t *testing.T) []byte {
	req, err := http.NewRequest("PUT", "http://localhost:"+httpPort+httpURI, bytes.NewReader([]byte(j)))
	if err != nil {
		t.Fatalf("PUT method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("PUT method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	if resp.StatusCode != status {
		t.Fatalf("PUT method returned wrong status code, want %d, got %d", status, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("PUT method failed to return body: %s", err.Error())
	}

	return b
}

func makePUTUpdateRequest(j string, id int64, customURI string, t *testing.T) []byte {
	uri := httpURI
	if customURI != "" {
		uri = customURI
	}

	url := "http://localhost:" + httpPort + uri + fmt.Sprintf("%d", id)
	req, err := http.NewRequest("PUT", url, bytes.NewReader([]byte(j)))
	if err != nil {
		t.Fatalf("PUT method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("PUT method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("url: %s", url)
		log.Printf("response body: %s", string(b))

		t.Fatalf("PUT method returned wrong status code, want %d, got %d", http.StatusOK, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("PUT method failed to return body: %s", err.Error())
	}
	return b
}

func makeDELETERequest(id int64, t *testing.T) {
	req, err := http.NewRequest("DELETE", "http://localhost:"+httpPort+httpURI+fmt.Sprintf("%d", id), bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("DELETE method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("DELETE method failed on HTTP server with handler from GetHTTPHandler: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE method returned wrong status code, want %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func makeGETReadRequest(id int64, t *testing.T) *http.Response {
	req, err := http.NewRequest("GET", "http://localhost:"+httpPort+httpURI+fmt.Sprintf("%d", id), bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	return resp
}

func makeGETListRequest(uriParams map[string]string, t *testing.T) []byte {
	uriParamString := ""
	for k, v := range uriParams {
		uriParamString = addWithAmpersand(uriParamString, k+"="+url.QueryEscape(v))
	}

	req, err := http.NewRequest("GET", "http://localhost:"+httpPort+httpURI+"?"+uriParamString, bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET method returned wrong status code, want %d, got %d", http.StatusOK, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("GET method failed to return body: %s", err.Error())
	}

	return b
}

func addWithAmpersand(s string, v string) string {
	if s != "" {
		s += "&"
	}
	s += v
	return s
}

func isInTheList(xs []string, v string) bool {
	for _, s := range xs {
		if s == v {
			return true
		}
	}
	return false
}

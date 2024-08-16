package struct2db

import (
	"fmt"
	"html"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test struct for all the tests
type TestStruct struct {
	ID    int64 `json:"test_struct_id"`
	Flags int64 `json:"test_struct_flags"`

	// Test email validation
	PrimaryEmail   string `json:"email" 2db:"req"`
	EmailSecondary string `json:"email2" 2db:"req email"`

	// Test length validation
	FirstName string `json:"first_name" 2db:"req lenmin:2 lenmax:30"`
	LastName  string `json:"last_name" 2db:"req lenmin:0 lenmax:255"`

	// Test int value validation
	Age   int `json:"age" 2db:"valmin:1 valmax:120"`
	Price int `json:"price" 2db:"valmin:0 valmax:999"`

	// Test regular expression
	PostCode  string `json:"post_code" 2db:"req lenmin:6 regexp:^[0-9]{2}\\-[0-9]{3}$"`
	PostCode2 string `json:"post_code2" 2db:"lenmin:6" 2db_regexp:"^[0-9]{2}\\-[0-9]{3}$"`

	// Some other fields
	Password        string `json:"password"`
	CreatedByUserID int64  `json:"created_by_user_id"`

	// Test unique tag
	Key string `json:"key" 2db:"req uniq lenmin:30 lenmax:255"`
}

func getTestStructWithData() *TestStruct {
	ts := &TestStruct{}
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

func recreateTestStructTable() {
	testController.DropTable(&TestStruct{})
	testController.CreateTable(&TestStruct{})
}

func areTestStructObjectsSame(ts1 *TestStruct, ts2 *TestStruct) bool {
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

// TestSave tests if Save properly inserts and updates object in the database
func TestSave(t *testing.T) {
	recreateTestStructTable()

	// Create an object in the database first
	ts := getTestStructWithData()
	err := testController.Save(ts)
	if err != nil {
		t.Fatalf("Save failed to insert struct to the table: %s", err.Op)
	}

	ts2 := &TestStruct{}
	ts2FieldsPtrs := []interface{}{&ts2.ID}
	ts2FieldsPtrs = append(ts2FieldsPtrs, testController.GetModelFieldInterfaces(ts2)...)
	err2 := dbConn.QueryRow("SELECT * FROM struct2db_test_structs ORDER BY test_struct_id DESC LIMIT 1").Scan(ts2FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to insert struct to the table: %s", err2.Error())
	}

	if ts2.ID == 0 || ts2.Flags != ts.Flags || ts2.PrimaryEmail != ts.PrimaryEmail || ts2.EmailSecondary != ts.EmailSecondary || ts2.FirstName != ts.FirstName || ts2.LastName != ts.LastName || ts2.Age != ts.Age || ts2.Price != ts.Price || ts2.PostCode != ts.PostCode || ts2.PostCode2 != ts.PostCode2 || ts2.CreatedByUserID != ts.CreatedByUserID || ts2.Key != ts.Key || ts2.Password != ts.Password {
		t.Fatalf("Save failed to insert struct to the table")
	}

	// Now, update the object in the database
	ts.Flags = 7
	ts.PrimaryEmail = "primary1@example.com"
	ts.EmailSecondary = "secondary2@example.com"
	ts.FirstName = "Johnny"
	ts.LastName = "Smithsy"
	ts.Age = 50
	ts.Price = 222
	ts.PostCode = "22-222"
	ts.PostCode2 = "33-333"
	ts.Password = "xxx"
	ts.CreatedByUserID = 7
	ts.Key = "123456789012345678901234567890aaa"

	err3 := testController.Save(ts)
	if err3 != nil {
		t.Fatalf("Save failed to update struct")
	}

	ts2 = &TestStruct{}
	ts2FieldsPtrs = []interface{}{&ts2.ID}
	ts2FieldsPtrs = append(ts2FieldsPtrs, testController.GetModelFieldInterfaces(ts2)...)
	err2 = dbConn.QueryRow("SELECT * FROM struct2db_test_structs ORDER BY test_struct_id DESC LIMIT 1").Scan(ts2FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to update struct in the table: %s", err2.Error())
	}

	if ts2.ID == 0 {
		t.Fatalf("Save failed to update struct in the table")
	}
}

// TestLoad tests if Load properly gets row from the database table and populate object fields with its value
func TestLoad(t *testing.T) {
	recreateTestStructTable()

	// Insert an object first
	ts := getTestStructWithData()
	testController.Save(ts)

	// Get the object
	ts2 := &TestStruct{}
	err := testController.Load(ts2, fmt.Sprintf("%d", ts.ID))
	if err != nil {
		t.Fatalf("Load failed to get data: %s", err.Op)
	}

	if !areTestStructObjectsSame(ts, ts2) {
		t.Fatalf("Load failed to set struct with data: %s", err.Op)
	}
}

// TestDelete tests if Delete removes object from the database
func TestDelete(t *testing.T) {
	recreateTestStructTable()

	// Insert an object first
	ts := getTestStructWithData()
	testController.Save(ts)

	// Delete it
	err := testController.Delete(ts)
	if err != nil {
		t.Fatalf("Delete failed to remove: %s", err.Op)
	}

	var cnt int64
	err2 := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) AS c FROM struct2db_test_structs WHERE test_struct_id = %d", ts.ID)).Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Delete failed to delete struct from the table")
	}
	if cnt > 0 {
		t.Fatalf("Delete failed to delete struct from the table")
	}
	if ts.ID != 0 {
		t.Fatalf("Delete failed to set ID to 0 on the struct")
	}
}

// TestGet tests if Get properly gets many objects from the database, filtered and ordered, with results limited to specific number
func TestGet(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should be ignored by Get later on
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.Price = 444
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts)
	}

	// Insert data that should be selected by filters
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		testController.Save(ts)
	}

	// Get the data from the database
	testStructs, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order: []string{"Age", "asc", "Price", "asc"},
		Limit: 10,
		Offset: 20,
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if len(testStructs) != 10 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 10, len(testStructs))
	}
	if testStructs[2].(*TestStruct).Age != 53 {
		t.Fatalf("Get failed to return correct list of objects, want %v, got %v", 53, testStructs[2].(*TestStruct).Age)
	}
}

// TestGetWithoutFilters tests if Get properly gets many objects from the database, without any filters
func TestGetWithoutFilters(t *testing.T) {
	recreateTestStructTable()

	// Insert data to the database
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		testController.Save(ts)
	}

	// Get the data
	testStructs, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order: []string{"Age", "asc", "Price", "asc"},
		Limit: 13,
		Offset: 14,
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if len(testStructs) != 13 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 10, len(testStructs))
	}
	if testStructs[2].(*TestStruct).Age != 47 {
		t.Fatalf("Get failed to return correct list of objects, want %v, got %v", 47, testStructs[2].(*TestStruct).Age)
	}
}

// TestGetWithRowObjTransformFunc tests if Get can properly return a list of custom elements (eg. string) 
// where each object (row from the database) is transform with a specific function
func TestGetWithRowObjTransformFunc(t *testing.T) {
	recreateTestStructTable()

	// Insert data to the database
	for i := 1; i < 3; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		ts.FirstName = fmt.Sprintf("%s %d", ts.FirstName, i)
		testController.Save(ts)
	}

	// Get the data
	testCustomList, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order: []string{"Age", "asc"},
		RowObjTransformFunc: func(obj interface{}) interface{} {
			out := "<tr>"

			v := reflect.ValueOf(obj)
			elem := v.Elem()
			i := reflect.Indirect(v)
			s := i.Type()
			for j := 0; j < s.NumField(); j++ {
				out += "<td>"
				field := s.Field(j)
				fieldType := field.Type.Kind()
				if fieldType == reflect.String {
					out += html.EscapeString(elem.Field(j).String())
				}
				if fieldType == reflect.Bool {
					out += fmt.Sprintf("%v", elem.Field(j).Bool())
				}
				if fieldType == reflect.Int || fieldType == reflect.Int64 {
					out += fmt.Sprintf("%d", elem.Field(j).Int())
				}
				out += "</td>"
			}

			out += "</tr>"

			return out
		},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects modified with transform func: %s", err.Op)
	}
	if len(testCustomList) != 2 {
		t.Fatalf("Get with transform func returned invalid number of objects, wanted %d got %d", 2, len(testCustomList))
	}
	// One of the columns is a random number so testing just the beginning
	if !strings.HasPrefix(testCustomList[0].(string), "<tr><td>1</td><td>4</td><td>primary@example.com</td><td>secondary@example.com</td><td>John 1</td><td>Smith</td><td>31</td><td>444</td><td>00-000</td><td>11-111</td><td>yyy</td><td>4</td><td>") || !strings.HasPrefix(testCustomList[1].(string), "<tr><td>2</td><td>4</td><td>primary@example.com</td><td>secondary@example.com</td><td>John 2</td><td>Smith</td><td>32</td><td>444</td><td>00-000</td><td>11-111</td><td>yyy</td><td>4</td><td>") {
		t.Fatalf("Get with transform func returned invalid objects")
	}
}

// TestGetCount tests if Get properly gets count of objects from the database, filtered
func TestGetCount(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should be ignored by GetCount later on
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.Price = 444
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts)
	}

	// Insert data that should be selected by filters
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts)
	}

	// Get the data from the database
	cnt, err := testController.GetCount(func() interface{} {
		return &TestStruct{}
	}, GetCountOptions{
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if cnt != 150 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 150, cnt)
	}
}

// TestDeleteMultiple tests if DeleteMultiple removes objects from database based on specified filters
func TestDeleteMultiple(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts)
	}

	// Insert data that should be delete
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts)
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(func() interface{} {
		return &TestStruct{}
	}, map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 50 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 50)
	}
}

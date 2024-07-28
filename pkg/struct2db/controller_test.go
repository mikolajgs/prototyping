package struct2db

import (
	"fmt"
	"log"
	"testing"
)

// TestGetModelIDInterface tests if GetModelIDInterface return pointer to ID field
func TestGetModelIDInterface(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	i := testController.GetModelIDInterface(ts)
	if *(i.(*int64)) != int64(123) {
		log.Fatalf("GetModelIDInterface failed to get ID")
	}
}

// TestGetModelIDValue tests if GetModelIDValue returns values of the ID field
func TestGetModelIDValue(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	v := testController.GetModelIDValue(ts)
	if v != 123 {
		log.Fatalf("GetModelIDValue failed to get ID")
	}
}

// TestGetModelFieldInterfaces tests if GetModelFieldInterfaces returns interfaces to object fields
func TestGetModelFieldInterfaces(t *testing.T) {
	// TODO
}

// TestResetFields tests if ResetFields zeroes object fields
func TestResetFields(t *testing.T) {
	// TODO
}

// TestCreateDBTables tests if CreateDBTables creates tables in the database
func TestCreateDBTables(t *testing.T) {
	ts := &TestStruct{}
	_ = testController.CreateDBTables(ts)

	cnt, err2 := getTableNameCnt("gen64_test_structs")
	if err2 != nil {
		t.Fatalf("CreateDBTables failed to create table for a struct: %s", err2.Error())
	}
	if cnt == 0 {
		t.Fatalf("CreateDBTables failed to create the table")
	}
}

// TestValidateWithValidStruct tests if Validate successfully validates object with valid values
func TestValidateWithValidStruct(t *testing.T) {
	ts := getTestStructWithData()
	b, failedFields, err := testController.Validate(ts, nil)
	if !b {
		t.Fatalf("Validate failed validate valid struct")
	}
	if len(failedFields) > 0 {
		t.Fatalf("Validate return non-empty failed field list when validating a valid struct")
	}
	if err != nil {
		t.Fatalf("Validate failed to validate valid struct: %s", err.Error())
	}
}

// TestValidateWithInvalidStruct tests if Validate invalidates object with invalid values
func TestValidateWithInvalidStruct(t *testing.T) {
	ts := getTestStructWithData()
	ts.PrimaryEmail = "invalidemail"
	ts.EmailSecondary = "invalidemail"
	ts.FirstName = "x"
	ts.LastName = "aFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmx"
	ts.Age = 0
	ts.Price = 1000
	ts.PostCode = "inv"
	ts.PostCode2 = "inv"
	ts.Key = "tooshort"
	b, failedFields, err := testController.Validate(ts, nil)
	if err != nil {
		t.Fatalf("Validate failed with an err")
	}
	if b {
		t.Fatalf("Validate failed to return false for struct with invalid field values")
	}
	for _, f := range []string{"PrimaryEmail", "EmailSecondary", "FirstName", "LastName", "Age", "Price", "PostCode", "PostCode2", "Key"} {
		if failedFields[f] == 0 {
			t.Fatalf(fmt.Sprintf("Validate failed to return field %s in failed fields", f))
		}
	}
}

// TestSaveToDB tests if SaveToDB properly inserts and updates object in the database
func TestSaveToDB(t *testing.T) {
	ts := getTestStructWithData()

	err := testController.SaveToDB(ts)
	if err != nil {
		t.Fatalf("SaveToDB failed to insert struct to the table: %s", err.Op)
	}
	id, flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err2 := getRow()
	if err2 != nil {
		t.Fatalf("SaveToDB failed to insert struct to the table: %s", err.Error())
	}
	if id == 0 || flags != ts.Flags || primaryEmail != ts.PrimaryEmail || emailSecondary != ts.EmailSecondary || firstName != ts.FirstName || lastName != ts.LastName || age != ts.Age || price != ts.Price || postCode != ts.PostCode || postCode2 != ts.PostCode2 || createdByUserID != ts.CreatedByUserID || key != ts.Key || password != ts.Password {
		t.Fatalf("SaveToDB failed to insert struct to the table")
	}

	ts.Flags = 7
	ts.PrimaryEmail = "primary1@gen64.net"
	ts.EmailSecondary = "secondary2@gen64.net"
	ts.FirstName = "Johnny"
	ts.LastName = "Smithsy"
	ts.Age = 50
	ts.Price = 222
	ts.PostCode = "22-222"
	ts.PostCode2 = "33-333"
	ts.Password = "xxx"
	ts.CreatedByUserID = 7
	ts.Key = "123456789012345678901234567890aaa"

	err3 := testController.SaveToDB(ts)
	if err3 != nil {
		t.Fatalf("SaveToDB failed to update struct")
	}

	flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err2 = getRowById(id)
	if err2 != nil {
		t.Fatalf("SaveToDB failed to update struct in the table: %s", err.Error())
	}
	if id == 0 || flags != ts.Flags || primaryEmail != ts.PrimaryEmail || emailSecondary != ts.EmailSecondary || firstName != ts.FirstName || lastName != ts.LastName || age != ts.Age || price != ts.Price || postCode != ts.PostCode || postCode2 != ts.PostCode2 || createdByUserID != ts.CreatedByUserID || key != ts.Key || password != ts.Password {
		t.Fatalf("SaveToDB failed to update struct to the table")
	}
}

// TestSetFromDB tests if SetFromDB properly gets row from the database table and populate object fields with its value
func TestSetFromDB(t *testing.T) {
	ts := getTestStructWithData()
	err := testController.SaveToDB(ts)
	if err != nil {
		t.Fatalf("SaveToDB in TestSetFromDB failed to insert struct to the table: %s", err.Op)
	}

	ts2 := &TestStruct{}
	err = testController.SetFromDB(ts2, fmt.Sprintf("%d", ts.ID))
	if err != nil {
		t.Fatalf("SetFromDB failed to get data: %s", err.Op)
	}

	if !areTestStructObjectSame(ts, ts2) {
		t.Fatalf("SetFromDB failed to set struct with data: %s", err.Op)
	}
}

// TestDeleteFromDB tests if DeleteFromDB removes object from the database
func TestDeleteFromDB(t *testing.T) {
	ts := getTestStructWithData()
	err := testController.SaveToDB(ts)
	if err != nil {
		t.Fatalf("SaveToDB in TestDeleteFromDB failed to insert struct to the table: %s", err.Op)
	}
	err = testController.DeleteFromDB(ts)
	if err != nil {
		t.Fatalf("DeleteFromDB failed to remove: %s", err.Op)
	}

	cnt, err2 := getRowCntById(ts.ID)
	if err2 != nil {
		t.Fatalf("DeleteFromDB failed to delete struct from the table")
	}
	if cnt > 0 {
		t.Fatalf("DeleteFromDB failed to delete struct from the table")
	}
	if ts.ID != 0 {
		t.Fatalf("DeleteFromDB failed to set ID to 0 on the struct")
	}
}

// TestGetFromDB tests if GetFromDB properly gets many objects from the database, filtered and ordered, with results limited to specific number
func TestGetFromDB(t *testing.T) {
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		testController.SaveToDB(ts)
	}

	testStructs, err := testController.GetFromDB(func() interface{} {
		return &TestStruct{}
	}, []string{"Age", "asc", "Price", "asc"}, 10, 20, map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@gen64.net"})
	if err != nil {
		t.Fatalf("GetFromDB failed to return list of objects: %s", err.Op)
	}
	if len(testStructs) != 10 {
		t.Fatalf("GetFromDB failed to return list of objects, want %v, got %v", 10, len(testStructs))
	}
	if testStructs[2].(*TestStruct).Age != 52 {
		t.Fatalf("GetFromDB failed to return correct list of objects, want %v, got %v", 52, testStructs[2].(*TestStruct).Age)
	}
}

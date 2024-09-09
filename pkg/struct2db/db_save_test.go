package struct2db

import "testing"

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

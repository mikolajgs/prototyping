package struct2db

import (
	"fmt"
	"testing"
)

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

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts)
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(func() interface{} {
		return &TestStruct{}
	}, DeleteMultipleOptions{
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 50 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 50)
	}
}

// TestDeleteMultipleWithRawQuery tests if DeleteMultiple removes objects from database based on specified filters,
// and a condition which is somewhat raw query
func TestDeleteMultipleWithRawQuery(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts)
	}

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts)
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(func() interface{} {
		return &TestStruct{}
	}, DeleteMultipleOptions{
		Filters: map[string]interface{}{
			"Price": 444,
			"PrimaryEmail": "primary@example.com",
			"_raw": []interface{}{
				".Age = ? OR .Age IN (?) OR (.Age = ? AND .PrimaryEmail = ?)",
				31,
				[]int{32,33,34},
				35,
				"miko@example.com",
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 45 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 45)
	}
}

// TestDeleteMultipleWithRawQueryOnly tests if DeleteMultiple removes objects from database based on a condition which is somewhat raw query
func TestDeleteMultipleWithRawQueryOnly(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts)
	}

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts)
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(func() interface{} {
		return &TestStruct{}
	}, DeleteMultipleOptions{
		Filters: map[string]interface{}{
			"_raw": []interface{}{
				"(.Price = ? AND .PrimaryEmail = ?) OR (.Age = ? OR .Age IN (?) OR (.Age = ? AND .PrimaryEmail = ?))",
				4444,
				"primary@example.com",
				31,
				[]int{32,33,34},
				35,
				"miko@example.com",
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 45 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 45)
	}
}

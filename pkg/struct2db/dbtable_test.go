package struct2db

import (
	"testing"
)

type TableTestStruct struct {
	ID    int64 `json:"db_table_struct_id"`
	Flags int64 `json:"db_table_struct_flags"`
}

// TestCreateDBTables tests if CreateDBTables creates tables in the database
func TestCreateDBTables(t *testing.T) {
	st := &TableTestStruct{}
	err := testController.CreateDBTables(st)
	if err != nil {
		t.Fatalf("CreateDBTables failed to create table for a struct: %s", err.Error())
	}

	cnt, err2 := getTableNameCnt("struct2db_table_test_structs")
	if err2 != nil {
		t.Fatalf("CreateDBTables failed to create table for a struct: %s", err2.Error())
	}
	if cnt == 0 {
		t.Fatalf("CreateDBTables failed to create the table")
	}
}

// TestDropDBTables tests if DropDBTables drops tables in the database
func TestDropDBTables(t *testing.T) {
	st := &TableTestStruct{}
	err := testController.DropDBTables(st)
	if err != nil {
		t.Fatalf("DropDBTables failed to drop table for a struct: %s", err.Error())
	}

	cnt, err2 := getTableNameCnt("struct2db_table_test_structs")
	if err2 != nil {
		t.Fatalf("DropDBTables failed to drop table for a struct: %s", err2.Error())
	}
	if cnt != 0 {
		t.Fatalf("DropDBTables failed to drop the table")
	}
}

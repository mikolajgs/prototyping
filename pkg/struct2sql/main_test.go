package struct2sql

import (
	"testing"
)

// Test struct for all the tests
type TestStruct struct {
	ID    int64 `json:"test_struct_id"`
	Flags int64 `json:"test_struct_flags"`

	// Test email validation
	PrimaryEmail   string `json:"email"`
	EmailSecondary string `json:"email2"`

	// Test length validation
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`

	// Test int value validation
	Age   int `json:"age"`
	Price int `json:"price"`

	// Test regular expression
	PostCode  string `json:"post_code"`
	PostCode2 string `json:"post_code2"`

	// Test HTTP endpoint tags
	Password        string `json:"password"`
	CreatedByUserID int64  `json:"created_by_user_id"`

	// Test unique tag
	Key string `json:"key" 2sql:"uniq db_type:varchar(2000)"`
}

// Instance of the test object
var testStructObj = &TestStruct{}

func TestSQLQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryDropTable()
	want := "DROP TABLE IF EXISTS test_structs"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h.GetQueryCreateTable()
	want = "CREATE TABLE test_structs (test_struct_id SERIAL PRIMARY KEY,test_struct_flags BIGINT DEFAULT 0,primary_email VARCHAR(255) DEFAULT '',email_secondary VARCHAR(255) DEFAULT '',first_name VARCHAR(255) DEFAULT '',last_name VARCHAR(255) DEFAULT '',age BIGINT DEFAULT 0,price BIGINT DEFAULT 0,post_code VARCHAR(255) DEFAULT '',post_code2 VARCHAR(255) DEFAULT '',password VARCHAR(255) DEFAULT '',created_by_user_id BIGINT DEFAULT 0,key VARCHAR(2000) DEFAULT '' UNIQUE)"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLInsertQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryInsert()
	want := "INSERT INTO test_structs(test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING test_struct_id"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLUpdateByIdQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryUpdateById()
	want := "UPDATE test_structs SET test_struct_flags=$1,primary_email=$2,email_secondary=$3,first_name=$4,last_name=$5,age=$6,price=$7,post_code=$8,post_code2=$9,password=$10,created_by_user_id=$11,key=$12 WHERE test_struct_id = $13"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLInsertOnConflictUpdateQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryInsertOnConflictUpdate()
	want := "INSERT INTO test_structs(test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key)"
	want += " VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"
	want += " ON CONFLICT (test_struct_id) DO UPDATE SET test_struct_flags=$14,primary_email=$15,email_secondary=$16,first_name=$17,last_name=$18,age=$19,price=$20,post_code=$21,post_code2=$22,password=$23,created_by_user_id=$24,key=$25"
	want += " RETURNING test_struct_id"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLDeleteQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryDeleteById()
	want := "DELETE FROM test_structs WHERE test_struct_id = $1"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLSelectQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQuerySelectById()
	want := "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs WHERE test_struct_id = $1"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h.GetQuerySelect(nil, 67, 13, nil, nil, nil)
	want = "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, nil, nil)
	want = "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs WHERE post_code2=$1 AND price=$2 ORDER BY email_secondary DESC,age ASC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"EmailSecondary": true}, nil)
	want = "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs WHERE post_code2=$1 AND price=$2 ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"EmailSecondary": true}, map[string]bool{"Price": true})
	want = "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs WHERE price=$1 ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "asc", "Age", "desc"}, 1, 3, map[string]interface{}{
		"Price": 33,
		"PostCode2": "11-111",
		"_raw": []interface{}{
			".Price=? OR (.EmailSecondary=? AND .Age IN (?)) OR (.Age IN (?)) OR (.EmailSecondary IN (?))",
			// We do not really care about the values, the query contains $x only symbols
			// However, we need to pass either value or an array so that an array can be extracted into multiple $x's
			0,
			0,
			[]int{0,0,0,0},
			[]int{0,0,0},
			[]int{0,0},
		},
		"_rawConjuction": RawConjuctionOR,
	}, map[string]bool{
		"EmailSecondary": true,
	}, map[string]bool{
		"Price": true,
	})
	want = "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key"
	want += " FROM test_structs WHERE"
	want += " (price=$1) OR (price=$2 OR (email_secondary=$3 AND age IN ($4,$5,$6,$7)) OR (age IN ($8,$9,$10)) OR (email_secondary IN ($11,$12)))"
	want += " ORDER BY email_secondary ASC LIMIT 1 OFFSET 3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLSelectCountQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQuerySelectCount(map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"Price": true})
	want := "SELECT COUNT(*) AS cnt FROM test_structs WHERE price=$1"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLDeleteWithFiltersQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryDelete(map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"Price": true})
	want := "DELETE FROM test_structs WHERE price=$1"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryDelete(map[string]interface{}{
		"Price": 4444,
		"PostCode2": "11-111",
		"_rawConjuction": RawConjuctionAND,
		"_raw": []interface{}{
			".Price=? OR .EmailSecondary=? OR .Age IN (?)",
			0,
			0,
			[]int{0,0,0},
		},
	}, map[string]bool{"Price": true})
	want = "DELETE FROM test_structs WHERE (price=$1) AND (price=$2 OR email_secondary=$3 OR age IN ($4,$5,$6))"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryDelete(map[string]interface{}{
		"_raw": []interface{}{
			".Price=? OR .EmailSecondary=? OR .Age IN (?)",
			0,
			0,
			[]int{0,0,0},
		},
	}, map[string]bool{"Price": true})
	want = "DELETE FROM test_structs WHERE (price=$1 OR email_secondary=$2 OR age IN ($3,$4,$5))"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLUpdateQueries(t *testing.T) {
	h := NewStruct2sql(testStructObj, Struct2sqlOptions{})

	got := h.GetQueryUpdate(
		map[string]interface{}{"Price": 1234, "PostCode2": "12-345"},
		map[string]interface{}{"PrimaryEmail": "primary@example.com"},
		nil,
		nil,
	)
	want := "UPDATE test_structs SET post_code2=$1,price=$2 WHERE primary_email=$3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryUpdate(
		map[string]interface{}{"Price": 1234, "PostCode2": "12-345", "FirstName": "Jane", "LastName": "Doe"},
		map[string]interface{}{"PrimaryEmail": "primary@example.com", "PostCode": "11-111"},
		map[string]bool{"FirstName": true, "LastName": true},
		map[string]bool{"PostCode": true},
	)
	want = "UPDATE test_structs SET first_name=$1,last_name=$2 WHERE post_code=$3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPluralName(t *testing.T) {
	type Category struct{}
	type Cross struct{}
	type ProductCategory struct{}
	type UserCart struct{}

	h1 := NewStruct2sql(&Category{}, Struct2sqlOptions{})
	h2 := NewStruct2sql(&Cross{}, Struct2sqlOptions{})
	h3 := NewStruct2sql(&ProductCategory{}, Struct2sqlOptions{})
	h4 := NewStruct2sql(&UserCart{}, Struct2sqlOptions{})

	got := h1.GetQueryDropTable()
	want := "DROP TABLE IF EXISTS categories"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h2.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS crosses"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h3.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS product_categories"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h4.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS user_carts"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

package crud

import (
	"fmt"
)

// DropDBTables drop tables in the database for specified objects (see DropDBTable for a single struct)
func (c Controller) DropDBTables(xobj ...interface{}) *ErrController {
	for _, obj := range xobj {
		err := c.DropDBTable(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateDBTables creates tables in the database for specified objects (see CreateDBTable for a single struct)
func (c Controller) CreateDBTables(xobj ...interface{}) *ErrController {
	for _, obj := range xobj {
		err := c.CreateDBTable(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateDBTable creates database table to store specified type of objects. It takes struct name and its fields,
// converts them into table and columns names (all lowercase with underscore), assigns column type based on the
// field type, and then executes "CREATE TABLE" query on attached DB connection
func (c Controller) CreateDBTable(obj interface{}) *ErrController {
	h, err := c.getHelper(obj)
	if err != nil {
		return err
	}

	_, err2 := c.dbConn.Exec(h.GetQueryCreateTable())
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	return nil
}

// DropDBTable drops database table used to store specified type of objects. It just takes struct name, converts
// it to lowercase-with-underscore table name and executes "DROP TABLE" query using attached DB connection
func (c Controller) DropDBTable(obj interface{}) *ErrController {
	h, err := c.getHelper(obj)
	if err != nil {
		return err
	}

	_, err2 := c.dbConn.Exec(h.GetQueryDropTable())
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	return nil
}

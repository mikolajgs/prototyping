# struct-sql-postgres

This package generates PostgreSQL SQL queries based on a struct instance. The concept is to define a struct, create a corresponding table to store its instances, and generate queries for managing the rows in that table, such as creating, updating, deleting, and selecting records.

The following queries can be generated:

* `CREATE TABLE`
* `DROP TABLE`
* `INSERT`
* `UPDATE ... WHERE id = ...`
* `INSERT ... ON CONFLICT UPDATE ...` (upsert)
* `SELECT ... WHERE id = ...`
* `DELETE ... WHERE id = ...`
* `SELECT ... WHERE ...`
* `DELETE ... WHERE ...`
* `UPDATE ... WHERE ...`

See `main_test.go` for a sample usage.


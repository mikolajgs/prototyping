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


## How to use

### Defining a struct

Create a struct to define an object to be stored in a database table.  In the example below, let's create a `Product`.

As of now, a field called `ID` is required.  Also, the corresponding table column name that is generated for it is always prefixed with struct name. Hence, for `ID` field in `Product` that would be `product_id`.

There is an another field called `Flags` that could be added, and it treated the same (meaning `product_flags` is generated, instead of `flags`).

````go
type Product struct {
  ID int64
  Flags int64
  Name string
  Description string `2sql:"db_type:varchar(2000)"` // "db_type" is used to force a specific string type for a column
  Code string `2sql:"uniq"` // "uniq" tells the module to make this column uniq
  ProductionYear int
  CreatedByUserID int64
  LastModifiedByUserID int64
}

#### Field tags

In the above definition, a special tag `2sql` is used to add specific configuration for the column when creating a table.

| Tag key | Description |
|---|-----------|
| `uniq` | When passed, the column will get a `UNIQUE` constraint|
| `db_type` | Overwrites default `VARCHAR(255)` column type for string field. Possible values are: `TEXT`, `BPCHAR(X)`, `CHAR(X)`, `VARCHAR(X)`, `CHARACTER VARYING(X)`, `CHARACTER(X)` where `X` is the size. See [PostgreSQL character types](https://www.postgresql.org/docs/current/datatype-character.html) for more information. |

A different than `2sql` tag can be used by passing `TagName` in `StructSQLOptions{}` when calling `NewStructSQL` function (see below.)

````

### Create a controller for the struct

To generate an SQL query based on a struct, a `StructSQL` object is used.  One per struct.

````go
s := NewStructSQL(Product{}, StructSQLOptions{})
````

#### StructSQLOptions

There are certain options that can be provided to the `NewStructSQL` function which change the functionally of the object.  See the table below.

| Key | Type | Description |
|---|---|---|
| DatabaseTablePrefix | `string` | Prefix for the table name, eg. `myprefix_` |
| ForceName | `string` | Table name is created out of the struct name, eg. for `MyProduct` that would be `my_products`. It is possible to overwrite the struct name, and further table name.  Hence, when `ForceName` is `AnotherStruct` then table name that will be used becomes `another_structs`. |
| TagName | `string` | Uses a different tag than `2sql`.  It is very useful when another module uses this module. |
| Base | `*StructSQL` | In some cases, we'd like to use already existing instance of this object as a base, instead of parsing the struct again.  For example, tags are already defined in another struct and should be re-used.  This is often used by other modules. |
| Joined | `map[string]*StructSQL` | When struct is used to describe a `SELECT` query with `INNER JOIN` to another structs (tables), this map can be used to overwrite `StructSQL` objects for children structs.  If not passed, then children structs that are meant to be used with `INNER JOIN` will be created using `NewStructSQL`.  See one of below sections on joined select queries for more details. |
| UseRootNameWhenJoinedPresent | bool | When struct is used to describe a `SELECT` query with `INNER JOIN` to another structs (tables) and the parent struct has a name like `Product_WithDetails` then it's so-called root name is `Product`, and that will be used as a base for the table name (so it'll be `products`). |

#### Get SQL queries

Use any of the following `GetQuery*` commands to get a desired SQL query.  See examples below.

````go
drop := s.GetQueryDropTable() // retursn 'DROP TABLE IF EXISTS products'

create := s.GetQueryCreateTable() // returns 'CREATE TABLE products (...)'

updateById := s.GetQueryUpdateById() // returns 'UPDATE products SET product_flags = $1, name = $2 ... WHERE product_id = $8
````

#### Get SQL queries with conditions


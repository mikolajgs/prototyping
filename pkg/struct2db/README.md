# struct2db

Package `struct2db` is meant to map structs to PostgreSQL tables (like ORM).

## Example usage
### Structs (models)
Models are defined with structs as follows (take a closer look at the tags):

```
type User struct {
	ID                 int    `json:"user_id"`
	Flags              int    `json:"flags"`
	Name               string `json:"name" validation:"req lenmin:2 lenmax:50"`
	Email              string `json:"email" validation:"req"`
	Password           string `json:"password"`
	EmailActivationKey string `json:"email_activation_key" validation:""`
	CreatedAt          int    `json:"created_at"`
	CreatedByUserID    int    `json:"created_by_user_id"`
}

type Session struct {
	ID                 int    `json:"session_id"`
	Flags              int    `json:"flags"`
	Key                string `json:"key" validation:"uniq lenmin:32 lenmax:50"`
	ExpiresAt          int    `json:"expires_at"`
	UserID             int    `json:"user_id" validation:"req"`
}

type Something struct {
	ID                 int    `json:"something_id"`
	Flags              int    `json:"flags"`
	Email              string `json:"email" validation:"req"`
	Age                int    `json:"age" validation:"req valmin:18 valmax:130 val:18"`
	Price              int    `json:"price" validation:"req valmin:0 valmax:9900 val:100"`
	CurrencyRate       int    `json:"currency_rate" validation:"req valmin:40000 valmax:61234 val:10000"`
	PostCode           string `json:"post_code" validation:"req val:32-600"`
}
```


#### Field tags
Struct tags define ORM behaviour. `struct2sql` parses tags such as `validation`, `http` and various tags starting with 
`validation_`. Apart from the last one, a tag define many properties which are separated with space char, and if they
contain a value other than bool (true, false), it is added after semicolon char.
See below list of all the tags with examples.

Tag | Example | Explanation
--- | --- | ---
`validation` | `validation:"req valmin:0 valmax:130 val:18"` | Struct field properties defining its valid value for model. See Field Properties for more info
`validation_regexp` | `validation_regexp:"^[0-9]{2}\\-[0-9]{3}$"` | Regular expression that struct field must match


##### Field Properties
Property | Explanation
--- | ---
`req` | Field is required
`uniq` | Field has to be unique (like `UNIQUE` on the database column)
`valmin` | If field is numeric, this is minimal value for the field
`valmax` | If field is numeric, this is maximal value for the field
`lenmin` | If field is string, this is a minimal length of the field value
`lenmax` | If field is string, this is a maximal length of the field value


### Database storage
Currently, `struct2db` supports only PostgreSQL as a storage for objects. 

#### Controller
To perform model database actions, a `Controller` object must be created. See below example that modify object(s) 
in the database.

```
// Create connection with sql
conn, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName))
defer conn.Close()

// Create CRUD controller and an instance of a struct
c := crud.NewController(conn, "app1_")
user := &User{}

err = c.CreateDBTable(user) // Run 'CREATE TABLE'

user.Email = "test@example.com"
user.Name = "Nicholas"
user.CreatedAt = time.Now().Unix()
err = c.SaveToDB(user) // Insert object to database table

user.Email = "newemail@example.com"
err = c.SaveToDB(user) // Update object in the database table

err = c.DeleteFromDB(user) // Delete object from the database table

err = c.DropDBTable(user) // Run 'DROP TABLE'
```

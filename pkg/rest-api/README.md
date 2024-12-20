# rest-api

Package `restapi` is meant to create REST API HTTP endpoint for simple data management.

HTTP endpoint can be set to allow creating, updating, removing new object, along with returning its details,
or list of objects. All requests and responses are in the JSON format.

## Example usage

### TL;DR
Check code in the `*_test.go` files, starting with `main_test.go` to get examples.

### Defining structs (models)
Models are defined with structs as follows (take a closer look at the tags):

```
type User struct {
	ID                 int    `json:"user_id"`
	Flags              int    `json:"flags"`
	Name               string `json:"name" restapi:"req lenmin:2 lenmax:50"`
	Email              string `json:"email" restapi:"req"`
	Password           string `json:"password"`
	EmailActivationKey string `json:"email_activation_key" restapi:""`
	CreatedAt          int    `json:"created_at"`
	CreatedByUserID    int    `json:"created_by_user_id"`
}

type Session struct {
	ID                 int    `json:"session_id"`
	Flags              int    `json:"flags"`
	Key                string `json:"key" restapi:"uniq lenmin:32 lenmax:50"`
	ExpiresAt          int    `json:"expires_at"`
	UserID             int    `json:"user_id" restapi:"req"`
}

type Something struct {
	ID                 int    `json:"something_id"`
	Flags              int    `json:"flags"`
	Email              string `json:"email" restapi:"req"`
	Age                int    `json:"age" restapi:"req valmin:18 valmax:130 val:18"`
	Price              int    `json:"price" restapi:"req valmin:0 valmax:9900 val:100"`
	CurrencyRate       int    `json:"currency_rate" restapi:"req valmin:40000 valmax:61234 val:10000"`
	PostCode           string `json:"post_code" restapi:"req val:32-600"`
}
```


#### Field tags
Struct tags define ORM behaviour. `restapi` parses tags such as `restapi`, `http` and various tags starting with 
`restapi_`. Apart from the last one, a tag define many properties which are separated with space char, and if they
contain a value other than bool (true, false), it is added after semicolon char.
See below list of all the tags with examples.

Tag | Example | Explanation
--- | --- | ---
`restapi` | `restapi:"req valmin:0 valmax:130 val:18"` | Struct field properties defining its valid value for model. See RESTAPI Field Properties for more info
`restapi_val` | `restapi_val:"Default value"` | Struct field default value
`restapi_regexp` | `restapi_regexp:"^[0-9]{2}\\-[0-9]{3}$"` | Regular expression that struct field must match
`restapi_testvalpattern` | `restapi_testvalpattern:DD-DDD` | Very simple pattern for generating valid test value (used for tests). In the string, `D` is replaced with a digit


##### CRUD Field Properties
Property | Explanation
--- | ---
`req` | Field is required
`uniq` | Field has to be unique (like `UNIQUE` on the database column)
`valmin` | If field is numeric, this is minimal value for the field
`valmax` | If field is numeric, this is maximal value for the field
`val` | Default value for the field. If the value is not a simple, short alphanumeric, use the `restapi_val` tag for it
`lenmin` | If field is string, this is a minimal length of the field value
`lenmax` | If field is string, this is a maximal length of the field value
`password` | Field is a password, and a function that generates it can be attached (see `ControllerConfig`)
`hidden` | Field's value will not be shown when listing item(s)


### Database storage
Currently, `restapi` supports only PostgreSQL as a storage for objects. 

#### Controller
To perform model database actions, a `Controller` object must be created. See below example that modify object(s) 
in the database.

```
import (
	restapi "github.com/mikolajgs/prototyping/pkg/rest-api"
)
```

```
// Create connection with sql
conn, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName))
defer conn.Close()

// Create RESTAPI controller and an instance of a struct
c := restapi.NewController(conn, "app1_", nil)
user := &User{}

err = c.CreateTable(user) // Run 'CREATE TABLE'
```

#### ControllerConfig

`restapi.&ControllerConfig{}` can be passed to a constructor. It contains the following fields:

* `TagName` can be used to change the name of the tag in struct (by the default it is `restapi`)
* `PasswordGenerator` - when a field is a `password` (see field tags above), a function can be attached, that generates the value to be put in the database

### HTTP Endpoints
With `restapi`, HTTP endpoints can be created to manage objects stored in the database.

If User struct is used for HTTP endpoint, fields such as `Password` will be present when listing users. Therefore, 
it's necessary to create new structs to define CRUD endpoints' input and/or output. These structs unfortunately need
validation tags (which can be different than the ones from "main" struct).

In below example, `User_Create` defines input fields when creating a User, `User_Update` defines fields that are 
meant to change when permorming update, `User_UpdatePassword` is an additional struct just for updating User 
password, and finally - fields in `User_List` will be visible when listing users or reading one user. (You can 
define these as you like).
```
type User_Create {
	ID       int    `json:"user_id"`
	Name     string `json:"name" restapi:"req lenmin:2 lenmax:50"`
	Email    string `json:"email" restapi:"req"`
	Password string `json:"password"`
}
type User_Update {
	ID       int    `json:"user_id"`
	Name     string `json:"name" restapi:"req lenmin:2 lenmax:50"`
	Email    string `json:"email" restapi:"req"`
}
type User_UpdatePassword {
	ID       int `json:"user_id"`
	Password string `json:"password"`
}
type User_List {
	ID       int    `json:"user_id"`
	Name     string `json:"name"
}
```

```
var parentFunc = func() interface{} { return &User; }
var createFunc = func() interface{} { return &User_Create; }
var readFunc   = func() interface{} { return &User_List; }
var updateFunc = func() interface{} { return &User_Update; }
var listFunc   = func() interface{} { return &User_List; }

var updatePasswordFunc = func() interface{} { return &User_UpdatePassword; }

http.HandleFunc("/users/", c.Handler("/users/", parentFunc, HandlerOptions{
	CreateConstructor: createFunc, // input fields (and JSON payload) for creating
	ReadConstructor: readFunc,     // output fields (and JSON output) for reading
	UpdateConstructor: updateFunc, // input fields (and JSON payload) for updating
	ListConstructor: listFunc,     // fields to appear when listing items (and JSON output)
}))
http.HandleFunc("/users/password/", c.Handler("/users/password/", parentFunc, HandlerOptions{
	UpdateConstructor: updatePasswordFunc, // input fields for that one updating endpoint
	Operations: OpUpdate, // only updating will be allowed
}))
log.Fatal(http.ListenAndServe(":9001", nil))
```

In the example, `/users/` CRUDL endpoint is created and it allows to:
* create new User by sending JSON payload using PUT method
* update existing User by sending JSON payload to `/users/:id` with PUT method
* get existing User details with making GET request to `/users/:id`
* delete existing User with DELETE request to `/users/:id`
* get list of Users with making GET request to `/users/` with optional query parameters such as `limit`, `offset` to slice the returned list and `filter_` params (eg. `filter_email`) to filter out records with by specific fields

When creating or updating an object, JSON payload with object details is
required. It should match the struct used for Create and Update operations.
In this case, `User_Create` and `User_Update`.

```
{
	"email": "test@example.com",
	"name": "Jane Doe",
	...
}
```

Output from the endpoint is in JSON format as well and it follows below
structure:

```
{
	"ok": 1,
	"err_text": "...",
	"data": {
		...
	}
}
```

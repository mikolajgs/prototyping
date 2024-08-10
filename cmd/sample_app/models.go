package main

type Person struct {
	ID int64
	Flags int64
	Name string `ui:"req lenmin:5 lenmax:200"`
	Age int `ui:"req valmin:0 valmax:150"`
	PostCode string `ui_regexp:"^[0-9][0-9]-[0-9][0-9][0-9]$"`
	Email string `ui:"req email"`
}

type Group struct {
	ID int64
	Flags int64
	Name string `ui:"req lenmin:3 lenmax:100"`
	Description string `ui:"lenmax:5000"`
}

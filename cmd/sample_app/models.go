package main

type Item struct {
	ID int64 `json:"item_id"`
	Flags int64 `json:"item_flags"`
	Title string `ui:"req lenmin:5 lenmax:200" json:"title"`
	Text string `ui:"lenmax:5000 db_type:VARCHAR(5000)" json:"text"`
}

type ItemGroup struct {
	ID int64 `json:"item_group_id"`
	Flags int64 `json:"item_group_flags"`
	Name string `ui:"req lenmin:3 lenmax:30" json:"name"`
	Description string `ui:"lenmax:255 db_type:VARCHAR(255)" json:"description"`
}

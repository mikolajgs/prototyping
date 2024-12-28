package prototyping

type Permission struct {
	ForType        int8   `json:"for_type"`
	ForID          int64  `json:"for_id"`
	Ops            int64  `json:"ops"`
	ToType         string `json:"to_type"`
	ToItem         int64  `json:"to_item"`
	CreatedAt      int64  `json:"created_at"`
	CreatedBy      int64  `json:"created_by"`
	LastModifiedAt int64  `json:"last_modified_at"`
	LastModifiedBy int64  `json:"last_modified_by"`
}

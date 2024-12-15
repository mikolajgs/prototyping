package umbrella

type Permission struct {
	ID             int64  `json:"id"`
	Flags          int64  `json:"flags"`
	ForType        int8   `json:"for_type"`
	ForItem        int64  `json:"for_item"`
	Ops            int64  `json:"ops"`
	ToType         string `json:"to_type"`
	ToItem         int64  `json:"to_item"`
	CreatedAt      int64  `json:"created_at"`
	CreatedBy      int64  `json:"created_by"`
	LastModifiedAt int64  `json:"last_modified_at"`
	LastModifiedBy int64  `json:"last_modified_by"`
}

const FlagTypeAllow = 4

const ForTypeEveryone = 1
const ForTypeUser = 4

const OpsCreate = 8
const OpsRead = 16
const OpsUpdate = 32
const OpsDelete = 64
const OpsList = 128

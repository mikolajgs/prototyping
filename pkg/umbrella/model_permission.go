package umbrella

import (
	sdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

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

type DefaultPermission struct {
	ctl        *sdb.Controller
	permission *Permission
}

func (g *DefaultPermission) CreateDBTable() error {
	permission := &Permission{}
	err := g.ctl.CreateTable(permission)
	if err != nil {
		return err
	}
	return nil
}

func (g *DefaultPermission) GetPermission() interface{} {
	return g.permission
}

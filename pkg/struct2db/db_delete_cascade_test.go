package struct2db

import (
	"testing"
)

/*                 max DelParent.Delete*() depth -> |
                                                    |
DelParent --> []DelChildNone                        |
          |-> []DelChildDelete --> []DelChildNone   |
					|	                   |-> []DelChildDelete +--> []DelChildDelete
					|									   |                    |`-> []DelChildUpdate
					|									   `-> []DelChildUpdate |
          `-> []DelChildUpdate --> []DelChildNone   |
						                   |-> []DelChildDelete |
														   `-> []DelChildUpdate +--> []DelChildDelete
															                      |`-> []DelChildUpdate
*/
type DelParent struct {
	ID int64
	Flags int64
	Name string
	// Expect all guys below to have a field DelParentID
	// no action should be taken for the 'None' one
	DelChildrenNone []*DelChildNone
	// all these guys whose parent is this instance (so DelParentID) should be deleted from the database
	DelChildrenDelete []*DelChildDelete `2db:"on_del:del"`
	// all these guys whose parent is this instance should have the field DelParentID updated to 0
	// (because sometimes we do not want to remove, more update options will be added later)
	DelChildrenUpdate []*DelChildUpdate `2db:"on_del:upd del_upd_field:DelParentID del_upd_val:0"`
}

type DelChildNone struct {
	ID int64
	DelParentID int64
}
type DelChildDelete struct {
	ID int64
	DelParentID int64
	// "cdel" stands for cascade delete (not "del")
	// if this struct has children they are linked by the DelParentID column which is specified in "cdel_field"
	DelChildrenDelete []*DelChildDelete `2db:"on_cdel:del cdel_field:DelParentID"`
	DelChildrenUpdate []*DelChildUpdate `2db:"on_cdel:upd cdel_field:DelParentID cdel_upd_field:DelChildDeleteID cdel_upd_val:0"`
}
type DelChildUpdate struct {
	ID int64
	DelParentID int64
	DelChildrenDelete []*DelChildDelete `2db:"on_cdel:del cdel_field:DelParentID"`
	DelChildrenUpdate []*DelChildUpdate `2db:"on_cdel:upd cdel_field:DelParentID cdel_upd_field:DelChildDeleteID cdel_upd_val:0"`
}

func TestDeleteCascade(t *testing.T) {
}

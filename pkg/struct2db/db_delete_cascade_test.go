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
	// Create a test parent (with children) and get its ID
	p := createTestDelParentWithChildren()

	// Delete the parent object
	_ = testController.Delete(p, DeleteOptions{
		Constructors: map[string]func() interface{}{
			"DelChildNone": func() interface{} { return &DelChildNone{}; },
			"DelChildDelete": func() interface{} { return &DelChildDelete{}; },
			"DelChildUpdate": func() interface{} { return &DelChildUpdate{}; },
		},
	})

	// Check things
}

func createTestDelParentWithChildren() interface{} {
	recreateTestDelTables();

	// create DelParent
	p := &DelParent{ Name: "Parent1" } // 1
	testController.Save(p, SaveOptions{})

	// create children
	for i:=0; i<2; i++ {
		cNone := &DelChildNone{ DelParentID: p.ID } // 1,2
		cDelete := &DelChildDelete{ DelParentID: p.ID } // 1,2
		cUpdate := &DelChildUpdate{ DelParentID: p.ID } // 1,2
		testController.Save(cNone, SaveOptions{})
		testController.Save(cDelete, SaveOptions{})
		testController.Save(cUpdate, SaveOptions{})
	}

	// create grandchildren
	for i:=0; i<2; i++ {
		for j:=0; j<2; j++ {
			cDeleteNone := &DelChildNone{ DelParentID: int64(i) } // 3,5, 7,9
			cDeleteDelete := &DelChildDelete{ DelParentID: int64(i) } // 3,5, 7,9
			cDeleteUpdate := &DelChildUpdate{ DelParentID: int64(i) } // 3,5, 7,9
			cUpdateNone := &DelChildNone{ DelParentID: int64(i) } // 4,6, 8,10
			cUpdateDelete := &DelChildDelete{ DelParentID: int64(i) } // 4,6, 8,10
			cUpdateUpdate := &DelChildUpdate{ DelParentID: int64(i) } // 4,6, 8,10
			testController.Save(cDeleteNone, SaveOptions{})
			testController.Save(cDeleteDelete, SaveOptions{})
			testController.Save(cDeleteUpdate, SaveOptions{})
			testController.Save(cUpdateNone, SaveOptions{})
			testController.Save(cUpdateDelete, SaveOptions{})
			testController.Save(cUpdateUpdate, SaveOptions{})
		}
	}

	// create grandgrandchildren
	cDeleteDeleteDelete := &DelChildDelete{ DelParentID: int64(9) } // 11
	cDeleteDeleteUpdate := &DelChildUpdate{ DelParentID: int64(9) } // 11
	cUpdateUpdateDelete := &DelChildDelete{ DelParentID: int64(10) } // 12
	cUpdateUpdateUpdate := &DelChildUpdate{ DelParentID: int64(10) } // 12
	testController.Save(cDeleteDeleteDelete, SaveOptions{})
	testController.Save(cDeleteDeleteUpdate, SaveOptions{})
	testController.Save(cUpdateUpdateUpdate, SaveOptions{})
	testController.Save(cUpdateUpdateDelete, SaveOptions{})

	return p
}

func recreateTestDelTables() {
	testController.DropTable(&DelParent{});
	testController.DropTable(&DelChildNone{});
	testController.DropTable(&DelChildDelete{});
	testController.DropTable(&DelChildUpdate{});
	testController.CreateTable(&DelParent{});
	testController.CreateTable(&DelChildNone{});
	testController.CreateTable(&DelChildDelete{});
	testController.CreateTable(&DelChildUpdate{});
}

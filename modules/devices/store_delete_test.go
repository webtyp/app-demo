package devices

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
)

// The plural contract, through the real presenter: Delete ships ids (N=1 and
// N>1 in the same shape) and the store removes exactly those rows in one
// statement. Regression net for the silent no-op, where the store asserted
// the old singular *Device payload and every delete errored unheard.
func TestMemCallerBulkDeleteRemovesOnlyMarked(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&deviceStore{db: db}, &Device{}, view.WithTitle("t"))
	deleter, ok := pres.(view.Deleter)
	if !ok {
		t.Fatal("presenter must implement view.Deleter")
	}
	var rerr error
	pres.Reload(func(err error) { rerr = err })
	if rerr != nil {
		t.Fatalf("reload failed: %v", rerr)
	}

	var derr error
	deleter.Delete([]string{"10", "11"}, func(err error) { derr = err })
	if derr != nil {
		t.Fatalf("bulk delete failed: %v", derr)
	}

	var ids []string
	err := db.Query(&Device{}).ReadAll(
		func() model.Model { return &Device{} },
		func(m model.Model) { ids = append(ids, m.(*Device).Id) },
	)
	if err != nil {
		t.Fatalf("re-read failed: %v", err)
	}
	if len(ids) != 13 {
		t.Errorf("expected 13 rows left after deleting 2 of 15, got %d", len(ids))
	}
	for _, id := range ids {
		if id == "10" || id == "11" {
			t.Errorf("deleted id %q is still in the store", id)
		}
	}

	// N=1 ships the same shape: a single delete is a batch of one.
	deleter.Delete([]string{"12"}, func(err error) { derr = err })
	if derr != nil {
		t.Fatalf("single delete failed: %v", derr)
	}
	if findErr := db.Query(&Device{}).Where("id").Eq("12").ReadOne(); findErr == nil {
		t.Error("id 12 must be gone after the single delete")
	}
}
package medicalhistory

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
)

// Regression net for the silent save breakage (see devices' equivalent): the
// store asserted *Visit but view ships saveArgs{records}. Now the wire is read.
func TestMemCallerSaveThroughThePresenter(t *testing.T) {
	db := newSeededVisitDB()
	pres := requirePatient{view.New(&visitStore{db: db}, &Visit{}, view.WithTitle("t"))}

	var saver view.Saver = pres
	var rerr error
	pres.Reload(func(err error) { rerr = err })
	if rerr != nil {
		t.Fatalf("reload failed: %v", rerr)
	}

	var serr error
	saver.Save([]model.Model{&Visit{Id: "v9", Patient: "Ana Lima", Doctor: "dr. X", Date: "2026-09-01", Reason: "Control", Diagnosis: ""}}, func(err error) { serr = err })
	if serr != nil {
		t.Fatalf("create-via-save failed: %v", serr)
	}
	got := readAll(t, db)
	if got["v9"] == nil || got["v9"].Patient != "Ana Lima" {
		t.Errorf("new record not persisted: %+v", got["v9"])
	}

	saver.Save([]model.Model{&Visit{Id: "v1", Patient: "Juan Pérez", Doctor: "dra. Nueva", Date: "2026-07-20", Reason: "Control", Diagnosis: "Actualizado"}}, func(err error) { serr = err })
	if serr != nil {
		t.Fatalf("update-via-save failed: %v", serr)
	}
	got = readAll(t, db)
	if got["v1"].Doctor != "dra. Nueva" || got["v1"].Diagnosis != "Actualizado" {
		t.Errorf("existing record not replaced: %+v", got["v1"])
	}
}
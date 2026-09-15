package devices

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
)

// Regression net for the silent save breakage: the store used to assert the
// singular *Device payload, but view ships saveArgs{records}, so every save
// errored unheard and the "Guardado" toast never fired. Now the wire is read.
func TestMemCallerSaveThroughThePresenter(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&deviceStore{db: db}, &Device{}, view.WithTitle("t"))
	saver, ok := pres.(view.Saver)
	if !ok {
		t.Fatal("presenter must implement view.Saver")
	}
	var rerr error
	pres.Reload(func(err error) { rerr = err })
	if rerr != nil {
		t.Fatalf("reload failed: %v", rerr)
	}

	// A brand-new record lands.
	var serr error
	saver.Save([]model.Model{&Device{Id: "99", Name: "Nuevo", Ip: "10.0.0.99"}}, func(err error) { serr = err })
	if serr != nil {
		t.Fatalf("create-via-save failed: %v", serr)
	}
	got := readAll(t, db)
	if got["99"] == nil || got["99"].Name != "Nuevo" {
		t.Errorf("new record not persisted: %+v", got["99"])
	}

	// Saving an existing id replaces the whole record.
	saver.Save([]model.Model{&Device{Id: "10", Name: "Renombrado", Ip: "10.0.0.10"}}, func(err error) { serr = err })
	if serr != nil {
		t.Fatalf("update-via-save failed: %v", serr)
	}
	got = readAll(t, db)
	if got["10"].Name != "Renombrado" || got["10"].Ip != "10.0.0.10" {
		t.Errorf("existing record not replaced: %+v", got["10"])
	}
}
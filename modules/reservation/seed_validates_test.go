package reservation

import (
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
	"webtyp.com/model"
	"webtyp.com/unixid"
)

// Every value a seed row actually carries must validate against the module's
// own schema: a format the schema rejects (a dash in a Text, a bad RUT check
// digit) is a record that can never be saved again.
//
// Only carried values are asserted: birthday and contact have no seed source
// (appointment_booking rows carry neither) and arrive empty, and an empty
// value fails under ANY minimum-bearing widget — Text today included. That
// optional-empty gap lives in the input library, out of scope here; what this
// guards is the schema rejecting its own real data.
func TestSeedsValidateAgainstSchema(t *testing.T) {
	env := testEnv(t)
	store := &reservationStore{env: env, staff: dom.NewString("staff-natasha"), staffId: "staff-natasha"}

	var rows []model.Model
	var lerr error
	store.List(func(r []model.Model, err error) { rows = r; lerr = err })
	if lerr != nil {
		t.Fatalf("List: %v", lerr)
	}
	if len(rows) == 0 {
		t.Fatal("expected seeded reservations, got none")
	}

	ids, err := unixid.NewUnixID()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range rows {
		r := m.(*Reservation)
		f, err := form.New("test", &Reservation{}, ids)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.LoadValues(r); err != nil {
			t.Fatalf("seed %q: load: %v", r.Id, err)
		}
		for _, inp := range f.Inputs {
			vals := inp.GetValues()
			if len(vals) == 0 || vals[0] == "" {
				continue // no seed value — see the comment above
			}
			if verr := inp.Validate(vals[0]); verr != nil {
				t.Errorf("seed %q field %q (%q): invalid against its own schema: %v",
					r.Id, inp.FieldName(), vals[0], verr)
			}
		}
	}
}

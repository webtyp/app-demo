package medicalhistory

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/unixid"
)

// Every seed row must validate against the module's own schema: the form the
// user edits is built from this same model, so a seed the schema rejects is
// a record that can never be saved again.
func TestSeedsValidateAgainstSchema(t *testing.T) {
	ids, err := unixid.NewUnixID()
	if err != nil {
		t.Fatal(err)
	}
	for id, v := range readAll(t, newSeededVisitDB()) {
		f, err := form.New("test", &Visit{}, ids)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.LoadValues(v); err != nil {
			t.Fatalf("seed %q: load: %v", id, err)
		}
		if err := f.Validate(); err != nil {
			t.Errorf("seed %q: invalid against its own schema: %v", id, err)
		}
	}
}

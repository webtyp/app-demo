//go:build !wasm

package reservation

import (
	"testing"

	"webtyp.com/time"
	"webtyp.com/view"
)

func TestByDayFilter(t *testing.T) {
	pres := byDay{view.New(&reservationStore{db: reservationDB}, &Reservation{}, view.WithTitle("t"))}

	if items := pres.Filter(""); items != nil {
		t.Fatalf("expected nil items for empty filter term, got %d items", len(items))
	}

	items := pres.Filter("2026-09-10")
	if len(items) != 4 {
		t.Fatalf("expected 4 items for 2026-09-10, got %d", len(items))
	}

	foundConfirmed := false
	foundAttended := false
	for _, it := range items {
		if it.LeadMain == "" {
			t.Errorf("expected LeadMain (hour) to be set on view.Item, got empty")
		}
		if it.Description == "Confirmada" {
			foundConfirmed = true
		}
		if it.Description == "Atendida" {
			foundAttended = true
		}
	}

	if !foundConfirmed {
		t.Errorf("expected at least one item with description 'Confirmada'")
	}
	if !foundAttended {
		t.Errorf("expected at least one item with description 'Atendida'")
	}

	itemsOther := pres.Filter("2026-09-12")
	if len(itemsOther) != 1 {
		t.Fatalf("expected 1 item for 2026-09-12, got %d", len(itemsOther))
	}
	if itemsOther[0].Label != "Diego Castro" {
		t.Errorf("expected label 'Diego Castro', got %q", itemsOther[0].Label)
	}
}

// TestSeedIncludesToday cubre el bug "hoy no se puede elegir": el calendario
// solo hace clicables los días con al menos una reserva, así que el seed debe
// traer reservas dated today — relativas a time.Now, nunca fijas, o el demo
// abre cada día con el hoy inerte.
func TestSeedIncludesToday(t *testing.T) {
	pres := byDay{view.New(&reservationStore{db: reservationDB}, &Reservation{}, view.WithTitle("t"))}
	today := time.FormatDate(time.Now())
	items := pres.Filter(today)
	if len(items) == 0 {
		t.Fatalf("expected at least 1 seeded item for today (%s), got none", today)
	}
}

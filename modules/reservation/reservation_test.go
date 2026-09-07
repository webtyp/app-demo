// Package reservation — pruebo el flujo sobre el módulo real: el Context filtra
// por área, la lista se scopes por médico y crear va por la op real.
package reservation

import (
	"strings"
	"testing"

	"webtyp.com/dom"

	ab "github.com/veltylabs/appointment_booking"

	"webtyp.com/app-demo/config"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testEnv(t *testing.T) *config.Env {
	t.Helper()
	return config.New()
}

func newContext(t *testing.T) *reservationContext {
	t.Helper()
	env := testEnv(t)
	ctx := &reservationContext{env: env, staffOptions: env.Staff()}
	ctx.Init(&emptyCtx{})
	return ctx
}

// TestContext_AreaFiltersStaff: elegir un área filtra el select de médico.
func TestContext_AreaFiltersStaff(t *testing.T) {
	ctx := newContext(t)

	if got := len(ctx.staffForArea()); got != 3 {
		t.Fatalf("empty area: expected 3 staff, got %d", got)
	}

	ctx.area.Set("medicina-general")
	ctx.rebuildStaffOptions()
	list := ctx.staffForArea()
	if len(list) != 1 || list[0].ID != "staff-natasha" {
		t.Fatalf("medicina-general should scope to Natasha, got %+v", list)
	}
	html := ctx.optionsHTML()
	if !strings.Contains(html, "Dra. Natasha Romanoff") {
		t.Errorf("expected only Natasha in the staff options for Medicina, got:\n%s", html)
	}
	if strings.Contains(html, "Dr. Tony Stark") || strings.Contains(html, "Dr. Thor Odinson") {
		t.Errorf("area filter must exclude other doctors:\n%s", html)
	}

	ctx.area.Set("traumatologia")
	ctx.rebuildStaffOptions()
	if list := ctx.staffForArea(); len(list) != 1 || list[0].ID != "staff-tony" {
		t.Fatalf("traumatologia should scope to Tony, got %+v", list)
	}
}

// optionsHTML aúna los <option> del select de médico (BindChildren -> nodes).
func (c *reservationContext) optionsHTML() string {
	nodes := c.staffNode.Get()
	parts := make([]string, 0, len(nodes))
	for _, n := range nodes {
		parts = append(parts, n.String())
	}
	return strings.Join(parts, "")
}

// TestList_ScopedByStaff: el store lista solo las reservas del médico scopeado.
func TestList_ScopedByStaff(t *testing.T) {
	env := testEnv(t)
	store := &reservationStore{env: env, staff: dom.NewString("")}

	store.staffId = ""
	rows, err := store.List()
	if err != nil {
		t.Fatalf("List with no staff: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected empty list without staff, got %d", len(rows))
	}

	store.staffId = "staff-natasha"
	rows, err = store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 reservations for Natasha, got %d", len(rows))
	}
	for _, m := range rows {
		r := m.(*Reservation)
		if r.Id == "" || r.Day == "" || r.Hour == "" {
			t.Errorf("reservation missing local fields: %+v", r)
		}
		if r.Day != "2026-09-10" && r.Day != "2026-09-11" {
			t.Errorf("unexpected reservation day %q", r.Day)
		}
	}
}

// TestCreate_GoesThroughRealOp: un save válido crea una reserva que aparece en
// OpListReservationsByStaff.
func TestCreate_GoesThroughRealOp(t *testing.T) {
	env := testEnv(t)
	store := &reservationStore{env: env, staff: dom.NewString("staff-natasha"), staffId: "staff-natasha"}

	err := store.Save(&Reservation{
		PatientRun:  "102030405",
		PatientName: "Nuevo Paciente",
		Day:         "2026-09-14", // lunes — Natasha agenda Lun–Vie
		Hour:        "15:00",
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	caller := env.Caller()
	out := &ab.ReservationList{}
	var callErr error
	caller.Call(
		ab.OpListReservationsByStaff,
		&ab.ListReservationsByStaffArgs{TenantId: env.TenantID(), StaffId: "staff-natasha", From: 0, To: 4102444800},
		out,
		func(e error) { callErr = e },
	)
	if callErr != nil {
		t.Fatalf("list after save: %v", callErr)
	}
	found := false
	for i := 0; i < out.Len(); i++ {
		r := out.At(i).(*ab.Reservation)
		if r.ClientId == "102030405" {
			found = true
			if r.Status != ab.StatusPending {
				t.Errorf("new reservation status = %q, want PENDING", r.Status)
			}
			// El slot persistido es exactamente el que se pidió (epoch UTC);
			// LocalStringTime es la proyección en la zona local de la máquina,
			// no un valor determinista — no se afinca en el test.
			if r.ReservationTime != slotStartUTC("2026-09-14", "15:00") {
				t.Errorf("reservation_time = %d, want %d", r.ReservationTime, slotStartUTC("2026-09-14", "15:00"))
			}
			if r.DurationMinSnapshot != 30 {
				t.Errorf("duration snapshot = %d, want 30", r.DurationMinSnapshot)
			}
		}
	}
	if !found {
		t.Fatal("reservation created via the real op not found in list")
	}
}

// TestCreate_SlotTaken: crear sobre un slot ocupado no crea una duplicada.
func TestCreate_SlotTaken(t *testing.T) {
	env := testEnv(t)
	store := &reservationStore{env: env, staff: dom.NewString("staff-natasha"), staffId: "staff-natasha"}

	err := store.Save(&Reservation{
		PatientRun:  "0001",
		PatientName: "Clon Burn",
		Day:         "2026-09-10",
		Hour:        "09:00",
	})
	if err == nil {
		t.Fatal("expected an error for an occupied slot")
	}

	rows, lerr := store.List()
	if lerr != nil {
		t.Fatalf("List: %v", lerr)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 reservations after the failed create, got %d", len(rows))
	}
}

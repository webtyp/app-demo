// Package demoenv — pruebo la composición: New no paniquea y las ops reales
// responden al seed. Consumer-shaped: usa los mismos wrappers públicos que el
// módulo agenda usará.
package demoenv

import (
	"testing"

	ab "github.com/veltylabs/appointment_booking"
)

// TestNew_DoesNotPanic: construir el entorno (DB + módulos + seed) no estalla.
func TestNew_DoesNotPanic(t *testing.T) {
	env := New()
	if env.Caller() == nil {
		t.Fatal("expected a Caller")
	}
	if env.Broker() == nil {
		t.Fatal("expected a Broker")
	}
	if env.TenantID() != "demo" {
		t.Fatalf("TenantID = %q, want demo", env.TenantID())
	}
	if len(env.Staff()) != 3 {
		t.Fatalf("expected 3 staff options, got %d", len(env.Staff()))
	}
	if len(env.Holidays2026()) == 0 {
		t.Fatal("expected national holidays to be seeded")
	}
}

// TestWeekly_SeededForNatasha: list_weekly_calendar devuelve las 5 filas de
// Natasha que el seed sembró vía las ops reales.
func TestWeekly_SeededForNatasha(t *testing.T) {
	env := New()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-natasha")

	var rows []ab.WorkCalendarWeekly
	client.Weekly(func(r []ab.WorkCalendarWeekly, err error) {
		if err != nil {
			t.Fatalf("Weekly: %v", err)
		}
		rows = r
	})
	if len(rows) != 5 {
		t.Fatalf("expected 5 weekly rows for Natasha, got %d", len(rows))
	}
	// All rows active with the seeded window 09:00–18:00 + break 13:00–14:00.
	for _, r := range rows {
		if !r.IsActive || r.WorkStart != 540 || r.WorkFinish != 1080 {
			t.Errorf("unexpected row: DayOfWeek=%d active=%v %d-%d",
				r.DayOfWeek, r.IsActive, r.WorkStart, r.WorkFinish)
		}
	}
}

// TestWeekly_TonyAndThor: Tony tiene 3 filas (Lun/Mié/Vie) sin colación; Thor
// ninguna (agenda recién creada).
func TestWeekly_TonyAndThor(t *testing.T) {
	env := New()

	tony := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-tony")
	var tRows []ab.WorkCalendarWeekly
	tony.Weekly(func(r []ab.WorkCalendarWeekly, err error) {
		if err != nil {
			t.Fatalf("weekly tony: %v", err)
		}
		tRows = r
	})
	if len(tRows) != 3 {
		t.Fatalf("expected 3 weekly rows for Tony, got %d", len(tRows))
	}
	for _, r := range tRows {
		if r.BreakStart != 0 || r.BreakFinish != 0 {
			t.Errorf("Tony must have no break: %+v", r)
		}
	}

	thor := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-thor")
	var hRows []ab.WorkCalendarWeekly
	thor.Weekly(func(r []ab.WorkCalendarWeekly, err error) {
		if err != nil {
			t.Fatalf("weekly thor: %v", err)
		}
		hRows = r
	})
	if len(hRows) != 0 {
		t.Fatalf("expected no weekly rows for Thor, got %d", len(hRows))
	}
}

// TestExceptions_SeededForNatasha: list_exceptions devuelve las 2 (18 y 19 sep).
func TestExceptions_SeededForNatasha(t *testing.T) {
	env := New()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-natasha")

	from := unixDay("2026-09-01")
	to := unixDay("2026-09-30")
	var excs []ab.WorkCalendarException
	client.Exceptions(from, to, func(r []ab.WorkCalendarException, err error) {
		if err != nil {
			t.Fatalf("exceptions: %v", err)
		}
		excs = r
	})
	if len(excs) != 2 {
		t.Fatalf("expected 2 exceptions for Natasha in Sept, got %d", len(excs))
	}
	for _, x := range excs {
		if x.ExceptionType != ab.ExcHoliday {
			t.Errorf("expected HOLIDAY, got %q", x.ExceptionType)
		}
	}
}

// TestAgendaRoundTrip: alta → list muestra la nueva → remove → ya no está.
// Consumer-shaped end-to-end contra el appointment_booking real.
func TestAgendaRoundTrip(t *testing.T) {
	env := New()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-thor")

	// Thor no tiene excepciones; agregar una y verificarla.
	date := unixDay("2026-11-20")
	client.AddException(ab.WorkCalendarException{
		TenantId:      env.TenantID(),
		StaffId:       "staff-thor",
		SpecificDate:  date,
		ExceptionType: ab.ExcBlocked,
		StartTime:     540,
		EndTime:       600,
	}, func(err error) {
		if err != nil {
			t.Fatalf("AddException: %v", err)
		}
	})

	var added ab.WorkCalendarException
	client.Exceptions(unixDay("2026-11-01"), unixDay("2026-11-30"), func(rows []ab.WorkCalendarException, err error) {
		if err != nil {
			t.Fatalf("list after add: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 exception after add, got %d", len(rows))
		}
		added = rows[0]
	})
	if added.ExceptionType != ab.ExcBlocked {
		t.Fatalf("expected BLOCKED, got %q", added.ExceptionType)
	}

	client.RemoveException(added.Id, func(err error) {
		if err != nil {
			t.Fatalf("RemoveException: %v", err)
		}
	})
	client.Exceptions(unixDay("2026-11-01"), unixDay("2026-11-30"), func(rows []ab.WorkCalendarException, err error) {
		if err != nil {
			t.Fatalf("list after remove: %v", err)
		}
		if len(rows) != 0 {
			t.Fatalf("expected no exceptions after remove, got %d", len(rows))
		}
	})
}

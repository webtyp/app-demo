// Package config — pruebo el entorno compartido: no paniquea y las ops reales
// responden al seed. Consumer-shaped: usa los mismos wrappers públicos que los
// módulos usarán.
package config

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
	if len(env.Specialties()) == 0 {
		t.Fatal("expected canonical specialties to be exposed")
	}
	if env.ESCForStaff("staff-natasha") == "" {
		t.Fatal("expected an employee_service_config for Natasha")
	}
}

// TestSeededBlocks_ForNatasha: list_blocks devuelve los 10 bloques de Natasha
// (2 bloques x 5 días con brecha de colación 13:00-14:00, 540-780 / 840-1080)
// que el seed sembró vía las ops reales.
func TestSeededBlocks_ForNatasha(t *testing.T) {
	env := New()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-natasha")

	var rows []ab.WorkCalendarBlock
	client.Blocks(func(r []ab.WorkCalendarBlock, err error) {
		if err != nil {
			t.Fatalf("Blocks: %v", err)
		}
		rows = r
	})
	if len(rows) != 10 {
		t.Fatalf("expected 10 blocks for Natasha, got %d", len(rows))
	}
	for _, r := range rows {
		if !r.IsActive {
			t.Errorf("unexpected block: DayOfWeek=%d active=%v %d-%d",
				r.DayOfWeek, r.IsActive, r.StartMin, r.EndMin)
		}
		if (r.StartMin != 540 || r.EndMin != 780) && (r.StartMin != 840 || r.EndMin != 1080) {
			t.Errorf("block time range out of bounds: %d-%d", r.StartMin, r.EndMin)
		}
	}
}

// TestSeededBlocks_TonyAndThor: Tony tiene 3 filas (Lun/Mié/Vie); Thor ninguna
// (agenda recién creada).
func TestSeededBlocks_TonyAndThor(t *testing.T) {
	env := New()

	tony := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-tony")
	var tRows []ab.WorkCalendarBlock
	tony.Blocks(func(r []ab.WorkCalendarBlock, err error) {
		if err != nil {
			t.Fatalf("blocks tony: %v", err)
		}
		tRows = r
	})
	if len(tRows) != 3 {
		t.Fatalf("expected 3 blocks for Tony, got %d", len(tRows))
	}

	thor := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-thor")
	var hRows []ab.WorkCalendarBlock
	thor.Blocks(func(r []ab.WorkCalendarBlock, err error) {
		if err != nil {
			t.Fatalf("blocks thor: %v", err)
		}
		hRows = r
	})
	if len(hRows) != 0 {
		t.Fatalf("expected no blocks for Thor, got %d", len(hRows))
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

// TestSpecialties_AreCanonical: las áreas vienen del catálogo, no inventadas.
func TestSpecialties_AreCanonical(t *testing.T) {
	env := New()
	specs := env.Specialties()
	found := false
	for _, s := range specs {
		if s == "medicina-general" {
			found = true
		}
		if s == "inventada" {
			t.Fatal("no invented slugs allowed")
		}
	}
	if !found {
		t.Fatal("expected medicina-general among the canonical specialties")
	}
}

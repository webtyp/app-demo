// Package config is the demo's composition root. It owns three concerns, all
// in one place: the shared runtime environment (env.go — the in-memory
// database, the veltylabs domain modules mounted on a router/loopback caller,
// the events broker and the seed), the visual theme (css.go), and the Spanish
// dictionary (lang.go).
package config

// The two brand colors the primary surface is built from:
//
//   - WASMViolet: the official WebAssembly logo purple.
//   - GoCyan:     the Go gopher mascot's light blue.
//
// They are the endpoints of the primary gradient defined in css.go.
const (
	WASMViolet = "#654FF0"
	GoCyan     = "#00ADD8"
)

package carshop

import "embed"

// MigrationsFS holds the goose SQL migrations embedded into the binary so the
// server can run them on boot without needing the source tree at runtime.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

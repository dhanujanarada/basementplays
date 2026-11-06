package migrations

import "embed"

// FS holds the plain SQL migrations applied by `ticketing migrate`.
//
//go:embed *.sql
var FS embed.FS

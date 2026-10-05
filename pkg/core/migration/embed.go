package migration

import "embed"

//go:embed sql
var migrationDir embed.FS

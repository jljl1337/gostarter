package sql

import "embed"

//go:embed migration
var MigrationDir embed.FS

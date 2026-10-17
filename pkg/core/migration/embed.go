package migration

import "embed"

//go:embed sql
var MigrationParentDir embed.FS

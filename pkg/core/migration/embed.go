package migration

import (
	"embed"
)

//go:embed sql
var migrationParentDir embed.FS

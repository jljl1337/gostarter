package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

func main() {
	migrationDir := flag.String("dir", "internal/sql/migration", "migration directory")
	migrationName := flag.String("name", "", "migration name")
	flag.Parse()

	if *migrationName == "" {
		fmt.Fprintln(os.Stderr, "Error: migration name is required")
		os.Exit(1)
	}

	// Create timestamp (in ISO 8601 format with numbers only)
	timestamp := generator.NowISO8601Number()

	// Create migration file paths
	migrationUpFile := filepath.Join(*migrationDir, fmt.Sprintf("%s-%s.up.sql", timestamp, *migrationName))
	migrationDownFile := filepath.Join(*migrationDir, fmt.Sprintf("%s-%s.down.sql", timestamp, *migrationName))

	// Ensure the directory exists
	dir := filepath.Dir(migrationUpFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Create the migration up file
	fileUp, err := os.Create(migrationUpFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating migration up file: %v\n", err)
		os.Exit(1)
	}
	fileUp.Close()

	// Create the migration down file
	fileDown, err := os.Create(migrationDownFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating migration down file: %v\n", err)
		os.Exit(1)
	}
	fileDown.Close()

	fmt.Printf("Created migration files:\n  %s\n  %s\n", migrationUpFile, migrationDownFile)
}

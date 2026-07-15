package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jljl1337/gostarter/generator"
	"github.com/jljl1337/gostarter/repository"
	"github.com/jljl1337/gostarter/sql"
)

/*
Migrate applies or rolls back database migrations based on the current state
of the database and the embedded migration files. The embedded migrations
are loaded from both the gostarter package and the appMigrationFS in the
parameter. The gostarter migrations can be overridden by the app migrations
individually by creating an app migration with the same corresponding ID.
*/
func Migrate(db *sqlx.DB, appMigrationFS embed.FS) error {
	ctx := context.Background()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	// Create the migrations table if it doesn't exist
	err = queries.CreateMigrationTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get the list of applied migrations
	appliedMigrations, err := queries.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get the list of embedded migrations (gostarter and app migrations)
	now := generator.NowISO8601()

	gostarterMigrationList, err := LoadMigrations(sql.MigrationDir, now)
	if err != nil {
		return fmt.Errorf("failed to load gostarter migrations: %w", err)
	}

	appMigrationList, err := LoadMigrations(appMigrationFS, now)
	if err != nil {
		return fmt.Errorf("failed to load app migrations: %w", err)
	}

	// Merge embedded migrations
	embeddedMigrationList := MergeMigrations(gostarterMigrationList, appMigrationList)

	// Compare the applied migrations with the embedded migrations
	if len(appliedMigrations) > len(embeddedMigrationList) {
		slog.Debug("Going to rollback applied migrations")
	} else if len(appliedMigrations) < len(embeddedMigrationList) {
		slog.Debug("Going to apply new migrations")
	} else {
		slog.Debug("Going to verify existing migrations")
	}

	minLen := min(len(appliedMigrations), len(embeddedMigrationList))

	// Verify overlap migrations
	for i := range minLen {
		slog.Debug("Verifying migration: " + appliedMigrations[i].ID)
		if appliedMigrations[i].ID != embeddedMigrationList[i].ID ||
			appliedMigrations[i].UpStatement != embeddedMigrationList[i].UpStatement ||
			appliedMigrations[i].DownStatement != embeddedMigrationList[i].DownStatement {
			return fmt.Errorf("applied migration does not match embedded migration at ID %s", appliedMigrations[i].ID)
		}
	}

	if len(appliedMigrations) < len(embeddedMigrationList) {
		// Apply new migrations
		for i := minLen; i < len(embeddedMigrationList); i++ {
			slog.Info("Applying migration: " + embeddedMigrationList[i].ID)
			_, err := tx.ExecContext(ctx, embeddedMigrationList[i].UpStatement)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", embeddedMigrationList[i].ID, err)
			}

			err = queries.InsertMigration(ctx, embeddedMigrationList[i])
			if err != nil {
				return fmt.Errorf("failed to record applied migration %s: %w", embeddedMigrationList[i].ID, err)
			}
		}
	} else if len(appliedMigrations) > len(embeddedMigrationList) {
		// Rollback applied migrations
		for i := len(appliedMigrations) - 1; i >= minLen; i-- {
			slog.Info("Rolling back migration: " + appliedMigrations[i].ID)
			_, err := tx.ExecContext(ctx, appliedMigrations[i].DownStatement)
			if err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", appliedMigrations[i].ID, err)
			}

			err = queries.DeleteMigration(ctx, appliedMigrations[i].ID)
			if err != nil {
				return fmt.Errorf("failed to remove migration record %s: %w", appliedMigrations[i].ID, err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

/*
LoadMigrations loads migrations from the embedded filesystem and returns a
slice of Migration structs.
*/
func LoadMigrations(fs embed.FS, now string) ([]repository.Migration, error) {
	// Get all migrations from the embedded filesystem
	dirEntryList, err := fs.ReadDir("migration")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	migrationMap := make(map[string]repository.Migration)
	migrationList := make([]repository.Migration, 0)

	for _, dirEntry := range dirEntryList {
		// Skip directories
		if dirEntry.IsDir() {
			slog.Warn("Skipping directory in migrations: " + dirEntry.Name())
			continue
		}

		// Get the migration statement
		filename := dirEntry.Name()

		// Remove .up.sql or .down.sql suffix to get the migration ID
		if len(filename) < 7 || (filename[len(filename)-7:] != ".up.sql" && filename[len(filename)-9:] != ".down.sql") {
			slog.Warn("Skipping file with invalid migration filename: " + filename)
			continue
		}

		// Read the migration file
		statementBytes, err := sql.MigrationDir.ReadFile("migration/" + filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}
		statement := string(statementBytes)

		// Determine if it's an up or down migration
		var migrationID string
		isUp := false

		if filename[len(filename)-7:] == ".up.sql" {
			migrationID = filename[:len(filename)-7]
			isUp = true
		} else if filename[len(filename)-9:] == ".down.sql" {
			migrationID = filename[:len(filename)-9]
		}

		// Update or create the migration entry
		migration, exists := migrationMap[migrationID]
		if !exists {
			migration = repository.Migration{
				ID:         migrationID,
				ExecutedAt: now,
			}
		}

		if isUp {
			migration.UpStatement = statement
		} else {
			migration.DownStatement = statement
		}

		migrationMap[migrationID] = migration
		if exists {
			migrationList = append(migrationList, migration)
		}
	}

	return migrationList, nil
}

/*
MergeMigrations merges two slices of migrations, gostarterMigrations and
appMigrations, into a single slice. If a migration with the same ID exists
in both slices, the one from appMigrations takes precedence. The resulting
slice should be sorted by migration ID in ascending order.
*/
func MergeMigrations(gostarterMigrations, appMigrations []repository.Migration) []repository.Migration {
	mergedMigrations := make([]repository.Migration, 0)

	gostarterIndex := 0
	appIndex := 0

	for gostarterIndex < len(gostarterMigrations) || appIndex < len(appMigrations) {
		if gostarterIndex < len(gostarterMigrations) && appIndex < len(appMigrations) {
			gostarterMigration := gostarterMigrations[gostarterIndex]
			appMigration := appMigrations[appIndex]

			if gostarterMigration.ID == appMigration.ID {
				mergedMigrations = append(mergedMigrations, appMigration)
				gostarterIndex++
				appIndex++
			} else if gostarterMigration.ID < appMigration.ID {
				mergedMigrations = append(mergedMigrations, gostarterMigration)
				gostarterIndex++
			} else {
				mergedMigrations = append(mergedMigrations, appMigration)
				appIndex++
			}
		} else if gostarterIndex < len(gostarterMigrations) {
			mergedMigrations = append(mergedMigrations, gostarterMigrations[gostarterIndex])
			gostarterIndex++
		} else if appIndex < len(appMigrations) {
			mergedMigrations = append(mergedMigrations, appMigrations[appIndex])
			appIndex++
		}
	}

	return mergedMigrations
}

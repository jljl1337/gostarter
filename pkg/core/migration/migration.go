package migration

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/sql"
)

// Run [MigrateAllContext] with a background context.
func MigrateAll(db *sqlx.DB, appMigrationFS embed.FS) error {
	ctx := context.Background()
	return MigrateAllContext(ctx, db, appMigrationFS)
}

// Run [MigrateContext] with gostarter migrations.
func MigrateAllContext(ctx context.Context, db *sqlx.DB, appMigrationFS embed.FS) error {
	return MigrateContext(ctx, db, true, appMigrationFS)
}

/*
Migrate applies or rolls back database migrations based on the current state
of the database and the embedded migration files. The embedded migrations
are loaded from both the gostarter package and the appMigrationFS in the
parameter.
*/
func MigrateContext(ctx context.Context, db *sqlx.DB, runGostarterMigration bool, appMigrationFS embed.FS) error {
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

	// Set the same timestamp for all migrations in this run
	now := generator.NowISO8601()

	// Get the list of embedded migrations (gostarter and app migrations)
	gostarterMigrationList, err := LoadMigrations(sql.MigrationDir, now)
	if err != nil {
		return fmt.Errorf("failed to load gostarter migrations: %w", err)
	}

	appMigrationList, err := LoadMigrations(appMigrationFS, now)
	if err != nil {
		return fmt.Errorf("failed to load app migrations: %w", err)
	}

	// Replace gostarter migrations with app migrations if they have the same ID
	gostarterMigrationList = replaceGostarterMigrations(gostarterMigrationList, appMigrationList)

	// Migrate gostarter migrations
	if runGostarterMigration {
		err = migrate(ctx, true, queries, gostarterMigrationList)
		if err != nil {
			return fmt.Errorf("failed to migrate gostarter migrations: %w", err)
		}
	}

	// Migrate app migrations
	err = migrate(ctx, false, queries, appMigrationList)
	if err != nil {
		return fmt.Errorf("failed to migrate app migrations: %w", err)
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
replaceGostarterMigrations replaces the gostarter migrations in the provided
slice with the app migrations if an app migration has the same ID as a
gostarter migration.
*/
func replaceGostarterMigrations(gostarterMigrations, appMigrations []repository.Migration) []repository.Migration {
	// Create a map of app migrations for quick lookup
	appMigrationMap := make(map[string]repository.Migration)
	for _, migration := range appMigrations {
		appMigrationMap[migration.ID] = migration
	}

	// Replace gostarter migrations with app migrations if they have the same ID
	for i, gostarterMigration := range gostarterMigrations {
		if appMigration, exists := appMigrationMap[gostarterMigration.ID]; exists {
			gostarterMigrations[i] = appMigration
		}
	}

	return gostarterMigrations
}

func migrate(
	ctx context.Context,
	isGostarterMigration bool,
	queries *repository.Queries,
	embeddedMigrationList []repository.Migration,
) error {
	var err error

	// Get the list of applied migrations
	var appliedMigrations []repository.Migration
	if isGostarterMigration {
		appliedMigrations, err = queries.GetAppliedGostarterMigrations(ctx)
	} else {
		appliedMigrations, err = queries.GetAppliedAppMigrations(ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

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

	// Apply or rollback migrations as needed
	if len(appliedMigrations) < len(embeddedMigrationList) {
		// Apply new migrations
		for i := minLen; i < len(embeddedMigrationList); i++ {
			slog.Info("Applying migration: " + embeddedMigrationList[i].ID)
			_, err := queries.ExecContext(ctx, embeddedMigrationList[i].UpStatement)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", embeddedMigrationList[i].ID, err)
			}

			if isGostarterMigration {
				err = queries.InsertGostarterMigration(ctx, embeddedMigrationList[i])
			} else {
				err = queries.InsertAppMigration(ctx, embeddedMigrationList[i])
			}
			if err != nil {
				return fmt.Errorf("failed to record applied migration %s: %w", embeddedMigrationList[i].ID, err)
			}
		}
	} else if len(appliedMigrations) > len(embeddedMigrationList) {
		// Rollback applied migrations
		for i := len(appliedMigrations) - 1; i >= minLen; i-- {
			slog.Info("Rolling back migration: " + appliedMigrations[i].ID)
			_, err := queries.ExecContext(ctx, appliedMigrations[i].DownStatement)
			if err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", appliedMigrations[i].ID, err)
			}

			if isGostarterMigration {
				err = queries.DeleteGostarterMigration(ctx, appliedMigrations[i].ID)
			} else {
				err = queries.DeleteAppMigration(ctx, appliedMigrations[i].ID)
			}
			if err != nil {
				return fmt.Errorf("failed to remove migration record %s: %w", appliedMigrations[i].ID, err)
			}
		}
	}

	return nil
}

package migration

import (
	"cmp"
	"context"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/jljl1337/gostarter/pkg/generator"
	"github.com/jljl1337/gostarter/pkg/log"
	"github.com/jljl1337/gostarter/pkg/repository"
)

type Candidate struct {
	name        string
	db          *repository.Database
	migrationFS fs.FS
}

type Service struct {
	newRepositoryTx        NewRepositoryTx
	migrationCandidateList []Candidate
}

func NewService() *Service {
	return &Service{
		newRepositoryTx: newDefaultRepositoryTx,
	}
}

func (s *Service) AddCandidate(name string, db *repository.Database, migrationFS fs.FS) {
	s.migrationCandidateList = append(s.migrationCandidateList, Candidate{
		name:        name,
		db:          db,
		migrationFS: migrationFS,
	})
}

func (s *Service) migrateAll(ctx context.Context) error {
	if len(s.migrationCandidateList) == 0 {
		return fmt.Errorf("no migration candidates found")
	}

	for _, candidate := range s.migrationCandidateList {
		log.Debugf("Starting migration for candidate: %s", candidate.name)
		if err := s.migrateCandidate(ctx, candidate); err != nil {
			return fmt.Errorf("failed to migrate candidate %s: %w", candidate.name, err)
		}
	}

	return nil
}

func (s *Service) migrateCandidate(ctx context.Context, candidate Candidate) error {
	repo, rollback, commit, err := s.newRepositoryTx(ctx, candidate.db, nil)
	if err != nil {
		return fmt.Errorf("failed to create repository with transaction: %w", err)
	}
	defer rollback()

	err = repo.CreateMigrationTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	now := generator.NowISO8601()

	embeddedMigrationList, err := s.loadMigrations(candidate.migrationFS, now)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrationList, err := repo.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	err = s.runMigration(ctx, repo, embeddedMigrationList, appliedMigrationList)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	// Commit the transaction
	if err := commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) loadMigrations(migrationFS fs.FS, now string) ([]Migration, error) {
	migrationMap := make(map[string]Migration)
	migrationList := make([]Migration, 0)

	// Get all migrations from the embedded filesystem
	err := fs.WalkDir(migrationFS, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk migration directory: %w", err)
		}

		// Skip directories
		if dirEntry.IsDir() {
			log.Warnf("Skipping directory in migrations: %s", dirEntry.Name())
			return nil
		}

		// Get the migration statement
		filename := dirEntry.Name()

		// Skip files that don't match the expected migration filename pattern
		if !strings.HasSuffix(filename, ".up.sql") && !strings.HasSuffix(filename, ".down.sql") {
			log.Warnf("Skipping file with invalid migration filename: %s", filename)
			return nil
		}

		// Read the migration file
		statementBytes, err := fs.ReadFile(migrationFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}
		statement := string(statementBytes)

		// Determine if it's an up or down migration
		var migrationID string
		isUp := false

		if before, ok := strings.CutSuffix(filename, ".up.sql"); ok {
			migrationID = before
			isUp = true
		} else {
			before, _ := strings.CutSuffix(filename, ".down.sql")
			migrationID = before
		}

		// Update or create the migration entry
		migration, exists := migrationMap[migrationID]
		if !exists {
			migration = Migration{
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

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Sort the migration list by ID to ensure they are applied in the correct
	// order
	slices.SortFunc(migrationList, func(a, b Migration) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return migrationList, nil
}

func (s *Service) runMigration(ctx context.Context, repo Repository, embeddedList, appliedList []Migration) error {
	embeddedLen := len(embeddedList)
	appliedLen := len(appliedList)

	if embeddedLen > appliedLen {
		log.Debugf("Applying %d new migrations", embeddedLen-appliedLen)
	} else if embeddedLen < appliedLen {
		log.Debugf("Rolling back %d migrations", appliedLen-embeddedLen)
	} else {
		log.Debugf("Verifying %d migrations", embeddedLen)
	}

	// Verify overlapping migrations
	minLen := min(embeddedLen, appliedLen)

	for i := range minLen {
		log.Debugf("Verifying migration %s", embeddedList[i].ID)
		idSame := embeddedList[i].ID == appliedList[i].ID
		upSame := embeddedList[i].UpStatement == appliedList[i].UpStatement
		downSame := embeddedList[i].DownStatement == appliedList[i].DownStatement

		if !idSame || !upSame || !downSame {
			return fmt.Errorf("applied migration %s does not match embedded migration %s", appliedList[i].ID, embeddedList[i].ID)
		}
	}

	// Apply or rollback migrations as needed
	if embeddedLen > appliedLen {
		// Apply new migrations
		for i := minLen; i < embeddedLen; i++ {
			log.Infof("Applying migration %s", embeddedList[i].ID)
			err := repo.RunStatement(ctx, embeddedList[i].UpStatement)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", embeddedList[i].ID, err)
			}

			err = repo.InsertMigration(ctx, embeddedList[i])
			if err != nil {
				return fmt.Errorf("failed to insert applied migration %s: %w", embeddedList[i].ID, err)
			}
		}
	} else if embeddedLen < appliedLen {
		// Rollback migrations
		for i := appliedLen - 1; i >= minLen; i-- {
			log.Infof("Rolling back migration %s", appliedList[i].ID)
			err := repo.RunStatement(ctx, appliedList[i].DownStatement)
			if err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", appliedList[i].ID, err)
			}

			err = repo.DeleteMigrationByID(ctx, appliedList[i].ID)
			if err != nil {
				return fmt.Errorf("failed to delete rolled back migration %s: %w", appliedList[i].ID, err)
			}
		}
	}

	return nil
}

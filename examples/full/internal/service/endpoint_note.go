package service

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/env"
	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

func (s *EndpointService) CreateNote(ctx context.Context, accountID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := generator.NowISO8601()

	queries := repository.NewQueries(tx)

	note := repository.Note{
		ID:         s.idGenerator(),
		AccountID:  accountID,
		Body:       "This is a new note.",
		Positivity: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err = queries.CreateNote(ctx, note)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to create note: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	if err := s.queueManager.Enqueue(ctx, env.QueueLaneNotePositivity, note.ID); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to enqueue note: %w", err)
	}

	return nil
}

func (s *EndpointService) GetNotesByAccountID(ctx context.Context, accountID string) ([]repository.Note, error) {
	queries := repository.NewQueries(s.db)

	notes, err := queries.GetNotesByAccountID(ctx, accountID)
	if err != nil {
		return nil, service.NewServiceErrorf(service.ErrCodeInternal, "failed to get notes by account ID: %w", err)
	}

	return notes, nil
}

func (s *EndpointService) UpdateNoteBodyByID(ctx context.Context, accountID string, noteID string, newBody string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := generator.NowISO8601()

	queries := repository.NewQueries(tx)

	notes, err := queries.GetNoteByID(ctx, noteID)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get note by ID: %w", err)
	}

	if len(notes) == 0 {
		return service.NewServiceErrorf(service.ErrCodeNotFound, "note with ID %s not found", noteID)
	}

	if len(notes) > 1 {
		return service.NewServiceErrorf(service.ErrCodeInternal, "multiple notes found with ID %s", noteID)
	}

	note := notes[0]
	if note.AccountID != accountID {
		return service.NewServiceErrorf(service.ErrCodeNotFound, "note with ID %s not found", noteID)
	}

	updateParams := repository.UpdateNoteBodyByIDParams{
		ID:        noteID,
		Body:      newBody,
		UpdatedAt: now,
	}

	err = queries.UpdateNoteBodyByID(ctx, updateParams)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update note by ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	if err := s.queueManager.Enqueue(ctx, env.QueueLaneNotePositivity, note.ID); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to enqueue note: %w", err)
	}

	return nil
}

func (s *EndpointService) DeleteNoteByID(ctx context.Context, accountID string, noteID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	notes, err := queries.GetNoteByID(ctx, noteID)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get note by ID: %w", err)
	}

	if len(notes) == 0 {
		return service.NewServiceErrorf(service.ErrCodeNotFound, "note with ID %s not found", noteID)
	}

	if len(notes) > 1 {
		return service.NewServiceErrorf(service.ErrCodeInternal, "multiple notes found with ID %s", noteID)
	}

	note := notes[0]
	if note.AccountID != accountID {
		return service.NewServiceErrorf(service.ErrCodeNotFound, "note with ID %s not found", noteID)
	}

	err = queries.DeleteNoteByID(ctx, noteID)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete note by ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	return nil
}

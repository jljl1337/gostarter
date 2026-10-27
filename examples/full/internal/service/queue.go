package service

import (
	"context"
	"strings"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/queue"
	gsRepository "github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/env"
	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

type QueueService struct {
	db *gsRepository.Database
}

func NewQueueService(db *gsRepository.Database) *QueueService {
	return &QueueService{
		db: db,
	}
}

func (s *QueueService) GetQueueLanes() []queue.Lane {
	return []queue.Lane{
		{
			Name:        env.QueueLaneNotePositivity,
			TaskHandler: s.UpdateNotePositivity,
		},
	}
}

func (s *QueueService) UpdateNotePositivity(payload string) error {
	time.Sleep(1 * time.Second) // Simulate some processing time

	var positiveWords = []string{"good", "great", "excellent", "awesome", "fantastic", "amazing", "wonderful", "positive", "happy", "joyful"}
	var negativeWords = []string{"bad", "terrible", "awful", "horrible", "negative", "sad", "unhappy", "angry", "frustrated", "disappointed"}

	ctx := context.Background()

	queries, rollback, commit, err := repository.NewQueriesTx(ctx, s.db, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to create queries with transaction: %w", err)
	}
	defer rollback()

	notes, err := queries.GetNoteByID(ctx, payload)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get note by ID: %w", err)
	}

	if len(notes) == 0 {
		return service.NewServiceErrorf(service.ErrCodeNotFound, "note with ID %s not found", payload)
	}

	if len(notes) > 1 {
		return service.NewServiceErrorf(service.ErrCodeInternal, "multiple notes found with ID %s", payload)
	}

	note := notes[0]
	positivity := 0

	for _, word := range positiveWords {
		count := strings.Count(note.Body, word)
		positivity += count
	}

	for _, word := range negativeWords {
		count := strings.Count(note.Body, word)
		positivity -= count
	}

	updateParams := repository.UpdateNotePositivityByIDParams{
		ID:         note.ID,
		Positivity: positivity,
		UpdatedAt:  generator.NowISO8601(),
	}

	err = queries.UpdateNotePositivityByID(ctx, updateParams)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update note positivity by ID: %w", err)
	}

	if err := commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	return nil
}

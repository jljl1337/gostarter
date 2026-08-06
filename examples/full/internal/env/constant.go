package env

import "github.com/jljl1337/gostarter/pkg/shared/env"

const (
	QueueLaneNotePositivity = "note_positivity"
)

var (
	DeleteNotesCronSchedule string
)

func MustSetConstants(files ...string) {
	env.MustSetConstantsWithoutPrefix(files...)

	DeleteNotesCronSchedule = env.MustGetString("DELETE_NOTES_CRON_SCHEDULE", "* * * * *")
}

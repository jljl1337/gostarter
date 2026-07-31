package env

import "github.com/jljl1337/gostarter/pkg/shared/env"

var (
	DeleteNotesCronSchedule string
)

func MustSetConstants() {
	env.MustSetConstants(false)

	DeleteNotesCronSchedule = env.MustGetString("DELETE_NOTES_CRON_SCHEDULE", "* * * * *")
}

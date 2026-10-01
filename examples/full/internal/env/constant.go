package env

import "github.com/jljl1337/gostarter/pkg/shared/env"

const (
	QueueLaneNotePositivity = "note_positivity"

	// RoleOwner is the highest role: it can manage the role of any lower role.
	RoleOwner = "owner"
	// RoleModerator can list and delete accounts of a lower role.
	RoleModerator = "moderator"
	// RoleUser is the lowest role, given to every account but the first one.
	RoleUser = "user"
)

var (
	DeleteNotesCronSchedule string
)

func MustSetConstants(files ...string) {
	env.MustSetConstantsWithoutPrefix(files...)

	DeleteNotesCronSchedule = env.MustGetString("DELETE_NOTES_CRON_SCHEDULE", "* * * * *")
}

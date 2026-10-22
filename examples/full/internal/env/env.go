package env

import (
	"fmt"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

const (
	QueueLaneNotePositivity = "note_positivity"

	// RoleOwner is the highest role: it can manage the role of any lower role.
	RoleOwner = "owner"
	// RoleModerator can list and delete accounts of a lower role.
	RoleModerator = "moderator"
	// RoleUser is the lowest role, given to every account but the first one.
	RoleUser = "user"
)

var Variables Config

type Config struct {
	DeleteNotesCronSchedule string `env:"DELETE_NOTES_CRON_SCHEDULE" default:"* * * * *"`
}

func SetVariables(files ...string) error {
	err := env.SetVariables(files...)
	if err != nil {
		return fmt.Errorf("failed to set constants: %w", err)
	}

	loader := &env.EnvLoader{}
	if err := loader.LoadIntoStruct(&Variables); err != nil {
		return fmt.Errorf("failed to load environment variables into struct: %w", err)
	}

	return nil
}

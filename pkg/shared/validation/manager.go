package validation

import "regexp"

type ValidationManager struct {
	usernameRegex *regexp.Regexp
	passwordRegex *regexp.Regexp
}

func NewValidationManager(usernameRegex string, passwordRegex string) (*ValidationManager, error) {
	usernameRegexp, err := regexp.Compile(usernameRegex)
	if err != nil {
		return nil, err
	}

	passwordRegexp, err := regexp.Compile(passwordRegex)
	if err != nil {
		return nil, err
	}

	return &ValidationManager{
		usernameRegex: usernameRegexp,
		passwordRegex: passwordRegexp,
	}, nil
}

func (v *ValidationManager) ValidateUsername(username string) bool {
	return v.usernameRegex.MatchString(username)
}

func (v *ValidationManager) ValidatePassword(password string) bool {
	return v.passwordRegex.MatchString(password)
}

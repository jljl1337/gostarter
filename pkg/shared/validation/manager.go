package validation

import "regexp"

type ValidationManager struct {
	languageCodeMap map[string]bool
	usernameRegex   *regexp.Regexp
	passwordRegex   *regexp.Regexp
}

func NewValidationManager(languageCodeList []string, usernameRegex string, passwordRegex string) (*ValidationManager, error) {
	usernameRegexp, err := regexp.Compile(usernameRegex)
	if err != nil {
		return nil, err
	}

	passwordRegexp, err := regexp.Compile(passwordRegex)
	if err != nil {
		return nil, err
	}

	languageCodeMap := make(map[string]bool)
	for _, code := range languageCodeList {
		languageCodeMap[code] = true
	}

	return &ValidationManager{
		languageCodeMap: languageCodeMap,
		usernameRegex:   usernameRegexp,
		passwordRegex:   passwordRegexp,
	}, nil
}

func (v *ValidationManager) ValidateLanguageCode(languageCode string) bool {
	return v.languageCodeMap[languageCode]
}

func (v *ValidationManager) ValidateUsername(username string) bool {
	return v.usernameRegex.MatchString(username)
}

func (v *ValidationManager) ValidatePassword(password string) bool {
	return v.passwordRegex.MatchString(password)
}

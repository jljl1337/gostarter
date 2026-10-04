package transport

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

type CookieGeneratorConfig struct {
	Name     string
	Secure   bool
	HttpOnly bool
	SameSite string
}

type CookieGenerator struct {
	name     string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func NewCookieGeneratorFromEnv() *CookieGenerator {
	return NewCookieGenerator(CookieGeneratorConfig{
		Name:     env.SessionCookieName,
		Secure:   env.SessionCookieSecure,
		HttpOnly: env.SessionCookieHttpOnly,
		SameSite: env.SessionCookieSameSite,
	})
}

func NewCookieGenerator(config CookieGeneratorConfig) *CookieGenerator {
	var sameSite http.SameSite

	switch config.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		log.Warnf("Invalid SameSite value '%s', defaulting to 'none'", config.SameSite)
		sameSite = http.SameSiteNoneMode
	}

	return &CookieGenerator{
		name:     config.Name,
		secure:   config.Secure,
		httpOnly: config.HttpOnly,
		sameSite: sameSite,
	}
}

func (cm *CookieGenerator) NewActiveSessionCookie(sessionToken string) *http.Cookie {
	return &http.Cookie{
		Name:     cm.name,
		Value:    sessionToken,
		Path:     "/",
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
}

func (cm *CookieGenerator) NewExpiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cm.name,
		Value:    "",
		Path:     "/",
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
		MaxAge:   -1,
	}
}

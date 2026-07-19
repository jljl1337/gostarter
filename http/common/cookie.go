package common

import (
	"net/http"

	"github.com/jljl1337/gostarter/env"
)

type CookieGenerator struct {
	name     string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func NewCookieGeneratorFromEnv() *CookieGenerator {
	return NewCookieGenerator(env.SessionCookieName, env.SessionCookieSecure, env.SessionCookieHttpOnly, env.SessionCookieSameSite)
}

func NewCookieGenerator(name string, secure bool, httpOnly bool, sameSite http.SameSite) *CookieGenerator {
	return &CookieGenerator{
		name:     name,
		secure:   secure,
		httpOnly: httpOnly,
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

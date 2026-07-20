package endpoint

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/http/middleware"
	"github.com/jljl1337/gostarter/pkg/core/service/endpoint"
	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type signUpRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LanguageCode string `json:"languageCode"`
}

type signInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signInPreSessionCSRFTokenResponse struct {
	CSRFToken string `json:"csrfToken"`
}

func (h *EndpointHandler) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/sign-up", h.signUp)
	mux.HandleFunc("POST /auth/pre-session", h.preSession)
	mux.HandleFunc("POST /auth/sign-in", h.signIn)
	mux.HandleFunc("POST /auth/sign-out", h.signOut)
	mux.HandleFunc("POST /auth/sign-out-all", h.signOutAll)
	mux.HandleFunc("GET /auth/csrf-token", h.csrfToken)
}

func (h *EndpointHandler) signUp(w http.ResponseWriter, r *http.Request) {
	// Input validation
	var req signUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessageResponse(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.responseHandler.WriteMessageResponse(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if req.LanguageCode == "" {
		h.responseHandler.WriteMessageResponse(w, "Language code is required", http.StatusBadRequest)
		return
	}

	if err := h.service.SignUp(r.Context(), endpoint.SignUpParams{
		Username:     req.Username,
		Password:     req.Password,
		LanguageCode: req.LanguageCode,
	}); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	h.responseHandler.WriteMessageResponse(w, "Account signed up successfully", http.StatusCreated)
}

func (h *EndpointHandler) preSession(w http.ResponseWriter, r *http.Request) {
	// Process the request
	sessionToken, CSRFToken, err := h.service.GetPreSession(r.Context())
	if err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	http.SetCookie(w, h.cookieGenerator.NewActiveSessionCookie(sessionToken))

	h.responseHandler.WriteJSONResponse(w, http.StatusOK, signInPreSessionCSRFTokenResponse{
		CSRFToken: CSRFToken,
	})
}

func (h *EndpointHandler) signIn(w http.ResponseWriter, r *http.Request) {
	// Input validation
	preSessionToken, err := r.Cookie(env.SessionCookieName)
	if err != nil {
		h.responseHandler.WriteMessageResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	preSessionCSRFToken := r.Header.Get("X-CSRF-Token")
	if preSessionCSRFToken == "" {
		h.responseHandler.WriteMessageResponse(w, "CSRF token is required", http.StatusUnauthorized)
		return
	}

	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessageResponse(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		h.responseHandler.WriteMessageResponse(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Process the request
	sessionToken, CSRFToken, err := h.service.SignIn(r.Context(), endpoint.SignInParams{
		PreSessionToken:     preSessionToken.Value,
		PreSessionCSRFToken: preSessionCSRFToken,
		Username:            req.Username,
		Password:            req.Password,
	})
	if err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	http.SetCookie(w, h.cookieGenerator.NewActiveSessionCookie(sessionToken))

	h.responseHandler.WriteJSONResponse(w, http.StatusOK, signInPreSessionCSRFTokenResponse{
		CSRFToken: CSRFToken,
	})
}

func (h *EndpointHandler) signOut(w http.ResponseWriter, r *http.Request) {
	// Input validation
	sessionToken, err := r.Cookie(env.SessionCookieName)
	if err != nil {
		h.responseHandler.WriteMessageResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Process the request
	if err := h.service.SignOut(r.Context(), sessionToken.Value); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	http.SetCookie(w, h.cookieGenerator.NewExpiredSessionCookie())

	h.responseHandler.WriteMessageResponse(w, "Account logged out successfully", http.StatusOK)
}

func (h *EndpointHandler) signOutAll(w http.ResponseWriter, r *http.Request) {
	// Process the request
	ctx := r.Context()
	account := middleware.GetAccountFromContext(ctx)
	if account == nil {
		slog.Error("Error getting account from context")
		h.responseHandler.WriteMessageResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.service.SignOutAllSession(r.Context(), *account); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	http.SetCookie(w, h.cookieGenerator.NewExpiredSessionCookie())

	h.responseHandler.WriteMessageResponse(w, "Account logged out from all sessions successfully", http.StatusOK)
}

func (h *EndpointHandler) csrfToken(w http.ResponseWriter, r *http.Request) {
	// Input validation
	sessionToken, err := r.Cookie(env.SessionCookieName)
	if err != nil {
		h.responseHandler.WriteMessageResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Process the request
	CSRFToken, err := h.service.CSRFToken(r.Context(), sessionToken.Value)
	if err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	// Respond to the client
	h.responseHandler.WriteJSONResponse(w, http.StatusOK, signInPreSessionCSRFTokenResponse{
		CSRFToken: CSRFToken,
	})
}

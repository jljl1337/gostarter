package transport

import (
	"encoding/json"
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/service"
)

type getCurrentAccountResponse struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	LanguageCode string `json:"languageCode"`
	CreatedAt    string `json:"createdAt"`
}

func (h *EndpointHandler) registerAccountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /accounts/me", h.getCurrentAccount)
	mux.HandleFunc("PATCH /accounts/me/username", h.updateUsername)
	mux.HandleFunc("PATCH /accounts/me/password", h.updatePassword)
	mux.HandleFunc("PATCH /accounts/me/language", h.updateLanguage)
	mux.HandleFunc("DELETE /accounts/me", h.deleteCurrentAccount)
}

func (h *EndpointHandler) getCurrentAccount(w http.ResponseWriter, r *http.Request) {
	// Process the request
	account := GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	// Respond to the client
	response := getCurrentAccountResponse{
		ID:           account.ID,
		Username:     account.Username,
		Role:         account.Role,
		LanguageCode: account.LanguageCode,
		CreatedAt:    account.CreatedAt,
	}
	h.responseHandler.WriteJSON(w, http.StatusOK, response)
}

func (h *EndpointHandler) updateUsername(w http.ResponseWriter, r *http.Request) {
	// Input validation
	var req struct {
		NewUsername string `json:"newUsername"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessage(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.NewUsername == "" {
		h.responseHandler.WriteMessage(w, "New username is required", http.StatusBadRequest)
		return
	}

	// Process the request
	account := GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	if err := h.service.UpdateUsernameByID(r.Context(), service.UpdateUsernameByIDParams{
		Account:     *account,
		NewUsername: req.NewUsername,
	}); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	// Respond to the client
	h.responseHandler.WriteMessage(w, "Username updated successfully", http.StatusOK)
}

func (h *EndpointHandler) updatePassword(w http.ResponseWriter, r *http.Request) {
	// Input validation
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessage(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.NewPassword == "" {
		h.responseHandler.WriteMessage(w, "New password is required", http.StatusBadRequest)
		return
	}

	// Process the request
	account := GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	if err := h.service.UpdatePasswordByID(r.Context(), service.UpdatePasswordByIDParams{
		Account:     *account,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	// Respond to the client
	h.responseHandler.WriteMessage(w, "Password updated successfully", http.StatusOK)
}

func (h *EndpointHandler) updateLanguage(w http.ResponseWriter, r *http.Request) {
	// Input validation
	var req struct {
		LanguageCode string `json:"languageCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessage(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.LanguageCode == "" {
		h.responseHandler.WriteMessage(w, "Language code is required", http.StatusBadRequest)
		return
	}

	// Process the request
	account := GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	if err := h.service.UpdateLanguageByID(r.Context(), service.UpdateLanguageByIDParams{
		Account:      *account,
		LanguageCode: req.LanguageCode,
	}); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	// Respond to the client
	h.responseHandler.WriteMessage(w, "Language updated successfully", http.StatusOK)
}

func (h *EndpointHandler) deleteCurrentAccount(w http.ResponseWriter, r *http.Request) {
	// Process the request
	account := GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	if err := h.service.DeleteAccountByID(r.Context(), *account); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	// Respond to the client
	http.SetCookie(w, h.cookieGenerator.NewExpiredSessionCookie())

	h.responseHandler.WriteMessage(w, "Account deleted successfully", http.StatusOK)
}

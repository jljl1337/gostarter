package transport

import (
	"encoding/json"
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/transport"

	"github.com/jljl1337/gostarter/examples/full/internal/env"
	"github.com/jljl1337/gostarter/examples/full/internal/repository"
	"github.com/jljl1337/gostarter/examples/full/internal/service"
)

type accountResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

func newAccountResponse(account repository.Account) accountResponse {
	return accountResponse{
		ID:        account.ID,
		Username:  account.Username,
		Role:      account.Role,
		CreatedAt: account.CreatedAt,
	}
}

// registerAccountRoleRoutes registers the role gated account routes. Access is
// granted by the role manager of the middleware based on the first path
// segment, so "/moderator" is reachable by moderators and owners only, and
// "/owner" by owners only. A lower role gets a not found response, as if the
// route did not exist.
func (h *EndpointHandler) registerAccountRoleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /"+env.RoleModerator+"/accounts", h.getAccounts)
	mux.HandleFunc("DELETE /"+env.RoleModerator+"/accounts/{id}", h.deleteAccountByID)
	mux.HandleFunc("PATCH /"+env.RoleOwner+"/accounts/{id}/role", h.updateAccountRoleByID)
}

func (h *EndpointHandler) getAccounts(w http.ResponseWriter, r *http.Request) {
	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	accounts, err := h.service.GetAccounts(r.Context())
	if err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	response := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		response = append(response, newAccountResponse(a))
	}

	h.responseHandler.WriteJSON(w, http.StatusOK, response)
}

func (h *EndpointHandler) updateAccountRoleByID(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		h.responseHandler.WriteMessage(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessage(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		h.responseHandler.WriteMessage(w, "Role is required", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAccountRoleByID(r.Context(), service.UpdateAccountRoleByIDParams{
		Actor:     *account,
		AccountID: accountID,
		NewRole:   req.Role,
	}); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	h.responseHandler.WriteMessage(w, "Account role updated successfully", http.StatusOK)
}

func (h *EndpointHandler) deleteAccountByID(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		h.responseHandler.WriteMessage(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorf(w, "failed to get account from context")
		return
	}

	if err := h.service.DeleteAccountByID(r.Context(), *account, accountID); err != nil {
		h.responseHandler.WriteServiceError(w, err)
		return
	}

	h.responseHandler.WriteMessage(w, "Account deleted successfully", http.StatusOK)
}

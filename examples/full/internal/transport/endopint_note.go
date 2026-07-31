package transport

import (
	"encoding/json"
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/transport"
)

func (h *EndpointHandler) registerNoteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /notes", h.createNote)
	mux.HandleFunc("GET /notes", h.getNotesOfAccount)
	mux.HandleFunc("PUT /notes/{id}", h.updateNoteByID)
	mux.HandleFunc("DELETE /notes/{id}", h.deleteNoteByID)
}

func (h *EndpointHandler) createNote(w http.ResponseWriter, r *http.Request) {
	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorResponsef(w, "failed to get account from context")
		return
	}

	if err := h.service.CreateNote(r.Context(), account.ID); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	h.responseHandler.WriteMessageResponse(w, "Note created successfully", http.StatusCreated)
}

func (h *EndpointHandler) getNotesOfAccount(w http.ResponseWriter, r *http.Request) {
	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorResponsef(w, "failed to get account from context")
		return
	}

	notes, err := h.service.GetNotesByAccountID(r.Context(), account.ID)
	if err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	h.responseHandler.WriteJSONResponse(w, http.StatusOK, notes)
}

func (h *EndpointHandler) updateNoteByID(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	noteID := r.PathValue("id")
	if noteID == "" {
		h.responseHandler.WriteMessageResponse(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorResponsef(w, "failed to get account from context")
		return
	}

	// Parse request body
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseHandler.WriteMessageResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the note
	if err := h.service.UpdateNoteByID(r.Context(), account.ID, noteID, req.Body); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	h.responseHandler.WriteMessageResponse(w, "Note updated successfully", http.StatusOK)
}

func (h *EndpointHandler) deleteNoteByID(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	noteID := r.PathValue("id")
	if noteID == "" {
		h.responseHandler.WriteMessageResponse(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	account := transport.GetAccountFromContext(r.Context())
	if account == nil {
		h.responseHandler.WriteErrorResponsef(w, "failed to get account from context")
		return
	}

	// Delete the note
	if err := h.service.DeleteNoteByID(r.Context(), account.ID, noteID); err != nil {
		h.responseHandler.WriteErrorResponse(w, err)
		return
	}

	h.responseHandler.WriteMessageResponse(w, "Note deleted successfully", http.StatusOK)
}

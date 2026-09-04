package handler

import (
	"net/http"

	"github.com/akeelnazir/osint-app/backend/internal/httperr"
	"github.com/akeelnazir/osint-app/backend/internal/middleware"
)

// UserHandler exposes /users/me.
type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

// Me returns the authenticated user's profile.
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}
	httperr.JSON(w, http.StatusOK, toUserDTO(u))
}

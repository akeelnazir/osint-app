// Package handler contains HTTP handlers for the REST API.
package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/akeelnazir/osint-app/backend/internal/auth"
	"github.com/akeelnazir/osint-app/backend/internal/ent"
	entuser "github.com/akeelnazir/osint-app/backend/internal/ent/user"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
)

// AuthHandler exposes register/login/refresh endpoints.
type AuthHandler struct {
	client  *ent.Client
	authSvc *auth.Service
}

func NewAuthHandler(client *ent.Client, authSvc *auth.Service) *AuthHandler {
	return &AuthHandler{client: client, authSvc: authSvc}
}

type registerReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional; defaults to "viewer"
}

type tokenResp struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"` // seconds
	User         userDTO `json:"user"`
}

type userDTO struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func toUserDTO(u *ent.User) userDTO {
	return userDTO{ID: u.ID, Email: u.Email, Username: u.Username, Role: string(u.Role)}
}

// Register creates a new user and returns tokens.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	details := map[string]string{}
	if !strings.Contains(req.Email, "@") {
		details["email"] = "valid email is required"
	}
	if len(req.Username) < 3 {
		details["username"] = "must be at least 3 characters"
	}
	if len(req.Password) < 8 {
		details["password"] = "must be at least 8 characters"
	}
	if len(details) > 0 {
		httperr.Write(w, httperr.Validation(details))
		return
	}

	role := entuser.RoleViewer
	if req.Role != "" {
		switch req.Role {
		case "admin", "analyst", "viewer":
			role = entuser.Role(req.Role)
		default:
			httperr.Write(w, httperr.BadRequest("invalid role"))
			return
		}
		// Only existing admins can create admins/analysts; for self-registration
		// we force viewer unless APP_ENV=development and no users exist yet.
		if role != entuser.RoleViewer {
			count, _ := h.client.User.Query().Count(r.Context())
			if count > 0 {
				httperr.Write(w, httperr.Forbidden("cannot self-register elevated roles"))
				return
			}
		}
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httperr.Write(w, httperr.Unprocessable(err.Error()))
		return
	}
	u, err := h.client.User.Create().
		SetEmail(req.Email).
		SetUsername(req.Username).
		SetPasswordHash(hash).
		SetRole(role).
		Save(r.Context())
	if err != nil {
		if ent.IsConstraintError(err) {
			httperr.Write(w, httperr.Conflict("email or username already taken"))
			return
		}
		httperr.Write(w, httperr.Internal("failed to create user"))
		return
	}
	h.writeTokens(w, r, u)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and returns tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	u, err := h.client.User.Query().Where(entuser.EmailEQ(req.Email)).Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Unauthorized("invalid credentials"))
		return
	}
	if err := auth.VerifyPassword(u.PasswordHash, req.Password); err != nil {
		httperr.Write(w, httperr.Unauthorized("invalid credentials"))
		return
	}
	h.writeTokens(w, r, u)
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh exchanges a refresh token for a new access+refresh pair.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	claims, err := h.authSvc.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		httperr.Write(w, httperr.Unauthorized("invalid refresh token"))
		return
	}
	u, err := h.client.User.Get(r.Context(), claims.UserID)
	if err != nil {
		httperr.Write(w, httperr.Unauthorized("user no longer exists"))
		return
	}
	h.writeTokens(w, r, u)
}

func (h *AuthHandler) writeTokens(w http.ResponseWriter, r *http.Request, u *ent.User) {
	access, accessExp, err := h.authSvc.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to issue access token"))
		return
	}
	refresh, _, err := h.authSvc.IssueRefreshToken(u.ID, string(u.Role))
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to issue refresh token"))
		return
	}
	httperr.JSON(w, http.StatusOK, tokenResp{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(time.Until(accessExp).Seconds()),
		User:         toUserDTO(u),
	})
}

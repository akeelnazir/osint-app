package handler

import (
	"net/http"
	"time"

	"github.com/akeelnazir/osint-app/backend/internal/ent"
	entcaserecord "github.com/akeelnazir/osint-app/backend/internal/ent/caserecord"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
	"github.com/akeelnazir/osint-app/backend/internal/middleware"

	"entgo.io/ent/dialect/sql"
)

// DashboardHandler provides aggregate stats for the landing page.
type DashboardHandler struct {
	client *ent.Client
}

func NewDashboardHandler(client *ent.Client) *DashboardHandler {
	return &DashboardHandler{client: client}
}

type dashboardStats struct {
	Cases    int `json:"cases"`
	Evidence int `json:"evidence"`
	Entities int `json:"entities"`
	Users    int `json:"users"`
}

type recentCaseDTO struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type activityFeedDTO struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

// Stats returns dashboard counts, recent cases, and activity feed.
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}
	role, _ := r.Context().Value(middleware.CtxRole).(string)

	// Counts.
	caseQ := h.client.CaseRecord.Query()
	if role != "admin" {
		caseQ = caseQ.Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.EQ("owner_id", userID),
				sql.In("id", s.Select("case_id").From(sql.Table("case_members")).Where(sql.EQ("user_id", userID))),
				sql.EQ("visibility", "public"),
			))
		})
	}
	caseCount, err := caseQ.Count(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to count cases"))
		return
	}
	evidenceCount, err := h.client.Evidence.Query().Count(r.Context())
	if err != nil {
		evidenceCount = 0
	}
	entityCount, err := h.client.Entity.Query().Count(r.Context())
	if err != nil {
		entityCount = 0
	}
	userCount, err := h.client.User.Query().Count(r.Context())
	if err != nil {
		userCount = 0
	}

	// Recent cases.
	recentCases, err := caseQ.Order(ent.Desc(entcaserecord.FieldUpdatedAt)).Limit(5).All(r.Context())
	if err != nil {
		recentCases = nil
	}
	recent := make([]recentCaseDTO, 0, len(recentCases))
	for _, c := range recentCases {
		recent = append(recent, recentCaseDTO{ID: c.ID, Title: c.Title, Status: string(c.Status), UpdatedAt: c.UpdatedAt})
	}

	// Activity feed (global, last 20).
	acts, err := h.client.ActivityLog.Query().WithUser().Order(ent.Desc("created_at")).Limit(20).All(r.Context())
	if err != nil {
		acts = nil
	}
	feed := make([]activityFeedDTO, 0, len(acts))
	for _, a := range acts {
		dto := activityFeedDTO{ID: a.ID, Action: a.Action, CreatedAt: a.CreatedAt}
		if a.Edges.User != nil {
			dto.Username = a.Edges.User.Username
		}
		feed = append(feed, dto)
	}

	httperr.JSON(w, http.StatusOK, map[string]any{
		"stats": dashboardStats{
			Cases:    caseCount,
			Evidence: evidenceCount,
			Entities: entityCount,
			Users:    userCount,
		},
		"recent_cases": recent,
		"activity":     feed,
	})
}

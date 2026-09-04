package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/akeelnazir/osint-app/backend/internal/ent"
	"github.com/akeelnazir/osint-app/backend/internal/geo"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"

	"entgo.io/ent/dialect/sql"
)

// SearchHandler handles global search across cases and evidence.
type SearchHandler struct {
	client *ent.Client
	drv    *sql.Driver
}

func NewSearchHandler(client *ent.Client, drv *sql.Driver) *SearchHandler {
	return &SearchHandler{client: client, drv: drv}
}

type searchResult struct {
	Type        string     `json:"type"` // "case" | "evidence"
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CaseID      int        `json:"case_id,omitempty"`
	EvidenceType string    `json:"evidence_type,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
}

// Search runs a full-text + ILIKE search across cases and evidence.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		httperr.Write(w, httperr.BadRequest("query parameter 'q' is required"))
		return
	}
	resultType := r.URL.Query().Get("type")
	from := r.URL.Query().Get("date_from")
	to := r.URL.Query().Get("date_to")
	lat := queryFloat(r, "lat", 0)
	lng := queryFloat(r, "lng", 0)
	radius := queryFloat(r, "radius", 0)
	tag := r.URL.Query().Get("tag")

	page := queryInt(r, "page", 1)
	if page < 1 {
		page = 1
	}
	limit := queryInt(r, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var results []searchResult

	// Search cases.
	if resultType == "" || resultType == "case" {
		cq := h.client.CaseRecord.Query().Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.Like("title", "%"+q+"%"),
				sql.Like("description", "%"+q+"%"),
			))
		})
		if tag != "" {
			cq = cq.Where(func(s *sql.Selector) {
				s.Where(sql.In("id", s.Select("case_tags").From(sql.Table("case_tags")).Where(sql.EQ("tag_id",
					s.Select("id").From(sql.Table("tags")).Where(sql.EQ("name", tag))))))
			})
		}
		cases, err := cq.WithOwner().Limit(limit).Offset(offset).All(r.Context())
		if err != nil {
			httperr.Write(w, httperr.Internal("case search failed"))
			return
		}
		for _, c := range cases {
			results = append(results, searchResult{
				Type:        "case",
				ID:          c.ID,
				Title:       c.Title,
				Description: c.Description,
			})
		}
	}

	// Search evidence (with FTS via tsvector if available, else ILIKE).
	if resultType == "" || resultType == "evidence" {
		eq := h.client.Evidence.Query().Where(func(s *sql.Selector) {
			// Use ILIKE fallback; FTS would use to_tsvector @@ plainto_tsquery.
			s.Where(sql.Or(
				sql.Like("title", "%"+q+"%"),
				sql.Like("description", "%"+q+"%"),
				sql.Like("content", "%"+q+"%"),
			))
		})
		if from != "" {
			if t, err := time.Parse(time.RFC3339, from); err == nil {
				eq = eq.Where(func(s *sql.Selector) { s.Where(sql.GTE("evidence_date", t)) })
			}
		}
		if to != "" {
			if t, err := time.Parse(time.RFC3339, to); err == nil {
				eq = eq.Where(func(s *sql.Selector) { s.Where(sql.LTE("evidence_date", t)) })
			}
		}
		if tag != "" {
			eq = eq.Where(func(s *sql.Selector) {
				s.Where(sql.In("id", s.Select("evidence_tags").From(sql.Table("evidence_tags")).Where(sql.EQ("tag_id",
					s.Select("id").From(sql.Table("tags")).Where(sql.EQ("name", tag))))))
			})
		}
		// Geo radius filter.
		if radius > 0 && (lat != 0 || lng != 0) {
			ids, err := geo.WithinRadius(r.Context(), h.drv, lat, lng, radius)
			if err == nil && len(ids) > 0 {
				eq = eq.Where(func(s *sql.Selector) { s.Where(sql.In("id", ids)) })
			} else if err == nil {
				// no geo matches -> empty evidence results
				eq = eq.Where(func(s *sql.Selector) { s.Where(sql.EQ("id", -1)) })
			}
		}
		evs, err := eq.WithCase().Limit(limit).Offset(offset).All(r.Context())
		if err != nil {
			httperr.Write(w, httperr.Internal("evidence search failed"))
			return
		}
		for _, e := range evs {
			sr := searchResult{
				Type:         "evidence",
				ID:           e.ID,
				Title:        e.Title,
				Description:  e.Description,
				EvidenceType: string(e.Type),
				Date:         e.EvidenceDate,
			}
			if e.Edges.Case != nil {
				sr.CaseID = e.Edges.Case.ID
			}
			results = append(results, sr)
		}
	}

	httperr.JSON(w, http.StatusOK, map[string]any{
		"data":  results,
		"query": q,
		"page":  page,
		"limit": limit,
	})
}

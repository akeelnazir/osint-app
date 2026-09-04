package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/akeelnazir/osint-app/backend/internal/ent"
	entcasemember "github.com/akeelnazir/osint-app/backend/internal/ent/casemember"
	entcaserecord "github.com/akeelnazir/osint-app/backend/internal/ent/caserecord"
	entuser "github.com/akeelnazir/osint-app/backend/internal/ent/user"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
	"github.com/akeelnazir/osint-app/backend/internal/middleware"

	"entgo.io/ent/dialect/sql"
)

// CaseHandler handles case CRUD, members, timeline, comments, activity.
type CaseHandler struct {
	client *ent.Client
	drv    *sql.Driver
}

func NewCaseHandler(client *ent.Client, drv *sql.Driver) *CaseHandler {
	return &CaseHandler{client: client, drv: drv}
}

// --- DTOs ---

type caseDTO struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Visibility  string    `json:"visibility"`
	OwnerID     int       `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCaseDTO(c *ent.CaseRecord) caseDTO {
	dto := caseDTO{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Status:      string(c.Status),
		Visibility:  string(c.Visibility),
		OwnerID:     c.Edges.Owner.ID,
		OwnerName:   c.Edges.Owner.Username,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	for _, t := range c.Edges.Tags {
		dto.Tags = append(dto.Tags, t.Name)
	}
	return dto
}

type createCaseReq struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Tags        []string `json:"tags"`
}

type updateCaseReq struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Visibility  *string  `json:"visibility"`
	Tags        []string `json:"tags"`
}

// List returns cases the user can see (owned, member of, or public).
func (h *CaseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}
	role, _ := r.Context().Value(middleware.CtxRole).(string)

	q := h.client.CaseRecord.Query()
	// Non-admins see: owned, member-of, or public.
	if role != "admin" {
		q = q.Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.EQ("owner_id", userID),
				sql.In(
					"id",
					s.Select("case_id").From(sql.Table("case_members")).Where(sql.EQ("user_id", userID)),
				),
				sql.EQ("visibility", "public"),
			))
		})
	}

	// Filters.
	if status := r.URL.Query().Get("status"); status != "" {
		q = q.Where(entcaserecord.StatusEQ(entcaserecord.Status(status)))
	}
	if vis := r.URL.Query().Get("visibility"); vis != "" {
		q = q.Where(entcaserecord.VisibilityEQ(entcaserecord.Visibility(vis)))
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		q = q.Where(entcaserecord.HasTagsWith(func(s *sql.Selector) {
			s.Where(sql.EQ("name", tag))
		}))
	}
	if q2 := r.URL.Query().Get("q"); q2 != "" {
		q = q.Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.Like("title", "%"+q2+"%"),
				sql.Like("description", "%"+q2+"%"),
			))
		})
	}

	// Pagination.
	page := queryInt(r, "page", 1)
	if page < 1 {
		page = 1
	}
	limit := queryInt(r, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	sort := r.URL.Query().Get("sort")
	switch sort {
	case "created_at":
		q = q.Order(ent.Asc(entcaserecord.FieldCreatedAt))
	case "-created_at":
		q = q.Order(ent.Desc(entcaserecord.FieldCreatedAt))
	case "title":
		q = q.Order(ent.Asc(entcaserecord.FieldTitle))
	default:
		q = q.Order(ent.Desc(entcaserecord.FieldUpdatedAt))
	}

	total, err := q.Count(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to count cases"))
		return
	}
	cases, err := q.
		WithOwner().
		WithTags().
		Limit(limit).
		Offset(offset).
		All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to list cases"))
		return
	}

	dtos := make([]caseDTO, 0, len(cases))
	for _, c := range cases {
		dtos = append(dtos, toCaseDTO(c))
	}
	httperr.JSON(w, http.StatusOK, map[string]any{
		"data":  dtos,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Create makes a new case owned by the authenticated user.
func (h *CaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}
	var req createCaseReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		httperr.Write(w, httperr.Validation(map[string]string{"title": "is required"}))
		return
	}
	vis := entcaserecord.VisibilityPrivate
	if req.Visibility == "public" {
		vis = entcaserecord.VisibilityPublic
	}

	builder := h.client.CaseRecord.Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetVisibility(vis).
		SetOwnerID(userID)
	for _, t := range req.Tags {
		if name := strings.TrimSpace(t); name != "" {
			builder.AddTagIDs(h.ensureTagID(r, name))
		}
	}

	c, err := builder.Save(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to create case"))
		return
	}
	c, err = h.client.CaseRecord.Query().Where(entcaserecord.IDEQ(c.ID)).WithOwner().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload case"))
		return
	}
	httperr.JSON(w, http.StatusCreated, toCaseDTO(c))
}

// Get returns a single case with access control.
func (h *CaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return // error already written
	}
	httperr.JSON(w, http.StatusOK, toCaseDTO(c))
}

// Update modifies a case (owner/admin/editors only).
func (h *CaseHandler) Update(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	if !h.canEdit(r, c) {
		httperr.Write(w, httperr.Forbidden("only the owner or editors can update this case"))
		return
	}
	var req updateCaseReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	builder := h.client.CaseRecord.UpdateOneID(c.ID)
	changed := false
	if req.Title != nil {
		builder.SetTitle(*req.Title)
		changed = true
	}
	if req.Description != nil {
		builder.SetDescription(*req.Description)
		changed = true
	}
	if req.Status != nil {
		switch *req.Status {
		case "open", "closed", "archived":
			builder.SetStatus(entcaserecord.Status(*req.Status))
			changed = true
		default:
			httperr.Write(w, httperr.BadRequest("invalid status"))
			return
		}
	}
	if req.Visibility != nil {
		switch *req.Visibility {
		case "private", "public":
			builder.SetVisibility(entcaserecord.Visibility(*req.Visibility))
			changed = true
		default:
			httperr.Write(w, httperr.BadRequest("invalid visibility"))
			return
		}
	}
	if req.Tags != nil {
		// Clear and re-add tags.
		if err := h.client.CaseRecord.UpdateOneID(c.ID).ClearTags().Exec(r.Context()); err != nil {
			httperr.Write(w, httperr.Internal("failed to clear tags"))
			return
		}
		for _, t := range req.Tags {
			if name := strings.TrimSpace(t); name != "" {
				builder.AddTagIDs(h.ensureTagID(r, name))
			}
		}
		changed = true
	}
	if !changed {
		httperr.Write(w, httperr.BadRequest("no fields to update"))
		return
	}
	if err := builder.Exec(r.Context()); err != nil {
		httperr.Write(w, httperr.Internal("failed to update case"))
		return
	}
	c, err = h.client.CaseRecord.Query().Where(entcaserecord.IDEQ(c.ID)).WithOwner().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload case"))
		return
	}
	httperr.JSON(w, http.StatusOK, toCaseDTO(c))
}

// Delete removes a case (owner/admin only).
func (h *CaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	if !h.canEdit(r, c) {
		httperr.Write(w, httperr.Forbidden("only the owner or admins can delete this case"))
		return
	}
	if err := h.client.CaseRecord.DeleteOneID(c.ID).Exec(r.Context()); err != nil {
		httperr.Write(w, httperr.Internal("failed to delete case"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Members ---

type addMemberReq struct {
	Username string `json:"username"`
	Role     string `json:"role"` // editor | viewer
}

type memberDTO struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (h *CaseHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	members, err := c.QueryMembers().WithUser().All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to list members"))
		return
	}
	dtos := make([]memberDTO, 0, len(members))
	for _, m := range members {
		u := m.Edges.User
		dtos = append(dtos, memberDTO{UserID: u.ID, Username: u.Username, Email: u.Email, Role: string(m.Role)})
	}
	httperr.JSON(w, http.StatusOK, dtos)
}

func (h *CaseHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	if !h.canEdit(r, c) {
		httperr.Write(w, httperr.Forbidden("only the owner or editors can add members"))
		return
	}
	var req addMemberReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	if req.Username == "" {
		httperr.Write(w, httperr.Validation(map[string]string{"username": "is required"}))
		return
	}
	role := entcasemember.RoleViewer
	if req.Role == "editor" {
		role = entcasemember.RoleEditor
	}
	target, err := h.client.User.Query().Where(entuser.UsernameEQ(req.Username)).Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.NotFound("user not found"))
		return
	}
	_, err = h.client.CaseMember.Create().
		SetCaseID(c.ID).
		SetUserID(target.ID).
		SetRole(role).
		Save(r.Context())
	if err != nil {
		if ent.IsConstraintError(err) {
			httperr.Write(w, httperr.Conflict("user is already a member"))
			return
		}
		httperr.Write(w, httperr.Internal("failed to add member"))
		return
	}
	httperr.JSON(w, http.StatusCreated, memberDTO{UserID: target.ID, Username: target.Username, Email: target.Email, Role: string(role)})
}

func (h *CaseHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	if !h.canEdit(r, c) {
		httperr.Write(w, httperr.Forbidden("only the owner or editors can remove members"))
		return
	}
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid user id"))
		return
	}
	if userID == c.Edges.Owner.ID {
		httperr.Write(w, httperr.BadRequest("cannot remove the owner"))
		return
	}
	if _, err := h.client.CaseMember.Delete().Where(
		func(s *sql.Selector) { s.Where(sql.And(sql.EQ("case_id", c.ID), sql.EQ("user_id", userID))) },
	).Exec(r.Context()); err != nil {
		httperr.Write(w, httperr.Internal("failed to remove member"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Timeline ---

type timelineEvent struct {
	EvidenceID  int        `json:"evidence_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	Date        *time.Time `json:"date"`
	Description string     `json:"description"`
}

func (h *CaseHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	q := c.QueryEvidence().Where(func(s *sql.Selector) {
		s.Where(sql.NotNull("evidence_date"))
	})
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where(func(s *sql.Selector) { s.Where(sql.GTE("evidence_date", t)) })
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where(func(s *sql.Selector) { s.Where(sql.LTE("evidence_date", t)) })
		}
	}
	if t := r.URL.Query().Get("type"); t != "" {
		q = q.Where(func(s *sql.Selector) { s.Where(sql.EQ("type", t)) })
	}
	evs, err := q.Order(func(s *sql.Selector) { s.OrderBy(sql.Asc("evidence_date")) }).All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load timeline"))
		return
	}
	events := make([]timelineEvent, 0, len(evs))
	for _, e := range evs {
		events = append(events, timelineEvent{
			EvidenceID:  e.ID,
			Title:       e.Title,
			Type:        string(e.Type),
			Date:        e.EvidenceDate,
			Description: e.Description,
		})
	}
	httperr.JSON(w, http.StatusOK, events)
}

// --- Comments (case-level) ---

type commentDTO struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	Mentions  []string  `json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
}

func toCommentDTO(c *ent.Comment) commentDTO {
	dto := commentDTO{
		ID:        c.ID,
		AuthorID:  c.Edges.Author.ID,
		Author:    c.Edges.Author.Username,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
	if c.Mentions != nil {
		dto.Mentions = c.Mentions
	}
	return dto
}

type createCommentReq struct {
	Body string `json:"body"`
}

func (h *CaseHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	comments, err := c.QueryComments().WithAuthor().Order(ent.Desc("created_at")).All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load comments"))
		return
	}
	dtos := make([]commentDTO, 0, len(comments))
	for _, c := range comments {
		dtos = append(dtos, toCommentDTO(c))
	}
	httperr.JSON(w, http.StatusOK, dtos)
}

func (h *CaseHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}
	var req createCommentReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		httperr.Write(w, httperr.Validation(map[string]string{"body": "is required"}))
		return
	}
	mentions := extractMentions(req.Body)
	com, err := h.client.Comment.Create().
		SetBody(req.Body).
		SetAuthorID(userID).
		SetCaseID(c.ID).
		SetMentions(mentions).
		Save(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to create comment"))
		return
	}
	com, err = h.client.Comment.Query().Where(func(s *sql.Selector) { s.Where(sql.EQ("id", com.ID)) }).WithAuthor().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload comment"))
		return
	}
	httperr.JSON(w, http.StatusCreated, toCommentDTO(com))
}

// --- Activity ---

type activityDTO struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *CaseHandler) ListActivity(w http.ResponseWriter, r *http.Request) {
	c, err := h.loadCase(w, r)
	if err != nil {
		return
	}
	acts, err := c.QueryActivity().WithUser().Order(ent.Desc("created_at")).Limit(50).All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load activity"))
		return
	}
	dtos := make([]activityDTO, 0, len(acts))
	for _, a := range acts {
		dto := activityDTO{
			ID:         a.ID,
			Action:     a.Action,
			TargetType: a.TargetType,
			TargetID:   a.TargetID,
			CreatedAt:  a.CreatedAt,
		}
		if a.Edges.User != nil {
			dto.UserID = a.Edges.User.ID
			dto.Username = a.Edges.User.Username
		}
		dtos = append(dtos, dto)
	}
	httperr.JSON(w, http.StatusOK, dtos)
}

// --- helpers ---

// loadCase fetches a case by URL param and enforces read access.
func (h *CaseHandler) loadCase(w http.ResponseWriter, r *http.Request) (*ent.CaseRecord, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return nil, err
	}
	c, err := h.client.CaseRecord.Query().
		Where(entcaserecord.IDEQ(id)).
		WithOwner().
		WithTags().
		Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.NotFound("case not found"))
		return nil, err
	}
	if !h.canView(r, c) {
		httperr.Write(w, httperr.Forbidden("you do not have access to this case"))
		return nil, httperr.Forbidden("access denied")
	}
	return c, nil
}

// canView returns true if the user can view the case.
func (h *CaseHandler) canView(r *http.Request, c *ent.CaseRecord) bool {
	role, _ := r.Context().Value(middleware.CtxRole).(string)
	if role == "admin" {
		return true
	}
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		return false
	}
	if c.Edges.Owner != nil && c.Edges.Owner.ID == userID {
		return true
	}
	if c.Visibility == entcaserecord.VisibilityPublic {
		return true
	}
	// Check membership.
	count, _ := h.client.CaseMember.Query().Where(
		func(s *sql.Selector) { s.Where(sql.And(sql.EQ("case_id", c.ID), sql.EQ("user_id", userID))) },
	).Count(r.Context())
	return count > 0
}

// canEdit returns true if the user can edit the case (owner, admin, or editor member).
func (h *CaseHandler) canEdit(r *http.Request, c *ent.CaseRecord) bool {
	role, _ := r.Context().Value(middleware.CtxRole).(string)
	if role == "admin" {
		return true
	}
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		return false
	}
	if c.Edges.Owner != nil && c.Edges.Owner.ID == userID {
		return true
	}
	m, err := h.client.CaseMember.Query().Where(
		func(s *sql.Selector) { s.Where(sql.And(sql.EQ("case_id", c.ID), sql.EQ("user_id", userID))) },
	).Only(r.Context())
	if err != nil {
		return false
	}
	return m.Role == "editor"
}

// ensureTagID returns the ID of the tag with the given name, creating it if needed.
func (h *CaseHandler) ensureTagID(r *http.Request, name string) int {
	t, err := h.client.Tag.Create().SetName(name).Save(r.Context())
	if err == nil {
		return t.ID
	}
	// Probably already exists.
	t, err = h.client.Tag.Query().Where(func(s *sql.Selector) { s.Where(sql.EQ("name", name)) }).Only(r.Context())
	if err != nil {
		return 0
	}
	return t.ID
}

// extractMentions finds @username mentions in a comment body.
func extractMentions(body string) []string {
	var mentions []string
	words := strings.Fields(body)
	for _, w := range words {
		if strings.HasPrefix(w, "@") {
			mentions = append(mentions, strings.TrimPrefix(w, "@"))
		}
	}
	return mentions
}

// getWriter is retained for backward compatibility but no longer used;
// loadCase now takes the ResponseWriter directly.
func getWriter(_ *http.Request) http.ResponseWriter { return nil }

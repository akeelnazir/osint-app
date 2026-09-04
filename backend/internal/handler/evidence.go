package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/akeelnazir/osint-app/backend/internal/ent"
	ententity "github.com/akeelnazir/osint-app/backend/internal/ent/entity"
	entevidence "github.com/akeelnazir/osint-app/backend/internal/ent/evidence"
	"github.com/akeelnazir/osint-app/backend/internal/geo"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
	"github.com/akeelnazir/osint-app/backend/internal/middleware"
	"github.com/akeelnazir/osint-app/backend/internal/nlp"
	"github.com/akeelnazir/osint-app/backend/internal/storage"

	"entgo.io/ent/dialect/sql"
)

// EvidenceHandler handles evidence CRUD, GeoJSON, entities, comments.
type EvidenceHandler struct {
	client  *ent.Client
	drv     *sql.Driver
	storage storage.Storage
}

func NewEvidenceHandler(client *ent.Client, drv *sql.Driver, store storage.Storage) *EvidenceHandler {
	return &EvidenceHandler{client: client, drv: drv, storage: store}
}

// --- DTOs ---

type evidenceDTO struct {
	ID           int            `json:"id"`
	CaseID       int            `json:"case_id"`
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	FilePath     string         `json:"file_path,omitempty"`
	MimeType     string         `json:"mime_type,omitempty"`
	Source       string         `json:"source,omitempty"`
	EvidenceDate *time.Time     `json:"evidence_date,omitempty"`
	Latitude     *float64       `json:"latitude,omitempty"`
	Longitude    *float64       `json:"longitude,omitempty"`
	Description  string         `json:"description,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedBy    int            `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Tags         []string       `json:"tags"`
}

func toEvidenceDTO(e *ent.Evidence) evidenceDTO {
	dto := evidenceDTO{
		ID:           e.ID,
		Type:         string(e.Type),
		Title:        e.Title,
		Content:      e.Content,
		FilePath:     e.FilePath,
		MimeType:     e.MimeType,
		Source:       e.Source,
		EvidenceDate: e.EvidenceDate,
		Latitude:     e.Latitude,
		Longitude:    e.Longitude,
		Description:  e.Description,
		Metadata:     e.Metadata,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		Tags:         []string{},
	}
	if e.Edges.Case != nil {
		dto.CaseID = e.Edges.Case.ID
	}
	if e.Edges.Creator != nil {
		dto.CreatedBy = e.Edges.Creator.ID
	}
	for _, t := range e.Edges.Tags {
		dto.Tags = append(dto.Tags, t.Name)
	}
	return dto
}

// List returns evidence for a case with filters.
func (h *EvidenceHandler) List(w http.ResponseWriter, r *http.Request) {
	caseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return
	}
	q := h.client.Evidence.Query().Where(entevidence.HasCaseWith(func(s *sql.Selector) {
		s.Where(sql.EQ("id", caseID))
	}))

	// Filters.
	if t := r.URL.Query().Get("type"); t != "" {
		q = q.Where(entevidence.TypeEQ(entevidence.Type(t)))
	}
	if from := r.URL.Query().Get("date_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where(func(s *sql.Selector) { s.Where(sql.GTE("evidence_date", t)) })
		}
	}
	if to := r.URL.Query().Get("date_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where(func(s *sql.Selector) { s.Where(sql.LTE("evidence_date", t)) })
		}
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		q = q.Where(entevidence.HasTagsWith(func(s *sql.Selector) {
			s.Where(sql.EQ("name", tag))
		}))
	}
	// Geo radius filter.
	lat := queryFloat(r, "lat", 0)
	lng := queryFloat(r, "lng", 0)
	radius := queryFloat(r, "radius", 0)
	if radius > 0 && (lat != 0 || lng != 0) {
		ids, err := geo.WithinRadius(r.Context(), h.drv, lat, lng, radius)
		if err != nil {
			httperr.Write(w, httperr.Internal("geo query failed"))
			return
		}
		if len(ids) == 0 {
			httperr.JSON(w, http.StatusOK, map[string]any{"data": []evidenceDTO{}})
			return
		}
		q = q.Where(entevidence.IDIn(ids...))
	}

	page := queryInt(r, "page", 1)
	if page < 1 {
		page = 1
	}
	limit := queryInt(r, "limit", 50)
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	total, err := q.Count(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to count evidence"))
		return
	}
	evs, err := q.WithCreator().WithTags().Order(ent.Desc("created_at")).Limit(limit).Offset(offset).All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to list evidence"))
		return
	}
	dtos := make([]evidenceDTO, 0, len(evs))
	for _, e := range evs {
		dtos = append(dtos, toEvidenceDTO(e))
	}
	httperr.JSON(w, http.StatusOK, map[string]any{
		"data":  dtos,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Create adds evidence to a case. Supports multipart (file) or JSON (text/url/raw).
func (h *EvidenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	caseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return
	}
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		httperr.Write(w, httperr.Unauthorized("not authenticated"))
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		h.createFromFile(w, r, caseID, userID)
		return
	}
	h.createFromJSON(w, r, caseID, userID)
}

type createEvidenceJSONReq struct {
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Source       string         `json:"source"`
	EvidenceDate *time.Time     `json:"evidence_date"`
	Latitude     *float64       `json:"latitude"`
	Longitude    *float64       `json:"longitude"`
	Description  string         `json:"description"`
	Metadata     map[string]any `json:"metadata"`
	Tags         []string       `json:"tags"`
	URL          string         `json:"url"` // for type=url
}

func (h *EvidenceHandler) createFromJSON(w http.ResponseWriter, r *http.Request, caseID, userID int) {
	var req createEvidenceJSONReq
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		httperr.Write(w, httperr.Validation(map[string]string{"title": "is required"}))
		return
	}
	evType := entevidence.Type(req.Type)
	if err := entevidence.TypeValidator(evType); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid evidence type"))
		return
	}

	builder := h.client.Evidence.Create().
		SetCaseID(caseID).
		SetType(evType).
		SetTitle(req.Title).
		SetSource(req.Source).
		SetDescription(req.Description).
		SetCreatorID(userID)

	if req.Content != "" {
		builder.SetContent(req.Content)
	}
	if req.EvidenceDate != nil {
		builder.SetEvidenceDate(*req.EvidenceDate)
	}
	if req.Latitude != nil && req.Longitude != nil {
		builder.SetLatitude(*req.Latitude).SetLongitude(*req.Longitude)
	}
	if req.Metadata != nil {
		builder.SetMetadata(req.Metadata)
	}

	// For URL evidence, fetch metadata.
	if evType == entevidence.TypeURL {
		url := req.URL
		if url == "" {
			url = req.Content
		}
		if url == "" {
			httperr.Write(w, httperr.Validation(map[string]string{"url": "is required for url evidence"}))
			return
		}
		if req.Content == "" {
			builder.SetContent(url)
		}
		meta := fetchURLMetadata(url)
		if req.Metadata == nil {
			builder.SetMetadata(meta)
		}
		if title := meta["title"]; title != "" && req.Title == "" {
			builder.SetTitle(title.(string))
		}
	}

	for _, t := range req.Tags {
		if name := strings.TrimSpace(t); name != "" {
			builder.AddTagIDs(h.ensureTagID(r, name))
		}
	}

	e, err := builder.Save(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to create evidence"))
		return
	}
	// Sync geom column.
	if req.Latitude != nil && req.Longitude != nil {
		_ = geo.SetGeom(r.Context(), h.drv, e.ID, req.Latitude, req.Longitude)
	}
	e, err = h.client.Evidence.Query().Where(entevidence.IDEQ(e.ID)).WithCreator().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload evidence"))
		return
	}
	httperr.JSON(w, http.StatusCreated, toEvidenceDTO(e))
}

func (h *EvidenceHandler) createFromFile(w http.ResponseWriter, r *http.Request, caseID, userID int) {
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MiB max
		httperr.Write(w, httperr.BadRequest("failed to parse multipart form"))
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		httperr.Write(w, httperr.Validation(map[string]string{"title": "is required"}))
		return
	}
	evTypeStr := r.FormValue("type")
	if evTypeStr == "" {
		// Infer from file content type.
		evTypeStr = "document"
	}
	evType := entevidence.Type(evTypeStr)
	if err := entevidence.TypeValidator(evType); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid evidence type"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httperr.Write(w, httperr.Validation(map[string]string{"file": "is required for file uploads"}))
		return
	}
	defer file.Close()

	path, err := h.storage.Save(header.Filename, file)
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to store file"))
		return
	}

	builder := h.client.Evidence.Create().
		SetCaseID(caseID).
		SetType(evType).
		SetTitle(title).
		SetFilePath(path).
		SetMimeType(header.Header.Get("Content-Type")).
		SetSource(r.FormValue("source")).
		SetDescription(r.FormValue("description")).
		SetCreatorID(userID)

	if dateStr := r.FormValue("evidence_date"); dateStr != "" {
		if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
			builder.SetEvidenceDate(t)
		}
	}
	if latStr := r.FormValue("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			builder.SetLatitude(lat)
			if lng, err := strconv.ParseFloat(r.FormValue("longitude"), 64); err == nil {
				builder.SetLongitude(lng)
			}
		}
	}

	e, err := builder.Save(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to create evidence"))
		return
	}
	if e.Latitude != nil && e.Longitude != nil {
		_ = geo.SetGeom(r.Context(), h.drv, e.ID, e.Latitude, e.Longitude)
	}
	e, err = h.client.Evidence.Query().Where(entevidence.IDEQ(e.ID)).WithCreator().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload evidence"))
		return
	}
	httperr.JSON(w, http.StatusCreated, toEvidenceDTO(e))
}

// Get returns a single evidence item.
func (h *EvidenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	httperr.JSON(w, http.StatusOK, toEvidenceDTO(e))
}

// Update modifies an evidence item.
func (h *EvidenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	var req struct {
		Title        *string        `json:"title"`
		Content      *string        `json:"content"`
		Source       *string        `json:"source"`
		Description  *string        `json:"description"`
		EvidenceDate *time.Time     `json:"evidence_date"`
		Latitude     *float64       `json:"latitude"`
		Longitude    *float64       `json:"longitude"`
		Metadata     map[string]any `json:"metadata"`
		Tags         []string       `json:"tags"`
	}
	if err := decodeJSON(r, &req); err != nil {
		httperr.Write(w, httperr.BadRequest("invalid JSON body"))
		return
	}
	builder := h.client.Evidence.UpdateOneID(e.ID)
	changed := false
	if req.Title != nil {
		builder.SetTitle(*req.Title)
		changed = true
	}
	if req.Content != nil {
		builder.SetContent(*req.Content)
		changed = true
	}
	if req.Source != nil {
		builder.SetSource(*req.Source)
		changed = true
	}
	if req.Description != nil {
		builder.SetDescription(*req.Description)
		changed = true
	}
	if req.EvidenceDate != nil {
		builder.SetEvidenceDate(*req.EvidenceDate)
		changed = true
	}
	if req.Latitude != nil && req.Longitude != nil {
		builder.SetLatitude(*req.Latitude).SetLongitude(*req.Longitude)
		changed = true
		_ = geo.SetGeom(r.Context(), h.drv, e.ID, req.Latitude, req.Longitude)
	}
	if req.Metadata != nil {
		builder.SetMetadata(req.Metadata)
		changed = true
	}
	if req.Tags != nil {
		_ = h.client.Evidence.UpdateOneID(e.ID).ClearTags().Exec(r.Context())
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
		httperr.Write(w, httperr.Internal("failed to update evidence"))
		return
	}
	e, err = h.client.Evidence.Query().Where(entevidence.IDEQ(e.ID)).WithCreator().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to reload evidence"))
		return
	}
	httperr.JSON(w, http.StatusOK, toEvidenceDTO(e))
}

// Delete removes an evidence item and its file.
func (h *EvidenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	if e.FilePath != "" {
		_ = h.storage.Delete(e.FilePath)
	}
	if err := h.client.Evidence.DeleteOneID(e.ID).Exec(r.Context()); err != nil {
		httperr.Write(w, httperr.Internal("failed to delete evidence"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- GeoJSON ---

type geoJSONFeature struct {
	Type       string         `json:"type"`
	Geometry   map[string]any `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type geoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

// GeoJSON returns all geolocated evidence for a case as a FeatureCollection.
func (h *EvidenceHandler) GeoJSON(w http.ResponseWriter, r *http.Request) {
	caseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return
	}
	evs, err := h.client.Evidence.Query().
		Where(entevidence.HasCaseWith(func(s *sql.Selector) {
			s.Where(sql.And(sql.EQ("id", caseID), sql.NotNull("latitude"), sql.NotNull("longitude")))
		})).
		All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load evidence"))
		return
	}
	fc := geoJSONFeatureCollection{Type: "FeatureCollection"}
	for _, e := range evs {
		fc.Features = append(fc.Features, geoJSONFeature{
			Type: "Feature",
			Geometry: map[string]any{
				"type":        "Point",
				"coordinates": []float64{*e.Longitude, *e.Latitude},
			},
			Properties: map[string]any{
				"id":    e.ID,
				"title": e.Title,
				"type":  string(e.Type),
				"date":  e.EvidenceDate,
			},
		})
	}
	httperr.JSON(w, http.StatusOK, fc)
}

// --- Entities ---

type entityDTO struct {
	ID         int     `json:"id"`
	EvidenceID int     `json:"evidence_id"`
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

func toEntityDTO(e *ent.Entity) entityDTO {
	dto := entityDTO{ID: e.ID, Type: string(e.Type), Value: e.Value, Confidence: e.Confidence}
	if e.Edges.Evidence != nil {
		dto.EvidenceID = e.Edges.Evidence.ID
	}
	return dto
}

// ExtractEntities runs regex+gazetteer NER on the evidence text and stores results.
func (h *EvidenceHandler) ExtractEntities(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	text := strings.Join([]string{e.Title, e.Description, e.Content}, "\n")
	entities := nlp.Extract(text)
	// Clear existing entities for this evidence.
	_, _ = h.client.Entity.Delete().Where(func(s *sql.Selector) {
		s.Where(sql.EQ("evidence_id", e.ID))
	}).Exec(r.Context())
	builders := make([]*ent.EntityCreate, 0, len(entities))
	for _, ent := range entities {
		builders = append(builders, h.client.Entity.Create().
			SetEvidenceID(e.ID).
			SetType(entype(ent.Type)).
			SetValue(ent.Value).
			SetConfidence(ent.Confidence))
	}
	saved, err := h.client.Entity.CreateBulk(builders...).Save(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to save entities"))
		return
	}
	dtos := make([]entityDTO, 0, len(saved))
	for _, e := range saved {
		dtos = append(dtos, toEntityDTO(e))
	}
	httperr.JSON(w, http.StatusOK, dtos)
}

func entype(s string) ententity.Type {
	switch s {
	case "person":
		return ententity.TypePerson
	case "organization":
		return ententity.TypeOrganization
	case "location":
		return ententity.TypeLocation
	default:
		return ententity.TypeOther
	}
}

// ListEntities returns entities for an evidence item.
func (h *EvidenceHandler) ListEntities(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	entities, err := e.QueryEntities().All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load entities"))
		return
	}
	dtos := make([]entityDTO, 0, len(entities))
	for _, e := range entities {
		dtos = append(dtos, toEntityDTO(e))
	}
	httperr.JSON(w, http.StatusOK, dtos)
}

// --- Comments (evidence-level) ---

func (h *EvidenceHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
	if err != nil {
		return
	}
	comments, err := e.QueryComments().WithAuthor().Order(ent.Desc("created_at")).All(r.Context())
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

func (h *EvidenceHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadEvidence(w, r)
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
		SetEvidenceID(e.ID).
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

// --- helpers ---

func (h *EvidenceHandler) loadEvidence(w http.ResponseWriter, r *http.Request) (*ent.Evidence, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "evidenceId"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid evidence id"))
		return nil, err
	}
	e, err := h.client.Evidence.Query().
		Where(entevidence.IDEQ(id)).
		WithCreator().
		WithTags().
		WithCase().
		Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.NotFound("evidence not found"))
		return nil, err
	}
	return e, nil
}

// ensureTagID returns the ID of the tag with the given name, creating it if needed.
func (h *EvidenceHandler) ensureTagID(r *http.Request, name string) int {
	t, err := h.client.Tag.Create().SetName(name).Save(r.Context())
	if err == nil {
		return t.ID
	}
	t, err = h.client.Tag.Query().Where(func(s *sql.Selector) { s.Where(sql.EQ("name", name)) }).Only(r.Context())
	if err != nil {
		return 0
	}
	return t.ID
}

// fetchURLMetadata scrapes basic metadata (title, description, og:image) from a URL.
func fetchURLMetadata(rawURL string) map[string]any {
	meta := map[string]any{"url": rawURL}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return meta
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return meta
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB
	if err != nil {
		return meta
	}
	html := string(body)
	if title := extractHTMLTag(html, "title"); title != "" {
		meta["title"] = title
	}
	if desc := extractMetaContent(html, "description"); desc != "" {
		meta["description"] = desc
	}
	if ogTitle := extractMetaProperty(html, "og:title"); ogTitle != "" {
		meta["title"] = ogTitle
	}
	if ogDesc := extractMetaProperty(html, "og:description"); ogDesc != "" {
		meta["description"] = ogDesc
	}
	if ogImage := extractMetaProperty(html, "og:image"); ogImage != "" {
		meta["preview_image"] = ogImage
	}
	return meta
}

func extractHTMLTag(html, tag string) string {
	lower := strings.ToLower(html)
	start := strings.Index(lower, "<"+tag+">")
	if start < 0 {
		return ""
	}
	start += len(tag) + 2
	end := strings.Index(lower[start:], "</"+tag+">")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(html[start : start+end])
}

func extractMetaContent(html, name string) string {
	lower := strings.ToLower(html)
	needle := `name="` + name + `"`
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return ""
	}
	// Find content="..." after this.
	rest := html[idx:]
	ci := strings.Index(strings.ToLower(rest), `content="`)
	if ci < 0 {
		return ""
	}
	ci += len(`content="`)
	end := strings.Index(rest[ci:], `"`)
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[ci : ci+end])
}

func extractMetaProperty(html, prop string) string {
	lower := strings.ToLower(html)
	needle := `property="` + prop + `"`
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return ""
	}
	rest := html[idx:]
	ci := strings.Index(strings.ToLower(rest), `content="`)
	if ci < 0 {
		return ""
	}
	ci += len(`content="`)
	end := strings.Index(rest[ci:], `"`)
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[ci : ci+end])
}

// ensureJSON re-encodes a value as JSON if needed (helper for metadata).
func ensureJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

var _ = fmt.Sprintf // keep fmt import if needed later

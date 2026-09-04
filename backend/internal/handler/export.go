package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pdf/fpdf"

	"github.com/akeelnazir/osint-app/backend/internal/ent"
	entcaserecord "github.com/akeelnazir/osint-app/backend/internal/ent/caserecord"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
)

// ExportHandler generates PDF and CSV reports.
type ExportHandler struct {
	client *ent.Client
}

func NewExportHandler(client *ent.Client) *ExportHandler {
	return &ExportHandler{client: client}
}

// CasePDF generates a PDF report for a case.
func (h *ExportHandler) CasePDF(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return
	}
	c, err := h.client.CaseRecord.Query().Where(entcaserecord.IDEQ(id)).WithOwner().WithTags().Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.NotFound("case not found"))
		return
	}
	evs, err := c.QueryEvidence().WithCreator().Order(ent.Asc("evidence_date")).All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load evidence"))
		return
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title.
	pdf.SetFont("Arial", "B", 20)
	pdf.MultiCell(0, 10, c.Title, "", "L", false)
	pdf.Ln(2)

	// Metadata.
	pdf.SetFont("Arial", "", 10)
	owner := "unknown"
	if c.Edges.Owner != nil {
		owner = c.Edges.Owner.Username
	}
	meta := fmt.Sprintf("Owner: %s    Status: %s    Visibility: %s    Created: %s",
		owner, c.Status, c.Visibility, c.CreatedAt.Format("2006-01-02"))
	pdf.MultiCell(0, 5, meta, "", "L", false)
	pdf.Ln(2)

	// Tags.
	if len(c.Edges.Tags) > 0 {
		tags := []string{}
		for _, t := range c.Edges.Tags {
			tags = append(tags, t.Name)
		}
		pdf.SetFont("Arial", "I", 9)
		pdf.MultiCell(0, 5, "Tags: "+strings.Join(tags, ", "), "", "L", false)
		pdf.Ln(2)
	}

	// Description.
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 7, "Summary", "", "L", false)
	pdf.SetFont("Arial", "", 10)
	desc := c.Description
	if desc == "" {
		desc = "(no description)"
	}
	pdf.MultiCell(0, 5, desc, "", "L", false)
	pdf.Ln(4)

	// Evidence table.
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 7, fmt.Sprintf("Evidence (%d items)", len(evs)), "", "L", false)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(20, 7, "Type", "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 7, "Title", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 7, "Source", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 7, "Date", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 7, "Location", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	for _, e := range evs {
		dateStr := ""
		if e.EvidenceDate != nil {
			dateStr = e.EvidenceDate.Format("2006-01-02")
		}
		locStr := ""
		if e.Latitude != nil && e.Longitude != nil {
			locStr = fmt.Sprintf("%.4f,%.4f", *e.Latitude, *e.Longitude)
		}
		pdf.CellFormat(20, 6, string(e.Type), "1", 0, "L", false, 0, "")
		pdf.CellFormat(70, 6, truncate(e.Title, 40), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, truncate(e.Source, 22), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, dateStr, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, locStr, "1", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	// Timeline (textual).
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 7, "Timeline", "", "L", false)
	pdf.SetFont("Arial", "", 9)
	for _, e := range evs {
		if e.EvidenceDate == nil {
			continue
		}
		line := fmt.Sprintf("%s  [%s]  %s", e.EvidenceDate.Format("2006-01-02"), e.Type, e.Title)
		pdf.MultiCell(0, 5, line, "", "L", false)
	}

	// Map snapshot: draw a simple schematic with point markers.
	if hasGeo(evs) {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)
		pdf.MultiCell(0, 7, "Map (schematic)", "", "L", false)
		pdf.SetFont("Arial", "", 9)
		drawSchematicMap(pdf, evs)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="case-%d-report.pdf"`, id))
	if err := pdf.Output(w); err != nil {
		httperr.Write(w, httperr.Internal("pdf generation failed"))
		return
	}
}

// EvidenceCSV exports a case's evidence as CSV.
func (h *ExportHandler) EvidenceCSV(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, httperr.BadRequest("invalid case id"))
		return
	}
	c, err := h.client.CaseRecord.Query().Where(entcaserecord.IDEQ(id)).Only(r.Context())
	if err != nil {
		httperr.Write(w, httperr.NotFound("case not found"))
		return
	}
	evs, err := c.QueryEvidence().All(r.Context())
	if err != nil {
		httperr.Write(w, httperr.Internal("failed to load evidence"))
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="case-%d-evidence.csv"`, id))
	// Write CSV manually (small, no need for encoding/csv import overhead).
	fmt.Fprintln(w, "id,type,title,source,date,latitude,longitude,description,created_at")
	for _, e := range evs {
		dateStr := ""
		if e.EvidenceDate != nil {
			dateStr = e.EvidenceDate.Format(time.RFC3339)
		}
		lat, lng := "", ""
		if e.Latitude != nil {
			lat = fmt.Sprintf("%f", *e.Latitude)
		}
		if e.Longitude != nil {
			lng = fmt.Sprintf("%f", *e.Longitude)
		}
		fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s,%s,%s,%s\n",
			e.ID, e.Type, csvEscape(e.Title), csvEscape(e.Source), dateStr, lat, lng,
			csvEscape(e.Description), e.CreatedAt.Format(time.RFC3339))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "..."
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func hasGeo(evs []*ent.Evidence) bool {
	for _, e := range evs {
		if e.Latitude != nil && e.Longitude != nil {
			return true
		}
	}
	return false
}

// drawSchematicMap draws a simple bounding-box map with point markers using fpdf vector primitives.
func drawSchematicMap(pdf *fpdf.Fpdf, evs []*ent.Evidence) {
	if len(evs) == 0 {
		return
	}
	// Compute bounds.
	minLat, maxLat, minLng, maxLng := 90.0, -90.0, 180.0, -180.0
	for _, e := range evs {
		if e.Latitude == nil || e.Longitude == nil {
			continue
		}
		lat, lng := *e.Latitude, *e.Longitude
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
	}
	if maxLat <= minLat {
		maxLat = minLat + 0.01
	}
	if maxLng <= minLng {
		maxLng = minLng + 0.01
	}
	// Draw a 150x100 mm box.
	x0, y0 := 25.0, 40.0
	w, h := 150.0, 100.0
	pdf.Rect(x0, y0, w, h, "D")
	// Plot points.
	for _, e := range evs {
		if e.Latitude == nil || e.Longitude == nil {
			continue
		}
		px := x0 + w*(*e.Longitude-minLng)/(maxLng-minLng)
		py := y0 + h*(maxLat-*e.Latitude)/(maxLat-minLat)
		pdf.Circle(px, py, 1.5, "F")
	}
	pdf.SetFont("Arial", "", 8)
	pdf.Text(x0, y0-2, fmt.Sprintf("Bounds: [%.4f,%.4f] to [%.4f,%.4f]", minLng, minLat, maxLng, maxLat))
}

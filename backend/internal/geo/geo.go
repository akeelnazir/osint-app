// Package geo provides raw-SQL helpers for PostGIS geography columns on the
// evidence table. Ent has no native geometry type, so we keep lat/lng as
// float fields and mirror them into a `geom geography(Point,4326)` column.
package geo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
)

// SetGeom updates the geom column for an evidence row from lat/lng.
// If either lat or lng is nil, the geom is set to NULL.
func SetGeom(ctx context.Context, drv *sql.Driver, evidenceID int, lat, lng *float64) error {
	if lat == nil || lng == nil {
		_, err := drv.ExecContext(ctx, "UPDATE evidences SET geom = NULL WHERE id = $1", evidenceID)
		return err
	}
	_, err := drv.ExecContext(ctx,
		"UPDATE evidences SET geom = ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography WHERE id = $3",
		*lng, *lat, evidenceID)
	return err
}

// WithinRadius returns evidence IDs within `radiusMeters` of (lat,lng).
func WithinRadius(ctx context.Context, drv *sql.Driver, lat, lng, radiusMeters float64) ([]int, error) {
	rows, err := drv.QueryContext(ctx,
		`SELECT id FROM evidences
		 WHERE geom IS NOT NULL
		   AND ST_DWithin(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)`,
		lng, lat, radiusMeters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// WithinPolygon returns evidence IDs inside a GeoJSON polygon (WKT or GeoJSON
// polygon string). The polygon must be a closed ring.
func WithinPolygon(ctx context.Context, drv *sql.Driver, geojsonPolygon string) ([]int, error) {
	rows, err := drv.QueryContext(ctx,
		`SELECT id FROM evidences
		 WHERE geom IS NOT NULL
		   AND ST_Contains(
		        ST_GeomFromGeoJSON($1)::geography::geometry,
		        geom::geometry
		       )`,
		geojsonPolygon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// WithinBBox returns evidence IDs within a bounding box (minLng, minLat, maxLng, maxLat).
func WithinBBox(ctx context.Context, drv *sql.Driver, minLng, minLat, maxLng, maxLat float64) ([]int, error) {
	rows, err := drv.QueryContext(ctx,
		`SELECT id FROM evidences
		 WHERE geom IS NOT NULL
		   AND ST_Within(geom::geometry,
		        ST_MakeEnvelope($1, $2, $3, $4, 4326))`,
		minLng, minLat, maxLng, maxLat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// IDListLiteral builds a comma-separated, SQL-safe literal for an IN clause.
// Returns "NULL" for empty input so the query matches nothing.
func IDListLiteral(ids []int) string {
	if len(ids) == 0 {
		return "NULL"
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ",")
}

// ErrNoPostGIS is returned when a geo query is attempted but the driver is not Postgres.
var ErrNoPostGIS = fmt.Errorf("postgis queries require a postgres driver")

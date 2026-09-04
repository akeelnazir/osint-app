// Package migrate contains custom SQL migrations that run alongside Ent's
// auto-generated schema. Specifically it enables the PostGIS extension and
// adds a geography(Point,4326) column + GiST index to the evidence table so
// the service layer can issue ST_DWithin / bounding-box queries.
package migrate

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// evidenceTable is the table used by the Evidence entity. Ent's default
// pluralization names it "evidences"; the raw SQL helpers below must match.
const evidenceTable = "evidences"

// PostGIS returns a set of SQL statements that:
//  1. Enable the postgis extension.
//  2. Add a `geom geography(Point,4326)` column to the evidence table.
//  3. Create a GiST index on it.
//  4. Add a generated tsvector column for full-text search on (title, description, content).
//
// Statements are idempotent (IF NOT EXISTS / OR REPLACE where supported).
func PostGIS() []string {
	return []string{
		`CREATE EXTENSION IF NOT EXISTS postgis`,
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS geom geography(Point,4326)`, evidenceTable),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_evidence_geom ON %s USING GIST (geom)`, evidenceTable),
		// Full-text search generated column (Postgres 12+).
		fmt.Sprintf(
			`ALTER TABLE %s ADD COLUMN IF NOT EXISTS tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'') || ' ' || coalesce(content,'')) ) STORED`,
			evidenceTable,
		),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_evidence_tsv ON %s USING GIN (tsv)`, evidenceTable),
	}
}

// Apply runs the PostGIS migrations against the given SQL client.
func Apply(ctx context.Context, drv *sql.Driver) error {
	// Only relevant for Postgres.
	if drv.Dialect() != dialect.Postgres {
		return nil
	}
	for _, stmt := range PostGIS() {
		if _, err := drv.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("postgis migration failed (%q): %w", stmt, err)
		}
	}
	return nil
}

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Evidence struct {
	ent.Schema
}

func (Evidence) Annotations() []schema.Annotation {
	// Store the geometry column as PostGIS Point (SRID 4326) via a raw column.
	// Ent has no native geometry type; we use a string column "geom_wkt" holding
	// WKT and rely on raw SQL helpers in the service layer to insert/query with
	// ST_GeomFromText / ST_DWithin. We also annotate a generated geography column
	// via a custom migration (see internal/server/migrate.go).
	return []schema.Annotation{
		entsql.Annotation{},
	}
}

func (Evidence) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").
			Values("text", "url", "image", "video", "document", "raw"),
		field.String("title").NotEmpty(),
		field.Text("content").Optional(),
		field.String("file_path").Optional(),
		field.String("mime_type").Optional(),
		field.String("source").Optional(),
		field.Time("evidence_date").Optional().Nillable(),
		// latitude / longitude stored as floats for convenience + indexing;
		// the service layer also writes a PostGIS geography(Point,4326) column
		// named "geom" via raw SQL for spatial queries.
		field.Float("latitude").Optional().Nillable(),
		field.Float("longitude").Optional().Nillable(),
		field.Text("description").Optional(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Evidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("case", CaseRecord.Type).
			Ref("evidence").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("created_evidence").
			Unique(),
		edge.To("tags", Tag.Type),
		edge.To("entities", Entity.Type),
		edge.To("comments", Comment.Type),
	}
}

func (Evidence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("evidence_date"),
	}
}

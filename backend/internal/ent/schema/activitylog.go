package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ActivityLog struct {
	ent.Schema
}

func (ActivityLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("action").NotEmpty(),
		field.String("target_type").Optional(),
		field.String("target_id").Optional(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (ActivityLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("activity").
			Unique(),
		edge.From("case", CaseRecord.Type).
			Ref("activity").
			Unique(),
	}
}

func (ActivityLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}

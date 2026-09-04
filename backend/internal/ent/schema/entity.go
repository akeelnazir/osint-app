package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Entity struct {
	ent.Schema
}

func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").Values("person", "organization", "location", "other"),
		field.String("value").NotEmpty(),
		field.Float("confidence").Default(1.0),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Entity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("evidence", Evidence.Type).
			Ref("entities").
			Unique().
			Required(),
	}
}

func (Entity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "value"),
	}
}

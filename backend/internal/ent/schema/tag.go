package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Tag struct {
	ent.Schema
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().NotEmpty(),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cases", CaseRecord.Type).Ref("tags"),
		edge.From("evidence", Evidence.Type).Ref("tags"),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
	}
}
